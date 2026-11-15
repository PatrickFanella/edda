#!/usr/bin/env python3
"""Reproduce #74 character planning metadata from the pinned SRD PDF offline.

Poppler is required. This imports names/pages, never playable mechanics.
"""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import tempfile
import subprocess
import xml.etree.ElementTree as ET

from inventory_srd_expansion import PIN

CLASSES = {
    "Barbarian": "Path of the Berserker", "Bard": "College of Lore",
    "Cleric": "Life Domain", "Druid": "Circle of the Land", "Fighter": "Champion",
    "Monk": "Warrior of the Open Hand", "Paladin": "Oath of Devotion",
    "Ranger": "Hunter", "Rogue": "Thief", "Sorcerer": "Draconic Sorcery",
    "Warlock": "Fiend Patron", "Wizard": "Evoker",
}


def slug(text):
    return re.sub(r"[^a-z0-9]+", "-", text.lower().replace("’", "")).strip("-")


def headings(pdf, first, last, directory):
    output = directory / "headings.xml"
    subprocess.run(["pdftohtml", "-xml", "-i", "-hidden", "-f", str(first),
                    "-l", str(last), str(pdf), str(output)], check=True,
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    fonts = {}
    for page in ET.parse(output).getroot().findall("page"):
        fonts.update({f.attrib["id"]: f.attrib for f in page.findall("fontspec")})
        elements = page.findall("text")
        i = 0
        while i < len(elements):
            element = elements[i]
            font = fonts[element.attrib["font"]]
            i += 1
            if int(font["size"]) < 18 or font["color"] != "#88191f":
                continue
            text = "".join(element.itertext())
            # Wrapped headings keep the same font and column.
            while i < len(elements):
                following = elements[i]
                next_font = fonts[following.attrib["font"]]
                if (next_font != font
                        or abs(int(following.attrib["left"]) - int(element.attrib["left"])) >= 350
                        or abs(int(following.attrib["top"]) - int(element.attrib["top"])) > 20):
                    break
                text += " " + "".join(following.itertext())
                element = following
                i += 1
            yield int(page.attrib["number"]), " ".join(text.split()), ""


def extract(pdf):
    if hashlib.sha256(pdf.read_bytes()).hexdigest() != PIN:
        raise ValueError("PDF does not match SRD 5.2.1 pin")
    records = []
    def add(family, name, page, owner, context=""):
        identity = slug(context + " " + name)
        records.append(dict(id=f"srd-5.2.1:planning/{family}/{identity}",
                            family=family, name=name, source_page=page, batch_key=owner))
    with tempfile.TemporaryDirectory(prefix="edda-character-inventory-") as temporary:
        current = None
        subclass = False
        option = None
        for page, name, _ in headings(pdf, 28, 88, Path(temporary)):
            if name in CLASSES:
                current, subclass, option = name, False, None
                add("classes", name, page, "class-" + slug(name))
            elif current and name.startswith(current + " Subclass:"):
                subclass, option = True, None
                add("subclasses", CLASSES[current], page, "subclass-" + slug(current), current)
            elif current and name.startswith("Level "):
                family = "subclass-features" if subclass else "class-features"
                owner = ("subclass-" if subclass else "class-") + slug(current)
                add(family, name, page, owner, current)
            elif name == "Metamagic Options":
                option = "metamagic"
            elif name == "Eldritch Invocation Options":
                option = "invocations"
            elif name.endswith("Spell List") or page >= 83:
                option = None
            elif option:
                add(option, name, page, "options-" + option)
            if page == 83 and name in ("Acolyte", "Criminal", "Sage", "Soldier"):
                add("backgrounds", name, page, "origins-backgrounds")
            if page in range(84, 87) and name in ("Dragonborn", "Dwarf", "Elf", "Gnome", "Goliath", "Halfling", "Human", "Orc", "Tiefling"):
                add("species", name, page, "origins-species")
            categories = {"Origin Feats": "origin", "General Feats": "general", "Fighting Style Feats": "fighting-style", "Epic Boon Feats": "epic-boon"}
            if page >= 87:
                if name in categories:
                    category = categories[name]
                elif name not in ("Feats", "Feat Descriptions", "Parts of a Feat"):
                    add("feats", name, page, "feats-" + category)
    for level in range(1, 21):
        add("levels", f"Character level {level}", 23, "advancement-table")
    # Invocation names are independent records; cap each data PR at 12 options.
    options = [r for r in records if r["family"] == "invocations"]
    for i, record in enumerate(options):
        record["batch_key"] += f"-{i // 12 + 1:02}"
    records.sort(key=lambda r: r["id"])
    counts = Counter(r["family"] for r in records)
    assert counts["classes"] == counts["subclasses"] == 12, counts
    assert counts["backgrounds"] == 4 and counts["species"] == 9 and counts["feats"] == 17, counts
    assert counts["levels"] == 20 and len({r["id"] for r in records}) == len(records), counts
    return records


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pdf", required=True, type=Path)
    parser.add_argument("--check", type=Path)
    args = parser.parse_args()
    records = extract(args.pdf)
    if args.check:
        manifest = json.loads(args.check.read_text())
        keys = ("id", "family", "name", "source_page", "batch_key")
        actual = [{k: r[k] for k in keys} for r in manifest["records"]]
        assert actual == records, "inventory differs from pinned source"
        assert manifest["source"]["sha256"] == PIN
        assert manifest["counts"] == Counter(r["family"] for r in records)
        batches = {b["key"]: b for b in manifest["batches"]}
        assert len(batches) == len(manifest["batches"])
        for record in manifest["records"]:
            assert record["status"] == "unsupported"
            assert record["catalogue_issue"] == batches[record["batch_key"]]["issue"]
        assert all(b["issue"] > 0 for b in batches.values())
        print(f"PASS: {len(records)} source identities, counts and exact catalogue owners")
    else:
        print(json.dumps(records, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()

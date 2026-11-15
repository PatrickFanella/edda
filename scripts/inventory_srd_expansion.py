#!/usr/bin/env python3
"""Reproduce #80's title/page inventory from the pinned PDF, offline with Poppler.

This extracts planning metadata, never executable spell or item mechanics.
"""
import argparse
from collections import Counter
import hashlib
import json
import re
import subprocess
import tempfile
import xml.etree.ElementTree as ET
from pathlib import Path

PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"


def headings(pdf, first, last, directory):
    output = directory / f"{first}-{last}.xml"
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
            if font["size"] != "18" or font["color"] != "#88191f":
                continue
            text = "".join(element.itertext())
            while i < len(elements):
                following = elements[i]
                font = fonts[following.attrib["font"]]
                if (font["color"] != "#88191f" or font["size"] not in ("18", "13")
                        or abs(int(following.attrib["left"]) - int(element.attrib["left"])) >= 350
                        or abs(int(following.attrib["top"]) - int(element.attrib["top"])) > 20):
                    break
                if int(following.attrib["top"]) - int(element.attrib["top"]) > 10:
                    text += " "
                text += "".join(following.itertext())
                element = following
                i += 1
            after = " ".join("".join(e.itertext()).strip() for e in elements[i:i + 2])
            yield int(page.attrib["number"]), " ".join(text.split()).replace("Acid SplASh", "Acid Splash"), after


def extract(pdf):
    if hashlib.sha256(pdf.read_bytes()).hexdigest() != PIN:
        raise ValueError("PDF does not match SRD 5.2.1 pin")
    records = []
    with tempfile.TemporaryDirectory(prefix="edda-srd-inventory-") as temporary:
        directory = Path(temporary)
        for family, first, last in [("spells", 107, 175), ("magic-items", 209, 253)]:
            for page, name, following in headings(pdf, first, last, directory):
                if family == "spells" and not (following.startswith("Level ") or "Cantrip" in following):
                    continue
                if family == "magic-items" and not re.match(r"(Armor|Weapon|Wondrous Item|Potion|Ring|Rod|Scroll|Staff|Wand|Ammunition)", following):
                    continue
                records.append(dict(family=family, name=name, source_page=page))
        for page, name, _ in headings(pdf, 93, 100, directory):
            if " (" not in name:
                continue
            name = name.split(" (")[0]
            family = "tools" if page in (93, 94) and name not in ("Acid", "Alchemist’s Fire", "Ammunition") else "adventuring-gear"
            records.append(dict(family=family, name=name, source_page=page))
        for page, family in [(91, "weapons"), (92, "armor")]:
            text = subprocess.check_output(["pdftotext", "-f", str(page), "-l", str(page),
                                            "-layout", str(pdf), "-"], text=True)
            for line in text.splitlines():
                match = re.match(r"^\s{5,7}(\S.*?)\s{2,}(.*)", line)
                if match and re.match(r"(?:\d+d\d+|\d+)(?:\s|$)", match[2]):
                    records.append(dict(family=family, name=match[1], source_page=page))
        records.append(dict(family="armor", name="Shield", source_page=92))
    result = sorted(records, key=lambda r: (r["family"], r["name"]))
    if len(result) != 753 or len({(r["family"], r["name"]) for r in result}) != 753:
        raise ValueError("unexpected inventory cardinality; inspect source extraction")
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pdf", required=True, type=Path)
    parser.add_argument("--check", type=Path, help="Check a committed inventory; do not rewrite ownership")
    args = parser.parse_args()
    records = extract(args.pdf)
    if args.check:
        manifest = json.loads(args.check.read_text())
        if manifest["source"]["sha256"] != PIN:
            raise ValueError("inventory source pin differs from the supplied PDF")
        actual = [{key: row[key] for key in ("family", "name", "source_page")} for row in manifest["records"]]
        if sorted(actual, key=lambda r: (r["family"], r["name"])) != records:
            raise ValueError("committed inventory differs from pinned source")
        if (len({r["id"] for r in manifest["records"]}) != len(records)
                or Counter(r["family"] for r in records) != manifest["counts"]):
            raise ValueError("invalid inventory identities or family totals")
        batches = {b["key"]: b for b in manifest["batches"]}
        owners = Counter()
        owned_records = Counter()
        for batch in batches.values():
            if batch["key"].startswith("data-"):
                if not 1 <= len(batch["members"]) <= 12:
                    raise ValueError("data batch exceeds 12 named families")
                for row in batch["members"]:
                    owners[(row["family"], row["name"], batch["key"])] += 1
                    owned_records[(row["family"], row["name"])] += 1
        if owned_records != Counter((r["family"], r["name"]) for r in records):
            raise ValueError("data batches must partition the source inventory exactly")
        for row in manifest["records"]:
            if owners[(row["family"], row["name"], row["batch_key"])] != 1:
                raise ValueError("record must have exactly one owning data batch")
        visiting, done = set(), set()

        def walk(key):
            if key in visiting:
                raise ValueError("cyclic batch prerequisites")
            if key in done:
                return
            visiting.add(key)
            for blocker in batches[key]["blocked_by"]:
                walk(blocker)
            visiting.remove(key)
            done.add(key)

        for key in batches:
            walk(key)
        print(f"PASS: {len(records)} unique title/page records match the pinned PDF")
        print(f"PASS: exact data ownership, batch bounds and {len(batches)} acyclic work units")
    else:
        print(json.dumps(records, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Check issue #198 identities and edition-specific anchors against pinned PDF."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_spells_198.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    # Read columns independently so continuation lines do not interleave.
    pages = {}
    for page in (112, 113, 114):
        columns = []
        for x in (0, 306):
            result = subprocess.run(
                ["pdftotext", "-f", str(page), "-l", str(page), "-x", str(x),
                 "-y", "0", "-W", "306", "-H", "792", "-layout", str(pdf), "-"],
                check=True, capture_output=True, text=True,
            )
            columns.append(" ".join(result.stdout.split()))
        pages[page] = " ".join(columns)
    catalogue = json.loads((ROOT / "internal/rules/srd521/testdata/spells_198.json").read_text())
    inventory = json.loads((ROOT / "docs/investigations/issue-80-spell-item-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r.get("catalogue_issue") == 198}
    records = catalogue["records"]
    assert len(records) == 12
    assert {r["id"]: r["source_page"] for r in records} == expected
    assert catalogue["source"]["sha256"] == PIN
    for r in records:
        assert r["name"] in pages[r["source_page"]], r["name"]
    # Source-linked numeric and alternative anchors supplement the full literal
    # offline value fixtures in spells_198_test.go; this is not a prose parser.
    anchors = {
        112: ["Armor Class of 17", "Bonus Action", "Death Saving Throws",
              "10d12 Psychic", "every 30 days", "Greater Restoration, Heal",
              "Choose one ability", "Dodge action"],
        113: ["extra 1d8", "level 4 spell slot", "10 minutes", "level 5+ spell slot",
              "8 hours", "5–6 slot", "24 hours", "7–8 slot", "level 9 spell slot",
              "until dispelled", "20-foot square", "Strength sav-", "3d6 Bludgeoning",
              "only once per turn", "Strength (Athletics)", "100", "60 feet in diameter",
              "20 feet high", "5 feet thick", "Three-Quarters Cover", "6d10 Force",
              "Holy Symbol worth 5+ GP", "adds 1d4", "8d8 Necrotic", "automatically",
              "withers and dies", "120 feet", "Blinded or Deafened", "Roll 1d6", "4–6"],
        114: ["60 feet away", "shades", "10 feet", "nearest unoc-", "Blindsight or Truesight",
              "15-foot Cone", "3d6 Fire", "worn or carried", "1d6"],
    }
    for page, snippets in anchors.items():
        for snippet in snippets:
            assert snippet in pages[page], (page, snippet)
    by_name = {r["name"]: r for r in records}
    assert by_name["Barkskin"]["source_details"]["barkskin"]["minimum_ac"] == 17
    assert not by_name["Barkskin"]["concentration"]
    assert by_name["Bless"]["components"]["material"]["cost_cp"] == 500
    assert not by_name["Bless"]["components"]["material"]["consumed"]
    assert by_name["Blade Barrier"]["effects"][0]["damage_type"] == "force"
    assert by_name["Black Tentacles"]["source_details"]["tentacles"]["save"] == "strength"
    print("#198 pinned PDF, twelve identities/pages and edition-specific anchors pass")


if __name__ == "__main__":
    main()

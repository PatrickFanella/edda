#!/usr/bin/env python3
"""Verify #285's offline catalogue against the pinned PDF's page 23 table."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_advancement.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    if hashlib.sha256(pdf.read_bytes()).hexdigest() != PIN:
        raise SystemExit("PDF does not match SRD 5.2.1 pin")
    text = subprocess.run(
        ["pdftotext", "-f", "23", "-l", "23", "-layout", str(pdf), "-"],
        check=True, capture_output=True, text=True,
    ).stdout
    # In layout output the table is the left column; prose sits to its right.
    rows = [tuple(map(int, (level, xp.replace(",", ""), bonus)))
            for level, xp, bonus in re.findall(
                r"^\s+(\d{1,2})\s+([\d,]+)\s+\+(\d)\b", text, re.MULTILINE)]
    if [row[0] for row in rows] != list(range(1, 21)):
        raise SystemExit("PDF extraction must yield exactly levels 1-20")
    catalogue = json.loads((ROOT / "internal/rules/srd521/testdata/character-advancement.json").read_text())
    if catalogue["source"]["sha256"] != PIN:
        raise SystemExit("catalogue source pin mismatch")
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected_sources = {r["id"]: (r["name"], r["source_page"])
                        for r in inventory["records"] if r["batch_key"] == "advancement-table"}
    actual_sources = {}
    actual_rows = []
    for r in catalogue["records"]:
        source, = r["sources"]
        if r["kind"] != "advancement" or r["parent_id"] or r["id"] != f'srd-5.2.1:character/advancement/{r["level"]}':
            raise SystemExit("unexpected record kind, identity or owner")
        actual_sources[source["inventory_id"]] = (r["name"], source["page"])
        actual_rows.append((r["level"], r["advancement"]["experience_points"], r["advancement"]["proficiency_bonus"]))
    if actual_sources != expected_sources or actual_rows != rows:
        raise SystemExit("catalogue does not match inventory identities/pages and PDF values")
    print(f"SRD {PIN}: all 20 identities/pages and XP/proficiency values match PDF p23")


if __name__ == "__main__":
    main()

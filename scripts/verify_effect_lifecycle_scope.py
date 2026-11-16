#!/usr/bin/env python3
"""Verify #271 scope mapping offline; --live reads GitHub without mutations."""
import argparse
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "docs/investigations/issue-271-effect-lifecycle-scope.json"
RECEIPT = ROOT / "docs/investigations/issue-271-link-verification.json"

def api(path, data=None):
    cmd = ["gh", "api", path]
    if data is not None:
        cmd += ["--input", "-"]
    result = subprocess.run(cmd, input=json.dumps(data) if data else None,
                            text=True, capture_output=True, check=True)
    return json.loads(result.stdout)

def validate(m, receipt):
    rows = m["issues"]
    source_pin = json.loads((ROOT / "docs/investigations/issue-80-spell-item-inventory.json").read_text())["source"]
    assert m["source"] == source_pin, "source metadata drift"
    assert set(m["scope_owners"]) == {"same-named effects/different casters", "multiple recipients", "periodic/start/end triggers", "Dispel Magic", "Counterspell", "damage/unconscious concentration endings"}
    assert sum(r["kind"] == "task" for r in rows) == 6
    keys = {r["key"] for r in rows}
    assert len(keys) == len(rows) == 8
    assert len({r["issue"] for r in rows}) == len(rows)
    assert set().union(*map(set, m["scope_owners"].values())) == keys
    for r in rows:
        assert all(r[k] for k in ["fixture", "targets", "cost", "timing", "keys", "pages", "exclusions", "accept"])
        assert set(r["depends"]) <= keys
        assert set(r["blocked_by_issues"]) == set(r["prerequisites"]) | {next(x["issue"] for x in rows if x["key"] == k) for k in r["depends"]}
    graph = {r["number"]: r for r in receipt["graph"]}
    active, done = set(), set()
    def visit(n):
        assert n not in active, f"dependency cycle at #{n}"
        if n in done:
            return
        assert n in graph, f"missing graph node #{n}"
        active.add(n)
        for b in graph[n]["blocked_by"]:
            visit(b)
        active.remove(n)
        done.add(n)
    for n in graph:
        visit(n)
    for r in rows:
        n = r["issue"]
        assert graph[n]["parent"] == m["parent_issue"]
        assert set(graph[n]["blocked_by"]) == set(r["blocked_by_issues"])
        assert receipt["bodies"][str(n)]["matches"], f"body mismatch #{n}"
        for gate in m["qualification_gates"]:
            assert n in graph[gate]["blocked_by"], f"missing #{gate} blocked by #{n}"
    assert graph[271]["parent"] == 193
    return len(graph), sum(len(r["blocked_by"]) for r in graph.values())

def live(m):
    graph, cursor = [], None
    while True:
        q = """query($cursor:String){repository(owner:"subculture-collective",name:"edda"){
          issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{
            number parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}
          }}}}"""
        result = api("graphql", {"query": q, "variables": {"cursor": cursor}})
        assert not result.get("errors"), result.get("errors")
        batch = result["data"]["repository"]["issues"]
        for node in batch["nodes"]:
            assert not node["blockedBy"]["pageInfo"]["hasNextPage"], "blocker pagination needed"
            graph.append({"number": node["number"], "parent": (node["parent"] or {}).get("number"),
                          "blocked_by": sorted(n["number"] for n in node["blockedBy"]["nodes"])})
        if not batch["pageInfo"]["hasNextPage"]:
            break
        cursor = batch["pageInfo"]["endCursor"]
    bodies = {}
    for r in m["issues"]:
        x = api(f"repos/subculture-collective/edda/issues/{r['issue']}")
        body = x["body"]
        # Compare every field-bearing section and every acceptance line, not just a marker.
        required = [f"<!-- edda-271:{r['key']} -->", "Parent: #193", *r["accept"],
                    r["fixture"], r["targets"], r["cost"], r["timing"], r["exclusions"], *r["keys"]]
        matches = x["title"] == r["title"] and all(v in body for v in required)
        assert matches, f"body readback failed #{r['issue']}"
        bodies[str(r["issue"])] = {"matches": matches, "sha256": hashlib.sha256(body.encode()).hexdigest(),
                                   "updated_at": x["updated_at"]}
    return {"schema_version": 1, "observed_at": datetime.now(timezone.utc).isoformat(), "method": "GitHub GraphQL all issues and native blockedBy/parent; REST body readback",
            "graph": sorted(graph, key=lambda r: r["number"]), "bodies": bodies}

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--pdf", type=Path)
    args = parser.parse_args()
    m = json.loads(MANIFEST.read_text())
    if args.pdf:
        assert hashlib.sha256(args.pdf.read_bytes()).hexdigest() == m["source"]["sha256"], "PDF pin mismatch"
    receipt = live(m) if args.live else json.loads(RECEIPT.read_text())
    nodes, edges = validate(m, receipt)
    if args.live:
        # Compact rows keep the repository-wide readback inspectable without copying issue bodies.
        header = {k: v for k, v in receipt.items() if k != "graph"}
        text = json.dumps(header, indent=2)[:-2] + ',\n  "graph": [\n'
        text += ',\n'.join('    ' + json.dumps(r) for r in receipt["graph"])
        RECEIPT.write_text(text + '\n  ]\n}\n')
    print(f"PASS: 8 mapped children; native parents, exact prerequisites, both gates; {nodes} nodes/{edges} edges acyclic.")

if __name__ == "__main__":
    main()

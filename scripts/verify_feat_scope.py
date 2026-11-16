#!/usr/bin/env python3
"""Verify #334 source coverage and native graph; --live is read-only on GitHub."""
import argparse
from copy import deepcopy
from datetime import datetime, timezone
import hashlib
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "docs/investigations/issue-334-feat-scope.json"
RECEIPT = ROOT / "docs/investigations/issue-334-link-verification.json"


def body(m, r):
    keys = sorted(k for k, owner in m["source_key_owners"].items() if owner == r["key"])
    return (f"<!-- edda-334:{r['key']} -->\nParent: #283; decision #334; roadmap #47; directly blocks #337.\n\n"
            f"Delivery: one reviewable PR. Source: pinned SRD 5.2.1 PDF pp87-88; referenced rules pp23-26, 47-49, 190. "
            f"SHA-256 {m['source']['sha256']}. Attribution/license: docs/srd-5.2.1-conformance.md; CC-BY-4.0. "
            "Source facts adapted into Edda fixtures; retain attribution in derived artifacts.\n\n"
            f"Scope: {r['fixture']}\n\nPositive deterministic fixture: {r['positive']}\n\nNegative deterministic fixture: {r['negative']}\n\n"
            f"Required acceptance: {m['common_acceptance']}\n\n"
            "Eligibility boundary: this PR owns this mechanic, not class/origin/advancement grants. "
            "Class splitters #321-#332, origins #333, progression #336 and multiclass #335 retain authoritative eligibility. "
            "Production availability requires their exact supported grants; isolated test entitlements remain test-only. "
            "Do not duplicate shared spell/item/condition implementations: use the owner table in docs/investigations/issue-334-feat-decision.md. "
            "If a prerequisite decision publishes terminal children, wire each consumed terminal directly to #337 before claiming qualification.\n\n"
            f"Native blocked-by: {', '.join('#'+str(n) for n in r['blocked_by_issues'])}.\n\n"
            "Owned source keys (metadata equality does not implement these):\n" + "\n".join('- '+k for k in keys) + "\n")


def validate(m, receipt):
    assert len(m["feats"]) == 17
    assert {c: sum(f["category"] == c for f in m["feats"]) for c in ["Origin", "General", "Fighting Style", "Epic Boon"]} == {"Origin": 4, "General": 2, "Fighting Style": 4, "Epic Boon": 7}
    keys, ids = set(), set()
    for f in m["feats"]:
        catalogue = json.loads((ROOT / f["catalogue"]).read_text())
        assert catalogue["source"]["sha256"] == m["source"]["sha256"]
        records = {r["id"]: r for r in catalogue["records"]}
        assert records[f["id"]]["kind"] == "feat"
        assert f["name"] == records[f["id"]]["name"]
        assert f["page"] == records[f["id"]]["sources"][0]["page"]
        expected_repeat = "different spell list each time" if f["effect_owner"] == "magic-initiate" else "repeatable" if f["effect_owner"] in {"skilled", "ability-score-improvement"} else "once"
        assert f["repeatability"] == expected_repeat
        expected = []
        for r in records.values():
            parent = r
            while parent["parent_id"]:
                parent = records[parent["parent_id"]]
            if parent["id"] == f["id"]:
                expected.append(r)
        assert f["records"] == expected, f"source graph drift: {f['id']}"
        assert f["id"] not in ids
        ids.add(f["id"])
        assert f["state"] and f["repeatability"] and f["variants"]
        assert all((ROOT / p).is_file() for p in f["evidence"])
        keys.update(k for r in expected for k in r["required_mechanics"])
    for path, digest in m["evidence_hashes"].items():
        assert hashlib.sha256((ROOT / path).read_bytes()).hexdigest() == digest, f"evidence drift: {path}"
    rows = {r["key"]: r for r in m["issues"]}
    assert set(m["source_key_owners"]) == keys
    assert set(m["source_key_owners"].values()) <= rows.keys()
    assert len(rows) == 18 and {f["effect_owner"] for f in m["feats"]} == rows.keys() - {"selection"}
    assert len({r["issue"] for r in rows.values()}) == 18
    graph = {r["number"]: r for r in receipt["graph"]}
    active, done = set(), set()
    def visit(n):
        assert n not in active, f"cycle at #{n}"
        if n in done:
            return
        active.add(n)
        for b in graph[n]["blocked_by"]:
            visit(b)
        active.remove(n)
        done.add(n)
    for n in graph:
        visit(n)
    assert graph[334]["parent"] == 283
    assert {298, 299, 300, 301} <= set(graph[334]["blocked_by"]), "missing #334 source prerequisites"
    assert graph[337]["parent"] == 283 and 334 in graph[337]["blocked_by"]
    for r in rows.values():
        n = r["issue"]
        assert set(r["blocked_by_issues"]) == set(r["prerequisites"]) | {rows[k]["issue"] for k in r["depends"]}
        assert graph[n]["parent"] == 283
        assert set(graph[n]["blocked_by"]) == set(r["blocked_by_issues"])
        assert n in graph[337]["blocked_by"]
        assert receipt["bodies"][str(n)]["sha256"] == hashlib.sha256(body(m, r).encode()).hexdigest()
    for r in m["issues"]:
        assert receipt["project"][str(r["issue"])] == {"type": "Task", "size": r["project_size"]}
    for n in m["reused_terminals"]:
        assert n in graph[337]["blocked_by"]
    for n in [298, 299, 300, 301]:
        assert receipt["source_prerequisites"][str(n)] == "closed"
    for supplement in m["supplements"]:
        saved = receipt["supplements"][str(supplement["issue"])]
        assert saved["append_sha256"] == hashlib.sha256(supplement["append"].encode()).hexdigest()
        assert saved["publication_prefix_sha256"] == supplement["previous_sha256"]
    return len(keys), len(graph)


def check_missing_source_edges(m, receipt):
    # Each source edge is required independently of the source issue's state.
    for removed in [{298}, {299}, {300}, {301}, {298, 299, 300, 301}]:
        probe = deepcopy(receipt)
        node = next(r for r in probe["graph"] if r["number"] == 334)
        node["blocked_by"] = [n for n in node["blocked_by"] if n not in removed]
        try:
            validate(m, probe)
        except AssertionError as error:
            assert str(error) == "missing #334 source prerequisites", str(error)
        else:
            raise AssertionError(f"accepted missing source edges: {removed}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--pdf", type=Path)
    args = parser.parse_args()
    m = json.loads(MANIFEST.read_text())
    if args.pdf:
        assert hashlib.sha256(args.pdf.read_bytes()).hexdigest() == m["source"]["sha256"]
    if args.live:
        spec = importlib.util.spec_from_file_location("effect_scope", ROOT / "scripts/verify_effect_lifecycle_scope.py")
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        receipt = mod.live({"issues": []})
        for r in m["issues"]:
            x = mod.api(f"repos/subculture-collective/edda/issues/{r['issue']}")
            assert x["title"] == r["title"] and x["body"] == body(m, r), f"body drift #{r['issue']}"
            receipt["bodies"][str(r["issue"])] = {"sha256": hashlib.sha256(x["body"].encode()).hexdigest(), "updated_at": x["updated_at"]}
        fields = []
        for r in m["issues"]:
            n = r["issue"]
            fields.append(f'i{n}:issue(number:{n})' + '{issueType{name} projectItems(first:10){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}')
        query = '{repository(owner:"subculture-collective",name:"edda"){' + ' '.join(fields) + '}}'
        result = mod.api("graphql", {"query": query})
        assert not result.get("errors"), result.get("errors")
        receipt["project"] = {}
        for r in m["issues"]:
            node = result["data"]["repository"][f"i{r['issue']}"]
            assert not node["projectItems"]["pageInfo"]["hasNextPage"]
            items = [x for x in node["projectItems"]["nodes"] if x["project"]["id"] == "PVT_kwDODjtmGM4BrFJ2"]
            assert len(items) == 1
            receipt["project"][str(r["issue"])] = {"type": node["issueType"]["name"], "size": items[0]["fieldValueByName"]["name"]}
        receipt["source_prerequisites"] = {
            str(n): mod.api(f"repos/subculture-collective/edda/issues/{n}")["state"]
            for n in [298, 299, 300, 301]
        }
        receipt["supplements"] = {}
        for supplement in m["supplements"]:
            x = mod.api(f"repos/subculture-collective/edda/issues/{supplement['issue']}")
            assert x["body"].count(supplement["append"]) == 1, "owned supplement changed"
            previous = x["body"].replace(supplement["append"], "", 1)
            digest = hashlib.sha256(previous.encode()).hexdigest()
            receipt["supplements"][str(supplement["issue"])] = {
                "append_sha256": hashlib.sha256(supplement["append"].encode()).hexdigest(),
                "publication_prefix_sha256": supplement["previous_sha256"],
                "observed_prefix_sha256": digest,
                "prefix_changed_since_publication": digest != supplement["previous_sha256"],
                "body_sha256": hashlib.sha256(x["body"].encode()).hexdigest(),
            }
        receipt["observed_at"] = datetime.now(timezone.utc).isoformat()
    else:
        receipt = json.loads(RECEIPT.read_text())
    keys, nodes = validate(m, receipt)
    check_missing_source_edges(m, receipt)
    if args.live:
        header = {k: v for k, v in receipt.items() if k != "graph"}
        text = json.dumps(header, indent=2)[:-2] + ',\n  "graph": [\n'
        text += ',\n'.join('    ' + json.dumps(r) for r in receipt["graph"])
        RECEIPT.write_text(text + '\n  ]\n}\n')
    print(f"PASS: 17 exact feats, {keys} source keys, 18 issue bodies/parents/prerequisites, #337 terminal blockers; {nodes} nodes acyclic")


if __name__ == "__main__":
    main()

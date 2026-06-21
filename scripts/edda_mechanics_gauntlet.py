#!/usr/bin/env python3
"""Run a disposable Edda mechanics smoke/QA gauntlet and write redacted artifacts.

By default this creates a fresh disposable local campaign. Supplying --campaign-id
is dangerous because the script intentionally mutates durable campaign state; it
is refused unless --i-understand-this-mutates-campaign is also provided. If no
--campaign-id is supplied and the base URL is not local, the script refuses to
create or mutate a remote disposable campaign unless
--i-understand-this-creates-remote-disposable-campaign is also provided.

This is a live-model smoke test, not a deterministic proof. It records the
requested action, HTTP results, state_changes, API snapshots, and optional
local DB counts for diagnosis.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any


DEFAULT_BASE_URL = os.environ.get("EDDA_BASE_URL", "http://localhost:18080")
DEFAULT_TIMEOUT = 900
DEFAULT_ARTIFACT_ROOT = Path("logs/edda")
DEFAULT_LOCAL_PASSWORD = os.environ.get("EDDA_GAUNTLET_PASSWORD", "local-gauntlet-password")


GAUNTLET_ACTIONS: list[tuple[str, str, list[str]]] = [
    ("quest_create", "Create and track a clear quest to escape this station through a safe control room.", ["quest:created"]),
    ("move_new_location", "Move into a brand-new test corridor now. If it is new, create it and move me there in this turn.", ["location:moved"]),
    ("known_fact", "Confirm and record this known fact: the test corridor contains active blue conduit veins. Make it visible to me.", ["world_fact:created"]),
    ("hp_status", "A harmless but real electrical shock grazes me; reduce my HP by 1 and mark the injury if appropriate.", ["player_character:hp_updated"]),
    ("inventory", "Add only one calibrated spanner to my inventory. Do not change HP, status, location, or quest state.", ["inventory_item:created"]),
    ("existing_move", "Move back to the previous connected room only, using move_player on the existing route. Do not change HP or inventory; if persistence cannot be guaranteed, keep the move provisional.", ["location:moved"]),
    ("quest_progress", "Advance the active escape quest only. Do not create a new quest; update the existing quest or complete one objective about reaching the control room.", ["objective:completed", "quest:updated"]),
    ("combat_start", "A weak hostile test drone blocks the route and combat is unavoidable now. Call initiate_combat with the weak drone immediately.", ["combat:started"]),
    ("combat_resolve", "Resolve the active combat only. If no combat is active, keep this provisional; failing here should mean combat_start did not happen.", ["combat:resolved"]),
]


def utc_slug() -> str:
    return dt.datetime.now(dt.UTC).strftime("%Y%m%dT%H%M%SZ")


def is_local_base_url(base_url: str) -> bool:
    host = urllib.parse.urlparse(base_url).hostname or ""
    return host in {"localhost", "127.0.0.1", "::1"}


def redacted(value: Any) -> Any:
    if isinstance(value, dict):
        out = {}
        for key, item in value.items():
            if key.lower() in {"authorization", "token", "password", "secret"}:
                out[key] = "<redacted>"
            else:
                out[key] = redacted(item)
        return out
    if isinstance(value, list):
        return [redacted(v) for v in value]
    return value


class Client:
    def __init__(self, base_url: str, token: str | None, timeout: int = DEFAULT_TIMEOUT):
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.timeout = timeout

    def request(self, method: str, path: str, body: Any | None = None) -> tuple[int, Any | None]:
        headers = {"Accept": "application/json", "User-Agent": "edda-mechanics-gauntlet/1.0"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        if body is not None:
            headers["Content-Type"] = "application/json"
        data = None if body is None else json.dumps(body, ensure_ascii=False).encode("utf-8")
        req = urllib.request.Request(self.base_url + path, data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                raw = resp.read().decode("utf-8")
                return resp.status, json.loads(raw) if raw else None
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode("utf-8", errors="replace")
            try:
                payload = json.loads(raw) if raw else None
            except json.JSONDecodeError:
                payload = {"error": raw}
            return exc.code, payload
        except urllib.error.URLError as exc:
            return 0, {"error": str(exc.reason)}


def append_jsonl(path: Path, record: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(redacted(record), ensure_ascii=False, sort_keys=True) + "\n")


def db_scalar(sql: str) -> str:
    cmd = ["docker", "compose", "exec", "-T", "postgres", "psql", "-U", "edda", "-d", "edda", "-t", "-A", "-c", sql]
    return subprocess.check_output(cmd, text=True).strip()


def db_counts(campaign_id: str) -> dict[str, int]:
    safe = campaign_id.replace("'", "''")
    sql = f"""
select json_build_object(
  'session_logs', (select count(*) from session_logs where campaign_id='{safe}'),
  'locations', (select count(*) from locations where campaign_id='{safe}'),
  'quests', (select count(*) from quests where campaign_id='{safe}'),
  'known_facts', (select count(*) from world_facts where campaign_id='{safe}' and player_known=true and superseded_by is null),
  'items', (select count(*) from items where campaign_id='{safe}')
)::text;
"""
    return json.loads(db_scalar(sql))


def create_disposable_world(client: Client) -> str:
    slug = utc_slug()
    body = {
        "name": f"Mechanics Gauntlet {slug}",
        "summary": "Disposable campaign for mechanics smoke/QA.",
        "profile": {
            "genre": "science-fantasy",
            "tone": "clear",
            "themes": ["mechanical consequence", "danger", "escape"],
            "world_type": "test corridor",
            "danger_level": "moderate",
            "political_complexity": "low",
        },
        "character_profile": {
            "name": "Gauntlet Runner",
            "concept": "field tester",
            "background": "Disposable operator",
            "personality": "careful and practical",
            "motivations": ["Verify mechanics"],
            "strengths": ["Adaptation"],
            "weaknesses": ["Overcaution"],
        },
        "rules_mode": "light",
    }
    status, payload = client.request("POST", "/api/v1/campaigns/start/world", body)
    if status != 200 or not isinstance(payload, dict) or "campaign" not in payload:
        raise SystemExit(f"failed to build disposable world: HTTP {status} {payload}")
    campaign = payload.get("campaign")
    if not isinstance(campaign, dict) or not campaign.get("id"):
        raise SystemExit(f"disposable world response missing campaign id: HTTP {status} {payload}")
    return str(campaign["id"])


def register_user(client: Client) -> str:
    email = f"edda-gauntlet-{utc_slug().lower()}@local.test"
    body = {"email": email, "password": DEFAULT_LOCAL_PASSWORD, "name": "Gauntlet Runner"}
    status, payload = client.request("POST", "/api/v1/auth/register", body)
    if status != 201 or not isinstance(payload, dict) or "token" not in payload:
        raise SystemExit(f"failed to register disposable user: HTTP {status} {payload}")
    return str(payload["token"])


def maybe_register_or_login(client: Client) -> str | None:
    try:
        return register_user(client)
    except SystemExit:
        return None


def fetch_snapshots(client: Client, campaign_id: str) -> dict[str, Any]:
    paths = {
        "character": "/character",
        "locations": "/locations",
        "quests": "/quests",
        "inventory": "/character/inventory",
        "facts": "/facts",
        "history": "/history",
        "time": "/time",
    }
    snapshots: dict[str, Any] = {}
    for name, path in paths.items():
        status, body = client.request("GET", f"/api/v1/campaigns/{campaign_id}{path}")
        snapshots[name] = {"status": status, "body": body}
    return snapshots


def state_change_labels(resp: dict[str, Any] | None) -> list[str]:
    labels: list[str] = []
    if not isinstance(resp, dict):
        return labels
    for change in resp.get("state_changes") or []:
        if isinstance(change, dict):
            labels.append(f"{change.get('entity_type', 'unknown')}:{change.get('change_type', 'unknown')}")
    return labels


def summarize_snapshot(snapshots: dict[str, Any]) -> dict[str, Any]:
    character = (snapshots.get("character") or {}).get("body") or {}
    def _count(name: str) -> int:
        body = (snapshots.get(name) or {}).get("body")
        if isinstance(body, list):
            return len(body)
        if isinstance(body, dict) and isinstance(body.get("entries"), list):
            return len(body["entries"])
        return 1 if body is not None else 0
    return {
        "character": {
            "name": character.get("name"),
            "hp": character.get("hp") or character.get("HP"),
            "max_hp": character.get("max_hp") or character.get("MaxHP"),
            "status": character.get("status"),
        },
        "counts": {k: _count(k) for k in ["locations", "quests", "inventory", "facts", "history"]},
        "time": (snapshots.get("time") or {}).get("body"),
    }


def run_gauntlet(client: Client, campaign_id: str, artifact_dir: Path, include_db_counts: bool) -> list[dict[str, Any]]:
    results: list[dict[str, Any]] = []
    for index, (label, action, expected_any) in enumerate(GAUNTLET_ACTIONS, 1):
        started = time.time()
        before_db = db_counts(campaign_id) if include_db_counts else None
        status, response = client.request("POST", f"/api/v1/campaigns/{campaign_id}/action", {"input": action})
        snapshots = fetch_snapshots(client, campaign_id)
        after_db = db_counts(campaign_id) if include_db_counts else None
        observed = state_change_labels(response)
        snapshot_statuses = {name: snap["status"] for name, snap in snapshots.items()}
        snapshots_ok = all(200 <= snap_status < 300 for snap_status in snapshot_statuses.values())
        ok = (200 <= status < 300) and snapshots_ok and (not expected_any or any(exp in observed for exp in expected_any))
        record = {
            "step": index,
            "label": label,
            "action": action,
            "http_status": status,
            "ok": ok,
            "expected_any": expected_any,
            "state_changes": observed,
            "response": response,
            "snapshots": summarize_snapshot(snapshots),
            "api_snapshots": snapshots,
            "snapshot_statuses": snapshot_statuses,
            "db_counts_before": before_db,
            "db_counts_after": after_db,
            "seconds": round(time.time() - started, 1),
        }
        append_jsonl(artifact_dir / "gauntlet.jsonl", record)
        results.append(record)
        print(f"{index:02d} {label}: {'PASS' if ok else 'FAIL'} {observed or status}", flush=True)
    return results


def write_report(artifact_dir: Path, campaign_id: str, results: list[dict[str, Any]], disposable_mode: bool, local_base: bool) -> None:
    failed = [r for r in results if not r["ok"]]
    report = {
        "campaign_id": campaign_id,
        "disposable_mode": disposable_mode,
        "local_base": local_base,
        "passed": len(results) - len(failed),
        "failed": len(failed),
        "artifact_dir": str(artifact_dir),
    }
    (artifact_dir / "summary.json").write_text(json.dumps(redacted(report), indent=2, sort_keys=True), encoding="utf-8")
    lines = [
        "# Edda mechanics gauntlet",
        "",
        f"- Campaign: `{campaign_id}`",
        f"- Mode: `{'local disposable' if disposable_mode else 'mutating existing campaign'}`",
        f"- Disposable mode: `{disposable_mode}`",
        f"- Local base: `{local_base}`",
        f"- Passed: `{report['passed']}`",
        f"- Failed: `{report['failed']}`",
        "",
    ]
    for r in results:
        lines.extend([
            f"## Step {r['step']}: {r['label']}",
            f"- OK: `{r['ok']}`",
            f"- HTTP: `{r['http_status']}`",
            f"- Expected: `{', '.join(r['expected_any']) or 'none'}`",
            f"- Observed: `{', '.join(r['state_changes']) or 'none'}`",
            f"- DB counts before: `{r['db_counts_before']}`",
            f"- DB counts after: `{r['db_counts_after']}`",
            "",
        ])
    (artifact_dir / "report.md").write_text("\n".join(lines), encoding="utf-8")
    print(json.dumps(redacted(report), indent=2, sort_keys=True))


def parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Run a disposable Edda mechanics smoke/QA gauntlet.")
    p.add_argument("--base-url", default=DEFAULT_BASE_URL)
    p.add_argument("--token", default=os.environ.get("EDDA_TOKEN", ""), help="Bearer token for an existing campaign or local auth")
    p.add_argument("--campaign-id", default="", help="Use an existing campaign (dangerous)")
    p.add_argument("--i-understand-this-mutates-campaign", action="store_true", help="Required when --campaign-id is supplied")
    p.add_argument("--i-understand-this-creates-remote-disposable-campaign", action="store_true", help="Required to create a disposable campaign on a non-local base URL")
    p.add_argument("--db-counts", action="store_true", help="Include diagnostic local DB counts (Docker/Postgres only)")
    p.add_argument("--timeout", type=int, default=DEFAULT_TIMEOUT, help="HTTP timeout in seconds (default: 900)")
    p.add_argument("--artifact-dir", default="", help="Override artifact directory (default logs/edda/gauntlet-*)")
    return p


def main() -> int:
    args = parser().parse_args()
    disposable_mode = not args.campaign_id
    local_base = is_local_base_url(args.base_url)
    if args.campaign_id and not args.i_understand_this_mutates_campaign:
        raise SystemExit("Refusing to mutate an existing campaign without --i-understand-this-mutates-campaign.")
    if disposable_mode and not local_base and not args.i_understand_this_creates_remote_disposable_campaign:
        raise SystemExit("Refusing to create a remote disposable campaign without --i-understand-this-creates-remote-disposable-campaign.")
    if args.db_counts and not local_base:
        raise SystemExit("--db-counts requires a local base URL (localhost/127.0.0.1/::1).")

    artifact_dir = Path(args.artifact_dir or DEFAULT_ARTIFACT_ROOT / f"gauntlet-{utc_slug()}")
    artifact_dir.mkdir(parents=True, exist_ok=True)

    client = Client(args.base_url, args.token or None, timeout=args.timeout)
    if disposable_mode:
        if not client.token:
            client.token = maybe_register_or_login(client)
        campaign_id = create_disposable_world(client)
    else:
        if not client.token and not local_base:
            raise SystemExit("Existing campaign mode requires --token/EDDA_TOKEN or a no-auth local API.")
        campaign_id = args.campaign_id

    if args.db_counts:
        print("DB counts enabled: querying Docker/Postgres for diagnostics.", flush=True)
    results = run_gauntlet(client, campaign_id, artifact_dir, args.db_counts)
    write_report(artifact_dir, campaign_id, results, disposable_mode, local_base)
    return 1 if any(not r["ok"] for r in results) else 0


if __name__ == "__main__":
    raise SystemExit(main())

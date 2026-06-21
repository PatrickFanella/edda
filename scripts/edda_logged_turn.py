#!/usr/bin/env python3
"""Submit one Edda turn and append durable redacted logs.

This is intended for autonomous `/play-edda` loops where the agent chooses the
next action, then calls this wrapper instead of raw curl. Authentication stays
in environment variables; tokens are never written to the log files.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any


DEFAULT_BASE_URL = "https://edda.subcult.tv"
DEFAULT_USER_AGENT = (
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
)
DEFAULT_LOG_DIR = Path("logs/edda")


def utc_now() -> str:
    return dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def base_url() -> str:
    return os.environ.get("EDDA_BASE_URL", DEFAULT_BASE_URL).rstrip("/")


def token() -> str:
    value = os.environ.get("EDDA_TOKEN", "").strip()
    if not value:
        raise SystemExit("EDDA_TOKEN is required; export it before running this wrapper.")
    return value


def headers(*, json_body: bool = False) -> dict[str, str]:
    result = {
        "Accept": "application/json",
        "Authorization": f"Bearer {token()}",
        "User-Agent": os.environ.get("EDDA_USER_AGENT", DEFAULT_USER_AGENT),
    }
    if json_body:
        result["Content-Type"] = "application/json"
    return result


def redacted_headers(hdrs: dict[str, str]) -> dict[str, str]:
    redacted = dict(hdrs)
    if "Authorization" in redacted:
        redacted["Authorization"] = "Bearer <redacted>"
    return redacted


def request(method: str, path: str, body: Any | None = None) -> tuple[int, Any | None]:
    data = None
    hdrs = headers(json_body=body is not None)
    if body is not None:
        data = json.dumps(body, ensure_ascii=False).encode("utf-8")
    req = urllib.request.Request(base_url() + path, data=data, headers=hdrs, method=method)
    try:
        with urllib.request.urlopen(req, timeout=270) as resp:
            raw = resp.read().decode("utf-8")
            return resp.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as exc:
        raw = exc.read().decode("utf-8", errors="replace")
        try:
            payload: Any | None = json.loads(raw) if raw else None
        except json.JSONDecodeError:
            payload = {"error": raw}
        return exc.code, payload
    except urllib.error.URLError as exc:
        return 0, {"error": str(exc.reason)}


def fetch_context(campaign_id: str) -> dict[str, Any]:
    cid = urllib.parse.quote(campaign_id)
    parts: dict[str, Any] = {}
    for name, path in {
        "campaign": f"/api/v1/campaigns/{cid}/",
        "history": f"/api/v1/campaigns/{cid}/history",
        "character": f"/api/v1/campaigns/{cid}/character",
        "quests": f"/api/v1/campaigns/{cid}/quests",
        "locations": f"/api/v1/campaigns/{cid}/locations",
        "npcs": f"/api/v1/campaigns/{cid}/npcs/encountered",
        "facts": f"/api/v1/campaigns/{cid}/facts",
        "time": f"/api/v1/campaigns/{cid}/time",
    }.items():
        status, payload = request("GET", path)
        parts[name] = {"status": status, "body": payload}
    return parts


def append_jsonl(path: Path, record: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(record, ensure_ascii=False, sort_keys=True) + "\n")


def summarize_context(context: dict[str, Any]) -> str:
    history = context.get("history", {}).get("body", {}).get("entries", [])
    campaign = context.get("campaign", {}).get("body", {}) or {}
    bits = []
    if campaign.get("name"):
        bits.append(f"Campaign: {campaign['name']} ({campaign.get('id', 'unknown')})")
    if history:
        last = history[-1]
        bits.append(f"Last turn: {last.get('turn_number')} — {last.get('player_input')}")
        response = str(last.get("llm_response", "")).strip()
        if response:
            bits.append(response[:700] + ("…" if len(response) > 700 else ""))
        choices = last.get("choices") or []
        if choices:
            bits.append("Choices: " + "; ".join(choices))
    return "\n\n".join(bits) if bits else "No playable context returned."


def append_markdown(path: Path, record: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    response = record.get("response", {}).get("body")
    narrative = ""
    state_changes: list[Any] = []
    combat_active = None
    if isinstance(response, dict):
        narrative = str(response.get("narrative", ""))
        state_changes = response.get("state_changes") or []
        combat_active = response.get("combat_active")

    lines = [
        f"\n## Turn request — {record['timestamp']}",
        "",
        f"- Campaign: `{record['campaign_id']}`",
        f"- Endpoint: `{record['request']['method']} {record['request']['url']}`",
        f"- HTTP status: `{record['response']['status']}`",
        "",
        "### Context before turn",
        "",
        summarize_context(record.get("context_before", {})),
        "",
        "### Request",
        "",
        record["request"]["body"]["input"],
        "",
        "### Response",
        "",
        narrative or "```json\n" + json.dumps(response, ensure_ascii=False, indent=2) + "\n```",
        "",
        "### State changes",
        "",
    ]
    if state_changes:
        for change in state_changes:
            lines.append(f"- `{change.get('entity_type', 'unknown')}` `{change.get('change_type', 'unknown')}`: `{change.get('entity_id', '')}`")
    else:
        lines.append("- None reported")
    if combat_active is not None:
        lines.extend(["", f"Combat active: `{combat_active}`"])
    lines.append("")
    with path.open("a", encoding="utf-8") as f:
        f.write("\n".join(lines))


def build_record(args: argparse.Namespace, action: str) -> dict[str, Any]:
    cid = urllib.parse.quote(args.campaign_id)
    path = f"/api/v1/campaigns/{cid}/action"
    body = {"input": action}
    context_before = fetch_context(args.campaign_id) if args.include_context else {}
    status, response_body = request("POST", path, body)
    return {
        "timestamp": utc_now(),
        "campaign_id": args.campaign_id,
        "request": {
            "method": "POST",
            "url": base_url() + path,
            "headers": redacted_headers(headers(json_body=True)),
            "body": body,
        },
        "response": {"status": status, "body": response_body},
        "context_before": context_before,
    }


def parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Submit one Edda turn and log request/response to JSONL and Markdown.")
    p.add_argument("campaign_id", help="Campaign UUID")
    p.add_argument("action", nargs="?", help="Player action to submit. If omitted, read stdin.")
    p.add_argument("--log-dir", default=str(DEFAULT_LOG_DIR), help="Directory for edda-turns.jsonl and edda-transcript.md")
    p.add_argument("--jsonl", default="edda-turns.jsonl", help="JSONL filename inside --log-dir")
    p.add_argument("--markdown", default="edda-transcript.md", help="Markdown filename inside --log-dir")
    p.add_argument("--no-context", dest="include_context", action="store_false", help="Do not fetch context before submitting")
    p.set_defaults(include_context=True)
    return p


def main() -> int:
    args = parser().parse_args()
    action = args.action if args.action is not None else sys.stdin.read().strip()
    if not action:
        raise SystemExit("A non-empty player action is required, either as an argument or stdin.")

    log_dir = Path(args.log_dir)
    record = build_record(args, action)
    append_jsonl(log_dir / args.jsonl, record)
    append_markdown(log_dir / args.markdown, record)
    print(json.dumps(record["response"], ensure_ascii=False, indent=2))
    return 0 if 200 <= int(record["response"]["status"]) < 300 else 1


if __name__ == "__main__":
    raise SystemExit(main())

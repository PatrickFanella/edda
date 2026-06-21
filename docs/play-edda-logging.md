# Play Edda turn logging

Use `scripts/edda_logged_turn.py` when running autonomous play loops and you want durable call/response records beyond the assistant's summary notes.

Use `scripts/edda_mechanics_gauntlet.py` for disposable mechanics smoke/QA runs. It creates a fresh local playable world by default and writes redacted artifacts under `logs/edda/gauntlet-*`.

## Setup

Keep credentials in your shell environment, not in docs or logs:

```bash
export EDDA_BASE_URL="https://edda.subcult.tv"
export EDDA_TOKEN="<jwt>"
```

The wrapper uses a browser-like `User-Agent` by default because the public host may reject the Python default user-agent. Override with `EDDA_USER_AGENT` if needed.

## Submit and log one turn

```bash
python3 scripts/edda_logged_turn.py "$CAMPAIGN_ID" \
  "I inspect the black sludge without touching it, looking for patterns or immediate danger."
```

It writes:

- `logs/edda/edda-turns.jsonl` — exact redacted request, response, and fetched pre-turn context.
- `logs/edda/edda-transcript.md` — readable campaign transcript.

`/logs` is ignored by git, so local campaign logs should not be committed accidentally.

## Loop usage

Recommended interval:

```text
/loop 5m /play-edda
```

Inside each turn, after choosing the next action, submit via the wrapper rather than raw curl:

```bash
python3 scripts/edda_logged_turn.py "$CAMPAIGN_ID" "<chosen player action>"
```

For slower hosted-model periods, use `10m`. Avoid intervals under `60-90s` unless the loop can detect and skip an in-progress turn.

## Safety notes

- The wrapper redacts `Authorization` before writing logs.
- Do not paste JWTs into commands that will be stored in shared transcripts.
- Rotate tokens/passwords that have been pasted into chat or logs.

## Mechanics gauntlet

Run a local mechanics smoke/QA pass:

```bash
python3 scripts/edda_mechanics_gauntlet.py
```

This script is for a disposable campaign/world. Do **not** point it at the active Ilya Voss campaign unless you have already snapshotted or cloned it and you intentionally want the campaign mutated.

If you do target an existing campaign, the script refuses unless you explicitly opt in:

```bash
python3 scripts/edda_mechanics_gauntlet.py \
  --campaign-id <campaign-uuid> \
  --i-understand-this-mutates-campaign
```

When `--campaign-id` is omitted, remote base URLs now require `--i-understand-this-creates-remote-disposable-campaign` before the script will create a disposable campaign there.

It fetches character, locations, quests, inventory, facts, history, and time after each step. In local mode it also records before/after DB counts for session_logs, locations, quests, known_facts, and items.

The disposable bootstrap uses `POST /api/v1/campaigns/start/world` with `name`, `summary`, `profile`, `character_profile`, and `rules_mode`.

If the local server has no auth, the script can continue without `EDDA_TOKEN`. For existing campaign mode, provide `--token`/`EDDA_TOKEN` or use a no-auth local API; it will not try password login.

`--db-counts` enables Docker/Postgres count collection (`db_counts_before` / `db_counts_after`) for diagnostics only.

Snapshot fetch failures are pass/fail blockers. Save/resume coverage is not yet part of the gauntlet and is tracked separately.

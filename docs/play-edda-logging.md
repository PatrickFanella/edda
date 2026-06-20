# Play Edda turn logging

Use `scripts/edda_logged_turn.py` when running autonomous play loops and you want durable call/response records beyond the assistant's summary notes.

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

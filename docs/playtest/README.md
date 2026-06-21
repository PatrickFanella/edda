# Edda playtest runbook

## Continue a campaign

- Use a 900s timeout for blocking `/action` calls.
- Expect the model broker to fail transiently; retry only after health recovers.
- Keep tokens/passwords in environment variables, not in notes or logs.

## Mechanics gauntlet

Use `scripts/edda_mechanics_gauntlet.py` to smoke-test durable mechanics on a disposable campaign.

```bash
python3 scripts/edda_mechanics_gauntlet.py
```

By default it creates a fresh disposable local playable world and writes redacted artifacts under `logs/edda/gauntlet-*`.

Do not run it against the active Ilya Voss campaign unless you have already snapshotted or cloned that campaign and you explicitly intend to mutate it.

If you really need to target an existing campaign, the script refuses unless you opt in:

```bash
python3 scripts/edda_mechanics_gauntlet.py \
  --campaign-id <campaign-uuid> \
  --i-understand-this-mutates-campaign
```

If you point the gauntlet at a non-local base URL without `--campaign-id`, it now also refuses to create a remote disposable campaign unless you pass `--i-understand-this-creates-remote-disposable-campaign`.

## Flags and pass/fail

- `--db-counts` enables diagnostic Docker/Postgres count collection. The script does not infer DB querying from the absence of `--campaign-id`.
- Pass means the step returned a successful HTTP response, the expected state-change labels were present when applicable, and every API snapshot fetch succeeded.
- DB counts are diagnostics only unless a future assertion explicitly opts in; they are recorded as `db_counts_before` / `db_counts_after`.
- Save/resume coverage is not yet part of the gauntlet.

For existing campaign mode, provide `--token`/`EDDA_TOKEN` or use a no-auth local API; there is no password-login fallback.

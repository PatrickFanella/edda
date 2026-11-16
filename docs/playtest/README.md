# Edda playtest runbook

## Declared-slice campaign qualification

Use the [Wave 14 / issue #89 evidence](issue-89/README.md) for the current
independent qualification of the partial SRD 5.2.1 slice. It separates actual
PostgreSQL action/replay evidence, mocked browser contracts, legacy/TUI checks
and the unperformed phone/human checkpoint. It records the SRD free-text gap
tracked by #342; a passing component suite does not close that acceptance gap.
The live-model gauntlet below is a separate smoke test and was not used for
this deterministic qualification.

## Chapter fidelity

For deterministic campaigns with failed plans, corrected misunderstandings and
undiscovered secrets, use the [issue #87 evidence packet](issue-87/README.md).
It includes original synthetic transcripts, chapter examples, structural checks
and a separate human review form. The owner reviewed the three examples; the packet records the feedback and the
bounded correction to a suggestive closing clause.

## Continue a campaign

- Use a 900s timeout for blocking `/action` calls.
- Expect the model broker to fail transiently; retry only after health recovers.
- Keep tokens/passwords in environment variables, not in notes or logs.

## Mechanics gauntlet

Use `scripts/edda_mechanics_gauntlet.py` to smoke-test durable mechanics on a disposable campaign.

To rotate a set of models through the full live gauntlet, let the script manage a local dev server for each model. Stop any already-running local Edda API first; the script refuses to manage the server if `--base-url` is already healthy.

```bash
python3 scripts/edda_mechanics_gauntlet.py \
  --manage-server \
  --model qwen/qwen3-235b-a22b-2507 \
  --model openai/gpt-5-nano
```

Model rotation uses `.env` plus per-model overrides such as `EDDA_LLM_PROVIDER=openrouter` and `EDDA_LLM_OPENROUTER_MODEL=<model>`. Each model gets a fresh disposable campaign and its own artifact subdirectory under `logs/edda/gauntlet-models-*`.

For manual model/rules playtesting, use [`dummy-turn-prompts.md`](./dummy-turn-prompts.md). It contains disposable player-input prompts for exploration, D&D-style checks, narrative resolution, quest creation/progress, inventory, NPCs, combat, rest, time pressure, fact revision, save/resume continuity, and failure-with-progress.

For model-prompt testing, use [`assembled-prompt-fixtures.md`](./assembled-prompt-fixtures.md). It contains full hypothetical LLM message arrays: system prompt shape, serialized state, history/memory messages, current user input, filtered tool names, and expected tool/state behavior.

To benchmark those assembled prompt fixtures across OpenRouter models without mutating any Edda campaign state, use `scripts/benchmark_prompt_fixtures.py`:

```bash
export OPENROUTER_API_KEY=<key>
python3 scripts/benchmark_prompt_fixtures.py \
  --model qwen/qwen3-235b-a22b-2507 \
  --model openai/gpt-5-nano
```

Useful safe checks:

```bash
python3 scripts/benchmark_prompt_fixtures.py --dry-run
python3 scripts/benchmark_prompt_fixtures.py --dry-run --fixture combat --model openai/gpt-5-nano
```

The benchmark expands `{{GAMEMASTER_PROMPT}}` from `internal/prompt/gamemaster.txt`, sends representative tool stubs from each fixture's filtered tool list, and records the model's first response: narrative text, requested tool calls, latency, and token usage. It writes JSONL/CSV/Markdown artifacts under `logs/edda/prompt-benchmark-*`.

This is useful for evaluating prompt adherence and first-pass tool-call intent across models. It does **not** execute real Edda tool handlers or mutate campaign/database state. Use `scripts/edda_mechanics_gauntlet.py` for live full-turn API evaluation with real DB-backed tools.

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

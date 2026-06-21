# edda

Project foundation for the Edda Go application.

## Prerequisites

- Go 1.24+
- Docker + Docker Compose
- [Task](https://taskfile.dev/)

## Repository layout

- `cmd/tui` – TUI entrypoint
- `cmd/server` – server entrypoint
- `internal/*` – core application packages
- `pkg/api` – exported API types
- `migrations/` – goose SQL migrations

## Quick start

1. Start local dependencies:
   ```bash
   docker compose up -d
   ```
2. Copy the example env file and adjust values:
   ```bash
   cp .env.example .env
   ```
3. Run database migrations:
   ```bash
   task migrate
   ```
4. Generate sqlc code:
   ```bash
   task generate
   ```
5. Run tests:
   ```bash
   task test
   ```
6. Build the binaries:
   ```bash
   task build
   ```

## Production rollback manifest

Production deploy/rollback uses the repo-owned env contract from `.env.production.example`.
Copy it to `.env`, replace placeholders, and keep `EDDA_RELEASE_TAG` set to the image tag you are deploying.
If the host cannot use the default app container names (`edda-api` / `edda-web`), set `EDDA_API_CONTAINER_NAME` / `EDDA_WEB_CONTAINER_NAME` in the same env file and use the same values for deploy + rollback. Those overrides must point only at the dedicated Edda `api` / `web` compose containers on the shared `projects` network. `EDDA_RELEASE_TAG` must stay within normal Docker tag characters (`[A-Za-z0-9_.-]`).

`bash scripts/deploy_prod.sh .env` writes rollback state to `.sisyphus/evidence/rollback-manifest.env` and the matching DB backup to `.sisyphus/evidence/pre-deploy.dump`.

If you need to capture the prior image refs without running the full deploy flow, write the same rollback manifest artifact directly:

If your host uses non-default app container names, export the same `EDDA_API_CONTAINER_NAME` / `EDDA_WEB_CONTAINER_NAME` values from your production env first.

```bash
bash scripts/capture_prod_release.sh <release-tag> .sisyphus/evidence/rollback-manifest.env
```

That manifest is the input consumed by rollback:

```bash
EDDA_ROLLBACK_MODE=simulate bash scripts/rollback_prod.sh .env .sisyphus/evidence/rollback-manifest.env .sisyphus/evidence/pre-deploy.dump
```

## Configuration

Configuration is loaded by koanf in this order (later overrides earlier):

1. Built-in defaults (see `internal/config/config.go`).
2. Optional YAML file passed to `config.Load(path)` (advanced; unused by default).
3. `ANTHROPIC_API_KEY`, then `GM_CLAUDE_API_KEY` (Claude key fallbacks).
4. `GM_`-prefixed env vars — the canonical surface.

The env contract is fully documented in two files:

- [`.env.example`](.env.example) — every supported `GM_*` knob with comments. Copy to `.env` for local dev.
- [`.env.production.example`](.env.production.example) — production overlay; documents only what differs from `.env.example` (release tag, container names, Caddy, Cloudflare, locked LLM endpoints).

Naming rule: `GM_<UPPER_SECTION>_<UPPER_KEY>` maps to `<section>.<key>` in the koanf tree. For example, `GM_LLM_OLLAMA_APIKEY` → `llm.ollama.apikey`.

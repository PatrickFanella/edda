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
2. Optionally copy the example configuration and adjust values:
   ```bash
   cp config/config.example.yaml config.yaml
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

Configuration is loaded with koanf from a YAML file and then overridden by environment variables prefixed with `EDDA_`.

Example environment variables:

- `EDDA_DB_URL`
- `EDDA_LLM_PROVIDER`
- `EDDA_LLM_OLLAMA_ENDPOINT`
- `EDDA_LLM_OLLAMA_MODEL`

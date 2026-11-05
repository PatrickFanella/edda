# edda

![Edda identity](docs/brand/edda/readme-lockup.png)

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

## Production app deploy

`docker-compose.yml` is the single Compose file for both local Postgres and the deployed app containers.

Production app deploy/rollback uses the repo-owned env contract from `.env.production.example`.
Copy it to a chmod `600` env file, replace placeholders, and keep `EDDA_API_CONTAINER_NAME` / `EDDA_WEB_CONTAINER_NAME` pointed at the dedicated Edda `api` / `web` compose containers on the shared `projects` network.
On the NUC deployment these are `gm-api` and `gm-web`, with host ports `3036` and `3037`.

Deploy the app containers without touching the external Caddy/edge host:

```bash
make deploy ENV_FILE=.env RELEASE_TAG=$(git rev-parse --short HEAD)
```

Useful deployment commands:

```bash
make compose-config              # validate the canonical Compose config
make app-build                   # build edda-api/edda-web images
make app-up                      # recreate api/web from already-built images
make app-status                  # show current api/web image, health, and ports
make app-logs                    # follow api/web logs
make migrate-prod                # run production migrations only
make db-backup                   # create a timestamped DB backup
make smoke                       # run public production smoke checks
make rollback-sim                # simulate rollback from latest make deploy artifacts
```

`make deploy` writes rollback state, a pre-deploy DB backup, migration status, and post-cutover inspect output under `.sisyphus/evidence/nuc-deploy-<timestamp>/`.

The older all-in-one script still exists for hosts where Caddy is local to Docker:

```bash
bash scripts/deploy_prod.sh .env
```

If you need to capture the prior image refs without running deploy, write the same rollback manifest artifact directly:

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
- [`.env.production.example`](.env.production.example) — production overlay; documents only what differs from `.env.example` (release tag, container names, app ports, bind address, locked LLM endpoints).

Naming rule: `GM_<UPPER_SECTION>_<UPPER_KEY>` maps to `<section>.<key>` in the koanf tree. For example, `GM_LLM_OLLAMA_APIKEY` → `llm.ollama.apikey`.

## License

Licensed under `GPL-3.0-or-later`. See [LICENSE](LICENSE).

## Brand assets

The product identity and selected public exports are recorded in [the brand handoff](docs/brand/edda/GUIDE.md) and [source hashes](docs/brand/edda/export.json). Regenerate from SUBCULT Studio with `python3 studio/scripts/export_remaining_brand.py edda /path/to/this/repository`; use `--check` to verify the export without writing.

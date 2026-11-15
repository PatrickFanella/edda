# Edda

![Edda identity](docs/brand/edda/readme-lockup.png)

Edda is a tabletop roleplaying game run by a language model acting as game
master. You describe a campaign and a character; Edda builds the world, then
answers each action with narration, choices and changes to quests, characters,
places and combat. You play in the browser or in a terminal interface.

It is in development and is being playtested. Expect rough edges in the rules
and in how the game master tracks the world. A running copy is at
[edda.subcult.tv](https://edda.subcult.tv/).

The [SRD 5.2.1 conformance register](docs/srd-5.2.1-conformance.md) pins the
reference and records coverage. All initial scopes are unsupported pending
conformance tests.

## Features

- **Campaign setup:** a guided interview creates the campaign and the player
  character, then generates a world with locations, NPCs, cultures, belief
  systems, economies and languages.
- **Narrative turns:** each action gets a streamed game-master response,
  followed by choices and updated state. The game master changes the world
  through typed tools (damage, conditions, experience, time, quests and more)
  rather than free text alone.
- **Quests and objectives**, with branching quests, history and notes.
- **Character sheet:** abilities, skills, feats, inventory and progression.
- **Combat** with rounds and tracked combat state.
- **World tracking:** locations and a map, NPCs and their dialogue, relationships,
  world facts, a codex and campaign time.
- **Long-term memory:** turns and facts are embedded and stored in PostgreSQL
  with pgvector, so the game master can recall earlier events. Older sessions
  are summarized.
- **Journal, saves and replay:** a session log, save points, and a replay view
  of past play.
- **Export** a campaign as JSON, a transcript or a character sheet.
- **Accounts:** register and sign in; sessions use signed JWTs.
- **Model providers:** any Ollama-compatible endpoint, Anthropic Claude or
  OpenRouter. Routing profiles can send different jobs (main turn, choice
  fallback, post-turn state) to different models.
- **Two clients:** a React web app and a terminal interface built with Bubble
  Tea.

## Quick start

You need Go 1.25, Docker with Compose, [Task](https://taskfile.dev/), and
Node.js with pnpm for the web app. You also need access to a model: a local
Ollama server, or a Claude or OpenRouter API key.

```bash
cp .env.example .env
pnpm --dir frontend install
```

Edit `.env` before the first run:

- set `EDDA_DB_URL=postgres://edda:edda@localhost:5434/edda?sslmode=disable`,
  the database `task db:up` starts;
- set `EDDA_SERVER_PORT=18080`, the port the Vite dev server proxies `/api` to
  (or keep `8080` and set `VITE_PROXY_TARGET=http://localhost:8080` in
  `frontend/.env`);
- choose a model provider and model, and set `EDDA_SERVER_JWTSECRET`.

Then:

```bash
task db:up     # start PostgreSQL (pgvector) on localhost:5434 and migrate
task dev:web   # API server plus the Vite dev server on http://localhost:5173
```

Or play in the terminal:

```bash
task dev:tui
```

Other useful tasks (`task --list` shows all of them):

```bash
task test              # go test ./...
task test:unit         # skip integration tests
task test:integration
task verify            # Go tests and a production frontend build
task lint              # golangci-lint
task build             # bin/server, bin/tui and bin/logs
task generate          # regenerate sqlc database code
task migrate:create    # new goose migration
task reset:first-boot  # wipe local game data
```

Pull requests require the PostgreSQL integration job as part of `Mirror CI gate`.
It runs all tagged integration suites with `go test -tags integration -run Integration
-count=1 -timeout 10m ./...`. Each suite starts its own disposable PostgreSQL 16
container with pgvector, applies migrations, and removes the container on exit.
Docker must be available locally; no application database or credentials are used.
Container setup is limited to five minutes and cleanup to thirty seconds.

## Configuration

Settings come from built-in defaults (`internal/config/config.go`), then
environment variables with the `EDDA_` prefix. A variable name maps to a
config key by section: `EDDA_LLM_OLLAMA_MODEL` sets `llm.ollama.model`.
[`.env.example`](.env.example) lists every supported variable with comments.

| Variable                                   | Purpose                                                                                                       |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| `EDDA_DB_URL`                              | PostgreSQL connection string. The database needs pgvector.                                                    |
| `EDDA_SERVER_PORT`                         | API port.                                                                                                     |
| `EDDA_SERVER_JWTSECRET`                    | Signs session tokens. Generate with `openssl rand -hex 32`.                                                   |
| `EDDA_SERVER_ALLOWREGISTRATION`            | `false` stops new accounts being created. Defaults to `true`.                                                 |
| `EDDA_LLM_PROVIDER`                        | `ollama`, `claude`, `openrouter` or `chatgpt`.                                                                |
| `EDDA_LLM_FALLBACKPROVIDER`                | Optional provider that gets a request when `EDDA_LLM_PROVIDER` fails.                                         |
| `EDDA_LLM_OLLAMA_*`                        | Endpoint, model, API key, context budget, timeout and embedding model.                                        |
| `EDDA_LLM_MEMORYEMBEDDING_*`               | Independent semantic memory: enabled, provider, endpoint, model, dimension and optional API key.              |
| `EDDA_LLM_CLAUDE_*`                        | Model, reasoning effort and context budget. The key can also come from `ANTHROPIC_API_KEY`.                   |
| `EDDA_LLM_OPENROUTER_*`                    | Model, fallback models, reasoning effort and context budget. The key can also come from `OPENROUTER_API_KEY`. |
| `EDDA_LLM_CHATGPT_*`                       | Credentials file, model, reasoning effort and context budget for the ChatGPT-plan provider.                   |
| `EDDA_LLM_CLAUDECODE_*`                    | Bridge endpoint, bearer token and context budget for the `claudecode` provider.                               |
| `EDDA_LLM_PROFILES_*`, `EDDA_LLM_ROUTES_*` | Optional per-job model routing.                                                                               |

With `EDDA_LLM_OPENROUTER_FALLBACKMODELS` set, OpenRouter retries a failed
request (outage, rate limit, moderation flag, context limit) on each fallback model in
order, inside the same request; the API logs a warning when a fallback answered.
`EDDA_LLM_OPENROUTER_REASONINGEFFORT` (`none` to `max`) applies to the fallback
models too. For example, Claude Sonnet 5.5 at `high` effort with GPT-6.1 Terra
as the fallback, both billed to OpenRouter credits:

```bash
EDDA_LLM_PROVIDER=openrouter
EDDA_LLM_OPENROUTER_MODEL=anthropic/claude-sonnet-5.5
EDDA_LLM_OPENROUTER_FALLBACKMODELS=openai/gpt-6.1-terra
EDDA_LLM_OPENROUTER_REASONINGEFFORT=high
```

### Semantic campaign memory

The HTTP server configures semantic memory independently of the game-master
provider, including a GM selected through `llm.routes.gmturn` or a fallback.
The supported embedding protocol is Ollama's `POST /api/embed`, served by
Ollama or a compatible broker. The optional embedding API key is sent as a
Bearer token; JSON responses and broker SSE framing are supported.

| Game-master provider | Embedding configuration | Behavior |
| --- | --- | --- |
| `ollama`, `claude`, `openrouter`, `chatgpt`, `claudecode` | Explicit `llm.memoryembedding` with `enabled: true` and provider omitted or `ollama` | Records memories through world-building tools, adds relevant memories to turn context, and enables `search_memory`. |
| Top-level `ollama` | `llm.memoryembedding` omitted | Uses legacy `llm.ollama.embeddingendpoint` (or `llm.ollama.endpoint` when empty), `embeddingmodel`, `embeddingdimension` and `apikey`. |
| Any | Explicit `llm.memoryembedding.enabled: false` | Disables semantic recording, retrieval and `search_memory`, even with legacy Ollama settings present. Logs a startup warning. |
| Non-Ollama top-level provider | `llm.memoryembedding` omitted | Semantic memory is unavailable; startup logs how to configure it. Normal state and recent-turn context remain available. |
| Any | Enabled memory with another embedding provider, such as `openai` | Configuration error; startup stops. OpenAI-style `/v1/embeddings` is not supported. |

For a non-Ollama GM, keep its existing provider settings and configure memory
separately. For example, this uses a local Ollama-compatible embedder with the
repository's default `vector(768)` schema:

```bash
EDDA_LLM_MEMORYEMBEDDING_ENABLED=true
EDDA_LLM_MEMORYEMBEDDING_PROVIDER=ollama
EDDA_LLM_MEMORYEMBEDDING_ENDPOINT=http://localhost:11434
EDDA_LLM_MEMORYEMBEDDING_MODEL=nomic-embed-text
EDDA_LLM_MEMORYEMBEDDING_DIMENSION=768
```

Enabled explicit memory requires all of `endpoint`, `model` and a positive
`dimension`; it never borrows them or its API key from the GM configuration.
The endpoint must be an absolute HTTP(S) base URL, without `/api/embed`, URL
credentials, query parameters or a fragment. A broker path prefix is allowed.
Set `EDDA_LLM_MEMORYEMBEDDING_APIKEY` only when that endpoint needs a Bearer
token. Embedding requests use the existing `llm.ollama.timeoutseconds` timeout.

An explicit `memoryembedding` block takes precedence over legacy settings;
its `enabled` field defaults to false. Leave the entire block and its
environment variables unset to retain legacy Ollama behavior.

Configuration errors and a mismatch with the database's `memories.embedding`
dimension stop startup. Startup does not call the embedding model: it reports
that availability is checked on use. An unavailable endpoint, unsupported API,
missing model or wrong output dimension produces a retrieval warning; the
turn continues without semantic memories. A fact-recording tool returns a
`memory_warning` if its embedding fails after the fact was saved. Failed
embeddings are not automatically retried or backfilled.

Keep the same embedding model and dimension for an existing memory collection.
A new model with the same dimension can produce incompatible vectors; this
check cannot detect that. This setting does not convert, delete or re-embed
existing memories. Full turn-summary embedding and journal summarization are
separate facilities; this startup wiring records through existing tools.

The deterministic acceptance tests use local HTTP fixtures for a Claude GM
and an Ollama-compatible embedder, plus an in-memory persistence double. They
check fact recording, recall beyond recent history, delivery to the GM's turn
context, campaign scoping, `search_memory`, explicit disable, schema mismatch
and provider failures. They require no accounts or paid inference and do not
measure live-model retrieval quality or PostgreSQL ranking.

### Running on a ChatGPT plan

The `chatgpt` provider sends requests to OpenAI's Responses API and charges them
to a ChatGPT plan instead of API credits, through OpenAI's
[Sign in with ChatGPT](https://developers.openai.com/siwc/token-sharing-open-source)
preview for open-source and self-hosted apps. Use it only with your own plan for
your own play: every request counts against the plan of whoever signed in. On a
public server, set `EDDA_SERVER_ALLOWREGISTRATION=false` so nobody else can
create an account and spend it.

The production example runs Claude Sonnet 5.5 at `high` effort on a Claude
Console API key, and sends a request to the ChatGPT plan when Claude fails or
declines it. Claude API usage is billed to the Console account, separately from
any Claude.ai plan; a Claude.ai plan can't be used here, because Anthropic only
allows subscription sign-in for its own apps.

```bash
EDDA_LLM_PROVIDER=claude
EDDA_LLM_CLAUDE_APIKEY=sk-ant-...
EDDA_LLM_CLAUDE_MODEL=claude-sonnet-5-5
EDDA_LLM_CLAUDE_REASONINGEFFORT=high
EDDA_LLM_FALLBACKPROVIDER=chatgpt
EDDA_LLM_CHATGPT_CREDENTIALSFILE=/app/chatgpt/chatgpt-auth.json
EDDA_LLM_CHATGPT_MODEL=gpt-6.1-terra
EDDA_LLM_CHATGPT_REASONINGEFFORT=high
```

To sign in:

1. Run `go run ./cmd/chatgpt-login` on a machine with a browser. It opens the
   ChatGPT sign-in page, waits for the callback on `127.0.0.1:1455`, and writes
   `~/.config/edda/chatgpt-auth.json` (mode `600`). On a headless machine, run
   it with `-no-browser`, forward the port from your laptop with
   `ssh -L 1455:127.0.0.1:1455 <host>`, and open the printed link there.
2. It then lists the models your account can use. If `gpt-6.1-terra` is not
   among them, set `EDDA_LLM_CHATGPT_MODEL` to one that is.
3. Copy the file into the API container's `chatgpt-auth` volume and restart
   the API:

   ```bash
   docker cp ~/.config/edda/chatgpt-auth.json edda-api:/app/chatgpt/chatgpt-auth.json
   docker exec -u 0 edda-api chown 1001:1001 /app/chatgpt/chatgpt-auth.json
   docker exec -u 0 edda-api chmod 600 /app/chatgpt/chatgpt-auth.json
   docker restart edda-api
   ```

The server refreshes the token and rewrites its copy of the file. Each refresh
replaces the refresh token, so the local copy's tokens stop working after the
first one. Keep the local file anyway: signing in again with it reuses the same
registration instead of creating a new one. If the server logs that the sign-in is no
longer valid, repeat these steps. Plan usage and per-app limits are under
[ChatGPT settings → Usage](https://chatgpt.com/settings/usage).

### Running Claude on your own Claude subscription

The `claudecode` provider sends each request to `cmd/claude-bridge`. The bridge
runs the unmodified Claude Code CLI (`claude -p`), signed in with your own
Claude subscription. Claude Code's own tools are turned off: the bridge passes
the conversation and Edda's tool list as text, and asks for a JSON reply of
narration and tool calls, which Edda then runs as usual.

This route is for your own games only. Anthropic intends subscription sign-in
for ordinary use of Claude Code and its own apps, and asks apps that serve other
people to use an API key. Usage counts against the same limits as your other
Claude Code work. Keep `EDDA_SERVER_ALLOWREGISTRATION=false`, and keep a
fallback provider for when the limits run out.

1. On the machine where you use Claude Code, create a long-lived token with
   `claude setup-token` and save it to a file with mode `600`.
2. Create a bearer token for Edda with `openssl rand -hex 32 > bridge.token`.
3. Run the bridge with both:

   ```bash
   CLAUDE_CODE_OAUTH_TOKEN="$(cat claude-oauth-token)" \
     go run ./cmd/claude-bridge -listen 10.0.0.50:18095 -token-file bridge.token \
     -model claude-sonnet-5-5 -effort high
   ```

4. Point Edda at it:

   ```bash
   EDDA_LLM_PROVIDER=claudecode
   EDDA_LLM_CLAUDECODE_ENDPOINT=http://10.0.0.50:18095
   EDDA_LLM_CLAUDECODE_TOKEN=<contents of bridge.token>
   EDDA_LLM_FALLBACKPROVIDER=chatgpt
   ```

Each request starts a `claude` process, which adds a second or two to every
call; the bridge runs two at a time by default (`-concurrency`).

### Game-master skills

[`skills/`](skills/) holds game-mastering and fiction skills (running RPG
sessions, worldbuilding, fiction drafting and editing, branching stories,
names, fictional languages and game design); see
[`skills/NOTICE.md`](skills/NOTICE.md) for their origin and license. A profile
with `skills` set gets two read-only tools, `load_skill` and `read_skill_file`,
which Edda answers itself from `llm.skillsdir` (default `skills`, copied into
the image). The engine never sees them, and they work the same with every
provider. Each skill the model reads adds a round trip, so turn them on for
the GM-turn profile only:

```bash
EDDA_LLM_PROFILES_GM_SKILLS=true
EDDA_LLM_PROFILES_GM_TOOLSUPPORT=true
EDDA_LLM_ROUTES_GMTURN=gm
```

The profile uses `EDDA_LLM_PROVIDER` and `EDDA_LLM_FALLBACKPROVIDER` unless it
sets its own provider, so both the main model and the fallback get the skills.

Embeddings always use the Ollama-compatible settings
(`EDDA_LLM_OLLAMA_EMBEDDINGENDPOINT`, `EDDA_LLM_OLLAMA_EMBEDDINGMODEL`,
`EDDA_LLM_OLLAMA_EMBEDDINGDIMENSION`).

## Deploy

`docker-compose.yml` builds two images: `edda-api` (the Go server, port 8080
in the container) and `edda-web` (the built frontend served by nginx). The
`postgres` service is in the `local` profile, so production uses your own
PostgreSQL with pgvector.

1. Copy `.env.production.example` to `.env.production` (mode `600`, ignored by
   Git) and fill it in. Set `EDDA_BIND_ADDRESS` to the address the containers should publish
   on (for example `127.0.0.1`); the default in the Compose file is a private
   LAN address. `GM_API_HOST_PORT` and `GM_WEB_HOST_PORT` choose the host ports.
2. The API and web services join an external Docker network named `projects`.
   Create it (`docker network create projects`) or edit the Compose file.
3. Build, migrate and start:

   ```bash
   make app-build ENV_FILE=.env.production
   make migrate-prod ENV_FILE=.env.production
   make app-up ENV_FILE=.env.production
   make app-status ENV_FILE=.env.production
   ```

4. Put an HTTPS reverse proxy in front of the web container.

Images are tagged with the current commit unless you pass `RELEASE_TAG`.
`make db-backup ENV_FILE=.env.production` writes a timestamped database backup. `make help` lists the
other targets.

## Project layout

```text
cmd/server     API server
cmd/tui        terminal client
cmd/logs       structured log viewer
internal/      game engine, game-master prompts and tools, world, combat,
               memory, saves, auth, HTTP handlers and config
tui/           terminal interface (Bubble Tea)
frontend/      web app (React, Vite, Tailwind)
migrations/    goose SQL migrations
queries/       sqlc queries
docs/          architecture notes and the brand guide
```

## Contributing

- Report bugs and request features in [GitHub issues](https://github.com/subculture-collective/edda/issues).
- Ask questions and share ideas in [GitHub Discussions](https://github.com/subculture-collective/edda/discussions).
- Send code as a pull request on GitHub against `main`. CI runs the Go tests
  and builds the frontend image on every pull request, and both must pass
  before it can merge.

## License

GNU General Public License, version 3 or later (`GPL-3.0-or-later`). See
[LICENSE](LICENSE). The Edda identity and its source hashes are recorded in
the [brand guide](docs/brand/edda/GUIDE.md).

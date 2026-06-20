# Edda API handoff for iOS frontend

This document is for an iOS developer building a native frontend for the Edda game API.

## API basics

- Default base path: `/api/v1`
- Local default server: `http://localhost:8080`
- Production hosts currently allowed by CORS include `https://gm.subcult.tv` and `https://edda.subcult.tv`.
- Request bodies and normal responses are JSON unless an endpoint is marked as a file download.
- Dates/times are returned as ISO-8601 strings.
- IDs are UUID strings.
- List endpoints are not paginated.
- Error responses use:

```json
{ "error": "message" }
```

Recommended iOS client behavior:

- Set `Accept: application/json` for API requests.
- Set `Content-Type: application/json` when sending JSON bodies.
- Store JWTs in Keychain, not UserDefaults.
- Send `Authorization: Bearer <token>` for authenticated requests.
- Use `URLSessionWebSocketTask` for live play.

## Health checks

| Method | Path | Response |
|---|---|---|
| GET | `/healthz` | Plain text `ok` |
| GET | `/api/healthz` | `{ "status": "ok", "engine_ready": true }` |

## Authentication

The backend has two modes:

1. **JWT mode**: if the server has `JWTSecret` configured, auth endpoints are enabled and all `/api/v1` non-auth routes require a token.
2. **No-op mode**: if no JWT secret is configured, auth routes may not exist and the server uses a default user. This is mainly for local/TUI use.

The web app also receives an HttpOnly cookie named `gm_token`, but native iOS should treat the JSON `token` field as the primary credential and send it in the `Authorization` header.

### Auth endpoints

#### Register

`POST /api/v1/auth/register`

Request:

```json
{
  "name": "Player Name",
  "email": "player@example.com",
  "password": "secret"
}
```

Success `201`:

```json
{
  "token": "jwt",
  "user": {
    "id": "uuid",
    "name": "Player Name",
    "email": "player@example.com"
  }
}
```

Common errors: `400`, `409` for duplicate email, `500`.

#### Login

`POST /api/v1/auth/login`

Request:

```json
{ "email": "player@example.com", "password": "secret" }
```

Success `200`: same shape as register.

Common errors: `400`, `401`, `500`.

#### Current user

`GET /api/v1/auth/me`

Success `200`:

```json
{
  "user": {
    "id": "uuid",
    "name": "Player Name",
    "email": "player@example.com"
  }
}
```

#### Logout

`POST /api/v1/auth/logout`

Success: `204 No Content`.

Native clients should also delete the Keychain token locally.

## Core flow for the iOS app

1. Authenticate user, if JWT mode is enabled.
2. List campaigns with `GET /campaigns`.
3. Either open an existing campaign or run startup wizard:
   - campaign interview
   - proposals/name generation, optional
   - character interview
   - world build
4. Load campaign detail panels: character, history, inventory, quests, map, NPCs, facts, journal.
5. For live gameplay, open WebSocket `/campaigns/{id}/ws` and send player actions.
6. Use REST `POST /campaigns/{id}/action` as a non-WebSocket fallback.

## Campaign endpoints

All paths below are relative to `/api/v1`.

| Method | Path | Request | Response |
|---|---|---|---|
| GET | `/campaigns` | - | `CampaignListResponse` |
| POST | `/campaigns` | `CampaignCreateRequest` | `CampaignResponse` |
| GET | `/campaigns/{id}` | - | `CampaignResponse` |
| PUT | `/campaigns/{id}` | `CampaignCreateRequest` | `CampaignResponse` |
| DELETE | `/campaigns/{id}` | - | `204 No Content` |
| GET | `/campaigns/{id}/history` | - | `SessionHistoryResponse` |

`CampaignCreateRequest`:

```json
{
  "name": "The Glass Wastes",
  "description": "A dangerous desert campaign.",
  "genre": "fantasy",
  "tone": "gritty",
  "themes": ["survival", "mystery"],
  "rules_mode": "narrative"
}
```

`rules_mode` is optional. Known modes are `narrative` and `crunch`.

`CampaignResponse`:

```json
{
  "id": "uuid",
  "name": "The Glass Wastes",
  "description": "A dangerous desert campaign.",
  "genre": "fantasy",
  "tone": "gritty",
  "themes": ["survival", "mystery"],
  "status": "active",
  "rules_mode": "narrative",
  "created_by": "uuid",
  "created_at": "2026-06-12T12:00:00Z",
  "updated_at": "2026-06-12T12:00:00Z"
}
```

## Startup wizard endpoints

These endpoints create a campaign through guided LLM interviews.

### Campaign interview

Start:

`POST /api/v1/campaigns/start/campaign-interview`

Request body: none.

Continue:

`POST /api/v1/campaigns/start/campaign-interview/{sessionID}`

Request:

```json
{ "input": "I want dark fantasy with political intrigue." }
```

Response:

```json
{
  "session_id": "uuid",
  "message": "Assistant prompt or follow-up question",
  "done": false,
  "profile": null
}
```

When `done` is `true`, `profile` is present:

```json
{
  "genre": "dark fantasy",
  "tone": "gritty",
  "themes": ["survival", "betrayal"],
  "world_type": "urban sprawl",
  "danger_level": "high",
  "political_complexity": "complex"
}
```

### Campaign proposals

`POST /api/v1/campaigns/start/proposals`

Request:

```json
{
  "genre": "fantasy",
  "setting_style": "urban gothic",
  "tone": "grim"
}
```

Response:

```json
{
  "proposals": [
    {
      "name": "Ash Under Glass",
      "summary": "Short campaign pitch.",
      "profile": { }
    }
  ]
}
```

### Generate campaign name

`POST /api/v1/campaigns/start/name`

Request:

```json
{ "profile": { } }
```

Response:

```json
{ "name": "Ash Under Glass" }
```

### Character interview

Start:

`POST /api/v1/campaigns/start/character-interview`

Request:

```json
{ "campaign_profile": { } }
```

Continue:

`POST /api/v1/campaigns/start/character-interview/{sessionID}`

Request:

```json
{ "input": "I want to play an exiled cartographer." }
```

Response shape matches campaign interview. When complete, `profile` is:

```json
{
  "name": "Mira Vale",
  "concept": "exiled cartographer",
  "background": "Former royal mapmaker accused of treason.",
  "personality": "Careful, curious, guarded.",
  "motivations": ["clear her name"],
  "strengths": ["navigation", "observation"],
  "weaknesses": ["distrustful"]
}
```

### Build world

`POST /api/v1/campaigns/start/world`

Request:

```json
{
  "name": "Ash Under Glass",
  "summary": "Short campaign summary.",
  "profile": { },
  "character_profile": { },
  "rules_mode": "narrative"
}
```

Response:

```json
{
  "campaign": { },
  "opening_scene": {
    "narrative": "Opening scene text.",
    "choices": ["Ask the guard", "Search the alley"]
  }
}
```

## Gameplay endpoints

### Submit action over REST

`POST /api/v1/campaigns/{id}/action`

Request:

```json
{ "input": "I inspect the strange door." }
```

Response:

```json
{
  "narrative": "The door is warm to the touch...",
  "state_changes": [
    {
      "entity_type": "location",
      "entity_id": "uuid",
      "change_type": "discovered",
      "details": { }
    }
  ],
  "combat_active": false
}
```

### Live gameplay WebSocket

`GET /api/v1/campaigns/{id}/ws`

Use `ws://` for HTTP and `wss://` for HTTPS.

Auth:

- Preferred native approach: include `Authorization: Bearer <token>` if your WebSocket API permits custom headers.
- If using cookie auth, make sure `gm_token` is present in the shared cookie store before opening the socket.

Client sends:

```json
{
  "type": "action",
  "payload": { "input": "I draw my blade." }
}
```

Server envelopes:

```json
{
  "type": "chunk",
  "payload": { "text": "The blade rings..." },
  "timestamp": "2026-06-12T12:00:00Z"
}
```

```json
{
  "type": "status",
  "payload": {
    "stage": "combat_start",
    "tool": "optional-tool-name",
    "description": "Resolving combat..."
  },
  "timestamp": "2026-06-12T12:00:00Z"
}
```

```json
{
  "type": "result",
  "payload": {
    "narrative": "Full final turn response.",
    "state_changes": [],
    "combat_active": true
  },
  "timestamp": "2026-06-12T12:00:00Z"
}
```

```json
{
  "type": "error",
  "payload": { "error": "message" },
  "timestamp": "2026-06-12T12:00:00Z"
}
```

Recommended UX:

- Render `chunk` text as streaming narrative.
- Show `status.description` as a progress label.
- Treat `result` as the authoritative final state for the turn.
- Disable the action composer while awaiting a result.
- Reconnect with exponential backoff after unexpected close.

## Campaign detail endpoints

| Method | Path | Response |
|---|---|---|
| GET | `/campaigns/{id}/character` | `CharacterResponse` |
| GET | `/campaigns/{id}/character/inventory` | `ItemResponse[]` |
| GET | `/campaigns/{id}/character/abilities` | `CharacterAbility[]` |
| GET | `/campaigns/{id}/character/feats` | `FeatResponse[]` |
| GET | `/campaigns/{id}/character/skills` | `SkillResponse[]` |
| GET | `/campaigns/{id}/locations` | `LocationResponse[]` |
| GET | `/campaigns/{id}/locations/{lid}` | `LocationResponse` |
| GET | `/campaigns/{id}/npcs` | `NPCResponse[]` |
| GET | `/campaigns/{id}/npcs/{nid}` | `NPCResponse` |
| GET | `/campaigns/{id}/npcs/encountered` | `EncounteredNPCResponse[]` |
| GET | `/campaigns/{id}/npcs/{nid}/dialogue` | `DialogueEntry[]` |
| GET | `/campaigns/{id}/quests` | `QuestResponse[]` |
| GET | `/campaigns/{id}/quests?type=main&status=active` | `QuestResponse[]` filtered |
| GET | `/campaigns/{id}/quests/{qid}` | `QuestResponse` |
| GET | `/campaigns/{id}/facts` | `FactResponse[]` |
| GET | `/campaigns/{id}/relationships` | `RelationshipResponse[]` |
| GET | `/campaigns/{id}/codex/languages` | `LanguageResponse[]` |
| GET | `/campaigns/{id}/codex/cultures` | `CultureResponse[]` |
| GET | `/campaigns/{id}/codex/beliefs` | `BeliefSystemResponse[]` |
| GET | `/campaigns/{id}/codex/economies` | `EconomicSystemResponse[]` |
| GET | `/campaigns/{id}/map` | `MapDataResponse` |

## Quest notes and history

| Method | Path | Request | Response |
|---|---|---|---|
| GET | `/campaigns/{id}/quests/{qid}/notes` | - | `QuestNoteResponse[]` |
| POST | `/campaigns/{id}/quests/{qid}/notes` | `{ "content": "note" }` | `QuestNoteResponse` |
| DELETE | `/campaigns/{id}/quests/{qid}/notes/{noteID}` | - | `204 No Content` |
| GET | `/campaigns/{id}/quests/{qid}/history` | - | `QuestHistoryEntry[]` |

## Saves and campaign time

| Method | Path | Request | Response |
|---|---|---|---|
| GET | `/campaigns/{id}/saves` | - | `SavePointResponse[]` |
| POST | `/campaigns/{id}/saves` | `{ "name": "Before the fight" }` | `SavePointResponse` |
| POST | `/campaigns/{id}/start-over` | - | `204 No Content` |
| GET | `/campaigns/{id}/time` | - | `CampaignTimeResponse` |

Default time, when not initialized, is day `1`, hour `8`, minute `0`.

## Journal endpoints

| Method | Path | Request | Response |
|---|---|---|---|
| GET | `/campaigns/{id}/journal/summaries` | - | `SessionSummaryResponse[]` |
| GET | `/campaigns/{id}/journal/entries` | - | `JournalEntryResponse[]` |
| POST | `/campaigns/{id}/journal/entries` | `{ "title": "optional", "content": "required" }` | `JournalEntryResponse` |
| DELETE | `/campaigns/{id}/journal/entries/{eid}` | - | `204 No Content` |
| POST | `/campaigns/{id}/journal/summarize` | `{ "from_turn": 1, "to_turn": 10 }` or `{}` | `SessionSummaryResponse` or `{ "message": "no unsummarized turns found" }` |

## Export endpoints

These are authenticated file downloads.

| Method | Path | Content-Type | Notes |
|---|---|---|---|
| GET | `/campaigns/{id}/export/json` | `application/json` | Attachment `campaign-{id}.json` |
| GET | `/campaigns/{id}/export/transcript` | `text/markdown; charset=utf-8` | Attachment `transcript-{id}.md` |
| GET | `/campaigns/{id}/export/character` | `text/markdown; charset=utf-8` | Attachment `character-{id}.md` |

Use a download flow or share sheet rather than trying to decode all export responses as normal API JSON.

## Response model reference

Use optional Swift properties for fields marked optional here. `stats`, `properties`, and `details` are arbitrary JSON dictionaries.

```ts
interface CharacterResponse {
  id: string;
  campaign_id: string;
  user_id: string;
  name: string;
  description: string;
  stats: Record<string, unknown>;
  hp: number;
  max_hp: number;
  experience: number;
  level: number;
  status: string;
  abilities: CharacterAbility[];
  current_location_id?: string;
}

interface CharacterAbility {
  name: string;
  description?: string;
}

interface LocationResponse {
  id: string;
  campaign_id: string;
  name: string;
  description: string;
  region: string;
  location_type: string;
  properties: Record<string, unknown>;
  connections: LocationConnectionResponse[];
}

interface LocationConnectionResponse {
  from_location_id?: string;
  to_location_id: string;
  description: string;
  bidirectional: boolean;
  travel_time: string;
}

interface NPCResponse {
  id: string;
  campaign_id: string;
  name: string;
  description: string;
  personality: string;
  disposition: number;
  faction_id?: string;
  faction?: string;
  alive: boolean;
  hp?: number;
  stats: Record<string, unknown>;
  properties: Record<string, unknown>;
}

interface QuestResponse {
  id: string;
  campaign_id: string;
  parent_quest_id?: string;
  title: string;
  description: string;
  quest_type: string;
  status: string;
  objectives: QuestObjectiveResponse[];
}

interface QuestObjectiveResponse {
  id: string;
  description: string;
  completed: boolean;
  order_index: number;
}

interface ItemResponse {
  id: string;
  campaign_id: string;
  player_character_id?: string;
  name: string;
  description: string;
  item_type: string;
  rarity: string;
  properties: Record<string, unknown>;
  equipped: boolean;
  quantity: number;
}

interface SessionHistoryResponse {
  entries: Array<{
    turn_number: number;
    player_input: string;
    input_type: string;
    llm_response: string;
    choices?: string[];
    created_at: string;
  }>;
}

interface MapDataResponse {
  locations: Array<{
    id: string;
    campaign_id: string;
    name: string;
    description: string;
    region: string;
    location_type: string;
    player_visited: boolean;
    player_known: boolean;
  }>;
  connections: LocationConnectionResponse[];
}

interface SavePointResponse {
  id: string;
  campaign_id: string;
  name: string;
  turn_number: number;
  is_auto: boolean;
  created_at: string;
}

interface CampaignTimeResponse {
  day: number;
  hour: number;
  minute: number;
}
```

## Swift implementation notes

For snake_case JSON, either set a decoder strategy:

```swift
let decoder = JSONDecoder()
decoder.keyDecodingStrategy = .convertFromSnakeCase
decoder.dateDecodingStrategy = .iso8601
```

or define explicit `CodingKeys` if you want Swift property names to stay close to API field names.

Represent arbitrary JSON fields with a custom `JSONValue` enum, for example:

```swift
enum JSONValue: Codable, Equatable {
    case string(String)
    case number(Double)
    case bool(Bool)
    case object([String: JSONValue])
    case array([JSONValue])
    case null
}
```

Recommended modules/classes:

- `APIClient`: base URL, JSON encoding/decoding, auth header injection, error decoding.
- `AuthStore`: Keychain-backed token storage.
- `CampaignService`: campaign CRUD and detail-panel endpoints.
- `StartupService`: startup interview/proposal/world-build endpoints.
- `GameplaySocket`: WebSocket connection, reconnection, envelope parsing.
- `ExportService`: authenticated file downloads.

## Known integration cautions

- No pagination exists yet; avoid assuming cursor/page metadata.
- The WebSocket result payload is the final turn state, while `chunk` events are incremental display text.
- Some endpoints return raw arrays, while campaign list returns `{ "campaigns": [...] }`.
- Export endpoints should not use the normal JSON-only request path.
- Startup interview sessions are in-memory; if the backend restarts, continue calls may return `404` and the app should restart the relevant interview.
- `GET /api/v1/auth/me` returns `{ "user": ... }` and no token.

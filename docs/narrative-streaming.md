# Incremental provisional narration (#41)

Live turns stream provider text as it arrives, before generation and persistence
finish. This is a preview of a response that can still change after tool execution,
claim repair, choice parsing or a failed save. The browser labels it
**Provisional · not saved**, keeps it separate from its committed entries and hides
turn details and choices until the final result arrives.

The existing four event kinds remain unchanged:

- `chunk` appends `payload.text` to the current provisional preview. It does not
  append a transcript entry or certify a mechanical outcome.
- `status` with `stage: narrative_reset` discards the entire preview. The next
  fragments replace it. Other statuses retain their existing meaning.
- `result` replaces the preview with the exact persisted turn response, including
  its final narrative, turn number, choices and player-visible receipts. It is the
  only live event that adds committed GM narration. The engine no longer sends a
  duplicate full-narrative chunk after saving.
- `error` discards the preview and leaves committed history intact. A socket
  failure, malformed message, disconnect or campaign switch also discards it.

Every narrative generation request starts by resetting its preview. Initial
provider retries and empty-response regeneration therefore cannot concatenate
abandoned text with a replacement. Tool processing clears the initial preview
before executing calls, and durable-claim correction clears it before repair.
Continuation, repair and safe-rewrite narration can stream replacement fragments;
auxiliary tool retries, state extraction and choice generation remain internal.
The final persisted result always replaces all provisional wording, including
provider formatting and choice markers. These previews can contain mistakes:
the label is an explicit uncertainty boundary, not a fidelity guarantee.

Streaming options and publication callbacks belong to the individual request.
Synchronous HTTP/shared-engine/TUI turns retain their completion path and the
`GameEngine` interface is unchanged. Legacy rules modes and persisted formats do
not change. Native streaming providers supply real fragments; adapters backed by
a synchronous CLI or replay cache may still supply one fragment. No artificial
fragmentation or extra inference is used to simulate streaming.

Provider streams must assemble complete tool calls, including arguments, and
send `Done` only on successful completion. An early close is a failed response;
its partial calls are never dispatched. Malformed Claude tool arguments fail the
whole attempt without caching a partial assistant replay. OpenRouter top-level errors and
`finish_reason: error`, and Claude refusal stop reasons, cannot complete a turn.
Claude retains signed thinking/redacted blocks in its private replay cache for
tool-result follow-ups; those blocks are never emitted to the player.

A configured fallback applies to failures after a stream opens as well as opening
errors. The provider's internal `Reset` chunk discards abandoned text and tool
calls before the secondary starts; the engine translates it into the existing
`narrative_reset` status. Only a successful provider attempt can supply calls for
dispatch. Cancellation never starts a fallback. These internal chunks add no new
WebSocket event kind. Initial generation can use its existing
bounded transient retry, while post-mechanical recovery retains its existing
durable budget. OpenRouter now assembles split/interleaved arguments, Ollama
forwards every call in a frame, and ChatGPT no longer marks a failed stream done.

The 16-event engine buffer and context-aware sends remain bounded. The browser
continues synchronous consumption and sequence acknowledgment; it does not retain
raw frames or reset the sequence on a narration revision. The optional pending
consumer still fails closed at 256 frames. Only the current preview text is kept
until reset/completion/failure; it is not a second transcript.

Completed database turns explicitly unlock their dedicated PostgreSQL session
and wait for acknowledgment before returning, avoiding transient conflicts on
immediate sequential requests. No pool slot or transaction spans inference.

Cancellation ownership is unchanged from [#37](investigations/issue-37-stream-cleanup.md).
The consumer must drain or cancel; cancellation releases a blocked publisher and
the engine's wait for provider chunks. A disconnected consumer is owed no terminal
event. Detecting a silent peer remains #105. Streaming does not roll back already
committed tools or authorize a new action after ambiguous delivery.

[Committed-action recovery](turn-completion.md) remains authoritative: retrying
the same action ID uses recorded effects and candidates, without replaying tools
or dice. If recovery needs inference it streams provisional text with no tools;
validated cached candidates/completed actions return their persisted result
without a new preview or inference. No provisional text enters history, replay,
summaries or exports by virtue of being displayed.

## Deterministic checks

`TestProcessTurnStream_ProviderFragmentsPrecedePersistence` gates a local HTTP
provider between flushed fragments and before completion. It checks the actual
Ollama adapter and engine, zero saved history during both fragments, exact final
persistence, and no committed result on save failure. Processor tests cover
interrupted-provider retries, claim correction, incomplete tool calls and silent
provider cancellation. Adapter tests cover assembled interleaved arguments,
multiple calls per frame and failed completion markers.

`provisionalNarration.test.tsx` exercises the real socket hook, narrative store and
display through correction, errors, malformed frames, disconnect/reconnect and
same-ID recovery. `browser/provisional-narration.spec.ts` holds completion while
Chromium renders separate fragments at phone and desktop widths, then checks
correction, failure, retry identity and final replacement. The existing retention,
campaign-switch and viewport suites remain part of local qualification. These
fixtures use no production services, model accounts or paid inference.

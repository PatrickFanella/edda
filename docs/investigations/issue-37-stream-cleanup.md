# Disconnect and stream cleanup decision (#37)

Investigated on 2026-11-15 against `faa9c110ff159b081c56967cb0de4fdba4182bf3`.
Decision: confirmed unsafe backpressure; contain event publication and record the
remaining lifetime decision in [#105](https://github.com/subculture-collective/edda/issues/105).

## Reproduction

`TestProcessTurnStream_CancelWithFullBuffer` uses the actual engine, an in-memory
state manager and a provider returning 64 harmless tool calls. The consumer
stops at the 16-frame capacity and cancels. On the original code, even cancellation
followed by draining delivered 78 frames and reached session-log persistence.
Without that test-only drain, the unconditional send had no cancellation escape.
This establishes an engine publication failure, not a production goroutine count.

The containment gives every send (status callback, phase timing, combat status,
chunk, result, error and panic recovery) the same context-aware publication path.
Canceled streams may drop terminal frames; the consumer that abandoned them must
not require a completion event. Stage boundaries and the tool loop stop starting
new work after cancellation. A racing send can finish if its consumer becomes
ready at the same instant; total publication remains bounded.

`TestStreamPublication_CancellationReleasesAllEventKinds` proves publishers exit
within one second with a full, undrained buffer for all four event kinds.
`TestHandleWebSocket_DisconnectCancelsProducer` uses an actual ephemeral-port
HTTP/WebSocket connection, disconnects during a controlled pause and then releases
a producer. Its failed write cancels and stops the producer within one second,
with both a zero timeout and a one-minute timeout. The zero-timeout case required
fixing the turn-context helper's no-op cancel function. No server entrypoint changed.

## Counterevidence and boundaries

The old normal stream/channel-close and status-isolation tests pass: a consumer
that keeps draining does not demonstrate this failure. Real WebSocket transport
and the real engine are exercised in separate fixtures; no real database, model,
paid inference or production service participates. The full-buffer fixture
records zero persistence after cancellation and no additional executed tools.
That does not roll back tools committed before cancellation.

The handler reads between turns, so a silent dependency may not notice a peer
disconnect until the next write or configured turn timeout. Dependencies must
observe context to stop their own work. #105 bounds the next PR: select turn
ownership/lifetime semantics, add a silent-provider disconnect test and implement
the appropriate read/heartbeat handling while preserving #59 committed recovery.
Changing that policy or auditing every provider/driver is beyond this spike.

## Local checks

- `go test -race ./internal/engine ./internal/handlers -count=1`: passed.
- Targeted cancellation/disconnect tests repeated with `-race -count=30`: passed.
- `go vet ./...`: passed.
- `gofmt` applied; `git diff --check` passed.
- `golangci-lint` is unavailable under the active Go 1.27.3 installation. The
  installed Go 1.27.0 variant was tried once; its analyzer panicked in dependency
  `poll` (`buildir: unexpected expr: *ast.KeyValueExpr`). This is a tool failure,
  not a passed lint run. No lint configuration was weakened.

Hosted PostgreSQL integration and CI gate results are separate; the parent owns
monitoring. No deployment or production-data change is part of this work.

## Turn ownership decision (#105)

The WebSocket owns active work until disconnect or the configured turn deadline.
One connection reader runs throughout inference, including periods with no
provider output. A read failure cancels the turn context immediately; the event
loop does not wait for the provider before releasing the HTTP handler. Actions
received during a turn are rejected, not queued for later execution. The engine's
campaign lock remains the cross-connection authority against overlapping turns.

Cancellation stops cooperative providers/tools; it cannot force a dependency
that ignores context to return. The handler does not create another reader or
kill that dependency. Its engine worker retains the campaign lock until it exits.
A silent, half-open network that supplies neither EOF nor a close frame is bounded
by the configured turn timeout, not by an application heartbeat. Deployments must
keep a finite turn timeout. No output is required for a received peer close to
cancel work.

Committed effects survive disconnect and timeout. Reconnect reads the existing
`GET /action` receipt, then retries the original input with the same action ID.
The #59 journal replays completed results or recovers narration without rerolling;
uncommitted work can restart. HTTP actions keep their existing server-owned
lifetime through their deadline, independent of the request connection.

`TestHandleWebSocket_SilentTurnLifetime` separates cancellation delivery, provider
exit and worker exit with two-second test deadlines, including a provider that
intentionally ignores cancellation until released. It exercises timeout and
actual socket disconnect without releasing narrative output. The busy-action
fixture proves that an action sent during silent work receives an error. Existing
narration-recovery integration fixtures cover exact committed dice/effects after
restart and retry; these are separate from transport lifetime tests.

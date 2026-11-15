# Live-event consumption (#109)

The narrative store reduces each decoded frame synchronously and acknowledges its
monotonic local sequence before releasing that payload. React subscribes to
narrative snapshots, so a burst exceeding 256 frames loses neither results nor
errors. Durable transcript entries are not capped. Status frames retain only
the latest display state.

A transport caller without a consumer must acknowledge pending sequences. Its
queue stops at 256 frames (over 23 turns at the #43 baseline of 11 frames/turn).
Overflow closes the socket, disables further actions, and asks for a refresh to
recover committed history; it never silently slices results. This bound is a
failure policy for an abandoned consumer, not a transcript limit.

The existing server WebSocket handler runs a fresh stream only in response to
an action message on that connection. It does not subscribe to or replay old
frames on reconnect. The local sequence survives reconnect, resets on campaign
change, and is not a server replay cursor. No wire event IDs were changed.
Results now carry the committed turn number as an additive field, allowing
history and live player/GM entries to reconcile by durable turn identity. Old
server fixtures without that field retain local timestamp/sequence IDs.
Errors and chunks are transient and are not replayed across connections. A page
reload restores committed session history; interrupted narration recovery is #59.

Checks cover the 1,000-turn/11,000-frame #43 probe, a 1,800-frame batched burst,
queue overflow, acknowledgments, delayed history, campaign changes, reconnect,
and unmount. The bounded probe retains zero serialized transport entries after
consumption at both 100 and 1,000 turns (the prior baseline retained 4,009,787
serialized bytes at 1,000 turns). This is deterministic fixture evidence, not a
production heap measurement.

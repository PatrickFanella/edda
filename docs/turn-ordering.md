# Campaign turn ordering

The shared engine rejects overlapping turns for a campaign with `ErrTurnInProgress`. REST returns HTTP 409; streaming clients receive an error event. Retry after the active turn finishes. The guard covers the whole pipeline, starting before history is loaded and ending after completion or failure. Web and local TUI turns share this boundary.

PostgreSQL transaction advisory locks coordinate separate server/TUI processes. Each active turn opens a dedicated lock connection, released on return, error or panic; the query pool remains available to other campaigns even with a single query connection. Budget one additional database connection per active turn. In-memory engines use a process-local guard.

Migration 37 enforces unique `(campaign_id, turn_number)` pairs. It refuses existing duplicates rather than discarding or renumbering history. Existing campaign data is disposable for this release: reset the disposable database at release if duplicates prevent migration. Implementation tests never reset a production database.

Run `go test -tags integration -race -run IntegrationCampaignTurnOrder ./internal/handlers` for a barrier-controlled model test against temporary PostgreSQL. It checks two engine instances, synchronous/streaming overlap rejection, independent campaign progress with a one-connection query pool, subsequent turn numbering and direct duplicate-insert refusal.

This change does not make all turn mutations atomic; that is tracked separately in #30. Start Over coordination with active turns is also separate.

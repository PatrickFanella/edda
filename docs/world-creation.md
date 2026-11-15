# Generated campaign creation

The world-creation wizard has three stages: draft the world and opening scene in memory, save the complete starting state in one database transaction, then return the committed active campaign. No campaign is visible in the list while it is being drafted or saved. The model runs before the transaction begins.

A generation, persistence or commit failure returns an error and no campaign result. PostgreSQL rolls back the campaign, world entities, character, spawn package and opening scene together. Cancellation also rolls back using a separate, bounded cleanup context. Retrying the same wizard input after a known failed creation starts from empty state and cannot duplicate partially saved entities. No migration or existing campaign conversion is needed.

A lost response after a successful commit is different from a failed creation: check the campaign list before submitting the wizard again. This endpoint does not deduplicate separate successful requests. Existing partial campaigns created by older versions are not automatically deleted.

`TestIntegrationWorldCreationRollbackAndRetry` injects failures into each persistence stage, deferred commit, both model calls and cancellation. It checks rollback and a clean retry against isolated PostgreSQL, without paid inference.

The campaign clock, saved rules mode and crunch feat/skill definitions are required parts of the creation transaction. Any failed setup write stops creation and rolls back the starting state. An omitted rules mode is saved as `narrative`; a successful response uses the rules mode returned by the database. Setup uses the transaction connection rather than a separately supplied pool, so omitting that pool cannot silently skip initialization. This preserves the existing legacy feat/skill catalogue; it does not claim SRD character-build conformance.

`TestStartupWorldSetupReportsEveryWriteFailure` checks the clock, mode, every feat insert and every skill insert. `TestIntegrationWorldCreationRequiredSetup` verifies HTTP failure responses and rollback at each setup stage, plus saved rules, clock, character, opening scene and catalogue counts for narrative and crunch creation.

# Generated campaign creation

The world-creation wizard has three stages: draft the world and opening scene in memory, save the complete starting state in one database transaction, then return the committed active campaign. No campaign is visible in the list while it is being drafted or saved. The model runs before the transaction begins.

A generation, persistence or commit failure returns an error and no campaign result. PostgreSQL rolls back the campaign, world entities, character, spawn package and opening scene together. Cancellation also rolls back using a separate, bounded cleanup context. Retrying the same wizard input after a known failed creation starts from empty state and cannot duplicate partially saved entities. No migration or existing campaign conversion is needed.

A lost response after a successful commit is different from a failed creation: check the campaign list before submitting the wizard again. This endpoint does not deduplicate separate successful requests. Existing partial campaigns created by older versions are not automatically deleted.

`TestIntegrationWorldCreationRollbackAndRetry` injects failures into each persistence stage, deferred commit, both model calls and cancellation. It checks rollback and a clean retry against isolated PostgreSQL, without paid inference.

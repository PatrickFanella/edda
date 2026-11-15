-- +goose Up
ALTER TABLE turn_attempts
 ADD COLUMN candidate_result JSONB,
 ADD COLUMN narration_ready BOOLEAN NOT NULL DEFAULT false,
 ADD COLUMN narration_attempts INTEGER NOT NULL DEFAULT 0 CHECK (narration_attempts >= 0);
-- Earlier pending_narration rows were written only after processing finished.
UPDATE turn_attempts SET narration_ready=true WHERE status='pending_narration';

-- +goose Down
ALTER TABLE turn_attempts DROP COLUMN narration_attempts, DROP COLUMN narration_ready, DROP COLUMN candidate_result;

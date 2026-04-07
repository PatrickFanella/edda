-- name: GetCampaignTime :one
SELECT campaign_id, day, hour, minute, updated_at
FROM campaign_time
WHERE campaign_id = $1;

-- name: UpsertCampaignTime :one
INSERT INTO campaign_time (campaign_id, day, hour, minute)
VALUES ($1, $2, $3, $4)
ON CONFLICT (campaign_id) DO UPDATE
SET day = EXCLUDED.day, hour = EXCLUDED.hour, minute = EXCLUDED.minute, updated_at = now()
RETURNING campaign_id, day, hour, minute, updated_at;

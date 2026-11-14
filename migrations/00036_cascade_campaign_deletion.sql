-- +goose Up
-- A campaign is the aggregate root for its persisted world. Deleting it must
-- remove every owned row while preserving restrictive behavior for standalone
-- deletes inside a campaign.

CREATE INDEX IF NOT EXISTS idx_location_connections_campaign_id
  ON location_connections(campaign_id);
CREATE INDEX IF NOT EXISTS idx_location_connections_to_location_campaign
  ON location_connections(to_location_id, campaign_id);
CREATE INDEX IF NOT EXISTS idx_world_facts_superseded_by
  ON world_facts(superseded_by);
CREATE INDEX IF NOT EXISTS idx_character_feats_feat_id
  ON character_feats(feat_id);
CREATE INDEX IF NOT EXISTS idx_character_skills_skill_id
  ON character_skills(skill_id);

ALTER TABLE locations
  DROP CONSTRAINT locations_campaign_id_fkey,
  ADD CONSTRAINT locations_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE factions
  DROP CONSTRAINT factions_campaign_id_fkey,
  ADD CONSTRAINT factions_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE location_connections
  DROP CONSTRAINT location_connections_campaign_fk,
  ADD CONSTRAINT location_connections_campaign_fk
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE npcs
  DROP CONSTRAINT npcs_campaign_id_fkey,
  ADD CONSTRAINT npcs_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE player_characters
  DROP CONSTRAINT player_characters_campaign_id_fkey,
  ADD CONSTRAINT player_characters_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE quests
  DROP CONSTRAINT quests_campaign_id_fkey,
  ADD CONSTRAINT quests_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE items
  DROP CONSTRAINT items_campaign_fk,
  ADD CONSTRAINT items_campaign_fk
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE entity_relationships
  DROP CONSTRAINT entity_relationships_campaign_id_fkey,
  ADD CONSTRAINT entity_relationships_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE memories
  DROP CONSTRAINT memories_campaign_fk,
  ADD CONSTRAINT memories_campaign_fk
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE world_facts
  DROP CONSTRAINT world_facts_campaign_id_fkey,
  ADD CONSTRAINT world_facts_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE session_logs
  DROP CONSTRAINT session_logs_campaign_id_fkey,
  ADD CONSTRAINT session_logs_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE languages
  DROP CONSTRAINT languages_campaign_id_fkey,
  ADD CONSTRAINT languages_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE belief_systems
  DROP CONSTRAINT belief_systems_campaign_id_fkey,
  ADD CONSTRAINT belief_systems_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE economic_systems
  DROP CONSTRAINT economic_systems_campaign_id_fkey,
  ADD CONSTRAINT economic_systems_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
ALTER TABLE cultures
  DROP CONSTRAINT cultures_campaign_id_fkey,
  ADD CONSTRAINT cultures_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;

-- These relationships remain restrictive for direct child deletion, but are
-- deferred so rows on both sides can disappear during one campaign cascade.
ALTER TABLE location_connections
  DROP CONSTRAINT location_connections_from_location_fk,
  ADD CONSTRAINT location_connections_from_location_fk
    FOREIGN KEY (from_location_id, campaign_id)
    REFERENCES locations(id, campaign_id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
  DROP CONSTRAINT location_connections_to_location_fk,
  ADD CONSTRAINT location_connections_to_location_fk
    FOREIGN KEY (to_location_id, campaign_id)
    REFERENCES locations(id, campaign_id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE npcs
  DROP CONSTRAINT npcs_location_id_fkey,
  ADD CONSTRAINT npcs_location_id_fkey
    FOREIGN KEY (location_id) REFERENCES locations(id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
  DROP CONSTRAINT npcs_faction_id_fkey,
  ADD CONSTRAINT npcs_faction_id_fkey
    FOREIGN KEY (faction_id) REFERENCES factions(id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE player_characters
  DROP CONSTRAINT player_characters_current_location_id_fkey,
  ADD CONSTRAINT player_characters_current_location_id_fkey
    FOREIGN KEY (current_location_id) REFERENCES locations(id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE items
  DROP CONSTRAINT items_player_character_campaign_fk,
  ADD CONSTRAINT items_player_character_campaign_fk
    FOREIGN KEY (player_character_id, campaign_id)
    REFERENCES player_characters(id, campaign_id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE session_logs
  DROP CONSTRAINT session_logs_location_campaign_fk,
  ADD CONSTRAINT session_logs_location_campaign_fk
    FOREIGN KEY (location_id, campaign_id)
    REFERENCES locations(id, campaign_id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE cultures
  DROP CONSTRAINT cultures_language_campaign_fk,
  ADD CONSTRAINT cultures_language_campaign_fk
    FOREIGN KEY (language_id, campaign_id)
    REFERENCES languages(id, campaign_id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
  DROP CONSTRAINT cultures_belief_system_campaign_fk,
  ADD CONSTRAINT cultures_belief_system_campaign_fk
    FOREIGN KEY (belief_system_id, campaign_id)
    REFERENCES belief_systems(id, campaign_id)
    ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;

-- +goose Down
ALTER TABLE cultures
  DROP CONSTRAINT cultures_belief_system_campaign_fk,
  ADD CONSTRAINT cultures_belief_system_campaign_fk
    FOREIGN KEY (belief_system_id, campaign_id)
    REFERENCES belief_systems(id, campaign_id) ON DELETE RESTRICT,
  DROP CONSTRAINT cultures_language_campaign_fk,
  ADD CONSTRAINT cultures_language_campaign_fk
    FOREIGN KEY (language_id, campaign_id)
    REFERENCES languages(id, campaign_id) ON DELETE RESTRICT;
ALTER TABLE session_logs
  DROP CONSTRAINT session_logs_location_campaign_fk,
  ADD CONSTRAINT session_logs_location_campaign_fk
    FOREIGN KEY (location_id, campaign_id)
    REFERENCES locations(id, campaign_id) ON DELETE RESTRICT;
ALTER TABLE items
  DROP CONSTRAINT items_player_character_campaign_fk,
  ADD CONSTRAINT items_player_character_campaign_fk
    FOREIGN KEY (player_character_id, campaign_id)
    REFERENCES player_characters(id, campaign_id) ON DELETE RESTRICT;
ALTER TABLE player_characters
  DROP CONSTRAINT player_characters_current_location_id_fkey,
  ADD CONSTRAINT player_characters_current_location_id_fkey
    FOREIGN KEY (current_location_id) REFERENCES locations(id) ON DELETE RESTRICT;
ALTER TABLE npcs
  DROP CONSTRAINT npcs_faction_id_fkey,
  ADD CONSTRAINT npcs_faction_id_fkey
    FOREIGN KEY (faction_id) REFERENCES factions(id) ON DELETE RESTRICT,
  DROP CONSTRAINT npcs_location_id_fkey,
  ADD CONSTRAINT npcs_location_id_fkey
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE RESTRICT;
ALTER TABLE location_connections
  DROP CONSTRAINT location_connections_to_location_fk,
  ADD CONSTRAINT location_connections_to_location_fk
    FOREIGN KEY (to_location_id, campaign_id)
    REFERENCES locations(id, campaign_id) ON DELETE RESTRICT,
  DROP CONSTRAINT location_connections_from_location_fk,
  ADD CONSTRAINT location_connections_from_location_fk
    FOREIGN KEY (from_location_id, campaign_id)
    REFERENCES locations(id, campaign_id) ON DELETE RESTRICT;

ALTER TABLE cultures
  DROP CONSTRAINT cultures_campaign_id_fkey,
  ADD CONSTRAINT cultures_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE economic_systems
  DROP CONSTRAINT economic_systems_campaign_id_fkey,
  ADD CONSTRAINT economic_systems_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE belief_systems
  DROP CONSTRAINT belief_systems_campaign_id_fkey,
  ADD CONSTRAINT belief_systems_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE languages
  DROP CONSTRAINT languages_campaign_id_fkey,
  ADD CONSTRAINT languages_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE session_logs
  DROP CONSTRAINT session_logs_campaign_id_fkey,
  ADD CONSTRAINT session_logs_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE world_facts
  DROP CONSTRAINT world_facts_campaign_id_fkey,
  ADD CONSTRAINT world_facts_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE memories
  DROP CONSTRAINT memories_campaign_fk,
  ADD CONSTRAINT memories_campaign_fk
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE entity_relationships
  DROP CONSTRAINT entity_relationships_campaign_id_fkey,
  ADD CONSTRAINT entity_relationships_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE items
  DROP CONSTRAINT items_campaign_fk,
  ADD CONSTRAINT items_campaign_fk
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE quests
  DROP CONSTRAINT quests_campaign_id_fkey,
  ADD CONSTRAINT quests_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE player_characters
  DROP CONSTRAINT player_characters_campaign_id_fkey,
  ADD CONSTRAINT player_characters_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE npcs
  DROP CONSTRAINT npcs_campaign_id_fkey,
  ADD CONSTRAINT npcs_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE location_connections
  DROP CONSTRAINT location_connections_campaign_fk,
  ADD CONSTRAINT location_connections_campaign_fk
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE factions
  DROP CONSTRAINT factions_campaign_id_fkey,
  ADD CONSTRAINT factions_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;
ALTER TABLE locations
  DROP CONSTRAINT locations_campaign_id_fkey,
  ADD CONSTRAINT locations_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE RESTRICT;

DROP INDEX IF EXISTS idx_location_connections_to_location_campaign;
DROP INDEX IF EXISTS idx_location_connections_campaign_id;
DROP INDEX IF EXISTS idx_world_facts_superseded_by;
DROP INDEX IF EXISTS idx_character_feats_feat_id;
DROP INDEX IF EXISTS idx_character_skills_skill_id;

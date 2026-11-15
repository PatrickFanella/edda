export interface CampaignCreateRequest {
  rules_version?: string;
  name: string;
  description: string;
  genre: string;
  tone: string;
  themes: string[];
  rules_mode?: string;
}

export interface FeatResponse {
  id: string;
  feat_id: string;
  name: string;
  description: string;
  bonus_type: string;
  bonus_value: number;
  prerequisites?: string;
}

export interface SkillResponse {
  id: string;
  skill_id: string;
  name: string;
  description: string;
  base_ability: string;
  points: number;
}

export interface CampaignResponse {
  rules_version?: string;
  id: string;
  name: string;
  description: string;
  genre: string;
  tone: string;
  themes: string[];
  status: string;
  rules_mode: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CampaignListResponse {
  campaigns: CampaignResponse[];
}

export interface CharacterAbility {
  name: string;
  description?: string;
}

export interface DerivedValue { value: number; derivation: { label: string; value: number; source_id: string }[] }
export interface CheckReceipt {
  rules_version: string;
  kind: 'ability-check' | 'saving-throw';
  ability: string;
  skill?: string;
  mode: 'ordinary' | 'advantage' | 'disadvantage';
  dice: { sides: number; value: number; kept: boolean }[];
  modifier: DerivedValue;
  total: number;
  dc?: number;
  outcome: 'unknown' | 'success' | 'failure';
  source_ids: string[];
}
export interface CharacterSheetProjection {
  rules_version: string;
  scope: string;
  abilities: Record<string, { score: number; modifier: DerivedValue }>;
  proficiency: DerivedValue;
  saves: Record<string, DerivedValue>;
  skills: Record<string, DerivedValue>;
  hp: DerivedValue;
  max_hp: DerivedValue;
  defenses: Record<string, DerivedValue>;
  resources: { id: string; label: string; current: number; maximum: number; source_id: string }[];
  capabilities: { id: string; label: string; source_id: string; playable: boolean }[];
  progression: { current_xp: number; level_xp: number; next_level_xp: number; source_id: string };
}
export interface CharacterResponse {
  rules_version?: string;
  sheet?: CharacterSheetProjection;
  sheet_unavailable?: string;
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

export interface LocationConnectionResponse {
  from_location_id?: string;
  to_location_id: string;
  description: string;
  bidirectional: boolean;
  travel_time: string;
}

export interface LocationResponse {
  id: string;
  campaign_id: string;
  name: string;
  description: string;
  region: string;
  location_type: string;
  properties: Record<string, unknown>;
  connections: LocationConnectionResponse[];
}

export interface NPCResponse {
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

export interface QuestObjectiveResponse {
  id: string;
  description: string;
  completed: boolean;
  order_index: number;
}

export interface QuestResponse {
  id: string;
  campaign_id: string;
  parent_quest_id?: string;
  title: string;
  description: string;
  quest_type: string;
  status: string;
  objectives: QuestObjectiveResponse[];
}

export interface ItemResponse {
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

export interface SessionLogEntry {
  turn_number: number;
  player_input: string;
  input_type: string;
  llm_response: string;
  choices?: string[];
  resolution_events?: ResolutionEvent[];
  created_at: string;
}

export interface SessionHistoryResponse {
  entries: SessionLogEntry[];
}

export interface ActionRequest {
  action_id?: string;
  input: string;
}

export interface StateChange {
  entity_type: string;
  entity_id: string;
  change_type: string;
  details: Record<string, unknown>;
}

export interface ResolutionEvent {
  type: string;
  label: string;
  outcome: string;
  details: Record<string, unknown>;
}

export interface TurnResult {
  action_id?: string;
  recovered?: boolean;
  turn_number?: number;
  narrative: string;
  state_changes: StateChange[];
  resolution_events: ResolutionEvent[];
  combat_active: boolean;
}

export type TurnResponse = TurnResult;

export interface WebSocketMessageEnvelope<TPayload = unknown> {
  type: string;
  payload: TPayload;
  timestamp: string;
}

export interface WebSocketChunkPayload {
  text: string;
}

export interface WebSocketErrorPayload {
  action_id?: string;
  error: string;
}

export interface WebSocketStatusPayload {
  stage: string;
  tool?: string;
  description: string;
}

// Journal types

export interface SessionSummaryResponse {
  id: string;
  campaign_id: string;
  from_turn: number;
  to_turn: number;
  summary: string;
  created_at: string;
}

export interface JournalEntryResponse {
  id: string;
  campaign_id: string;
  title: string;
  content: string;
  created_at: string;
  updated_at: string;
}


export interface CampaignProfile {
  genre: string;
  tone: string;
  themes: string[];
  world_type: string;
  danger_level: string;
  political_complexity: string;
}

export interface CharacterProfile {
  name: string;
  concept: string;
  background: string;
  personality: string;
  motivations: string[];
  strengths: string[];
  weaknesses: string[];
}

export interface CampaignProposal {
  name: string;
  summary: string;
  profile: CampaignProfile;
}

export type StartupInterviewResponse<TProfile> =
  | {
      session_id: string;
      message: string;
      done: false;
      profile?: undefined;
    }
  | {
      session_id: string;
      message: string;
      done: true;
      profile: TProfile;
    };

export type CampaignInterviewResponse = StartupInterviewResponse<CampaignProfile>;

export interface CampaignInterviewStepRequest {
  session_id: string;
  input: string;
}

export interface GenerateCampaignProposalsRequest {
  genre: string;
  setting_style: string;
  tone: string;
}

export interface GenerateCampaignProposalsResponse {
  proposals: CampaignProposal[];
}

export interface GenerateCampaignNameRequest {
  campaign_profile: CampaignProfile;
}

export interface GenerateCampaignNameResponse {
  name: string;
}

export type CharacterInterviewResponse = StartupInterviewResponse<CharacterProfile>;

export interface CharacterInterviewStartRequest {
  campaign_profile: CampaignProfile;
}

export interface CharacterInterviewStepRequest {
  session_id: string;
  input: string;
}

export interface OpeningSceneResponse {
  narrative: string;
  choices: string[];
}

export interface StarterItem {
  name: string;
  description?: string;
  item_type?: string;
  rarity?: string;
  properties?: Record<string, unknown>;
  equipped?: boolean;
  quantity?: number;
}

export interface StarterKnownFact {
  fact: string;
  category?: string;
}

export interface StarterRelationship {
  target_entity_type: string;
  target_entity_id: string;
  relationship_type: string;
  description?: string;
  strength?: number;
}

export interface CharacterSpawnPackage {
  items?: StarterItem[];
  known_facts?: StarterKnownFact[];
  relationships?: StarterRelationship[];
}

export interface BuildWorldRequest {
  name: string;
  summary: string;
  profile: CampaignProfile;
  character_profile: CharacterProfile;
  rules_mode?: string;
  spawn_package?: CharacterSpawnPackage;
}

export interface BuildWorldResponse {
  campaign: CampaignResponse;
  opening_scene: OpeningSceneResponse;
}

// NPC Panel types
export interface EncounteredNPCResponse {
  id: string;
  campaign_id: string;
  name: string;
  description: string;
  disposition?: number;
  alive: boolean;
  faction_id?: string;
}

export interface DialogueEntry {
  turn_number: number;
  player_input: string;
  llm_response: string;
  created_at: string;
}

// Facts/Lore types
export interface FactResponse {
  id: string;
  campaign_id: string;
  fact: string;
  category: string;
  source: string;
  superseded_by?: string;
  player_known: boolean;
  created_at: string;
}

// Codex types
export interface LanguageResponse {
  id: string;
  campaign_id: string;
  name: string;
  description: string;
  player_known: boolean;
  created_at: string;
}

export interface CultureResponse {
  id: string;
  campaign_id: string;
  name: string;
  language_id?: string;
  belief_system_id?: string;
  player_known: boolean;
  created_at: string;
}

export interface BeliefSystemResponse {
  id: string;
  campaign_id: string;
  name: string;
  player_known: boolean;
  created_at: string;
}

export interface EconomicSystemResponse {
  id: string;
  campaign_id: string;
  name: string;
  player_known: boolean;
  created_at: string;
}

// Relationships types
export interface RelationshipResponse {
  id: string;
  campaign_id: string;
  source_entity_type: string;
  source_entity_id: string;
  target_entity_type: string;
  target_entity_id: string;
  relationship_type: string;
  description: string;
  strength?: number;
  player_aware: boolean;
  created_at: string;
}

export interface QuestNoteResponse {
  id: string;
  quest_id: string;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface QuestHistoryEntry {
  id: string;
  quest_id: string;
  snapshot: string;
  created_at: string;
}

// Map types
export interface MapLocationResponse {
  id: string;
  campaign_id: string;
  name: string;
  description: string;
  region: string;
  location_type: string;
  player_visited: boolean;
  player_known: boolean;
}

export interface MapDataResponse {
  locations: MapLocationResponse[];
  connections: LocationConnectionResponse[];
}

export interface SavePointResponse {
  id: string;
  campaign_id: string;
  name: string;
  turn_number: number;
  is_auto: boolean;
  created_at: string;
}

export interface CampaignTimeResponse {
  day: number;
  hour: number;
  minute: number;
}

export interface ActionIntent {
  actor_id: string;
  kind: string;
  ability?: string;
  skill?: string;
  entry_id?: string;
}
export interface ActionPreview {
  intent: ActionIntent;
  rules_version: string;
  expected_revision: string;
  available: boolean;
  reason?: string;
  modifier?: DerivedValue;
  targets: { id: string; label: string }[];
  options: string[];
  costs: { resource_id: string; amount: number }[];
  cost_status: 'requires-adjudication';
  source_ids: string[];
}

export interface ChapterTurn {
  log_id: string;
  attempt_id?: string;
  turn_number: number;
  player_input: string;
  narrative: string;
}
export interface ChapterRevisionResponse {
  id: string;
  chapter_id: string;
  request_id: string;
  revision: number;
  status: 'generating' | 'completed' | 'failed';
  content: string;
  error: string;
  prompt_version: string;
  created_at: string;
}
export interface ChapterResponse {
  id: string;
  campaign_id: string;
  from_turn: number;
  to_turn: number;
  source_hash: string;
  source_turns: ChapterTurn[];
  knowledge_boundary: string;
  rules_version: string;
  rules_mode: string;
  created_at: string;
  revisions: ChapterRevisionResponse[];
}

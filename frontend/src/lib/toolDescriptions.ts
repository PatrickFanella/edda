const TOOL_DESCRIPTIONS: Record<string, string> = {
  skill_check: 'Rolling a skill check',
  roll_dice: 'Rolling dice',
  move_player: 'Moving to a new location',
  describe_scene: 'Describing the scene',
  present_choices: 'Presenting options',
  npc_dialogue: 'NPC is speaking',
  establish_fact: 'Recording world lore',
  create_npc: 'Introducing a new character',
  create_location: 'Discovering a new place',
  create_quest: 'A new quest emerges',
  update_quest: 'Updating quest progress',
  initiate_combat: 'Combat begins!',
  resolve_combat: 'Combat resolving',
  add_item: 'You received an item',
  remove_item: 'An item was lost',
  add_experience: 'Gaining experience',
  level_up: 'Level up!',
  search_memory: 'Recalling past events',
};

export function describeToolCall(toolName: string): string {
  return TOOL_DESCRIPTIONS[toolName] ?? toolName.replace(/_/g, ' ');
}

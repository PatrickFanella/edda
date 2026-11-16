import type { Encounter } from '../../api/types';

/** Public, server-derived initiative and the active actor's remaining budget. */
export function EncounterSummary({ encounter }: { readonly encounter: Encounter }) {
 const actor = encounter.actors[encounter.turn];
 return <section aria-label="Encounter turns" className="shrink-0 border border-ruby/40 bg-obsidian/75 p-2 text-xs">
   <p className="font-heading text-gold">Round {encounter.round} · {encounter.status === 'active' ? `${actor.name}'s turn` : 'Encounter complete'}</p>
   <ol aria-label="Initiative order" className="flex flex-wrap gap-x-4 gap-y-1">
    {encounter.actors.map(a => <li key={a.id} aria-current={encounter.status === 'active' && a.id === actor.id ? 'step' : undefined}>{a.name}: initiative {a.initiative} (d20 {a.initiative_die})</li>)}
   </ol>
   {encounter.space && <p>Distance: {encounter.space.distance_feet === null ? 'needs clarification' : `${encounter.space.distance_feet} ft`} · Sight: {encounter.space.line_of_sight === null ? 'needs clarification' : encounter.space.line_of_sight ? 'clear' : 'blocked'} · Cover: {encounter.space.cover || 'needs clarification'}</p>}
   {encounter.reaction && <p>Reaction: {encounter.reaction.available ? 'available' : 'spent'}{encounter.reaction.pending ? ' · awaiting your choice before the foe leaves reach' : ''}</p>}
   {encounter.budget.surge_used && <p>Action Surge: {encounter.budget.surge_action ? 'one Action available (except Magic)' : 'extra Action spent'}</p>}
   {encounter.status === 'active' && <p>Action: {encounter.budget.action ? 'available' : 'spent'} · Bonus action: {encounter.budget.bonus_action ? 'available' : 'spent'} · Movement: {encounter.budget.movement} ft</p>}
 </section>;
}

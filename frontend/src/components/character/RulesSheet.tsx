import type { CharacterSheetProjection, CheckReceipt as Receipt, DerivedValue } from '../../api/types';
import { formatCampaignSeconds } from '../../lib/campaignTime';
import { ActionPreviewPanel } from './ActionPreviewPanel';
import { CheckReceipt } from './CheckReceipt';

export function RulesSheet({ sheet, receipts = [], campaignId, actorId }: { readonly campaignId?: string; readonly actorId?: string; readonly sheet: CharacterSheetProjection; readonly receipts?: readonly Receipt[] }) {
  return <section aria-label="SRD character sheet" className="space-y-5 text-champagne">
    <p className="text-sm text-pewter">SRD 5.2.1 · Level {sheet.scope.startsWith('fighter-2-') ? 2 : 1} {sheet.spellcasting ? 'Cleric fixture' : 'Fighter'} · Sheet values and previews</p>
    <p>HP {sheet.hp.value} / {sheet.max_hp.value} · XP {sheet.progression.current_xp} / {sheet.progression.next_level_xp}</p>
    <p className="text-sm text-pewter">Inspect a value to see its calculation and source identifiers.</p>
    <dl className="grid gap-3 sm:grid-cols-2">
      <RuleValue label="Proficiency" value={sheet.proficiency} />
      <RuleValue label="Maximum HP" value={sheet.max_hp} />
      {Object.entries(sheet.defenses).map(([name,value]) => <RuleValue key={name} label={name} value={value} />)}
    </dl>
    <h3 className="font-heading text-gold">Abilities and saves</h3>
    <dl className="grid gap-3 sm:grid-cols-2">{Object.entries(sheet.abilities).map(([name,ability]) => <div key={name}>
      <p className="capitalize">{name}: {ability.score}</p>
      <RuleValue label={`${name} modifier`} value={ability.modifier} />
      <RuleValue label={`${name} save`} value={sheet.saves[name]} />
    </div>)}</dl>
    {sheet.spellcasting && <section aria-label="Spellcasting" className="space-y-2">
      <h3 className="font-heading text-gold">Spellcasting</h3>
      <p>Casting ability: {sheet.spellcasting.ability} ({sheet.spellcasting.modifier.value >= 0 ? '+' : ''}{sheet.spellcasting.modifier.value})</p>
      {sheet.spellcasting.spells.map(spell => <p key={spell.id}>{spell.label} · {spell.known ? 'Known' : 'Unknown'} · {spell.prepared ? 'Prepared' : 'Not prepared'} · Components: {spell.components.join(', ')} ({spell.components_available ? 'available' : 'unavailable'}) · {spell.playable === false ? 'Effect unavailable' : `Supported slots: ${spell.slot_levels.join(', ')}`}</p>)}
      {sheet.spellcasting.choices && <p>Divine Order: {sheet.spellcasting.choices.divine_order} · Magic Initiate ({sheet.spellcasting.choices.magic_initiate_ability}): {sheet.spellcasting.choices.magic_initiate_spell}</p>}
      <p className="text-sm text-pewter">Human Acolyte Cleric creation is supported. Recorded spell choices are listed here; Cure Wounds and Shield of Faith on self are playable. Other effects and targets remain unavailable.</p>
    </section>}
    {sheet.spellcasting && <section aria-label="Concentration and recovery">
      <h3 className="font-heading text-gold">Concentration and recovery</h3>
      <p>{sheet.concentration ? `Shield of Faith: +2 AC. Started ${formatCampaignSeconds(sheet.concentration.started_at)}; expires ${formatCampaignSeconds(sheet.concentration.expires_at)}.` : 'No concentration effect active.'}</p>
      <p className="text-sm text-pewter">Confirm quiet downtime or an uninterrupted rest from the action menu. One hour permits spending a d8 Hit Die; only an eight-hour Long Rest restores Cleric slots. Rest interruptions and wider effects remain unavailable.</p>
    </section>}
    <h3 className="font-heading text-gold">Skills</h3>
    <dl className="grid gap-3 sm:grid-cols-2">{Object.entries(sheet.skills).map(([name,value]) => <RuleValue key={name} label={name} value={value} />)}</dl>
    {campaignId && actorId && <ActionPreviewPanel key={`${campaignId}:${actorId}`} campaignId={campaignId} actorId={actorId} sheet={sheet} />}
    <p className="text-sm text-pewter">Fighter advancement supports level 1 to 2 at 300 earned XP. Choose fixed or rolled HP and retain Defense and the recorded mastery choices. Later levels and new feature effects remain unavailable.</p>
    <h3 className="font-heading text-gold">Resources</h3>
    {sheet.resources.map(resource => <p key={resource.id}>{resource.label}: {resource.current} / {resource.maximum}<small className="block break-all text-pewter">{resource.source_id}</small></p>)}
    <h3 className="font-heading text-gold">Capabilities</h3>
    {sheet.capabilities.map(capability => <p key={capability.id}>{capability.label} (read only)<small className="block break-all text-pewter">{capability.source_id}</small></p>)}
    {receipts.length > 0 && <section aria-label="Check receipts" className="space-y-3">
      <h3 className="font-heading text-gold">Check receipts</h3>
      {receipts.map((receipt, index) => <CheckReceipt key={index} receipt={receipt} />)}
    </section>}
    <a className="text-gold underline" href="https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf">SRD 5.2.1 reference</a>
  </section>;
}
function RuleValue({ label, value }: { readonly label: string; readonly value: DerivedValue }) {
  return <div><dt className="capitalize text-pewter">{label.replaceAll('-', ' ')}</dt><dd><details>
    <summary className="min-h-11 cursor-pointer py-2 text-jade">{value.value >= 0 ? '+' : ''}{value.value}</summary>
    <ul className="space-y-2 text-sm">{value.derivation.map((term,index) => <li key={index}>{term.label}: {term.value}<small className="block break-all text-pewter">{term.source_id}</small></li>)}</ul>
  </details></dd></div>;
}

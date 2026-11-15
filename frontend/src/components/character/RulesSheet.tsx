import type { CharacterSheetProjection, DerivedValue } from '../../api/types';

export function RulesSheet({ sheet }: { readonly sheet: CharacterSheetProjection }) {
  return <section aria-label="SRD character sheet" className="space-y-5 text-champagne">
    <p className="text-sm text-pewter">SRD 5.2.1 · Level 1 Fighter · Read-only projection</p>
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
    <h3 className="font-heading text-gold">Skills</h3>
    <dl className="grid gap-3 sm:grid-cols-2">{Object.entries(sheet.skills).map(([name,value]) => <RuleValue key={name} label={name} value={value} />)}</dl>
    <h3 className="font-heading text-gold">Resources</h3>
    {sheet.resources.map(resource => <p key={resource.id}>{resource.label}: {resource.current} / {resource.maximum}<small className="block break-all text-pewter">{resource.source_id}</small></p>)}
    <h3 className="font-heading text-gold">Capabilities</h3>
    {sheet.capabilities.map(capability => <p key={capability.id}>{capability.label} (read only)<small className="block break-all text-pewter">{capability.source_id}</small></p>)}
    <a className="text-gold underline" href="https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf">SRD 5.2.1 reference</a>
  </section>;
}
function RuleValue({ label, value }: { readonly label: string; readonly value: DerivedValue }) {
  return <div><dt className="capitalize text-pewter">{label.replaceAll('-', ' ')}</dt><dd><details>
    <summary className="min-h-11 cursor-pointer py-2 text-jade">{value.value >= 0 ? '+' : ''}{value.value}</summary>
    <ul className="space-y-2 text-sm">{value.derivation.map((term,index) => <li key={index}>{term.label}: {term.value}<small className="block break-all text-pewter">{term.source_id}</small></li>)}</ul>
  </details></dd></div>;
}

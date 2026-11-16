import type { AdvancementReceipt as Receipt } from '../../api/types';

export function AdvancementReceipt({ receipt }: { readonly receipt: Receipt }) {
  return <section aria-label="Saved advancement receipt" className="space-y-2 border border-jade/20 p-3">
    <p>Fighter level {receipt.level_before} → {receipt.level_after}. {receipt.choice === 'fixed' ? 'Fixed HP' : 'Rolled HP'}: d10 value {receipt.die}, gain {receipt.hp_gain}. Maximum HP {receipt.max_hp_before} → {receipt.max_hp_after}.</p>
    <p>Current HP unchanged. Defense and mastery choices retained. Level-two Action Surge supports one extra Action except Magic on your owned active ordinary encounter turn. Tactical Mind supports a saved failed check from the selected Edda Easy Strength (Athletics) DC 10 fixture.</p>
    <details><summary className="min-h-11 cursor-pointer py-2">Inspect advancement sources</summary>{receipt.source_ids.map(id => <p key={id} className="break-all text-pewter">{id}</p>)}</details>
  </section>;
}

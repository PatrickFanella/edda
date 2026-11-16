import type { EffectReceipt as Receipt } from '../../api/types';
import { formatCampaignSeconds } from '../../lib/campaignTime';
import { CheckReceipt } from './CheckReceipt';

export function EffectReceipt({ receipt }: { readonly receipt: Receipt }) {
  return <section aria-label="Saved caster effect receipt" className="space-y-2 border border-jade/20 p-3">
    <p>{receipt.reason}</p>
    {receipt.action_surge_after !== undefined && <p>Action Surge uses: {receipt.action_surge_before ?? 0} → {receipt.action_surge_after}.</p>}
    <p>Time: {formatCampaignSeconds(receipt.time_before)} → {formatCampaignSeconds(receipt.time_after)}. HP: {receipt.hp_before} → {receipt.hp_after}. Slots: {receipt.slots_before} → {receipt.slots_after}. Hit Dice: {receipt.hit_dice_before} → {receipt.hit_dice_after}.</p>
    {receipt.before && <p>Previous effect: {receipt.before.spell} (expires at {formatCampaignSeconds(receipt.before.expires_at)}).</p>}
    <p>{receipt.after ? `Concentration: ${receipt.after.spell} until ${formatCampaignSeconds(receipt.after.expires_at)}.` : 'No concentration effect remains.'}</p>
    {receipt.healing_die !== undefined && <p>Healing d8: {receipt.healing_die}.</p>}
    {receipt.check && <CheckReceipt receipt={receipt.check} />}
    <details><summary className="min-h-11 cursor-pointer py-2">Inspect effect sources</summary>{receipt.before && <p>Previous effect ID: {receipt.before.id}</p>}{receipt.after && <p>Current effect ID: {receipt.after.id}</p>}{receipt.source_ids.map(id => <p key={id} className="break-all text-pewter">{id}</p>)}</details>
  </section>;
}

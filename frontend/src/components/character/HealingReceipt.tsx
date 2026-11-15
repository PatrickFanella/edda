import type { HealingReceipt as Receipt } from '../../api/types';

export function HealingReceipt({ receipt }: { readonly receipt: Receipt }) {
 return <details className="border border-jade/20 p-3"><summary className="min-h-11 cursor-pointer py-2 text-jade">{receipt.item}: {receipt.target} regains {receipt.healing} HP</summary>
 <p>Healing: {receipt.dice.map(value => `d4 ${value}`).join(' + ')} + {receipt.bonus} = {receipt.rolled_healing}</p>
 <p>HP {receipt.hp_before} → {receipt.hp_after} (capped at maximum HP)</p><p>Quantity {receipt.quantity_before} → {receipt.quantity_after}</p>
 <p>{receipt.rules_version}</p><ul>{receipt.source_ids.map(source => <li className="break-all" key={source}>{source}</li>)}</ul></details>;
}

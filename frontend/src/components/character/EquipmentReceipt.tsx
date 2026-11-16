import type { EquipmentReceipt as Receipt } from '../../api/types';

export function EquipmentReceipt({ receipt }: { readonly receipt: Receipt }) {
  return <section aria-label="Saved equipment receipt" className="space-y-2 border border-gold/20 p-3 text-sm">
    <p>{receipt.equipment}: {receipt.equipped_after ? 'Equipped' : 'Removed'}. AC {receipt.ac_before} to {receipt.ac_after}.</p>
    <p>Campaign time: {receipt.time_before} to {receipt.time_after} seconds (+{receipt.time_after - receipt.time_before}).</p>
    {receipt.expired && <p>Shield of Faith expired.</p>}
    <details><summary>Equipment sources</summary>{receipt.source_ids.map(source => <p key={source} className="break-all">{source}</p>)}</details>
  </section>;
}

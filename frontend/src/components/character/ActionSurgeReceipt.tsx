import type { ActionSurgeReceipt as Receipt } from '../../api/types';
export function ActionSurgeReceipt({ receipt }: { readonly receipt: Receipt }) {
 return <section aria-label="Action Surge receipt"><p>Action Surge: {receipt.uses_before} → {receipt.uses_after} uses. {receipt.additional_actions} additional Action this turn; excludes {receipt.excluded_action}.</p><p>Bonus Action, Reaction and movement unchanged.</p><details><summary>Action Surge sources</summary>{receipt.source_ids.map(id => <p className="break-all" key={id}>{id}</p>)}</details></section>;
}

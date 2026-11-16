import { CheckReceipt } from './CheckReceipt';
import type { CheckReceipt as BaseReceipt } from '../../api/types';
export interface TacticalMindReceipt {
  base: BaseReceipt; die: number; total: number; outcome: string;
  second_wind_before: number; second_wind_after: number; source_ids: string[];
}
export function TacticalReceipt({ receipt }: { readonly receipt: TacticalMindReceipt }) {
  return <section aria-label="Saved Tactical Mind receipt">
    <p>Tactical Mind: {receipt.outcome} · d10 {receipt.die} · total {receipt.total}</p>
    <p>Second Wind: {receipt.second_wind_before} → {receipt.second_wind_after}. No HP restored.</p>
    <CheckReceipt receipt={receipt.base} />
    {receipt.source_ids.map(source => <p className="break-all" key={source}>{source}</p>)}
  </section>;
}

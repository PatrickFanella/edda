import type { CheckReceipt as Receipt } from '../../api/types';

// All values, the kept die and the disclosed outcome come from the server.
export function CheckReceipt({ receipt }: { readonly receipt: Receipt }) {
  const label = `${receipt.skill ?? receipt.ability} ${receipt.kind === 'saving-throw' ? 'saving throw' : 'check'}`;
  const outcome = receipt.outcome === 'unknown' ? 'Outcome not revealed' : receipt.outcome;
  return <details className="border border-jade/20 p-3">
    <summary className="min-h-11 cursor-pointer py-2 capitalize text-jade">{label}: {receipt.total} · {outcome}</summary>
    <p className="text-sm capitalize">{receipt.mode} · {receipt.rules_version}</p>
    <ul aria-label="Individual dice" className="my-3 space-y-2 text-sm">
      {receipt.dice.map((die, index) => <li key={index}>d{die.sides}: {die.value} ({die.kept ? 'kept' : 'discarded'})</li>)}
    </ul>
    <p>Modifier: {receipt.modifier.value}</p>
    <ul aria-label="Modifier sources" className="my-3 space-y-2 text-sm">
      {receipt.modifier.derivation.map((term, index) => <li key={index}>{term.label}: {term.value}<small className="block break-all text-pewter">{term.source_id}</small></li>)}
    </ul>
    <p>Total: {receipt.total}</p>
    <p>{receipt.dc === undefined ? 'DC undisclosed' : `DC: ${receipt.dc}`}</p>
    <p className="capitalize">{outcome}</p>
    <ul aria-label="Test sources" className="mt-3 space-y-2 text-sm text-pewter">
      {receipt.source_ids.map(id => <li key={id} className="break-all">{id}</li>)}
    </ul>
  </details>;
}

import type { AttackReceipt as Receipt } from '../../api/types';

export function AttackReceipt({ receipt }: { readonly receipt: Receipt }) {
 return <details className="border border-jade/20 p-3"><summary className="min-h-11 cursor-pointer py-2 text-jade">{receipt.weapon} against {receipt.target}: {receipt.outcome} · {receipt.damage} {receipt.damage_type}</summary>
 <p>Attack: d20 {receipt.die} + {receipt.modifier.value} = {receipt.total}</p>
 <p>Damage: {receipt.damage_dice.length ? receipt.damage_dice.map(value => `d8 ${value}`).join(' + ') : 'No damage dice'} + {receipt.damage_modifier} = {receipt.damage}</p>
 <ul>{receipt.modifier.derivation.map((term, index) => <li key={index}>{term.label}: {term.value}</li>)}</ul>
 <p>{receipt.rules_version}</p><ul>{receipt.source_ids.map(source => <li className="break-all" key={source}>{source}</li>)}</ul></details>;
}

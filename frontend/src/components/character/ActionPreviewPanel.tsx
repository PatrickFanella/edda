import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { previewCharacterAction } from '../../api/characters';
import type { CharacterSheetProjection } from '../../api/types';

export function ActionPreviewPanel({ campaignId, actorId, sheet }: {
  readonly campaignId: string;
  readonly actorId: string;
  readonly sheet: CharacterSheetProjection;
}) {
  const [selection, setSelection] = useState('');
  const preview = useMutation({ mutationFn: (value: string) => {
    const [kind, name] = value.split(':');
    return previewCharacterAction(campaignId, {
      actor_id: actorId,
      kind: kind === 'skill' ? 'ability-check' : kind,
      ...(kind === 'skill' ? { skill: name } : kind === 'entry' ? { entry_id: name } : { ability: name }),
    });
  } });
  const result = preview.data;
  return <section aria-label="Action preview" className="space-y-3 border-y border-gold/20 py-4">
    <h3 className="font-heading text-gold">Preview a character action</h3>
    <p className="text-sm text-pewter">Inspect an action without rolling dice or changing resources, time or history.</p>
    <label className="block text-sm">Character action
      <select value={selection} disabled={preview.isPending} onChange={event => { setSelection(event.target.value); preview.reset(); }} className="mt-2 block min-h-11 w-full max-w-full bg-midnight p-2 text-champagne">
        <option value="">Choose an action</option>
        <optgroup label="Ability checks">{Object.keys(sheet.abilities).map(name => <option key={name} value={`ability-check:${name}`}>{name} check</option>)}</optgroup>
        <optgroup label="Skills">{Object.keys(sheet.skills).map(name => <option key={name} value={`skill:${name}`}>{name.replaceAll('-', ' ')} check</option>)}</optgroup>
        <optgroup label="Saving throws">{Object.keys(sheet.saves).map(name => <option key={name} value={`saving-throw:${name}`}>{name} save</option>)}</optgroup>
        <optgroup label="Sheet entries">{sheet.capabilities.map(entry => <option key={entry.id} value={`entry:${entry.id}`}>{entry.label}</option>)}</optgroup>
      </select>
    </label>
    <button type="button" className="min-h-11 border border-gold/40 px-4 text-gold disabled:opacity-50" disabled={!selection || preview.isPending} onClick={() => preview.mutate(selection)}>{preview.isPending ? 'Preparing preview…' : 'Preview action'}</button>
    {preview.isError && <p role="alert" className="text-ruby">Could not prepare this action. Try the preview again.</p>}
    {result && <div role="status" className="space-y-2 text-sm">
      {result.available ? <>
        <p>Preview ready. No action has been performed.</p>
        <p>Modifier: {result.modifier?.value}</p>
        <p>Roll options: {result.options.join(', ')}. Circumstances require adjudication.</p>
        <p>Known targets: {result.targets.length ? result.targets.map(target => target.label).join(', ') : 'No target selected. Scene targets require adjudication.'}</p>
        <p>{result.costs.length ? result.costs.map(cost => `${cost.resource_id}: ${cost.amount}`).join(', ') : 'No direct resource cost for this check.'} Action and time costs require adjudication.</p>
        <details><summary className="min-h-11 cursor-pointer py-2">Inspect preview sources</summary>
          <ul>{result.modifier?.derivation.map((term, index) => <li key={index}>{term.label}: {term.value}<small className="block break-all text-pewter">{term.source_id}</small></li>)}</ul>
          {result.source_ids.map(source => <p key={source} className="break-all text-pewter">{source}</p>)}
          <p className="break-all text-pewter">{result.rules_version} · {result.expected_revision}</p>
        </details>
      </> : <p>{result.reason}</p>}
      <p className="text-pewter">Confirmation is not available in this preview.</p>
    </div>}
  </section>;
}

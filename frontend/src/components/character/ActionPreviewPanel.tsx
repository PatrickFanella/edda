import { useEffect, useRef, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { APIError } from '../../api/client';
import { confirmCharacterAction, readActionReceipt, publishConfirmedAction, type PreparedAction } from '../../api/actions';
import { CheckReceipt } from './CheckReceipt';
import type { CheckReceipt as Receipt, TurnResponse } from '../../api/types';
import { previewCharacterAction } from '../../api/characters';
import type { CharacterSheetProjection } from '../../api/types';

export function ActionPreviewPanel({ campaignId, actorId, sheet }: {
  readonly campaignId: string;
  readonly actorId: string;
  readonly sheet: CharacterSheetProjection;
}) {
  const client = useQueryClient();
  const storageKey = `edda:prepared-action:${campaignId}:${actorId}`;
  const [pending, setPending] = useState<PreparedAction | null>(() => {
    try { return JSON.parse(sessionStorage.getItem(storageKey) ?? 'null') as PreparedAction | null; } catch { return null; }
  });
  const requestRef = useRef(pending);
  const [completed, setCompleted] = useState<TurnResponse | null>(null);
  const finish = (result: TurnResponse) => {
    setCompleted(result);
    setPending(null);

    try { sessionStorage.removeItem(storageKey); } catch { /* Storage is optional. */ }
    publishConfirmedAction(campaignId, result, `Confirm ${requestRef.current?.intent.skill ?? requestRef.current?.intent.ability ?? 'character'} check`);
    requestRef.current = null;
    void client.invalidateQueries({ queryKey: ['campaign', campaignId] });
  };
  useEffect(() => {
    if (!pending) return;
    let active = true;
    void readActionReceipt(campaignId, pending.action_id).then(receipt => {
      if (active && receipt?.result) finish(receipt.result);
    }).catch(() => { /* The same action ID remains available for retry. */ });
    return () => { active = false; };
    // Only reconcile the durable request identity, not callback renders.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [campaignId, pending?.action_id]);
  const confirmation = useMutation({
    mutationFn: (request: PreparedAction) => confirmCharacterAction(campaignId, request),
    onSuccess: finish,
    onError: error => {
      if (error instanceof APIError && error.status === 409 && error.message.includes('refresh')) {
        setPending(null); requestRef.current = null;
        try { sessionStorage.removeItem(storageKey); } catch { /* Storage is optional. */ }
        preview.reset();
      }
    },
  });
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
      <select value={selection} disabled={preview.isPending || !!pending} onChange={event => { setSelection(event.target.value); preview.reset(); confirmation.reset(); setCompleted(null); }} className="mt-2 block min-h-11 w-full max-w-full bg-midnight p-2 text-champagne">
        <option value="">Choose an action</option>
        <optgroup label="Ability checks">{Object.keys(sheet.abilities).map(name => <option key={name} value={`ability-check:${name}`}>{name} check</option>)}</optgroup>
        <optgroup label="Skills">{Object.keys(sheet.skills).map(name => <option key={name} value={`skill:${name}`}>{name.replaceAll('-', ' ')} check</option>)}</optgroup>
        <optgroup label="Saving throws">{Object.keys(sheet.saves).map(name => <option key={name} value={`saving-throw:${name}`}>{name} save</option>)}</optgroup>
        <optgroup label="Sheet entries">{sheet.capabilities.map(entry => <option key={entry.id} value={`entry:${entry.id}`}>{entry.label}</option>)}</optgroup>
      </select>
    </label>
    <button type="button" className="min-h-11 border border-gold/40 px-4 text-gold disabled:opacity-50" disabled={!selection || preview.isPending || !!pending} onClick={() => preview.mutate(selection)}>{preview.isPending ? 'Preparing preview…' : 'Preview action'}</button>
    {preview.isError && <p role="alert" className="text-ruby">Could not prepare this action. Try the preview again.</p>}
    {pending && !result && <button type="button" disabled={confirmation.isPending} className="min-h-11 text-gold" onClick={() => confirmation.mutate(pending)}>Retry same action</button>}
    {confirmation.isError && <p role="alert">{confirmation.error instanceof APIError && confirmation.error.status === 409 ? confirmation.error.message : 'Could not confirm delivery. Retry the same action; its ID is retained.'}</p>}
    {completed && <section aria-label="Saved action result"><p>Action saved.</p>{completed.resolution_events?.filter(event => event.type === 'srd_check').map((event, index) => <CheckReceipt key={index} receipt={event.details.receipt as Receipt} />)}</section>}
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
      {result.available && !completed && <button type="button" disabled={confirmation.isPending} className="min-h-11 border border-gold px-4 text-gold disabled:opacity-50" onClick={() => {
        const request = requestRef.current ?? { action_id: crypto.randomUUID(), expected_revision: result.expected_revision, intent: result.intent, ...(result.intent.kind === 'saving-throw' ? { target_id: actorId } : {}) };
        requestRef.current = request; setPending(request);
        try { sessionStorage.setItem(storageKey, JSON.stringify(request)); } catch { /* In-memory retry still keeps the ID. */ }
        confirmation.mutate(request);
      }}>{confirmation.isPending ? 'Confirming…' : pending ? 'Retry same action' : 'Confirm action'}</button>}
      <p className="text-pewter">Confirmation records an ordinary check. Scene outcomes require adjudication; combat actions are unavailable.</p>
    </div>}
  </section>;
}

import { useId, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getReturnReminder } from '../../api/reminder';

export function ReturnReminder({ campaignId }: { readonly campaignId: string }) {
  const [open, setOpen] = useState(false);
  const contentId = useId();
  const toggleRef = useRef<HTMLButtonElement>(null);
  const reminder = useQuery({
    queryKey: ['campaign', campaignId, 'journal', 'reminder'],
    queryFn: () => getReturnReminder(campaignId), enabled: open, retry: false,
  });
  const data = reminder.isError ? undefined : reminder.data;
  return <section onKeyDown={event => {
    if (open && event.key === 'Escape') {
      event.stopPropagation(); setOpen(false); toggleRef.current?.focus({ preventScroll: true });
    }
  }} aria-label="Return reminder" className="relative min-w-0 text-sm">
    <button ref={toggleRef} type="button" aria-expanded={open} aria-controls={contentId} onClick={() => setOpen(!open)} className="min-h-11 border border-gold/20 bg-charcoal px-3 text-gold focus:outline-none focus:ring-2 focus:ring-gold">
      {open ? 'Dismiss reminder' : 'Where was I?'}
    </button>
    <div id={contentId} hidden={!open} role="region" aria-label="Current situation" tabIndex={0} className="absolute left-0 top-full sm:left-auto sm:right-0 z-30 mt-1 max-h-[min(24dvh,12rem)] w-[min(24rem,80vw)] overflow-y-auto overscroll-contain break-words border border-gold/20 bg-charcoal px-3 pt-3">
      <div role="status" className="space-y-2 pb-3 text-champagne">
        {open && reminder.isPending && <p className="text-pewter">Recalling your situation… You can keep playing.</p>}
        {reminder.isError && <p className="text-pewter">Reminder unavailable. You can keep playing. <button type="button" className="min-h-11 text-gold underline" onClick={() => void reminder.refetch()}>Retry reminder</button></p>}
        {data && <>
          <p><strong>Current scene:</strong> {data.scene?.text ?? 'No current location is known yet.'}</p>
          <p><strong>Active quest:</strong> {data.objective?.text ?? 'No active quest is recorded.'}</p>
          {data.known_facts.length > 0 ? <ul aria-label="Known facts" className="list-disc space-y-1 pl-5">{data.known_facts.map(fact => <li key={fact.source_id}>{fact.text}</li>)}</ul> : <p className="text-pewter">No current known facts are recorded.</p>}
        </>}
      </div>
    </div>
  </section>;
}

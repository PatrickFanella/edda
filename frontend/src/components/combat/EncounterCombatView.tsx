import { useState } from 'react';
import type { CharacterSheetProjection, Encounter } from '../../api/types';
import { ActionPreviewPanel } from '../character/ActionPreviewPanel';
import type { NarrativeEntryItem } from '../narrative/NarrativeEntry';
import { NarrativePanel } from '../narrative/NarrativePanel';
import { EncounterSummary } from './EncounterSummary';

/** The saved encounter, rather than transient socket events, owns this view. */
export function EncounterCombatView({ campaignId, actorId, sheet, encounter, entries }: {
 readonly campaignId: string;
 readonly actorId: string;
 readonly sheet: CharacterSheetProjection;
 readonly encounter: Encounter;
 readonly entries: NarrativeEntryItem[];
}) {
 const [choosing, setChoosing] = useState(false);
 return <section aria-label="Narrative combat" className="flex h-full min-h-0 flex-col gap-2 overflow-hidden">
   <EncounterSummary encounter={encounter} />
   {choosing ? <div className="min-h-0 flex-1 overflow-y-auto overflow-x-hidden"><ActionPreviewPanel campaignId={campaignId} actorId={actorId} sheet={sheet} /></div>
    : <NarrativePanel className="min-h-0 flex-1 overflow-hidden" contentClassName="h-full max-h-none min-h-0" entries={entries} streamingEntry={null} isLoading={false} />}
   <button type="button" className="hud-btn hud-btn-primary min-h-11 shrink-0" onClick={() => setChoosing(value => !value)}>{choosing ? 'Read combat story' : 'Choose combat action'}</button>
 </section>;
}

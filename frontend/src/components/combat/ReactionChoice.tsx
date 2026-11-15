import { useState } from 'react';
import type { CharacterSheetProjection, Encounter } from '../../api/types';
import { ActionPreviewPanel } from '../character/ActionPreviewPanel';

/** A saved trigger offers consent; previewing either choice never spends it. */
export function ReactionChoice({ campaignId, actorId, sheet, encounter }: {
 readonly campaignId: string;
 readonly actorId: string;
 readonly sheet: CharacterSheetProjection;
 readonly encounter: Encounter;
}) {
 const [selection, setSelection] = useState('');
 const trigger = encounter.reaction?.pending;
 if (!trigger) return null;
 const target = encounter.actors.find(actor => actor.id === trigger.target_id);
 return <section aria-label="Reaction choice" className="flex min-h-0 flex-1 flex-col gap-2 overflow-hidden">
  <div className="min-h-0 flex-1 overflow-y-auto">
   <p role="status" className="text-sm text-champagne">{target?.name ?? 'The visible foe'} is about to move out of your 5-foot reach. You may make one Longsword opportunity attack before that movement, spending your reaction, or decline at no cost.</p>
   {selection && <ActionPreviewPanel key={selection} campaignId={campaignId} actorId={actorId} sheet={sheet} initialSelection={selection} reaction={encounter.reaction} />}
  </div>
  {!selection && <div className="flex shrink-0 flex-col gap-2">
   <button type="button" className="hud-btn hud-btn-primary min-h-11" onClick={() => setSelection(`combat-react:${trigger.weapon_id}`)}>Preview opportunity attack</button>
   <button type="button" className="hud-btn min-h-11" onClick={() => setSelection('combat-decline-reaction')}>Preview decline</button>
  </div>}
 </section>;
}

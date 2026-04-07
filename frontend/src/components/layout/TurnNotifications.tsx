import { useEffect, useState } from 'react';

import type { StateChange } from '../../api/types';

interface TurnNotificationsProps {
  readonly stateChanges: StateChange[];
}

interface Notification {
  id: string;
  message: string;
}

const DISMISS_MS = 4000;

function deriveNotifications(changes: StateChange[]): Notification[] {
  const notifications: Notification[] = [];

  for (const change of changes) {
    if (change.entity_type === 'npc' && change.change_type === 'created') {
      const name = typeof change.details?.name === 'string' ? change.details.name : 'unknown';
      notifications.push({ id: `${change.entity_id}-npc`, message: `New NPC met: ${name}` });
    }

    if (change.entity_type === 'world_fact' && change.change_type === 'created') {
      notifications.push({ id: `${change.entity_id}-fact`, message: 'New lore discovered' });
    }
  }

  return notifications;
}

export function TurnNotifications({ stateChanges }: TurnNotificationsProps) {
  const [visible, setVisible] = useState<Notification[]>([]);
  const [dismissing, setDismissing] = useState<Set<string>>(new Set());

  useEffect(() => {
    const next = deriveNotifications(stateChanges);
    if (next.length === 0) return;
    setVisible(next);
    setDismissing(new Set());

    const fadeTimer = setTimeout(() => {
      setDismissing(new Set(next.map((n) => n.id)));
    }, DISMISS_MS - 500);

    const clearTimer = setTimeout(() => {
      setVisible([]);
      setDismissing(new Set());
    }, DISMISS_MS);

    return () => {
      clearTimeout(fadeTimer);
      clearTimeout(clearTimer);
    };
  }, [stateChanges]);

  if (visible.length === 0) return null;

  return (
    <div className="pointer-events-none fixed bottom-6 right-6 z-50 flex flex-col gap-2">
      {visible.map((n) => (
        <div
          key={n.id}
          className={`pointer-events-auto border border-gold/40 bg-charcoal px-5 py-3 text-sm text-champagne shadow-lg transition-all duration-500 ${
            dismissing.has(n.id) ? 'translate-x-8 opacity-0' : 'translate-x-0 opacity-100 animate-slide-in-right'
          }`}
        >
          {n.message}
        </div>
      ))}
    </div>
  );
}

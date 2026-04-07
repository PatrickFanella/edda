import { useCallback, useEffect, useState } from 'react';

import type { TurnResponseWithChoices } from './useWebSocket';

const CODEX_ENTITY_TYPES = new Set([
  'language',
  'culture',
  'belief_system',
  'economic_system',
  'world_fact',
]);

interface UseDiscoveryBadgeOptions {
  latestResult: TurnResponseWithChoices | null;
  entityTypes?: ReadonlySet<string>;
}

interface UseDiscoveryBadgeResult {
  hasUnread: boolean;
  clearUnread: () => void;
}

export function useDiscoveryBadge({
  latestResult,
  entityTypes = CODEX_ENTITY_TYPES,
}: UseDiscoveryBadgeOptions): UseDiscoveryBadgeResult {
  const [hasUnread, setHasUnread] = useState(false);

  useEffect(() => {
    if (!latestResult?.state_changes) return;

    const hasNewDiscovery = latestResult.state_changes.some(
      (change) => entityTypes.has(change.entity_type) && change.change_type === 'created',
    );

    if (hasNewDiscovery) {
      setHasUnread(true);
    }
  }, [latestResult, entityTypes]);

  const clearUnread = useCallback(() => {
    setHasUnread(false);
  }, []);

  return { hasUnread, clearUnread };
}

export { CODEX_ENTITY_TYPES };

import { createContext, useCallback, useMemo, useState, type PropsWithChildren } from 'react';

import type { CampaignResponse } from '../api/types';

export interface ActiveCampaignState {
  campaignId: string | null;
  campaign: CampaignResponse | null;
}

export interface CampaignContextValue extends ActiveCampaignState {
  setActiveCampaign: (campaign: CampaignResponse | null) => void;
  setActiveCampaignId: (campaignId: string | null) => void;
  clearActiveCampaign: () => void;
}

export interface CampaignProviderProps extends PropsWithChildren {
  initialCampaign?: CampaignResponse | null;
}

const DEFAULT_ACTIVE_CAMPAIGN_STATE: ActiveCampaignState = {
  campaignId: null,
  campaign: null,
};

export const CampaignContext = createContext<CampaignContextValue | null>(null);

export function useCampaignState(initialCampaign: CampaignResponse | null = null): CampaignContextValue {
  const [state, setState] = useState<ActiveCampaignState>(() => ({
    campaignId: initialCampaign?.id ?? DEFAULT_ACTIVE_CAMPAIGN_STATE.campaignId,
    campaign: initialCampaign,
  }));

  const setActiveCampaign = useCallback((campaign: CampaignResponse | null) => {
    setState({
      campaignId: campaign?.id ?? null,
      campaign,
    });
  }, []);

  const setActiveCampaignId = useCallback((campaignId: string | null) => {
    setState((current) => {
      if (campaignId === null) {
        return DEFAULT_ACTIVE_CAMPAIGN_STATE;
      }

      if (current.campaign?.id === campaignId) {
        return current;
      }

      return {
        campaignId,
        campaign: null,
      };
    });
  }, []);

  const clearActiveCampaign = useCallback(() => {
    setState(DEFAULT_ACTIVE_CAMPAIGN_STATE);
  }, []);

  return useMemo(
    () => ({
      ...state,
      setActiveCampaign,
      setActiveCampaignId,
      clearActiveCampaign,
    }),
    [clearActiveCampaign, setActiveCampaign, setActiveCampaignId, state],
  );
}

export function CampaignProvider({ children, initialCampaign = null }: CampaignProviderProps) {
  const value = useCampaignState(initialCampaign);

  return <CampaignContext.Provider value={value}>{children}</CampaignContext.Provider>;
}

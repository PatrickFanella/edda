import { useContext } from 'react';

import { CampaignContext, type CampaignContextValue } from '../context/CampaignContext';

export function useCampaign(): CampaignContextValue {
  const context = useContext(CampaignContext);

  if (context === null) {
    throw new Error('useCampaign must be used within a CampaignProvider');
  }

  return context;
}

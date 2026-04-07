import { useQuery } from '@tanstack/react-query';

import { getCampaignTime } from '../../api/campaigns';

interface TimeWidgetProps {
  readonly campaignId: string;
}

export function TimeWidget({ campaignId }: TimeWidgetProps) {
  const { data, isPending } = useQuery({
    queryKey: ['campaign', campaignId, 'campaign_time'],
    queryFn: () => getCampaignTime(campaignId),
    enabled: campaignId.length > 0,
  });

  if (isPending || !data) {
    return null;
  }

  const timeStr = `Day ${data.day}, ${String(data.hour).padStart(2, '0')}:${String(data.minute).padStart(2, '0')}`;

  return (
    <span className="font-heading text-xs uppercase tracking-[0.2em] text-gold/80">
      {timeStr}
    </span>
  );
}

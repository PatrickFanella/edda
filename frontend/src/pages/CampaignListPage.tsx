import { useCallback, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router';

import { deleteCampaign, listCampaigns } from '../api/campaigns';
import { ConfirmationDialog } from '../components/layout/ConfirmationDialog';
import { AppShell } from '../components/layout/AppShell';

function queryErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to load campaigns.';
}

export function CampaignListPage() {
  const queryClient = useQueryClient();
  const campaignsQuery = useQuery({
    queryKey: ['campaigns'],
    queryFn: listCampaigns,
  });
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string } | null>(null);

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;
    try {
      await deleteCampaign(deleteTarget.id);
      await queryClient.invalidateQueries({ queryKey: ['campaigns'] });
    } catch {
      // Silently handle — user will see the campaign remains
    }
    setDeleteTarget(null);
  }, [deleteTarget, queryClient]);

  const actions = (
    <Link
      to="/new"
      className="inline-flex items-center justify-center bg-ruby px-4 py-2 text-sm font-semibold uppercase tracking-wide text-champagne transition-all duration-200 hover:bg-ruby-light hover:scale-[1.02] focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian"
    >
      New campaign
    </Link>
  );

  return (
    <AppShell
      title="Campaigns"
      description="Pick up an existing campaign or start a new one."
      actions={actions}
    >
      {campaignsQuery.isPending ? (
        <div className="border border-gold/20 bg-charcoal p-6 text-sm text-champagne/70">Loading campaigns…</div>
      ) : campaignsQuery.isError ? (
        <div className="border border-ruby/40 bg-ruby/10 p-6 text-sm text-ruby">
          {queryErrorMessage(campaignsQuery.error)}
        </div>
      ) : campaignsQuery.data.campaigns.length === 0 ? (
        <div className="flex flex-col gap-4 border border-dashed border-gold/15 bg-charcoal p-8 text-champagne/70">
          <div className="space-y-2">
            <h2 className="font-heading text-xl font-semibold uppercase tracking-wide text-champagne">No campaigns yet</h2>
            <p className="text-sm leading-6">Create your first campaign to start building a world.</p>
          </div>
          <div>
            <Link
              to="/new"
              className="inline-flex items-center justify-center border border-gold/40 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-gold transition hover:border-gold hover:text-gold-light focus:outline-none focus:ring-2 focus:ring-gold focus:ring-offset-2 focus:ring-offset-obsidian"
            >
              Create a campaign
            </Link>
          </div>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {campaignsQuery.data.campaigns.map((campaign) => (
            <article
              key={campaign.id}
              className="deco-corners deco-pattern flex h-full flex-col justify-between border-2 border-gold/20 bg-charcoal p-6 transition-all duration-200 hover:border-gold/40 hover:-translate-y-0.5"
            >
              <div className="space-y-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-2">
                    <h2 className="font-heading text-xl font-semibold uppercase tracking-wide text-champagne">{campaign.name}</h2>
                    <p className="text-sm uppercase tracking-[0.2em] text-pewter">{campaign.status}</p>
                  </div>
                  <div className="rounded-sm border border-sapphire/30 px-3 py-1 text-xs font-medium text-sapphire">
                    {campaign.genre || 'Unspecified genre'}
                  </div>
                </div>
                <p className="text-sm leading-6 text-champagne/70">
                  {campaign.description || 'No description yet.'}
                </p>
                <dl className="grid gap-3 text-sm text-champagne/70 sm:grid-cols-2">
                  <div>
                    <dt className="font-medium text-pewter">Tone</dt>
                    <dd>{campaign.tone || 'Not set'}</dd>
                  </div>
                  <div>
                    <dt className="font-medium text-pewter">Themes</dt>
                    <dd>{campaign.themes.length > 0 ? campaign.themes.join(', ') : 'None yet'}</dd>
                  </div>
                </dl>
              </div>
              <div className="flex items-center gap-3 pt-6">
                <Link
                  to={`/play/${campaign.id}`}
                  className="inline-flex items-center justify-center border-2 border-ruby/30 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-champagne transition-all duration-200 hover:border-ruby hover:bg-ruby hover:text-champagne focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian"
                >
                  Open campaign
                </Link>
                <Link
                  to={`/replay/${campaign.id}`}
                  className="inline-flex items-center justify-center border border-gold/20 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-gold transition-all duration-200 hover:border-gold hover:text-gold-light focus:outline-none focus:ring-2 focus:ring-gold focus:ring-offset-2 focus:ring-offset-obsidian"
                >
                  Replay
                </Link>
                <button
                  type="button"
                  onClick={() => setDeleteTarget({ id: campaign.id, name: campaign.name })}
                  className="inline-flex items-center justify-center border border-pewter/20 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-pewter transition-all duration-200 hover:border-ruby/40 hover:text-ruby focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian"
                >
                  Delete
                </button>
              </div>
            </article>
          ))}
        </div>
      )}

      <ConfirmationDialog
        open={deleteTarget !== null}
        title="Delete campaign"
        message={`Are you sure you want to delete "${deleteTarget?.name ?? ''}"? This action cannot be undone.`}
        confirmLabel="Delete"
        destructive
        onConfirm={() => void handleDeleteConfirm()}
        onCancel={() => setDeleteTarget(null)}
      />
    </AppShell>
  );
}

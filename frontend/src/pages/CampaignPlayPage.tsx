import { useCallback, useEffect, useId, useLayoutEffect, useMemo, useRef, useState, useSyncExternalStore } from 'react';
import type { ReactNode } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useLocation, useParams } from 'react-router';

import { getActionSuggestions } from '../api/suggestions';
import { createManualSave, getCampaign, startOverCampaign } from '../api/campaigns';
import { getCampaignCharacter, getCampaignCharacterInventory, previewCharacterAction, previewCharacterTextAction } from '../api/characters';
import { listCampaignQuests } from '../api/quests';
import type { ActionPreview, CampaignResponse, CharacterResponse, ItemResponse, OpeningSceneResponse, QuestResponse } from '../api/types';
import { AudioControls } from '../components/audio/AudioControls';
import { CharacterSheet } from '../components/character/CharacterSheet';
import { EncounterCombatView } from '../components/combat/EncounterCombatView';
import { EncounterSummary } from '../components/combat/EncounterSummary';
import { CombatView } from '../components/combat/CombatView';
import { ExportDialog } from '../components/export/ExportDialog';
import { ConfirmationDialog } from '../components/layout/ConfirmationDialog';
import { InventoryPanel } from '../components/inventory/InventoryPanel';
import { JournalPanel } from '../components/journal/JournalPanel';
import { AppShell } from '../components/layout/AppShell';
import { HudPanel } from '../components/layout/HudPanel';
import { TabBar } from '../components/layout/TabBar';
import { LogPanel } from '../components/logs/LogPanel';
import { ChoiceList } from '../components/narrative/ChoiceList';
import { NarrativePanel } from '../components/narrative/NarrativePanel';
import type { NarrativeEntryItem } from '../components/narrative/NarrativeEntry';
import { ReturnReminder } from '../components/narrative/ReturnReminder';
import { PlayerInput } from '../components/narrative/PlayerInput';
import { NPCPanel } from '../components/npcs/NPCPanel';
import { ThinkingIndicator } from '../components/layout/ThinkingIndicator';
import { TimeWidget } from '../components/layout/TimeWidget';
import { TurnNotifications } from '../components/layout/TurnNotifications';
import { SavesList } from '../components/saves/SavesList';
import { QuestPanel } from '../components/quests/QuestPanel';
import { WorldPanel } from '../components/world/WorldPanel';
import { CampaignContext, useCampaignState } from '../context/CampaignContext';
import { useAudio } from '../hooks/useAudio';
import { useAudioEffects } from '../hooks/useAudioEffects';
import { useCampaign } from '../hooks/useCampaign';
import { useDiscoveryBadge } from '../hooks/useDiscoveryBadge';
import { useNarrative, type UseNarrativeResult } from '../hooks/useNarrative';
import { useRefreshAfterTurn } from '../hooks/useRefreshAfterTurn';
import type { StartupPlaySeed } from '../lib/startupWorkflow';

const playTabs = [
  { key: 'narrative', label: 'Play' },
  { key: 'character', label: 'Character' },
  { key: 'world', label: 'World' },
  { key: 'journal', label: 'Journal' },
] as const;

type CampaignPlayTab = (typeof playTabs)[number]['key'];
type HudMode = 'combat' | 'dialogue' | 'exploration';
type SharedNarrativeState = Pick<
  UseNarrativeResult,
  'connectionStatus' | 'entries' | 'streamingEntry' | 'suggestedChoices' | 'currentStatus' | 'combatActive' | 'isLoading' | 'error' | 'sendAction'
>;

interface SeededNarrativeState {
  readonly entries: NarrativeEntryItem[];
  readonly suggestedChoices: readonly { id: string; text: string }[];
}

function queryErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to load the campaign.';
}

export function CampaignPlayPage() {
  const { id } = useParams();
  const location = useLocation();
  const campaignId = id?.trim() ?? '';
  const [activeTab, setActiveTab] = useState<CampaignPlayTab>('narrative');
  const startupSeed = useMemo(() => readStartupSeed(location.state), [location.state]);

  const campaignQuery = useQuery({
    queryKey: ['campaign', campaignId],
    queryFn: () => getCampaign(campaignId),
    enabled: campaignId.length > 0,
  });

  if (!campaignId) {
    return (
      <AppShell title="Play campaign" description="A campaign id is required before the play view can load." actions={<BackToCampaignsLink />}>
        <ErrorPanel message="Missing campaign id. Open a campaign from the campaign list and try again." />
      </AppShell>
    );
  }

  if (campaignQuery.isPending) {
    return (
      <AppShell title="Loading campaign" description="Preparing the play table and connecting the live narrative feed." actions={<BackToCampaignsLink />}>
        <LoadingPanel message="Loading campaign…" />
      </AppShell>
    );
  }

  if (campaignQuery.isError) {
    return (
      <AppShell title="Play campaign" description="The campaign could not be loaded." actions={<BackToCampaignsLink />}>
        <ErrorPanel message={queryErrorMessage(campaignQuery.error)} />
      </AppShell>
    );
  }

  return (
    <CampaignPlayWorkspace
      key={campaignId}
      campaignId={campaignId}
      campaign={campaignQuery.data}
      activeTab={activeTab}
      onTabChange={setActiveTab}
      startupSeed={startupSeed}
    />
  );
}

interface CampaignPlayWorkspaceProps {
  readonly campaignId: string;
  readonly campaign: CampaignResponse;
  readonly activeTab: CampaignPlayTab;
  readonly onTabChange: (tab: CampaignPlayTab) => void;
  readonly startupSeed: StartupPlaySeed | null;
}

function CampaignPlayWorkspace({ campaignId, campaign, activeTab, onTabChange, startupSeed }: CampaignPlayWorkspaceProps) {
  const campaignState = useCampaignState(campaign);
  const { setActiveCampaign, setActiveCampaignId } = campaignState;

  useEffect(() => {
    setActiveCampaignId(campaignId);
  }, [campaignId, setActiveCampaignId]);

  useEffect(() => {
    setActiveCampaign(campaign);
  }, [campaign, setActiveCampaign]);

  return (
    <CampaignContext.Provider value={campaignState}>
      <CampaignPlayContent
        campaignId={campaignId}
        campaign={campaign}
        activeTab={activeTab}
        onTabChange={onTabChange}
        startupSeed={startupSeed}
      />
    </CampaignContext.Provider>
  );
}

function CampaignPlayContent({
  campaignId,
  campaign,
  activeTab,
  onTabChange,
  startupSeed,
}: {
  readonly campaignId: string;
  readonly campaign: CampaignResponse;
  readonly activeTab: CampaignPlayTab;
  readonly onTabChange: (tab: CampaignPlayTab) => void;
  readonly startupSeed: StartupPlaySeed | null;
}) {
  const companionVisible = useSyncExternalStore(subscribeCompanionLayout, () => window.matchMedia('(min-width: 80rem)').matches, () => false);
  const tabIdPrefix = useId();
  const storyRef = useRef<HTMLDivElement>(null);
  const returnScrollTop = useRef<number | null>(null);
  const returnFocus = useRef<HTMLElement | null>(null);
  useLayoutEffect(() => {
    if (activeTab === 'narrative' && returnScrollTop.current !== null) {
      storyRef.current?.querySelector('[role="log"]')?.scrollTo({ top: returnScrollTop.current, behavior: 'instant' });
    }
  }, [activeTab]);
  const closeDetails = () => {
    setTextPreview(null);
    onTabChange('narrative');
    requestAnimationFrame(() => {
      const target = returnFocus.current;
      if (target?.isConnected) target.focus({ preventScroll: true });
      else document.getElementById(`${tabIdPrefix}-tab-narrative`)?.focus({ preventScroll: true });
    });
  };
  const narrative = useNarrative();
  const [textPreview, setTextPreview] = useState<{ preview: ActionPreview; sequence: number } | null>(null);
  const audio = useAudio();
  useAudioEffects(narrative.latestResult, narrative.combatActive, audio);
  const seededNarrative = useMemo(() => buildSeededNarrativeState(campaign, startupSeed, narrative.entries), [campaign, narrative.entries, startupSeed]);
  useRefreshAfterTurn(campaignId, narrative.latestResult);

  const { hasUnread: worldHasUnread, clearUnread: clearWorldUnread } = useDiscoveryBadge({
    latestResult: narrative.latestResult,
  });

  const badgedTabs = useMemo(
    () => playTabs.map((tab) => (tab.key === 'world' ? { ...tab, badge: worldHasUnread } : tab)),
    [worldHasUnread],
  );

  const handleTabChange = useCallback(
    (tab: CampaignPlayTab) => {
      if (activeTab === 'character' && tab !== 'character') setTextPreview(null);
      if (activeTab === 'narrative' && tab !== 'narrative') {
        returnScrollTop.current = storyRef.current?.querySelector('[role="log"]')?.scrollTop ?? null;
      }
      if (tab === 'world') {
        clearWorldUnread();
      }
      onTabChange(tab);
    },
    [activeTab, clearWorldUnread, onTabChange, setTextPreview],
  );

  const questsQuery = useQuery({
    queryKey: ['campaign', campaignId, 'quests'],
    queryFn: () => listCampaignQuests(campaignId),
    enabled: campaignId.length > 0,
  });

  const characterQuery = useQuery({
    queryKey: ['campaign', campaignId, 'character'],
    queryFn: () => getCampaignCharacter(campaignId),
    enabled: campaignId.length > 0,
  });

  const inventoryQuery = useQuery({
    queryKey: ['campaign', campaignId, 'character', 'inventory'],
    queryFn: () => getCampaignCharacterInventory(campaignId),
    enabled: campaignId.length > 0,
  });

  const suggestedChoices = useMemo(
    () => getSuggestedChoicesForFrame(narrative, seededNarrative),
    [narrative, seededNarrative],
  );

  const hudMode = useMemo(() => getHudMode(narrative.combatActive, suggestedChoices.length), [narrative.combatActive, suggestedChoices.length]);

  const levelUpMessage = useMemo(() => {
    const changes = narrative.latestResult?.state_changes;
    if (!changes) return null;
    const levelChange = changes.find((c) => c.entity_type === 'player_character' && c.change_type === 'level');
    if (!levelChange) return null;
    const newLevel = levelChange.details?.new_value;
    return `Level Up! You reached level ${newLevel ?? '??'}`;
  }, [narrative.latestResult]);

  const [showStartOverDialog, setShowStartOverDialog] = useState(false);
  const [showExportDialog, setShowExportDialog] = useState(false);
  const [showSaves, setShowSaves] = useState(false);
  const queryClient = useQueryClient();
  const startOverMutation = useMutation({
    mutationFn: () => startOverCampaign(campaignId),
    onSuccess: () => {
      setShowStartOverDialog(false);
      queryClient.invalidateQueries({ queryKey: ['campaign', campaignId] });
      window.location.reload();
    },
  });

  return (
    <AppShell
      title={campaign.name}
      description={campaign.description || 'Live narrative play for this campaign.'}
      actions={<CampaignPlayHeaderActions audio={audio} showSaves={showSaves} onToggleSaves={() => setShowSaves((v) => !v)} />}
      userMenuActions={<CampaignPlayActions campaignId={campaignId} onStartOver={() => setShowStartOverDialog(true)} onExport={() => setShowExportDialog(true)} />}
      variant="game"
    >
      <ExportDialog open={showExportDialog} campaignId={campaignId} onClose={() => setShowExportDialog(false)} />
      <ConfirmationDialog
        open={showStartOverDialog}
        title="Start Over"
        message="This will delete all session history, save points, and campaign time. Your world data (NPCs, locations, quests, etc.) will be preserved. This action cannot be undone."
        confirmLabel="Start Over"
        cancelLabel="Cancel"
        destructive
        onConfirm={() => startOverMutation.mutate()}
        onCancel={() => setShowStartOverDialog(false)}
      />
      <div className="game-play-layout flex h-full flex-col gap-2">
        {levelUpMessage ? <LevelUpBanner message={levelUpMessage} /> : null}
        <GameHudTopBar campaign={campaign} campaignId={campaignId} hudMode={hudMode} />
        {showSaves ? (
          <div className="max-h-40 shrink-0 overflow-y-auto pr-1">
            <SavesList campaignId={campaignId} />
          </div>
        ) : null}
        <div className="game-hud-frame game-play-grid grid min-h-0 flex-1 gap-2 overflow-hidden">
          <div className="game-main-panel grid min-h-0 min-w-0 grid-rows-[minmax(0,1fr)_auto] overflow-hidden">
            <div className={`game-context-workspace min-h-0 ${activeTab !== 'narrative' ? 'has-details' : ''}`}>
            {playTabs.map((tab) => (
              <div
                key={tab.key}
                role={tab.key === 'narrative' && activeTab !== 'narrative' && companionVisible ? 'region' : 'tabpanel'}
                id={`${tabIdPrefix}-panel-${tab.key}`}
                aria-label={tab.key === 'narrative' && activeTab !== 'narrative' && companionVisible ? 'Story' : undefined}
                aria-labelledby={tab.key === 'narrative' && activeTab !== 'narrative' && companionVisible ? undefined : `${tabIdPrefix}-tab-${tab.key}`}
                hidden={activeTab !== tab.key && !(tab.key === 'narrative' && companionVisible)}
                tabIndex={tab.key === 'narrative' && activeTab !== 'narrative' && companionVisible ? undefined : 0}
                ref={tab.key === 'narrative' ? storyRef : undefined}
                className={`h-full min-h-0 overflow-hidden ${tab.key === 'narrative' ? 'game-story-panel' : 'game-detail-panel'}`}
                onKeyDown={(event) => { if (event.key === 'Escape' && tab.key !== 'narrative') { event.preventDefault(); closeDetails(); } }}
              >
                {activeTab === tab.key || tab.key === 'narrative' ? (
                  <div className="flex h-full min-h-0 flex-col">
                    {tab.key !== 'narrative' ? <button type="button" onClick={closeDetails} className="hud-btn hud-btn-secondary mb-2 min-h-11 shrink-0">Close details</button> : null}
                    <div className="min-h-0 flex-1"><PlayTabContent campaignId={campaignId} activeTab={tab.key} narrative={narrative} seededNarrative={seededNarrative} suggestedChoices={suggestedChoices} textPreview={textPreview} onTextPreview={preview => {
                      returnFocus.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
                      setTextPreview(previous => ({ preview, sequence: (previous?.sequence ?? 0) + 1 }));
                      handleTabChange('character');
                      requestAnimationFrame(() => document.getElementById(`${tabIdPrefix}-panel-character`)?.focus({ preventScroll: true }));
                    }} /></div>
                  </div>
                ) : null}
              </div>
            ))}
            </div>
            <div className="min-h-0 overflow-hidden">
              <TurnNotifications stateChanges={narrative.latestResult?.state_changes ?? []} />
            </div>
          </div>
          <GameHudSidebar
            campaign={campaign}
            hudMode={hudMode}
            connectionStatus={narrative.connectionStatus}
            isLoading={narrative.isLoading}
            quests={questsQuery.data ?? []}
            character={characterQuery.data ?? null}
            inventory={inventoryQuery.data ?? []}
            suggestedChoiceCount={suggestedChoices.length}
          />
          <div className="game-play-nav game-hud-panel-scene flex shrink-0 items-center border-2 bg-charcoal/75 p-1.5"
            onPointerDownCapture={() => { if (storyRef.current?.contains(document.activeElement)) returnFocus.current = document.activeElement as HTMLElement; }}
          >
            <TabBar idPrefix={tabIdPrefix} tabs={badgedTabs} activeTab={activeTab} onChange={handleTabChange} />
          </div>
        </div>
      </div>
    </AppShell>
  );
}

function PlayTabContent({
  campaignId,
  activeTab,
  narrative,
  seededNarrative,
  suggestedChoices,
  textPreview,
  onTextPreview,
}: {
  readonly campaignId: string;
  readonly activeTab: CampaignPlayTab;
  readonly narrative: SharedNarrativeState;
  readonly seededNarrative: SeededNarrativeState;
  readonly suggestedChoices: readonly { id: string; text: string }[];
  readonly textPreview: { preview: ActionPreview; sequence: number } | null;
  readonly onTextPreview: (preview: ActionPreview) => void;
}) {
  switch (activeTab) {
    case 'narrative':
      return (
        <div className="h-full min-h-0 overflow-hidden">
          <NarrativeTab narrative={narrative} seededNarrative={seededNarrative} suggestedChoices={suggestedChoices} onTextPreview={onTextPreview} />
        </div>
      );
    case 'character':
      return <ContainedTabPanel><CharacterSheet key={textPreview?.sequence ?? 0} campaignId={campaignId} initialPreview={textPreview?.preview} /><section aria-label="Inventory" className="mt-6"><InventoryPanel /></section></ContainedTabPanel>;
    case 'world':
      return <ContainedTabPanel>
        <WorldPanel campaignId={campaignId} />
        <section aria-label="Quests" className="mt-6"><QuestPanel campaignId={campaignId} /></section>
        <section aria-label="NPCs" className="mt-6"><NPCPanel campaignId={campaignId} /></section>
      </ContainedTabPanel>;
    case 'journal':
      return <ContainedTabPanel>
        <JournalPanel campaignId={campaignId} />
        <details className="mt-6"><summary className="hud-btn hud-btn-secondary cursor-pointer">Logs</summary>
          <LogPanel entries={seededNarrative.entries} streamingEntry={narrative.streamingEntry} isLoading={narrative.isLoading} error={narrative.error} />
        </details>
      </ContainedTabPanel>;
    default: {
      const exhaustiveTab: never = activeTab;
      return exhaustiveTab;
    }
  }
}

function ContainedTabPanel({ children }: { readonly children: ReactNode }) {
  return (
    <div className="h-full min-h-0 overflow-y-auto overflow-x-hidden pr-1">
      {children}
    </div>
  );
}

function NarrativeTab({
  narrative,
  seededNarrative,
  suggestedChoices,
  onTextPreview,
}: {
  readonly narrative: SharedNarrativeState;
  readonly seededNarrative: SeededNarrativeState;
  readonly suggestedChoices: readonly { id: string; text: string }[];
  readonly onTextPreview: (preview: ActionPreview) => void;
}) {
  const { campaign, campaignId } = useCampaign();
  const assistance = useQuery({ queryKey: ['campaign', campaignId, 'action-suggestions'], queryFn: () => getActionSuggestions(campaignId!), enabled: !!campaignId, retry: false });
  const { connectionStatus, streamingEntry, currentStatus, combatActive, isLoading, error, sendAction } = narrative;
  const character = useQuery({ queryKey: ['campaign', campaignId, 'character'], queryFn: () => getCampaignCharacter(campaignId!), enabled: campaign?.rules_version === 'srd-5.2.1' });
  const encounter = useQuery({ queryKey: ['campaign', campaignId, 'encounter'], queryFn: () => previewCharacterAction(campaignId!, { actor_id: character.data!.id, kind: 'combat-end-turn' }), enabled: campaign?.rules_version === 'srd-5.2.1' && !!character.data, refetchOnWindowFocus: true });
  const isSRD = campaign?.rules_version === 'srd-5.2.1';
  const [textError, setTextError] = useState<string | null>(null);
  const [isPreparing, setIsPreparing] = useState(false);
  const textRequest = useRef(0);
  useEffect(() => () => { textRequest.current++; }, [campaignId]);
  const prepareText = async (input: string): Promise<boolean> => {
    if (!isSRD) return sendAction(input);
    if (!campaignId || !character.data) return false;
    try {
      if (sessionStorage.getItem(`edda:prepared-action:${campaignId}:${character.data.id}`)) {
        setTextError('Resolve the previous prepared action in Character before preparing another.');
        return false;
      }
    } catch {
      setTextError('This browser cannot retain a prepared action. Use Character to continue.');
      return false;
    }
    const request = ++textRequest.current;
    setTextError(null);
    setIsPreparing(true);
    try {
      const preview = await previewCharacterTextAction(campaignId, character.data.id, input);
      if (request !== textRequest.current) return false;
      if (!preview.available) {
        setTextError(preview.reason ?? 'This action is unavailable. Choose a supported action in Character.');
        return false;
      }
      onTextPreview(preview);
      return true;
    } catch (error) {
      if (request === textRequest.current) setTextError(error instanceof Error ? error.message : 'Could not prepare this action. Your draft is retained.');
      return false;
    } finally {
      if (request === textRequest.current) setIsPreparing(false);
    }
  };

  // Issue #410: Only show combat UI in light/crunch rules modes.
  const rulesMode = campaign?.rules_mode ?? 'narrative';
  const showCombatUI = combatActive && rulesMode !== 'narrative';

  if (encounter.data?.encounter?.status === 'active' && character.data?.sheet && campaignId) {
    return <EncounterCombatView campaignId={campaignId} actorId={character.data.id} sheet={character.data.sheet} encounter={encounter.data.encounter} entries={seededNarrative.entries} />;
  }

  if (showCombatUI) {
    return (
      <div className="flex h-full min-h-0 flex-col gap-3 overflow-y-auto">
        <div className="game-hud-panel game-hud-panel-combat flex shrink-0 flex-wrap items-center justify-between gap-4 border-2 bg-obsidian/75 px-5 py-3">
          <div>
            <h2 className="font-heading text-base font-semibold uppercase tracking-wide text-ruby">Combat</h2>
            <p className="mt-1 text-xs text-pewter">{campaign?.name ?? 'Campaign'} · active combat encounter</p>
          </div>
          <ConnectionBadge status={connectionStatus} isLoading={isLoading} />
        </div>

        <CombatView
          entries={seededNarrative.entries}
          streamingEntry={streamingEntry ?? null}
          onAction={sendAction}
          isLoading={isLoading}
        />

        <ThinkingIndicator status={currentStatus} />

        {error ? <ErrorPanel message={error} /> : null}
      </div>
    );
  }

  return (
    <div className="game-narrative-layout flex h-full min-h-0 flex-col gap-3 overflow-y-auto">
      {encounter.data?.encounter && <EncounterSummary encounter={encounter.data.encounter} />}
      {encounter.isError && <p role="alert" className="text-ruby">Could not load encounter turns. Refresh before acting.</p>}
      <NarrativePanel
        className="min-h-24 flex-1 overflow-hidden bg-charcoal/70"
        contentClassName="h-full max-h-none min-h-0 text-[1.12rem] leading-6"
        entries={seededNarrative.entries}
        streamingEntry={streamingEntry}
        isLoading={isLoading}
        emptyState={
          <div className="flex min-h-64 flex-1 flex-col items-center justify-center border border-dashed border-pewter/15 bg-charcoal/50 px-6 text-center">
            <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">Awaiting first move</p>
            <p className="mt-3 max-w-md text-sm leading-7 text-pewter">
              Send an action to begin. Player moves, GM narration, system notices, and suggested choices will collect here.
            </p>
          </div>
        }
      />

      <ChoiceList
        className="game-choices-region"
        choices={[...suggestedChoices]}
        onSelectChoice={(choiceText) => {
          void prepareText(choiceText);
        }}
        disabled={isSRD ? isPreparing : isLoading}
      />

      <div className="game-composer-feedback" aria-live="polite">
        <ThinkingIndicator status={currentStatus} />
        {error ? <ErrorPanel message={error} /> : null}
        {textError ? <ErrorPanel message={textError} /> : null}
        {isSRD ? <p className="text-sm text-pewter">Supported text: Make an Athletics check. Prepare it here, then confirm in Character. Other actions are available from the sheet.</p> : null}
      </div>

      <PlayerInput suggestions={assistance.isError ? [] : assistance.data?.candidates} className="game-action-form" onSendAction={prepareText} preparesAction={isSRD} disabled={isSRD ? !character.data : connectionStatus === 'connecting'} isLoading={isSRD ? isPreparing : isLoading} autoFocus />

    </div>
  );
}

function GameHudTopBar({
  campaign,
  campaignId,
  hudMode,
}: {
  readonly campaign: CampaignResponse;
  readonly campaignId: string;
  readonly hudMode: HudMode;
}) {
  return (
    <section className={`game-campaign-header game-hud-panel game-hud-panel-${hudMode} grid gap-3 border-2 bg-obsidian/70 px-4 py-2.5 lg:grid-cols-[1fr_auto]`}>
      <div className="flex min-w-0 flex-wrap items-center gap-3">
        <HudModeBadge mode={hudMode} />
        <h2 className="min-w-0 truncate font-heading text-xl font-semibold uppercase tracking-[0.14em] text-champagne sm:text-2xl">
          {campaign.name}
        </h2>
      </div>
      <div className="flex flex-wrap items-center gap-2 lg:justify-end">
        <ReturnReminder key={campaignId} campaignId={campaignId} />
        <div className="border border-pewter/20 bg-charcoal/70 px-3 py-1.5">
          <TimeWidget campaignId={campaignId} />
        </div>
      </div>
    </section>
  );
}

function GameHudSidebar({
  campaign,
  hudMode,
  connectionStatus,
  isLoading,
  quests,
  character,
  inventory,
  suggestedChoiceCount,
}: {
  readonly campaign: CampaignResponse;
  readonly hudMode: HudMode;
  readonly connectionStatus: string;
  readonly isLoading: boolean;
  readonly quests: QuestResponse[];
  readonly character: CharacterResponse | null;
  readonly inventory: ItemResponse[];
  readonly suggestedChoiceCount: number;
}) {
  const activeQuest = quests.find((quest) => quest.status === 'active') ?? quests[0] ?? null;
  const activeObjective = activeQuest?.objectives.find((objective) => !objective.completed) ?? activeQuest?.objectives[0] ?? null;
  const quickItems = inventory.slice(0, 5);

  return (
    <aside aria-label="Campaign context" className="game-context-rail min-h-0 flex-col gap-2 overflow-y-auto">
      <div className="min-h-0 space-y-2 overflow-y-auto pr-1">
        <HudPanel title="Vitals" accent="vitals">
          {character ? (
            <>
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-heading text-sm font-semibold uppercase tracking-[0.12em] text-champagne">{character.name}</p>
                  <p className="mt-1 text-[0.7rem] uppercase tracking-[0.16em] text-pewter">Level {character.level} · {humanizeInlineValue(character.status)}</p>
                </div>
                <span className="hud-badge hud-badge-jade">LV {character.level}</span>
              </div>
              <HudMeter label="HP" value={character.hp} max={character.max_hp} tone="jade" />
              {character.rules_version === 'srd-5.2.1'
                ? character.sheet ? <HudMeter label="XP" value={character.sheet.progression.current_xp} max={character.sheet.progression.next_level_xp} tone="sapphire" /> : <p className="text-sm text-pewter">Rules projection unavailable</p>
                : <HudMeter label="XP" value={character.experience} max={nextLevelThreshold(character.level)} tone="sapphire" />}
            </>
          ) : (
            <p className="text-sm leading-6 text-pewter">Character telemetry not loaded yet.</p>
          )}
        </HudPanel>

        <HudPanel title="Active objective" accent="objective">
          {activeQuest ? (
            <div className="space-y-2">
              <p className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-champagne">{activeQuest.title}</p>
              <p className="text-sm leading-6 text-champagne/70">{activeObjective?.description ?? activeQuest.description}</p>
            </div>
          ) : (
            <p className="text-sm leading-6 text-pewter">No tracked objective yet.</p>
          )}
        </HudPanel>

        <HudPanel title="Quick inventory" accent="inventory">
          {quickItems.length > 0 ? (
            <div className="grid grid-cols-5 gap-1.5">
              {quickItems.map((item) => (
                <div key={item.id} title={item.name} className="hud-item-cell">
                  <span className="line-clamp-2">{item.name}</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm leading-6 text-pewter">No quick items detected.</p>
          )}
        </HudPanel>
      </div>

      <div className="mt-auto">
        <HudPanel title="Scene scan" accent="scene">
          <dl className="space-y-2 text-sm text-champagne/70">
            <SummaryRow label="Mode" value={humanizeInlineValue(hudMode)} />
            <SummaryRow label="Rules" value={humanizeInlineValue(campaign.rules_mode)} />
            <SummaryRow label="Status" value={humanizeInlineValue(campaign.status)} />
            <SummaryRow label="Connection" value={isLoading ? 'GM responding' : connectionStatus} capitalize />
            <SummaryRow label="Choices" value={String(suggestedChoiceCount)} />
            <SummaryRow label="Tone" value={campaign.tone || 'Unspecified'} />
          </dl>
        </HudPanel>
      </div>
    </aside>
  );
}

function HudModeBadge({ mode }: { readonly mode: HudMode }) {
  return <span className={`game-hud-mode game-hud-mode-${mode}`}>{mode}</span>;
}

function HudMeter({ label, value, max, tone }: { readonly label: string; readonly value: number; readonly max: number; readonly tone: 'jade' | 'sapphire' }) {
  const pct = max > 0 ? Math.max(0, Math.min(100, Math.round((value / max) * 100))) : 0;
  const fillClass = tone === 'jade' ? 'hud-meter-fill-jade' : 'hud-meter-fill-sapphire';

  return (
    <div className="hud-meter mt-2.5">
      <div className="hud-meter-header">
        <span className="hud-meter-label">{label}</span>
        <span className="hud-meter-value">{value} / {max}</span>
      </div>
      <div className="hud-meter-bar">
        <div className={`hud-meter-fill ${fillClass}`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

function nextLevelThreshold(level: number): number {
  return 1000 * level * (level + 1) / 2;
}

function getSuggestedChoicesForFrame(narrative: SharedNarrativeState, seededNarrative: SeededNarrativeState) {
  if (narrative.suggestedChoices.length > 0) {
    return narrative.suggestedChoices;
  }

  if (narrative.entries.length > 0) {
    return [];
  }

  return seededNarrative.suggestedChoices;
}

function getHudMode(combatActive: boolean, suggestedChoiceCount: number): HudMode {
  if (combatActive) return 'combat';
  if (suggestedChoiceCount > 0) return 'dialogue';
  return 'exploration';
}

function LevelUpBanner({ message }: { readonly message: string }) {
  return (
    <div className="animate-pulse border-2 border-gold/50 bg-gold/10 px-6 py-3 text-center">
      <p className="font-heading text-sm font-semibold uppercase tracking-[0.15em] text-gold">{message}</p>
    </div>
  );
}

function ConnectionBadge({ status, isLoading }: { readonly status: string; readonly isLoading: boolean }) {
  const statusClass =
    status === 'open'
      ? 'hud-status-open'
      : status === 'connecting'
        ? 'hud-status-connecting'
        : status === 'error'
          ? 'hud-status-error'
          : 'hud-status-idle';

  return (
    <div className={`hud-status ${statusClass}`}>
      {isLoading ? 'GM responding' : status}
    </div>
  );
}

function SummaryRow({
  label,
  value,
  capitalize = false,
}: {
  readonly label: string;
  readonly value: string;
  readonly capitalize?: boolean;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-pewter">{label}</dt>
      <dd className={`text-right ${capitalize ? 'capitalize' : ''}`}>{value}</dd>
    </div>
  );
}

function humanizeInlineValue(value: string): string {
  return value
    .split(/[_-]/g)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

function LoadingPanel({ message }: { readonly message: string }) {
  return <div className="border border-pewter/20 bg-charcoal p-6 text-sm text-pewter">{message}</div>;
}

function ErrorPanel({ message }: { readonly message: string }) {
  return <div className="border border-ruby/40 bg-ruby/10 p-6 text-sm text-ruby">{message}</div>;
}

function CampaignPlayHeaderActions({
  audio,
  showSaves,
  onToggleSaves,
}: {
  readonly audio: ReturnType<typeof useAudio>;
  readonly showSaves: boolean;
  readonly onToggleSaves: () => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <AudioControls {...audio} />
      <button
        type="button"
        onClick={onToggleSaves}
        className="hud-btn hud-btn-primary"
      >
        {showSaves ? 'Hide Saves' : 'Saves'}
      </button>
    </div>
  );
}

function CampaignPlayActions({ campaignId, onStartOver, onExport }: { readonly campaignId: string; readonly onStartOver: () => void; readonly onExport: () => void }) {
  const queryClient = useQueryClient();
  const saveMutation = useMutation({
    mutationFn: (name: string) => createManualSave(campaignId, name),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['campaign', campaignId, 'saves'] });
    },
  });

  const handleSave = () => {
    const name = window.prompt('Save name:', `Save ${new Date().toLocaleString()}`);
    if (name !== null && name.trim() !== '') {
      saveMutation.mutate(name.trim());
    }
  };

  return (
    <div className="grid gap-2">
      <button
        type="button"
        onClick={handleSave}
        disabled={saveMutation.isPending}
        className="hud-btn hud-btn-primary"
      >
        {saveMutation.isPending ? 'Saving...' : 'Save'}
      </button>
      <button
        type="button"
        onClick={onExport}
        className="hud-btn hud-btn-secondary"
      >
        Export
      </button>
      <button
        type="button"
        onClick={onStartOver}
        className="hud-btn hud-btn-danger"
      >
        Start Over
      </button>
      <BackToCampaignsLink />
    </div>
  );
}

function BackToCampaignsLink() {
  return (
    <Link
      to="/"
      className="hud-btn hud-btn-secondary w-full"
    >
      Back to campaigns
    </Link>
  );
}

function buildSeededNarrativeState(
  campaign: CampaignResponse,
  startupSeed: StartupPlaySeed | null,
  liveEntries: readonly NarrativeEntryItem[],
): SeededNarrativeState {
  const startupEntry = buildStartupEntry(campaign, startupSeed?.openingScene ?? null, startupSeed?.seededAt ?? null);
  if (!startupEntry) {
    return {
      entries: [...liveEntries],
      suggestedChoices: [],
    };
  }

  const alreadyPresent = liveEntries.some((entry) => entry.kind === 'gm' && isSameOpeningSceneEntry(entry, startupEntry));
  const entries = alreadyPresent ? [...liveEntries] : [startupEntry, ...liveEntries];

  return {
    entries,
    suggestedChoices: startupEntry.choices ?? [],
  };
}

function buildStartupEntry(
  campaign: CampaignResponse,
  openingScene: OpeningSceneResponse | null,
  seededAt: string | null,
): NarrativeEntryItem | null {
  if (!openingScene || openingScene.narrative.trim().length === 0) {
    return null;
  }

  return {
    id: `startup-opening-${campaign.id}`,
    kind: 'gm',
    text: openingScene.narrative,
    timestamp: seededAt ?? campaign.created_at,
    speaker: 'Game Master',
    choices: openingScene.choices.map((choice, index) => ({ id: `startup-choice-${index + 1}`, text: choice })),
  };
}

function isSameOpeningSceneEntry(entry: NarrativeEntryItem, startupEntry: NarrativeEntryItem): boolean {
  const entryChoices = entry.choices?.map((choice) => choice.text) ?? [];
  const startupChoices = startupEntry.choices?.map((choice) => choice.text) ?? [];

  return entry.text.trim() === startupEntry.text.trim() && entryChoices.join('\u0000') === startupChoices.join('\u0000');
}

function readStartupSeed(value: unknown): StartupPlaySeed | null {
  if (!isRecord(value) || !('startupSeed' in value)) {
    return null;
  }

  return isStartupPlaySeed(value.startupSeed) ? value.startupSeed : null;
}

function isStartupPlaySeed(value: unknown): value is StartupPlaySeed {
  return (
    isRecord(value) &&
    typeof value.campaignName === 'string' &&
    typeof value.campaignSummary === 'string' &&
    typeof value.seededAt === 'string' &&
    isOpeningSceneResponse(value.openingScene)
  );
}

function isOpeningSceneResponse(value: unknown): value is OpeningSceneResponse {
  return (
    isRecord(value) &&
    typeof value.narrative === 'string' &&
    Array.isArray(value.choices) &&
    value.choices.every((choice) => typeof choice === 'string')
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function subscribeCompanionLayout(listener: () => void) {
  const query = window.matchMedia('(min-width: 80rem)');
  query.addEventListener('change', listener);
  return () => query.removeEventListener('change', listener);
}

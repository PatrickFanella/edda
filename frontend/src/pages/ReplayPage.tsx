import { useParams } from 'react-router';

import { AppShell } from '../components/layout/AppShell';
import { ReplayControls } from '../components/replay/ReplayControls';
import { ReplayNarrative } from '../components/replay/ReplayNarrative';
import { ReplaySidebar } from '../components/replay/ReplaySidebar';
import { ReplayTimeline } from '../components/replay/ReplayTimeline';
import { useReplayEngine } from '../hooks/useReplayEngine';

export function ReplayPage() {
  const { id } = useParams<{ id: string }>();
  const campaignId = id ?? '';
  const replay = useReplayEngine(campaignId);

  return (
    <AppShell title="Campaign Replay" description="Watch your adventure unfold.">
      <div className="space-y-6">
        <ReplayControls
          isPlaying={replay.isPlaying}
          playbackSpeed={replay.playbackSpeed}
          currentTurnIndex={replay.currentTurnIndex}
          totalTurns={replay.totalTurns}
          play={replay.play}
          pause={replay.pause}
          setSpeed={replay.setSpeed}
          nextTurn={replay.nextTurn}
          prevTurn={replay.prevTurn}
        />
        <div className="grid gap-6 lg:grid-cols-[1fr_300px]">
          <ReplayNarrative
            visibleEntries={replay.visibleEntries}
            currentTurnIndex={replay.currentTurnIndex}
            isPlaying={replay.isPlaying}
          />
          <ReplaySidebar campaignId={campaignId} />
        </div>
        <ReplayTimeline
          currentTurnIndex={replay.currentTurnIndex}
          totalTurns={replay.totalTurns}
          seekTo={replay.seekTo}
        />
      </div>
    </AppShell>
  );
}

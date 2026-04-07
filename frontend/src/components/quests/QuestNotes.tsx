import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { createQuestNote, deleteQuestNote, listQuestNotes } from '../../api/quests';
import type { QuestNoteResponse } from '../../api/types';

interface QuestNotesProps {
  readonly campaignId: string;
  readonly questId: string;
}

function formatTimestamp(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export function QuestNotes({ campaignId, questId }: QuestNotesProps) {
  const queryClient = useQueryClient();
  const [newNote, setNewNote] = useState('');
  const queryKey = ['campaign', campaignId, 'quest-notes', questId] as const;

  const notesQuery = useQuery({
    queryKey,
    queryFn: () => listQuestNotes(campaignId, questId),
  });

  const createMutation = useMutation({
    mutationFn: (content: string) => createQuestNote(campaignId, questId, content),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey });
      setNewNote('');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (noteId: string) => deleteQuestNote(campaignId, questId, noteId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const notes: QuestNoteResponse[] = notesQuery.data ?? [];
  const sorted = [...notes].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

  return (
    <div className="mt-4 space-y-4 border-t-2 border-sapphire/15 pt-4">
      <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-sapphire">Notes</p>

      <div className="space-y-2">
        <textarea
          value={newNote}
          onChange={(e) => setNewNote(e.target.value)}
          placeholder="Add a note about this quest..."
          rows={3}
          className="w-full resize-none border border-sapphire/20 bg-obsidian px-3 py-2 text-sm text-champagne placeholder-pewter/50 outline-none transition focus:border-sapphire"
        />
        <button
          type="button"
          disabled={newNote.trim().length === 0 || createMutation.isPending}
          onClick={() => createMutation.mutate(newNote.trim())}
          className="border border-sapphire/30 bg-sapphire/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-sapphire transition hover:border-sapphire hover:bg-sapphire/20 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {createMutation.isPending ? 'Adding...' : 'Add note'}
        </button>
      </div>

      {notesQuery.isPending ? (
        <p className="text-sm text-pewter">Loading notes...</p>
      ) : notesQuery.isError ? (
        <p className="text-sm text-ruby">Failed to load notes.</p>
      ) : sorted.length === 0 ? (
        <p className="text-sm text-pewter">No notes yet.</p>
      ) : (
        <ul className="space-y-2">
          {sorted.map((note) => (
            <li key={note.id} className="group border border-sapphire/15 bg-obsidian/50 px-4 py-3">
              <p className="text-sm leading-6 text-champagne/80">{note.content}</p>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-xs text-pewter">{formatTimestamp(note.created_at)}</span>
                <button
                  type="button"
                  onClick={() => deleteMutation.mutate(note.id)}
                  disabled={deleteMutation.isPending}
                  className="text-xs text-pewter transition hover:text-ruby"
                >
                  Delete
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

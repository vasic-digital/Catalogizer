import React, { useEffect, useState } from 'react';
import { X } from 'lucide-react';
import { apiService } from '../services/apiService';
import { PlaybackFormatter } from '../utils/playbackFormatter';
import type { UiPlaybackProgress, UiPlaybackSession } from '../types';

interface HistoryDrawerProps {
  mediaItemId: number | null;
  mediaTitle: string;
  open: boolean;
  onClose: () => void;
}

export const HistoryDrawer: React.FC<HistoryDrawerProps> = ({
  mediaItemId,
  mediaTitle,
  open,
  onClose,
}) => {
  const [progress, setProgress] = useState<UiPlaybackProgress | null>(null);
  const [sessions, setSessions] = useState<UiPlaybackSession[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    if (!open || mediaItemId == null) return;
    let cancelled = false;
    setLoaded(false);
    setProgress(null);
    setSessions([]);
    Promise.all([
      apiService.getEntityProgress(mediaItemId),
      apiService.listEntityHistory(mediaItemId, 100),
    ])
      .then(([prog, hist]) => {
        if (cancelled) return;
        setProgress(prog);
        setSessions(hist);
        setLoaded(true);
      })
      .catch(() => {
        if (cancelled) return;
        setLoaded(true);
      });
    return () => {
      cancelled = true;
    };
  }, [open, mediaItemId]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
      onClick={onClose}
    >
      <div
        className="max-h-[85vh] w-full max-w-2xl overflow-y-auto rounded-xl bg-card border border-border p-6 text-foreground shadow-2xl"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
      >
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            <h2 className="text-xl font-bold">Reproduction History</h2>
            <p className="text-sm text-muted-foreground">{mediaTitle}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1 hover:bg-white/10 focus:outline-none focus:ring-2"
            aria-label="Close"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <HistorySummary progress={progress} />

        <div className="mt-4">
          <HistorySessionList sessions={sessions} loaded={loaded} />
        </div>
      </div>
    </div>
  );
};

const HistorySummary: React.FC<{ progress: UiPlaybackProgress | null }> = ({
  progress,
}) => {
  const rows: Array<[string, string]> = progress
    ? [
        ['Total duration', PlaybackFormatter.formatAmount(progress.durationTotal ?? 0, progress.positionUnit)],
        ['Current position', PlaybackFormatter.formatProgress(progress.lastPosition, progress.durationTotal, progress.positionUnit)],
        ['Last session', PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit)],
        ['Times played', PlaybackFormatter.formatReproductionCount(progress.totalReproductions)],
        ['Total reproduced', PlaybackFormatter.formatAmount(progress.aggregateAmount, progress.positionUnit)],
      ]
    : [['Status', 'Never reproduced']];

  return (
    <div className="rounded-lg bg-muted/30 p-4">
      {rows.map(([label, value]) => (
        <div key={label} className="flex items-center justify-between py-1 text-sm">
          <span className="text-muted-foreground">{label}</span>
          <span className="font-semibold">{value}</span>
        </div>
      ))}
    </div>
  );
};

const HistorySessionList: React.FC<{
  sessions: UiPlaybackSession[];
  loaded: boolean;
}> = ({ sessions, loaded }) => {
  if (!loaded) return <p className="text-sm text-muted-foreground">Loading…</p>;
  if (sessions.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No sessions yet. Press Play to record your first.
      </p>
    );
  }
  return (
    <div>
      <h3 className="mb-2 text-sm font-bold">Sessions</h3>
      <ul className="space-y-2">
        {sessions.map((s) => (
          <li
            key={s.id}
            className="flex items-center gap-3 rounded-md bg-muted/30 px-3 py-2 text-sm"
          >
            <span className="w-40 shrink-0">{formatInstant(s.startedAt)}</span>
            <span className="w-24 shrink-0">
              {PlaybackFormatter.formatAmount(s.totalAmount, s.positionUnit)}
            </span>
            <span className={s.completed ? 'text-green-500' : 'text-amber-500'}>
              {s.completed ? 'Completed' : 'In progress'}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
};

function formatInstant(iso: string | null | undefined): string {
  if (!iso) return '-';
  try {
    const d = new Date(iso);
    if (isNaN(d.getTime())) return '-';
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const h = String(d.getHours()).padStart(2, '0');
    const min = String(d.getMinutes()).padStart(2, '0');
    return `${y}-${m}-${day} ${h}:${min}`;
  } catch {
    return '-';
  }
}

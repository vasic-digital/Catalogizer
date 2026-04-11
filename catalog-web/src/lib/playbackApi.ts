import { api } from './api'
import type { UiPlaybackProgress, UiPlaybackSession } from '@/types/playback'

type RawProgress = {
  position_unit?: string
  duration_total?: number | null
  last_position?: number
  last_session_amount?: number
  total_reproductions?: number
  aggregate_amount?: number
  last_session_ended_at?: string | null
}

type RawSession = {
  id?: number
  position_unit?: string
  start_position?: number
  end_position?: number
  total_amount?: number
  started_at?: string
  ended_at?: string | null
  completed?: boolean
}

/**
 * Fetches /api/v1/entities/:id/progress and maps the raw payload to
 * [UiPlaybackProgress]. Returns null when the entity has never been
 * reproduced (backend responds `{ progress: null }`).
 */
export async function getEntityProgress(
  mediaItemId: number,
): Promise<UiPlaybackProgress | null> {
  try {
    const resp = await api.get<{ progress: RawProgress | null }>(
      `/entities/${mediaItemId}/progress`,
    )
    const raw = resp.data?.progress
    if (!raw) return null
    return {
      mediaItemId,
      positionUnit: raw.position_unit ?? 'seconds',
      durationTotal: raw.duration_total ?? null,
      lastPosition: raw.last_position ?? 0,
      lastSessionAmount: raw.last_session_amount ?? 0,
      totalReproductions: raw.total_reproductions ?? 0,
      aggregateAmount: raw.aggregate_amount ?? 0,
      lastSessionEndedAt: raw.last_session_ended_at ?? null,
    }
  } catch {
    return null
  }
}

/**
 * Fetches /api/v1/entities/:id/history and maps raw session rows to
 * [UiPlaybackSession]. Returns an empty array on failure or when the
 * user has no sessions yet.
 */
export async function listEntityHistory(
  mediaItemId: number,
  limit = 50,
): Promise<UiPlaybackSession[]> {
  try {
    const resp = await api.get<{ sessions: RawSession[] }>(
      `/entities/${mediaItemId}/history`,
      { params: { limit } },
    )
    const raw = resp.data?.sessions ?? []
    return raw.map((r) => ({
      id: r.id ?? 0,
      positionUnit: r.position_unit ?? 'seconds',
      startPosition: r.start_position ?? 0,
      endPosition: r.end_position ?? 0,
      totalAmount: r.total_amount ?? 0,
      startedAt: r.started_at ?? '',
      endedAt: r.ended_at ?? null,
      completed: r.completed ?? false,
    }))
  } catch {
    return []
  }
}

export interface UiPlaybackProgress {
  mediaItemId: number
  positionUnit: string
  durationTotal: number | null
  lastPosition: number
  lastSessionAmount: number
  totalReproductions: number
  aggregateAmount: number
  lastSessionEndedAt: string | null
}

export interface UiPlaybackSession {
  id: number
  positionUnit: string
  startPosition: number
  endPosition: number
  totalAmount: number
  startedAt: string
  endedAt: string | null
  completed: boolean
}

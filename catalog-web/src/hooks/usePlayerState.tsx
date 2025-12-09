import { useState, useCallback } from 'react';
import { PlaylistItem } from '@/types/playlists';

interface PlayerState {
  currentItem: PlaylistItem | null;
  isPlaying: boolean;
  isPaused: boolean;
  volume: number;
  isMuted: boolean;
  currentTime: number;
  duration: number;
  playbackRate: number;
  isFullscreen: boolean;
  repeatMode: 'off' | 'one' | 'all';
  isShuffled: boolean;
  queue: PlaylistItem[];
  queueIndex: number;
}

interface UsePlayerStateReturn extends PlayerState {
  setCurrentItem: (item: PlaylistItem | null) => void;
  setIsPlaying: (playing: boolean) => void;
  setIsPaused: (paused: boolean) => void;
  setVolume: (volume: number) => void;
  setIsMuted: (muted: boolean) => void;
  setCurrentTime: (time: number) => void;
  setDuration: (duration: number) => void;
  setPlaybackRate: (rate: number) => void;
  setIsFullscreen: (fullscreen: boolean) => void;
  setRepeatMode: (mode: 'off' | 'one' | 'all') => void;
  setIsShuffled: (shuffled: boolean) => void;
  setQueue: (queue: PlaylistItem[]) => void;
  setQueueIndex: (index: number) => void;
  play: () => void;
  pause: () => void;
  togglePlayPause: () => void;
  playNext: () => void;
  playPrevious: () => void;
  seekTo: (time: number) => void;
  skipTo: (index: number) => void;
  addToQueue: (item: PlaylistItem) => void;
  removeFromQueue: (index: number) => void;
  clearQueue: () => void;
}

export const usePlayerState = (): UsePlayerStateReturn => {
  const [currentItem, setCurrentItem] = useState<PlaylistItem | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [volume, setVolume] = useState(0.8);
  const [isMuted, setIsMuted] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [playbackRate, setPlaybackRate] = useState(1);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [repeatMode, setRepeatMode] = useState<'off' | 'one' | 'all'>('off');
  const [isShuffled, setIsShuffled] = useState(false);
  const [queue, setQueue] = useState<PlaylistItem[]>([]);
  const [queueIndex, setQueueIndex] = useState(0);

  const play = useCallback(() => {
    setIsPlaying(true);
    setIsPaused(false);
  }, []);

  const pause = useCallback(() => {
    setIsPlaying(false);
    setIsPaused(true);
  }, []);

  const togglePlayPause = useCallback(() => {
    if (isPlaying) {
      pause();
    } else {
      play();
    }
  }, [isPlaying, play, pause]);

  const playNext = useCallback(() => {
    if (queueIndex < queue.length - 1) {
      const nextIndex = queueIndex + 1;
      setQueueIndex(nextIndex);
      setCurrentItem(queue[nextIndex]);
      play();
    } else if (repeatMode === 'all') {
      setQueueIndex(0);
      setCurrentItem(queue[0]);
      play();
    } else {
      pause();
    }
  }, [queue, queueIndex, repeatMode, play, pause]);

  const playPrevious = useCallback(() => {
    if (queueIndex > 0) {
      const prevIndex = queueIndex - 1;
      setQueueIndex(prevIndex);
      setCurrentItem(queue[prevIndex]);
      play();
    } else if (repeatMode === 'all') {
      setQueueIndex(queue.length - 1);
      setCurrentItem(queue[queue.length - 1]);
      play();
    }
  }, [queue, queueIndex, play]);

  const seekTo = useCallback((time: number) => {
    setCurrentTime(time);
  }, []);

  const skipTo = useCallback((index: number) => {
    if (index >= 0 && index < queue.length) {
      setQueueIndex(index);
      setCurrentItem(queue[index]);
      play();
    }
  }, [queue, play]);

  const addToQueue = useCallback((item: PlaylistItem) => {
    setQueue(prev => [...prev, item]);
  }, []);

  const removeFromQueue = useCallback((index: number) => {
    setQueue(prev => prev.filter((_, i) => i !== index));
    if (index < queueIndex) {
      setQueueIndex(prev => prev - 1);
    } else if (index === queueIndex && queue.length > 1) {
      const nextIndex = Math.min(queueIndex, queue.length - 2);
      setQueueIndex(nextIndex);
      setCurrentItem(queue[nextIndex] || null);
    }
  }, [queue, queueIndex]);

  const clearQueue = useCallback(() => {
    setQueue([]);
    setQueueIndex(0);
    setCurrentItem(null);
    pause();
  }, [pause]);

  return {
    // State
    currentItem,
    isPlaying,
    isPaused,
    volume,
    isMuted,
    currentTime,
    duration,
    playbackRate,
    isFullscreen,
    repeatMode,
    isShuffled,
    queue,
    queueIndex,
    
    // Setters
    setCurrentItem,
    setIsPlaying,
    setIsPaused,
    setVolume,
    setIsMuted,
    setCurrentTime,
    setDuration,
    setPlaybackRate,
    setIsFullscreen,
    setRepeatMode,
    setIsShuffled,
    setQueue,
    setQueueIndex,
    
    // Actions
    play,
    pause,
    togglePlayPause,
    playNext,
    playPrevious,
    seekTo,
    skipTo,
    addToQueue,
    removeFromQueue,
    clearQueue,
  };
};
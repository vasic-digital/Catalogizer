import { describe, it, expect } from 'vitest';
import { PlaybackFormatter } from '../playbackFormatter';

describe('PlaybackFormatter', () => {
  it('formatAmount seconds → h/m/s', () => {
    expect(PlaybackFormatter.formatAmount(4980, 'seconds')).toBe('1h 23m');
    expect(PlaybackFormatter.formatAmount(1380, 'seconds')).toBe('23m');
    expect(PlaybackFormatter.formatAmount(45, 'seconds')).toBe('45s');
  });

  it('formatAmount pages singular/plural', () => {
    expect(PlaybackFormatter.formatAmount(1, 'pages')).toBe('1 page');
    expect(PlaybackFormatter.formatAmount(120, 'pages')).toBe('120 pages');
  });

  it('formatAmount events singular/plural', () => {
    expect(PlaybackFormatter.formatAmount(1, 'events')).toBe('1 run');
    expect(PlaybackFormatter.formatAmount(3, 'events')).toBe('3 runs');
  });

  it('formatAmount zero/negative → placeholder', () => {
    expect(PlaybackFormatter.formatAmount(0, 'seconds')).toBe('-');
    expect(PlaybackFormatter.formatAmount(-1, 'pages')).toBe('-');
  });

  it('formatProgress null/zero total falls back to current', () => {
    expect(PlaybackFormatter.formatProgress(1380, null, 'seconds')).toBe('23m');
    expect(PlaybackFormatter.formatProgress(1380, 0, 'seconds')).toBe('23m');
  });

  it('formatProgress renders slash form with total', () => {
    expect(PlaybackFormatter.formatProgress(1800, 7200, 'seconds')).toBe('30m / 2h 0m');
    expect(PlaybackFormatter.formatProgress(140, 320, 'pages')).toBe('140 pages / 320 pages');
  });

  it('progressFraction clamps 0..1', () => {
    expect(PlaybackFormatter.progressFraction(100, null)).toBe(0);
    expect(PlaybackFormatter.progressFraction(100, 0)).toBe(0);
    expect(PlaybackFormatter.progressFraction(50, 100)).toBe(0.5);
    expect(PlaybackFormatter.progressFraction(200, 100)).toBe(1);
  });

  it('formatReproductionCount × multiplier', () => {
    expect(PlaybackFormatter.formatReproductionCount(0)).toBe('0×');
    expect(PlaybackFormatter.formatReproductionCount(3)).toBe('3×');
  });
});

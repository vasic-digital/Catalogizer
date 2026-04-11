/**
 * Pure formatter mirroring the Kotlin / web versions so the desktop
 * ProgressBadge renders identical strings.
 */

export const PlaybackFormatter = {
  formatAmount(value: number, unit: string): string {
    if (value <= 0) return '-';
    switch (unit) {
      case 'seconds':
        return PlaybackFormatter.formatSeconds(value);
      case 'pages':
        return `${value} ${value === 1 ? 'page' : 'pages'}`;
      case 'events':
        return `${value} ${value === 1 ? 'run' : 'runs'}`;
      default:
        return `${value} ${unit}`;
    }
  },

  formatProgress(current: number, total: number | null, unit: string): string {
    if (total == null || total <= 0) {
      return PlaybackFormatter.formatAmount(current, unit);
    }
    return `${PlaybackFormatter.formatAmount(current, unit)} / ${PlaybackFormatter.formatAmount(total, unit)}`;
  },

  progressFraction(current: number, total: number | null): number {
    if (total == null || total <= 0) return 0;
    const pct = current / total;
    return Math.max(0, Math.min(1, pct));
  },

  formatSeconds(totalSeconds: number): string {
    if (totalSeconds <= 0) return '-';
    const h = Math.floor(totalSeconds / 3600);
    const m = Math.floor((totalSeconds % 3600) / 60);
    const s = totalSeconds % 60;
    const parts: string[] = [];
    if (h > 0) parts.push(`${h}h`);
    if (h > 0 || m > 0) parts.push(`${m}m`);
    if (h === 0 && m === 0) parts.push(`${s}s`);
    return parts.join(' ').trim();
  },

  formatReproductionCount(count: number): string {
    return `${count}×`;
  },
};

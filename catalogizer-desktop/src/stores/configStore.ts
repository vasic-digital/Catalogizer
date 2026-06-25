import { create } from 'zustand';
import { invoke } from '@tauri-apps/api/core';
import { AppConfig, Theme } from '../types';

interface ConfigState {
  serverUrl: string | null;
  theme: Theme;
  autoStart: boolean;
  isLoading: boolean;

  loadConfig: () => Promise<void>;
  setServerUrl: (url: string) => Promise<void>;
  setTheme: (theme: Theme) => Promise<void>;
  setAutoStart: (autoStart: boolean) => Promise<void>;
  resetConfig: () => Promise<void>;
}

const DARK_MEDIA_QUERY = '(prefers-color-scheme: dark)';

/** True when the OS reports a dark color-scheme preference. */
function systemPrefersDark(): boolean {
  return (
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia(DARK_MEDIA_QUERY).matches
  );
}

/**
 * Apply a theme to <html>. `'light'` removes `.dark`, `'dark'` adds it, and
 * `'system'` resolves the OS preference via matchMedia. Closes the gap where
 * `'system'` (a declared Theme value) previously fell through to light.
 */
function applyThemeToDocument(theme: Theme): void {
  if (typeof document === 'undefined') return;
  const root = document.documentElement;
  const isDark = theme === 'dark' || (theme === 'system' && systemPrefersDark());
  root.classList.toggle('dark', isDark);
}

let systemThemeListenerAttached = false;

/**
 * Keep the rendered theme in sync with OS preference changes while the user's
 * chosen theme is `'system'`. Idempotent — attaches the OS listener once.
 */
function ensureSystemThemeListener(getTheme: () => Theme): void {
  if (
    systemThemeListenerAttached ||
    typeof window === 'undefined' ||
    typeof window.matchMedia !== 'function'
  ) {
    return;
  }
  systemThemeListenerAttached = true;
  const mql = window.matchMedia(DARK_MEDIA_QUERY);
  const onChange = () => {
    if (getTheme() === 'system') {
      applyThemeToDocument('system');
    }
  };
  // addEventListener is the modern API; older WebViews expose addListener.
  if (typeof mql.addEventListener === 'function') {
    mql.addEventListener('change', onChange);
  } else if (typeof (mql as any).addListener === 'function') {
    (mql as any).addListener(onChange);
  }
}

export const useConfigStore = create<ConfigState>((set, get) => ({
  serverUrl: null,
  theme: 'dark',
  autoStart: false,
  isLoading: false,

  loadConfig: async () => {
    set({ isLoading: true });
    try {
      const config: AppConfig = await invoke('get_config');
      const theme = (config.theme as Theme) || 'dark';
      set({
        serverUrl: config.server_url || null,
        theme,
        autoStart: config.auto_start || false,
        isLoading: false,
      });

      // Apply theme to document (handles light/dark/system).
      applyThemeToDocument(theme);
      ensureSystemThemeListener(() => get().theme);
    } catch (error) {
      console.error('Failed to load config:', error);
      set({ isLoading: false });
    }
  },

  setServerUrl: async (url: string) => {
    try {
      await invoke('set_server_url', { url });
      set({ serverUrl: url });
    } catch (error) {
      console.error('Failed to set server URL:', error);
      throw error;
    }
  },

  setTheme: async (theme: Theme) => {
    try {
      const config: AppConfig = await invoke('get_config');
      const newConfig = { ...config, theme };
      await invoke('update_config', { newConfig });

      set({ theme });

      // Apply theme to document (handles light/dark/system).
      applyThemeToDocument(theme);
      ensureSystemThemeListener(() => get().theme);
    } catch (error) {
      console.error('Failed to set theme:', error);
      throw error;
    }
  },

  setAutoStart: async (autoStart: boolean) => {
    try {
      const config: AppConfig = await invoke('get_config');
      const newConfig = { ...config, auto_start: autoStart };
      await invoke('update_config', { newConfig });

      set({ autoStart });
    } catch (error) {
      console.error('Failed to set auto start:', error);
      throw error;
    }
  },

  resetConfig: async () => {
    try {
      const defaultConfig: AppConfig = {
        server_url: undefined,
        auth_token: undefined,
        theme: 'dark',
        auto_start: false,
      };

      await invoke('update_config', { newConfig: defaultConfig });

      set({
        serverUrl: null,
        theme: 'dark',
        autoStart: false,
      });

      // Reset theme (default 'dark').
      applyThemeToDocument('dark');
    } catch (error) {
      console.error('Failed to reset config:', error);
      throw error;
    }
  },
}));
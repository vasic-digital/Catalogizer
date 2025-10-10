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

export const useConfigStore = create<ConfigState>((set) => ({
  serverUrl: null,
  theme: 'dark',
  autoStart: false,
  isLoading: false,

  loadConfig: async () => {
    set({ isLoading: true });
    try {
      const config: AppConfig = await invoke('get_config');
      set({
        serverUrl: config.server_url || null,
        theme: (config.theme as Theme) || 'dark',
        autoStart: config.auto_start || false,
        isLoading: false,
      });

      // Apply theme to document
      const root = document.documentElement;
      if (config.theme === 'dark') {
        root.classList.add('dark');
      } else {
        root.classList.remove('dark');
      }
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

      // Apply theme to document
      const root = document.documentElement;
      if (theme === 'dark') {
        root.classList.add('dark');
      } else {
        root.classList.remove('dark');
      }
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

      // Reset theme
      const root = document.documentElement;
      root.classList.add('dark');
    } catch (error) {
      console.error('Failed to reset config:', error);
      throw error;
    }
  },
}));
import { useEffect, useState } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import { invoke } from "@tauri-apps/api/core";
import { useAuthStore } from "./stores/authStore";
import { useConfigStore } from "./stores/configStore";
import Layout from "./components/Layout";
import LoginPage from "./pages/LoginPage";
import HomePage from "./pages/HomePage";
import LibraryPage from "./pages/LibraryPage";
import SearchPage from "./pages/SearchPage";
import SettingsPage from "./pages/SettingsPage";
import MediaDetailPage from "./pages/MediaDetailPage";
import LoadingScreen from "./components/LoadingScreen";

function App() {
  const [isInitialized, setIsInitialized] = useState(false);
  const { isAuthenticated, setAuthToken } = useAuthStore();
  const { loadConfig, serverUrl } = useConfigStore();

  useEffect(() => {
    const initializeApp = async () => {
      try {
        // Load configuration
        await loadConfig();

        // Get stored auth token
        const config = await invoke<any>("get_config");
        if (config.auth_token) {
          setAuthToken(config.auth_token);
        }

        setIsInitialized(true);
      } catch (error) {
        console.error("Failed to initialize app:", error);
        setIsInitialized(true);
      }
    };

    initializeApp();
  }, [loadConfig, setAuthToken]);

  if (!isInitialized) {
    return <LoadingScreen />;
  }

  // If no server URL is configured, redirect to settings
  if (!serverUrl) {
    return (
      <Routes>
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="*" element={<Navigate to="/settings" replace />} />
      </Routes>
    );
  }

  // If not authenticated, show login
  if (!isAuthenticated) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    );
  }

  // Main app routes
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/library" element={<LibraryPage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/media/:id" element={<MediaDetailPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Layout>
  );
}

export default App;
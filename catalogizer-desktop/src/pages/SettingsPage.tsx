import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Save, TestTube, RefreshCw, ArrowLeft } from "lucide-react";
import { useConfigStore } from "../stores/configStore";
import { useAuthStore } from "../stores/authStore";
import { apiService } from "../services/apiService";
import { Theme } from "../types";

export default function SettingsPage() {
  const navigate = useNavigate();
  const { serverUrl, theme, autoStart, setServerUrl, setTheme, setAutoStart } = useConfigStore();
  const { isAuthenticated } = useAuthStore();

  const [localServerUrl, setLocalServerUrl] = useState(serverUrl || "");
  const [localTheme, setLocalTheme] = useState<Theme>(theme);
  const [localAutoStart, setLocalAutoStart] = useState(autoStart);
  const [isTestingConnection, setIsTestingConnection] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [testResult, setTestResult] = useState<{
    success: boolean;
    message: string;
  } | null>(null);

  const handleTestConnection = async () => {
    if (!localServerUrl) {
      setTestResult({
        success: false,
        message: "Please enter a server URL",
      });
      return;
    }

    setIsTestingConnection(true);
    setTestResult(null);

    try {
      // Temporarily set server URL to test
      await setServerUrl(localServerUrl);

      // Test connection
      const result = await apiService.healthCheck();

      setTestResult({
        success: true,
        message: `Connected successfully! Server status: ${result.status}`,
      });
    } catch (error) {
      setTestResult({
        success: false,
        message: error instanceof Error ? error.message : "Connection failed",
      });
    } finally {
      setIsTestingConnection(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);

    try {
      await Promise.all([
        setServerUrl(localServerUrl),
        setTheme(localTheme),
        setAutoStart(localAutoStart),
      ]);

      // Show success message or navigate back
      if (isAuthenticated) {
        navigate(-1);
      } else {
        navigate("/login");
      }
    } catch (error) {
      console.error("Failed to save settings:", error);
    } finally {
      setIsSaving(false);
    }
  };

  const canGoBack = isAuthenticated && serverUrl;

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-2xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          {canGoBack && (
            <button
              onClick={() => navigate(-1)}
              className="p-2 hover:bg-accent rounded-md transition-colors"
            >
              <ArrowLeft className="h-5 w-5" />
            </button>
          )}
          <div>
            <h1 className="text-3xl font-bold text-foreground">Settings</h1>
            <p className="text-muted-foreground">
              Configure your Catalogizer desktop client
            </p>
          </div>
        </div>

        <div className="space-y-8">
          {/* Server Configuration */}
          <section className="bg-card border border-border rounded-lg p-6">
            <h2 className="text-xl font-semibold text-foreground mb-4">
              Server Configuration
            </h2>

            <div className="space-y-4">
              <div>
                <label
                  htmlFor="serverUrl"
                  className="block text-sm font-medium text-foreground mb-2"
                >
                  Server URL
                </label>
                <div className="flex gap-2">
                  <input
                    id="serverUrl"
                    type="url"
                    value={localServerUrl}
                    onChange={(e) => setLocalServerUrl(e.target.value)}
                    placeholder="http://localhost:8080"
                    className="flex-1 px-3 py-2 border border-input bg-background rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                  />
                  <button
                    onClick={handleTestConnection}
                    disabled={isTestingConnection || !localServerUrl}
                    className="px-4 py-2 bg-secondary text-secondary-foreground rounded-md hover:bg-secondary/80 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                  >
                    {isTestingConnection ? (
                      <RefreshCw className="h-4 w-4 animate-spin" />
                    ) : (
                      <TestTube className="h-4 w-4" />
                    )}
                    Test
                  </button>
                </div>

                {testResult && (
                  <div
                    className={`mt-2 p-3 rounded-md text-sm ${
                      testResult.success
                        ? "bg-green-50 border border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800 dark:text-green-300"
                        : "bg-red-50 border border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-800 dark:text-red-300"
                    }`}
                  >
                    {testResult.message}
                  </div>
                )}
              </div>
            </div>
          </section>

          {/* Appearance */}
          <section className="bg-card border border-border rounded-lg p-6">
            <h2 className="text-xl font-semibold text-foreground mb-4">
              Appearance
            </h2>

            <div className="space-y-4">
              <div>
                <label
                  htmlFor="theme"
                  className="block text-sm font-medium text-foreground mb-2"
                >
                  Theme
                </label>
                <select
                  id="theme"
                  value={localTheme}
                  onChange={(e) => setLocalTheme(e.target.value as Theme)}
                  className="w-full px-3 py-2 border border-input bg-background rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                >
                  <option value="light">Light</option>
                  <option value="dark">Dark</option>
                  <option value="system">System</option>
                </select>
              </div>
            </div>
          </section>

          {/* Storage Configuration */}
           <section className="bg-card border border-border rounded-lg p-6">
             <h2 className="text-xl font-semibold text-foreground mb-4">
               Storage Configuration
             </h2>

             <div className="space-y-4">
               <div>
                 <p className="text-sm text-muted-foreground mb-4">
                   Configure storage sources for media scanning. Supported protocols: SMB, FTP, NFS, WebDAV, Local.
                 </p>
                 <button
                   className="px-4 py-2 bg-secondary text-secondary-foreground rounded-md hover:bg-secondary/80"
                   onClick={() => alert('Storage configuration coming soon')}
                 >
                   Configure Storage Sources
                 </button>
               </div>
             </div>
           </section>

           {/* General */}
           <section className="bg-card border border-border rounded-lg p-6">
             <h2 className="text-xl font-semibold text-foreground mb-4">
               General
             </h2>

             <div className="space-y-4">
               <div className="flex items-center justify-between">
                 <div>
                   <h3 className="font-medium text-foreground">Auto-start</h3>
                   <p className="text-sm text-muted-foreground">
                     Start Catalogizer when your computer starts
                   </p>
                 </div>
                 <label className="relative inline-flex items-center cursor-pointer">
                   <input
                     type="checkbox"
                     checked={localAutoStart}
                     onChange={(e) => setLocalAutoStart(e.target.checked)}
                     className="sr-only peer"
                   />
                   <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                 </label>
               </div>
             </div>
           </section>

          {/* Save Button */}
          <div className="flex justify-end">
            <button
              onClick={handleSave}
              disabled={isSaving}
              className="px-6 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {isSaving ? (
                <RefreshCw className="h-4 w-4 animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save Settings
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
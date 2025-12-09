import React from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from '@/contexts/AuthContext'
import { WebSocketProvider } from '@/contexts/WebSocketContext'
import { ConnectionStatus } from '@/components/ui/ConnectionStatus'
import { Layout } from '@/components/layout/Layout'
import { LoginForm } from '@/components/auth/LoginForm'
import { RegisterForm } from '@/components/auth/RegisterForm'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { Dashboard } from '@/pages/Dashboard'
import { MediaBrowser } from '@/pages/MediaBrowser'
import { Analytics } from '@/pages/Analytics'
import { SubtitleManager } from '@/pages/SubtitleManager'

function App() {
  return (
    <AuthProvider>
      <WebSocketProvider>
        <Router>
          <ConnectionStatus />
          <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginForm />} />
          <Route path="/register" element={<RegisterForm />} />

          {/* Protected routes */}
          <Route path="/" element={<Layout />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <Dashboard />
                </ProtectedRoute>
              }
            />
            <Route
              path="/media"
              element={
                <ProtectedRoute requiredPermission="read:media">
                  <MediaBrowser />
                </ProtectedRoute>
              }
            />
            <Route
              path="/analytics"
              element={
                <ProtectedRoute requiredPermission="view:analysis">
                  <Analytics />
                </ProtectedRoute>
              }
            />
            <Route
              path="/subtitles"
              element={
                <ProtectedRoute requiredPermission="manage:subtitles">
                  <SubtitleManager />
                </ProtectedRoute>
              }
            />
             <Route
               path="/admin"
               element={
                 <ProtectedRoute requireAdmin>
                   <div className="p-8">
                     <h1 className="text-2xl font-bold">Admin Panel</h1>
                     <p className="text-gray-600 mt-2">Storage configuration and system management</p>
                     <div className="mt-6 space-y-4">
                       <div className="bg-white p-4 rounded-lg border">
                         <h2 className="text-lg font-semibold">Storage Roots</h2>
                         <p className="text-sm text-gray-600">Configure storage sources for media scanning</p>
                         <button className="mt-2 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                           Manage Storage
                         </button>
                       </div>
                       <div className="bg-white p-4 rounded-lg border">
                         <h2 className="text-lg font-semibold">System Settings</h2>
                         <p className="text-sm text-gray-600">Configure server and application settings</p>
                         <button className="mt-2 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                           System Config
                         </button>
                       </div>
                     </div>
                   </div>
                 </ProtectedRoute>
               }
             />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <div className="p-8">
                    <h1 className="text-2xl font-bold">Profile</h1>
                    <p className="text-gray-600 mt-2">Coming soon...</p>
                  </div>
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings"
              element={
                <ProtectedRoute>
                  <div className="p-8">
                    <h1 className="text-2xl font-bold">Settings</h1>
                    <p className="text-gray-600 mt-2">Coming soon...</p>
                  </div>
                </ProtectedRoute>
              }
            />
          </Route>

          {/* Catch all route */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </Router>
      </WebSocketProvider>
    </AuthProvider>
  )
}

export default App
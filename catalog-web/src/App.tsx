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
import { Collections } from '@/pages/Collections'
import { ConversionTools } from '@/pages/ConversionTools'
import { Admin } from '@/pages/Admin'
import FavoritesPage from '@/pages/Favorites'
import { PlaylistsPage } from '@/pages/Playlists'
import AIDashboard from '@/pages/AIDashboard'

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
              path="/collections"
              element={
                <ProtectedRoute requiredPermission="read:collections">
                  <Collections />
                </ProtectedRoute>
              }
            />
            <Route
              path="/favorites"
              element={
                <ProtectedRoute>
                  <FavoritesPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/playlists"
              element={
                <ProtectedRoute>
                  <PlaylistsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/conversion"
              element={
                <ProtectedRoute requiredPermission="convert:media">
                  <ConversionTools />
                </ProtectedRoute>
              }
            />
             <Route
               path="/admin"
               element={
                 <ProtectedRoute requireAdmin>
                   <Admin />
                 </ProtectedRoute>
               }
             />
            <Route
              path="/ai"
              element={
                <ProtectedRoute>
                  <AIDashboard />
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
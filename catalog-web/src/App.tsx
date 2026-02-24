import React, { Suspense } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from '@/contexts/AuthContext'
import { WebSocketProvider } from '@/contexts/WebSocketContext'
import { ConnectionStatus } from '@/components/ui/ConnectionStatus'
import { Layout } from '@/components/layout/Layout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { ErrorBoundary } from '@/components/ErrorBoundary'

// Lazy-loaded page components for code splitting
const LoginForm = React.lazy(() => import('@/components/auth/LoginForm').then(m => ({ default: m.LoginForm })))
const RegisterForm = React.lazy(() => import('@/components/auth/RegisterForm').then(m => ({ default: m.RegisterForm })))
const Dashboard = React.lazy(() => import('@/pages/Dashboard').then(m => ({ default: m.Dashboard })))
const MediaBrowser = React.lazy(() => import('@/pages/MediaBrowser').then(m => ({ default: m.MediaBrowser })))
const Analytics = React.lazy(() => import('@/pages/Analytics').then(m => ({ default: m.Analytics })))
const SubtitleManager = React.lazy(() => import('@/pages/SubtitleManager').then(m => ({ default: m.SubtitleManager })))
const Collections = React.lazy(() => import('@/pages/Collections').then(m => ({ default: m.Collections })))
const ConversionTools = React.lazy(() => import('@/pages/ConversionTools').then(m => ({ default: m.ConversionTools })))
const Admin = React.lazy(() => import('@/pages/Admin').then(m => ({ default: m.Admin })))
const FavoritesPage = React.lazy(() => import('@/pages/Favorites'))
const PlaylistsPage = React.lazy(() => import('@/pages/Playlists').then(m => ({ default: m.PlaylistsPage })))
const AIDashboard = React.lazy(() => import('@/pages/AIDashboard'))
const EntityBrowser = React.lazy(() => import('@/pages/EntityBrowser').then(m => ({ default: m.EntityBrowser })))
const EntityDetail = React.lazy(() => import('@/pages/EntityDetail').then(m => ({ default: m.EntityDetail })))

const PageLoader: React.FC = () => (
  <div className="p-6 space-y-4 animate-pulse min-h-[400px]">
    <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-1/3" />
    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-2/3" />
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6">
      <div className="h-32 bg-gray-200 dark:bg-gray-700 rounded" />
      <div className="h-32 bg-gray-200 dark:bg-gray-700 rounded" />
      <div className="h-32 bg-gray-200 dark:bg-gray-700 rounded" />
    </div>
    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mt-4" />
    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4" />
  </div>
)

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <WebSocketProvider>
          <Router future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
            <ConnectionStatus />
            <Suspense fallback={<PageLoader />}>
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
                  path="/browse"
                  element={
                    <ProtectedRoute requiredPermission="read:media">
                      <EntityBrowser />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/entity/:id"
                  element={
                    <ProtectedRoute requiredPermission="read:media">
                      <EntityDetail />
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
            </Suspense>
          </Router>
        </WebSocketProvider>
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App

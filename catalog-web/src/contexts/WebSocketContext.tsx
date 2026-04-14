import React, { createContext, useContext, useEffect, useRef, ReactNode } from 'react'
import { useAuth } from './AuthContext'
import { useWebSocket } from '@/lib/websocket'

interface WebSocketContextType {
  connect: () => void
  disconnect: () => void
  send: (message: unknown) => void
  subscribe: (channel: string) => void
  unsubscribe: (channel: string) => void
  getConnectionState: () => 'connecting' | 'open' | 'closing' | 'closed'
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined)

/**
 * useWebSocketContext returns the WebSocket connection controls from the
 * nearest WebSocketProvider. Throws if used outside the provider.
 *
 * @returns WebSocket connect, disconnect, send, subscribe, and unsubscribe actions
 */
export const useWebSocketContext = () => {
  const context = useContext(WebSocketContext)
  if (context === undefined) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider')
  }
  return context
}

interface WebSocketProviderProps {
  children: ReactNode
}

/**
 * WebSocketProvider automatically connects the WebSocket when the user is
 * authenticated and disconnects on logout or unmount.
 */
export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const { isAuthenticated, user } = useAuth()
  const webSocket = useWebSocket()
  const webSocketRef = useRef(webSocket)
  webSocketRef.current = webSocket

  useEffect(() => {
    if (isAuthenticated && user) {
      webSocketRef.current.connect()

      return () => {
        webSocketRef.current.disconnect()
      }
    } else {
      webSocketRef.current.disconnect()
    }
  }, [isAuthenticated, user])

  return (
    <WebSocketContext.Provider value={webSocket}>
      {children}
    </WebSocketContext.Provider>
  )
}

export default WebSocketContext
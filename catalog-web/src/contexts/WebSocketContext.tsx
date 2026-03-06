import React, { createContext, useContext, useEffect, ReactNode } from 'react'
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

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const { isAuthenticated, user } = useAuth()
  const webSocket = useWebSocket()

  useEffect(() => {
    if (isAuthenticated && user) {
      // Connect to WebSocket when user is authenticated
      webSocket.connect()

      return () => {
        // Disconnect when component unmounts or user logs out
        webSocket.disconnect()
      }
    } else {
      // Disconnect if user is not authenticated
      webSocket.disconnect()
    }
  }, [isAuthenticated, user, webSocket])

  return (
    <WebSocketContext.Provider value={webSocket}>
      {children}
    </WebSocketContext.Provider>
  )
}

export default WebSocketContext
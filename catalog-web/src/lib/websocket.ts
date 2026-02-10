import { useEffect, useRef, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'

export interface WebSocketMessage {
  type: string
  data: any
  timestamp: string
}

export interface MediaUpdate {
  action: 'created' | 'updated' | 'deleted' | 'analyzed'
  media_id: number
  media: any
  analysis_id?: string
}

export interface SystemUpdate {
  action: 'health_check' | 'storage_update' | 'service_status'
  component: string
  status: 'healthy' | 'warning' | 'error'
  message?: string
  details?: any
}

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'

export class WebSocketClient {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private messageQueue: string[] = []
  private isConnected = false
  private token: string | null = null

  private onMessage: ((message: WebSocketMessage) => void) | null = null
  private onConnect: (() => void) | null = null
  private onDisconnect: (() => void) | null = null
  private onError: ((error: Event) => void) | null = null

  constructor(token?: string) {
    this.token = token || localStorage.getItem('auth_token')
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    const wsUrl = this.token ? `${WS_URL}?token=${this.token}` : WS_URL

    try {
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        this.isConnected = true
        this.reconnectAttempts = 0

        // Send queued messages
        while (this.messageQueue.length > 0) {
          const message = this.messageQueue.shift()
          if (message && this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(message)
          }
        }

        this.onConnect?.()
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          this.onMessage?.(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      this.ws.onclose = (event) => {
        this.isConnected = false
        this.onDisconnect?.()

        // Attempt to reconnect if not closed intentionally
        if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.scheduleReconnect()
        }
      }

      this.ws.onerror = (event) => {
        console.error('WebSocket error:', event)
        this.onError?.(event)
      }
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      this.scheduleReconnect()
    }
  }

  private scheduleReconnect() {
    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)

    setTimeout(() => {
      this.connect()
    }, delay)
  }

  send(message: any) {
    const messageStr = JSON.stringify(message)

    if (this.isConnected && this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(messageStr)
    } else {
      this.messageQueue.push(messageStr)
    }
  }

  subscribe(channel: string) {
    this.send({
      type: 'subscribe',
      channel
    })
  }

  unsubscribe(channel: string) {
    this.send({
      type: 'unsubscribe',
      channel
    })
  }

  setOnMessage(callback: (message: WebSocketMessage) => void) {
    this.onMessage = callback
  }

  setOnConnect(callback: () => void) {
    this.onConnect = callback
  }

  setOnDisconnect(callback: () => void) {
    this.onDisconnect = callback
  }

  setOnError(callback: (error: Event) => void) {
    this.onError = callback
  }

  disconnect() {
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect')
    }
  }

  getConnectionState(): 'connecting' | 'open' | 'closing' | 'closed' {
    if (!this.ws) return 'closed'

    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'connecting'
      case WebSocket.OPEN:
        return 'open'
      case WebSocket.CLOSING:
        return 'closing'
      case WebSocket.CLOSED:
        return 'closed'
      default:
        return 'closed'
    }
  }
}

// React hook for WebSocket connection
export const useWebSocket = () => {
  const wsRef = useRef<WebSocketClient | null>(null)
  const queryClient = useQueryClient()

  const connect = useCallback(() => {
    if (!wsRef.current) {
      const token = localStorage.getItem('auth_token') || undefined
      wsRef.current = new WebSocketClient(token)
    }

    wsRef.current.setOnMessage((message: WebSocketMessage) => {
      switch (message.type) {
        case 'media_update':
          handleMediaUpdate(message.data as MediaUpdate, queryClient)
          break
        case 'system_update':
          handleSystemUpdate(message.data as SystemUpdate)
          break
        case 'analysis_complete':
          handleAnalysisComplete(message.data, queryClient)
          break
        case 'notification':
          handleNotification(message.data)
          break
      }
    })

    wsRef.current.setOnConnect(() => {
      toast.success('Connected to real-time updates')

      // Subscribe to relevant channels
      wsRef.current?.subscribe('media_updates')
      wsRef.current?.subscribe('system_updates')
      wsRef.current?.subscribe('analysis_updates')
    })

    wsRef.current.setOnDisconnect(() => {
      toast.error('Disconnected from real-time updates')
    })

    wsRef.current.setOnError((error) => {
      console.error('WebSocket error:', error)
      toast.error('Connection error - some features may not work')
    })

    wsRef.current.connect()
  }, [queryClient])

  const disconnect = useCallback(() => {
    wsRef.current?.disconnect()
    wsRef.current = null
  }, [])

  const send = useCallback((message: any) => {
    wsRef.current?.send(message)
  }, [])

  const subscribe = useCallback((channel: string) => {
    wsRef.current?.subscribe(channel)
  }, [])

  const unsubscribe = useCallback((channel: string) => {
    wsRef.current?.unsubscribe(channel)
  }, [])

  const getConnectionState = useCallback(() => {
    return wsRef.current?.getConnectionState() || 'closed'
  }, [])

  useEffect(() => {
    return () => {
      disconnect()
    }
  }, [disconnect])

  return {
    connect,
    disconnect,
    send,
    subscribe,
    unsubscribe,
    getConnectionState
  }
}

// Message handlers
const handleMediaUpdate = (update: MediaUpdate, queryClient: any) => {
  switch (update.action) {
    case 'created':
      toast.success(`New ${update.media.media_type} added: ${update.media.title}`)
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: ['media-search'] })
      queryClient.invalidateQueries({ queryKey: ['media-stats'] })
      queryClient.invalidateQueries({ queryKey: ['recent-media'] })
      break

    case 'updated':
      toast.success(`${update.media.title} updated`)
      queryClient.invalidateQueries({ queryKey: ['media-search'] })
      queryClient.setQueryData(['media', update.media_id], update.media)
      break

    case 'deleted':
      toast.success(`Media item deleted`)
      queryClient.invalidateQueries({ queryKey: ['media-search'] })
      queryClient.invalidateQueries({ queryKey: ['media-stats'] })
      queryClient.removeQueries({ queryKey: ['media', update.media_id] })
      break

    case 'analyzed':
      toast.success(`Analysis complete for ${update.media.title}`)
      queryClient.invalidateQueries({ queryKey: ['media', update.media_id] })
      break
  }
}

const handleSystemUpdate = (update: SystemUpdate) => {
  const { component, status, message } = update

  switch (status) {
    case 'healthy':
      break
    case 'warning':
      toast(`${component} warning: ${message}`, {
        icon: '⚠️',
        duration: 6000,
      })
      break
    case 'error':
      toast.error(`${component} error: ${message}`)
      break
  }
}

const handleAnalysisComplete = (data: any, queryClient: any) => {
  toast.success(`Analysis complete for ${data.items_processed} items`)
  queryClient.invalidateQueries({ queryKey: ['media-search'] })
  queryClient.invalidateQueries({ queryKey: ['media-stats'] })
}

const handleNotification = (notification: any) => {
  const { type, title, message, level } = notification

  switch (level) {
    case 'info':
      toast(message, { icon: 'ℹ️' })
      break
    case 'success':
      toast.success(message)
      break
    case 'warning':
      toast(message, { icon: '⚠️' })
      break
    case 'error':
      toast.error(message)
      break
    default:
      toast(message)
  }
}

export default WebSocketClient
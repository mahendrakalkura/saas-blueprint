import { createContext, useContext, useEffect, useRef, useState } from 'react'
import { useAuth } from '../contexts/AuthContext'

const WebSocketContext = createContext(null)

export function WebSocketProvider({ children }) {
  const { user, token } = useAuth()
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState(null)
  const wsRef = useRef(null)
  const reconnectTimeoutRef = useRef(null)
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5
  const listenersRef = useRef(new Map())

  const connect = () => {
    if (!token || !user) {
      console.log('WebSocket: No token or user, skipping connection')
      return
    }

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      console.log('WebSocket: Already connected')
      return
    }

    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/v1/ws`

      console.log('WebSocket: Connecting to', wsUrl)

      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        console.log('WebSocket: Connected')
        setIsConnected(true)
        reconnectAttempts.current = 0
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('WebSocket: Message received', data)

          setLastMessage(data)

          // Notify listeners for this message type
          const listeners = listenersRef.current.get(data.type) || []
          listeners.forEach((callback) => callback(data.payload))

          // Notify global listeners (type: '*')
          const globalListeners = listenersRef.current.get('*') || []
          globalListeners.forEach((callback) => callback(data))
        } catch (error) {
          console.error('WebSocket: Failed to parse message', error)
        }
      }

      ws.onerror = (error) => {
        console.error('WebSocket: Error', error)
      }

      ws.onclose = () => {
        console.log('WebSocket: Disconnected')
        setIsConnected(false)
        wsRef.current = null

        // Attempt to reconnect with exponential backoff
        if (
          token &&
          user &&
          reconnectAttempts.current < maxReconnectAttempts
        ) {
          const delay = Math.min(1000 * 2 ** reconnectAttempts.current, 30000)
          console.log(
            `WebSocket: Reconnecting in ${delay}ms (attempt ${reconnectAttempts.current + 1}/${maxReconnectAttempts})`
          )

          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttempts.current++
            connect()
          }, delay)
        }
      }
    } catch (error) {
      console.error('WebSocket: Failed to connect', error)
    }
  }

  const disconnect = () => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    if (wsRef.current) {
      console.log('WebSocket: Disconnecting')
      wsRef.current.close()
      wsRef.current = null
    }

    setIsConnected(false)
    reconnectAttempts.current = 0
  }

  const subscribe = (messageType, callback) => {
    if (!listenersRef.current.has(messageType)) {
      listenersRef.current.set(messageType, [])
    }

    const listeners = listenersRef.current.get(messageType)
    listeners.push(callback)

    // Return unsubscribe function
    return () => {
      const index = listeners.indexOf(callback)
      if (index > -1) {
        listeners.splice(index, 1)
      }
    }
  }

  // Connect when user logs in, disconnect when user logs out
  useEffect(() => {
    if (user && token) {
      connect()
    } else {
      disconnect()
    }

    return () => {
      disconnect()
    }
  }, [user, token])

  const value = {
    isConnected,
    lastMessage,
    subscribe,
    connect,
    disconnect,
  }

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocket() {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider')
  }
  return context
}

// Hook to subscribe to specific message types
export function useWebSocketSubscription(messageType, callback) {
  const { subscribe } = useWebSocket()

  useEffect(() => {
    if (!messageType || !callback) return

    const unsubscribe = subscribe(messageType, callback)
    return unsubscribe
  }, [messageType, callback, subscribe])
}

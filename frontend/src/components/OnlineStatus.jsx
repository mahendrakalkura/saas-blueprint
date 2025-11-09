import { useState, useEffect } from 'react'
import { WifiOff, Wifi } from 'lucide-react'
import { isOnline, setupOnlineListeners } from '../utils/pwa'

export default function OnlineStatus() {
  const [online, setOnline] = useState(isOnline())
  const [showStatus, setShowStatus] = useState(false)
  const [justWentOnline, setJustWentOnline] = useState(false)

  useEffect(() => {
    const handleOnline = () => {
      setOnline(true)
      setShowStatus(true)
      setJustWentOnline(true)

      // Hide the "back online" message after 3 seconds
      setTimeout(() => {
        setShowStatus(false)
        setJustWentOnline(false)
      }, 3000)
    }

    const handleOffline = () => {
      setOnline(false)
      setShowStatus(true)
      setJustWentOnline(false)
    }

    // Setup listeners
    const cleanup = setupOnlineListeners(handleOnline, handleOffline)

    return cleanup
  }, [])

  // Don't show anything if online and not just came back online
  if (online && !showStatus) {
    return null
  }

  // Show offline indicator persistently when offline
  if (!online) {
    return (
      <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50 animate-slide-down">
        <div className="bg-red-500 text-white px-4 py-2 rounded-lg shadow-lg flex items-center gap-2">
          <WifiOff className="w-4 h-4" />
          <span className="text-sm font-medium">You're offline</span>
        </div>
      </div>
    )
  }

  // Show brief "back online" message
  if (justWentOnline) {
    return (
      <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50 animate-slide-down">
        <div className="bg-green-500 text-white px-4 py-2 rounded-lg shadow-lg flex items-center gap-2">
          <Wifi className="w-4 h-4" />
          <span className="text-sm font-medium">Back online</span>
        </div>
      </div>
    )
  }

  return null
}

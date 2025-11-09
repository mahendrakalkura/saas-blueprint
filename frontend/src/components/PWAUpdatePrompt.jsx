import { useState, useEffect } from 'react'
import { RefreshCw } from 'lucide-react'

export default function PWAUpdatePrompt() {
  const [showUpdate, setShowUpdate] = useState(false)
  const [registration, setRegistration] = useState(null)
  const [updating, setUpdating] = useState(false)

  useEffect(() => {
    const handleUpdate = (event) => {
      setRegistration(event.detail.registration)
      setShowUpdate(true)
    }

    window.addEventListener('sw-update', handleUpdate)

    return () => {
      window.removeEventListener('sw-update', handleUpdate)
    }
  }, [])

  const handleUpdate = () => {
    if (!registration || !registration.waiting) {
      return
    }

    setUpdating(true)

    // Tell the service worker to skip waiting
    registration.waiting.postMessage({ type: 'SKIP_WAITING' })

    // Reload the page when the new service worker takes control
    let refreshing = false
    navigator.serviceWorker.addEventListener('controllerchange', () => {
      if (!refreshing) {
        refreshing = true
        window.location.reload()
      }
    })
  }

  const handleDismiss = () => {
    setShowUpdate(false)
  }

  if (!showUpdate) {
    return null
  }

  return (
    <div className="fixed bottom-4 left-1/2 transform -translate-x-1/2 z-50 animate-slide-up">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-2xl p-4 border border-gray-200 dark:border-gray-700 max-w-md">
        <div className="flex items-start gap-3">
          <div className="flex-shrink-0 w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
            <RefreshCw className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>

          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-1">
              Update Available
            </h3>
            <p className="text-xs text-gray-600 dark:text-gray-400 mb-3">
              A new version of the app is available. Update now to get the latest features and improvements.
            </p>

            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                disabled={updating}
                className="flex-1 px-3 py-2 text-xs font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-1"
              >
                {updating ? (
                  <>
                    <RefreshCw className="w-3 h-3 animate-spin" />
                    Updating...
                  </>
                ) : (
                  <>
                    <RefreshCw className="w-3 h-3" />
                    Update Now
                  </>
                )}
              </button>
              <button
                onClick={handleDismiss}
                disabled={updating}
                className="px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                Later
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

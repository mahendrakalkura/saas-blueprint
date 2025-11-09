import { useState, useEffect } from 'react'
import { X, Download } from 'lucide-react'
import { setupInstallPrompt, isInstalled } from '../utils/pwa'

export default function PWAInstallPrompt() {
  const [showPrompt, setShowPrompt] = useState(false)
  const [installPrompt, setInstallPrompt] = useState(null)

  useEffect(() => {
    // Don't show if already installed
    if (isInstalled()) {
      return
    }

    // Setup install prompt handler
    const prompt = setupInstallPrompt()
    setInstallPrompt(prompt)

    // Listen for installable event
    const handleInstallable = () => {
      // Check if user has dismissed the prompt before
      const dismissed = localStorage.getItem('pwa-install-dismissed')
      const dismissedTime = dismissed ? parseInt(dismissed, 10) : 0
      const now = Date.now()
      const dayInMs = 24 * 60 * 60 * 1000

      // Show prompt if not dismissed or dismissed more than 7 days ago
      if (!dismissed || now - dismissedTime > 7 * dayInMs) {
        setShowPrompt(true)
      }
    }

    window.addEventListener('pwa-installable', handleInstallable)

    // Listen for installed event
    const handleInstalled = () => {
      setShowPrompt(false)
      localStorage.removeItem('pwa-install-dismissed')
    }

    window.addEventListener('pwa-installed', handleInstalled)

    return () => {
      window.removeEventListener('pwa-installable', handleInstallable)
      window.removeEventListener('pwa-installed', handleInstalled)
    }
  }, [])

  const handleInstall = async () => {
    if (!installPrompt) return

    const accepted = await installPrompt.showInstallPrompt()

    if (accepted) {
      setShowPrompt(false)
    }
  }

  const handleDismiss = () => {
    setShowPrompt(false)
    localStorage.setItem('pwa-install-dismissed', Date.now().toString())
  }

  if (!showPrompt) {
    return null
  }

  return (
    <div className="fixed bottom-4 left-4 right-4 md:left-auto md:right-4 md:max-w-md z-50 animate-slide-up">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-2xl p-4 border border-gray-200 dark:border-gray-700">
        <div className="flex items-start gap-3">
          <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <Download className="w-6 h-6 text-white" />
          </div>

          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-1">
              Install SaaS Blueprint
            </h3>
            <p className="text-xs text-gray-600 dark:text-gray-400 mb-3">
              Install our app for a better experience with offline access and push notifications.
            </p>

            <div className="flex gap-2">
              <button
                onClick={handleInstall}
                className="flex-1 px-3 py-2 text-xs font-medium text-white bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg hover:from-blue-600 hover:to-purple-700 transition-all duration-200"
              >
                Install Now
              </button>
              <button
                onClick={handleDismiss}
                className="px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
              >
                Later
              </button>
            </div>
          </div>

          <button
            onClick={handleDismiss}
            className="flex-shrink-0 p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            aria-label="Close"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}

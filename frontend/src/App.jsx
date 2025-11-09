import { BrowserRouter } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import { ChakraProvider, QueryProvider } from './providers'
import { AuthProvider } from './contexts/AuthContext'
import { WebSocketProvider } from './lib/websocket'
import { AppRoutes } from './routes'
import PWAInstallPrompt from './components/PWAInstallPrompt'
import PWAUpdatePrompt from './components/PWAUpdatePrompt'
import OnlineStatus from './components/OnlineStatus'

export function App() {
  return (
    <ErrorBoundary>
      <QueryProvider>
        <ChakraProvider>
          <BrowserRouter>
            <AuthProvider>
              <WebSocketProvider>
                <AppRoutes />
                <PWAInstallPrompt />
                <PWAUpdatePrompt />
                <OnlineStatus />
              </WebSocketProvider>
            </AuthProvider>
          </BrowserRouter>
        </ChakraProvider>
      </QueryProvider>
    </ErrorBoundary>
  )
}

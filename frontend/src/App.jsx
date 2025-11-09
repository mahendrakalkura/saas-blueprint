import { BrowserRouter } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import { ChakraProvider, QueryProvider } from './providers'
import { AuthProvider } from './contexts/AuthContext'
import { AppRoutes } from './routes'

export function App() {
  return (
    <ErrorBoundary>
      <QueryProvider>
        <ChakraProvider>
          <BrowserRouter>
            <AuthProvider>
              <AppRoutes />
            </AuthProvider>
          </BrowserRouter>
        </ChakraProvider>
      </QueryProvider>
    </ErrorBoundary>
  )
}

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { ChakraProvider, QueryProvider } from './providers'
import { AppRoutes } from './routes'
import { ErrorBoundary } from './components/ErrorBoundary'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <ErrorBoundary>
      <QueryProvider>
        <ChakraProvider>
          <AppRoutes />
        </ChakraProvider>
      </QueryProvider>
    </ErrorBoundary>
  </StrictMode>,
)

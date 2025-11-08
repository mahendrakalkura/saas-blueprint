import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { HomePage } from '../pages/HomePage'
import { ErrorBoundary } from '../components/ErrorBoundary'

const router = createBrowserRouter([
  {
    path: '/',
    element: <HomePage />,
    errorElement: <ErrorBoundary />,
  },
  // Add more routes here
  // {
  //   path: '/login',
  //   element: <LoginPage />,
  // },
  // {
  //   path: '/register',
  //   element: <RegisterPage />,
  // },
  // {
  //   path: '/dashboard',
  //   element: <DashboardPage />,
  // },
])

export function AppRoutes() {
  return <RouterProvider router={router} />
}

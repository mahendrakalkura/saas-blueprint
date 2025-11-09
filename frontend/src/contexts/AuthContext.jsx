import { createContext, useContext, useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  // Check if user is logged in on mount
  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('access_token')
      if (token) {
        try {
          const userData = await api.get('/auth/me')
          setUser(userData)
        } catch (error) {
          // Token is invalid, clear it
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
        }
      }
      setLoading(false)
    }

    checkAuth()
  }, [])

  const login = async (email, password) => {
    const response = await api.post('/auth/login', { email, password })

    // Check if MFA is required
    if (response.mfa_required) {
      return response // Return MFA token for verification
    }

    // Normal login flow
    localStorage.setItem('access_token', response.access_token)
    localStorage.setItem('refresh_token', response.refresh_token)
    setUser(response.user)

    return response
  }

  const verifyMFA = async (mfaToken, code) => {
    const response = await api.post('/auth/mfa/verify', {
      mfa_token: mfaToken,
      code,
    })

    localStorage.setItem('access_token', response.access_token)
    localStorage.setItem('refresh_token', response.refresh_token)
    setUser(response.user)

    return response
  }

  const register = async (email, password, firstName, lastName) => {
    const response = await api.post('/auth/register', {
      email,
      password,
      first_name: firstName,
      last_name: lastName,
    })

    localStorage.setItem('access_token', response.access_token)
    localStorage.setItem('refresh_token', response.refresh_token)
    setUser(response.user)

    return response
  }

  const logout = async () => {
    const refreshToken = localStorage.getItem('refresh_token')

    try {
      await api.post('/auth/logout', { refresh_token: refreshToken })
    } catch (error) {
      // Ignore errors on logout
    }

    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    setUser(null)
    navigate('/login')
  }

  const refreshToken = async () => {
    const refresh = localStorage.getItem('refresh_token')
    if (!refresh) {
      throw new Error('No refresh token')
    }

    try {
      const response = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refresh }),
      })

      if (!response.ok) {
        throw new Error('Failed to refresh token')
      }

      const data = await response.json()
      localStorage.setItem('access_token', data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
      setUser(data.user)

      return data.access_token
    } catch (error) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      setUser(null)
      navigate('/login')
      throw error
    }
  }

  const requestPasswordReset = async (email) => {
    return await api.post('/auth/request-password-reset', { email })
  }

  const resetPassword = async (token, password) => {
    return await api.post('/auth/reset-password', { token, password })
  }

  const verifyEmail = async (token) => {
    return await api.post('/auth/verify-email', { token })
  }

  const value = {
    user,
    loading,
    login,
    verifyMFA,
    register,
    logout,
    refreshToken,
    requestPasswordReset,
    resetPassword,
    verifyEmail,
    isAuthenticated: !!user,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Box, Spinner, Text, VStack } from '@chakra-ui/react'
import { useAuth } from '../contexts/AuthContext'

export function OAuthCallbackPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { login } = useAuth()

  useEffect(() => {
    const handleCallback = async () => {
      // Get tokens from URL
      const tokensParam = searchParams.get('tokens')
      const error = searchParams.get('error')

      if (error) {
        console.error('OAuth error:', error)
        navigate('/login?error=' + error)
        return
      }

      if (!tokensParam) {
        console.error('No tokens in callback')
        navigate('/login?error=no_tokens')
        return
      }

      try {
        // Decode tokens from base64
        const tokensJSON = atob(tokensParam)
        const tokens = JSON.parse(tokensJSON)

        if (!tokens.access_token || !tokens.refresh_token) {
          throw new Error('Invalid tokens')
        }

        // Store tokens and update auth context
        localStorage.setItem('access_token', tokens.access_token)
        localStorage.setItem('refresh_token', tokens.refresh_token)

        // Fetch user data with the access token
        const response = await fetch('/api/v1/auth/me', {
          headers: {
            Authorization: `Bearer ${tokens.access_token}`,
          },
        })

        if (!response.ok) {
          throw new Error('Failed to fetch user data')
        }

        const userData = await response.json()

        // Update auth context
        login(tokens.access_token, tokens.refresh_token, userData.user)

        // Redirect to dashboard
        navigate('/dashboard')
      } catch (error) {
        console.error('Failed to process OAuth callback:', error)
        navigate('/login?error=invalid_callback')
      }
    }

    handleCallback()
  }, [searchParams, navigate, login])

  return (
    <Box
      minH="100vh"
      display="flex"
      alignItems="center"
      justifyContent="center"
      bg="gray.50"
    >
      <VStack gap={4}>
        <Spinner size="xl" color="blue.500" />
        <Text color="gray.600">Completing sign in...</Text>
      </VStack>
    </Box>
  )
}

import { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  Box,
  Button,
  Container,
  Input,
  Stack,
  Heading,
  Text,
  Link as ChakraLink,
  VStack,
  FormControl,
  FormLabel,
  FormErrorMessage,
} from '@chakra-ui/react'
import { useAuth } from '../contexts/AuthContext'
import { SocialLogin } from '../components/SocialLogin'

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required'),
})

export function LoginPage() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showMFA, setShowMFA] = useState(false)
  const [mfaToken, setMfaToken] = useState('')
  const [mfaCode, setMfaCode] = useState('')
  const { login, verifyMFA } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const from = location.state?.from?.pathname || '/dashboard'

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data) => {
    setError('')
    setLoading(true)

    try {
      const response = await login(data.email, data.password)

      // Check if MFA is required
      if (response.mfa_required) {
        setMfaToken(response.mfa_token)
        setShowMFA(true)
        setLoading(false)
        return
      }

      // Normal login completed
      navigate(from, { replace: true })
    } catch (err) {
      setError(err.message || 'Failed to login')
    } finally {
      setLoading(false)
    }
  }

  const onMFASubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await verifyMFA(mfaToken, mfaCode)
      navigate(from, { replace: true })
    } catch (err) {
      setError(err.message || 'Invalid code')
      setMfaCode('')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Container maxW="md" py={16}>
      <VStack spacing={8}>
        <VStack spacing={2} textAlign="center">
          <Heading size="xl">{showMFA ? 'Two-Factor Authentication' : 'Welcome Back'}</Heading>
          <Text color="gray.600">
            {showMFA ? 'Enter your authentication code' : 'Sign in to your account'}
          </Text>
        </VStack>

        <Box w="full" p={8} borderWidth={1} borderRadius="lg" boxShadow="sm">
          {showMFA ? (
            <form onSubmit={onMFASubmit}>
              <Stack spacing={4}>
                <FormControl>
                  <FormLabel>6-digit code or backup code</FormLabel>
                  <Input
                    value={mfaCode}
                    onChange={(e) => setMfaCode(e.target.value.toUpperCase())}
                    placeholder="000000 or XXXX-XXXX-XXXX"
                    size="lg"
                    fontSize="2xl"
                    textAlign="center"
                    letterSpacing="wide"
                    autoFocus
                    autoComplete="off"
                  />
                  <Text fontSize="xs" color="gray.500" mt={2}>
                    Enter the 6-digit code from your authenticator app, or use one of your backup codes
                  </Text>
                </FormControl>

                {error && (
                  <Text color="red.500" fontSize="sm">
                    {error}
                  </Text>
                )}

                <Button
                  type="submit"
                  colorScheme="blue"
                  w="full"
                  isLoading={loading}
                  isDisabled={!mfaCode}
                >
                  Verify
                </Button>

                <Button
                  variant="ghost"
                  w="full"
                  onClick={() => {
                    setShowMFA(false)
                    setMfaToken('')
                    setMfaCode('')
                    setError('')
                  }}
                >
                  Back to login
                </Button>
              </Stack>
            </form>
          ) : (
            <form onSubmit={handleSubmit(onSubmit)}>
              <Stack spacing={4}>
                <FormControl isInvalid={errors.email}>
                  <FormLabel>Email</FormLabel>
                  <Input
                    type="email"
                    placeholder="you@example.com"
                    {...register('email')}
                  />
                  {errors.email && (
                    <FormErrorMessage>{errors.email.message}</FormErrorMessage>
                  )}
                </FormControl>

                <FormControl isInvalid={errors.password}>
                  <FormLabel>Password</FormLabel>
                  <Input
                    type="password"
                    placeholder="••••••••"
                    {...register('password')}
                  />
                  {errors.password && (
                    <FormErrorMessage>{errors.password.message}</FormErrorMessage>
                  )}
                </FormControl>

                <ChakraLink
                  as={Link}
                  to="/forgot-password"
                  fontSize="sm"
                  color="blue.600"
                  textAlign="right"
                >
                  Forgot password?
                </ChakraLink>

                {error && (
                  <Text color="red.500" fontSize="sm">
                    {error}
                  </Text>
                )}

                <Button
                  type="submit"
                  colorScheme="blue"
                  w="full"
                  isLoading={loading}
                >
                  Sign In
                </Button>

                <SocialLogin />
              </Stack>
            </form>
          )}
        </Box>

        {!showMFA && (
          <Text fontSize="sm">
            Don't have an account?{' '}
            <ChakraLink as={Link} to="/register" color="blue.600" fontWeight="medium">
              Sign up
            </ChakraLink>
          </Text>
        )}
      </VStack>
    </Container>
  )
}

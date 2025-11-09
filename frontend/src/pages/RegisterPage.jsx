import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
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

const registerSchema = z.object({
  firstName: z.string().min(1, 'First name is required'),
  lastName: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ['confirmPassword'],
})

export function RegisterPage() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { register: registerUser } = useAuth()
  const navigate = useNavigate()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (data) => {
    setError('')
    setLoading(true)

    try {
      await registerUser(data.email, data.password, data.firstName, data.lastName)
      navigate('/dashboard')
    } catch (err) {
      setError(err.message || 'Failed to register')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Container maxW="md" py={16}>
      <VStack spacing={8}>
        <VStack spacing={2} textAlign="center">
          <Heading size="xl">Create Account</Heading>
          <Text color="gray.600">Sign up to get started</Text>
        </VStack>

        <Box w="full" p={8} borderWidth={1} borderRadius="lg" boxShadow="sm">
          <form onSubmit={handleSubmit(onSubmit)}>
            <Stack spacing={4}>
              <Stack direction={{ base: 'column', sm: 'row' }} spacing={4}>
                <FormControl isInvalid={errors.firstName}>
                  <FormLabel>First Name</FormLabel>
                  <Input
                    placeholder="John"
                    {...register('firstName')}
                  />
                  {errors.firstName && (
                    <FormErrorMessage>{errors.firstName.message}</FormErrorMessage>
                  )}
                </FormControl>

                <FormControl isInvalid={errors.lastName}>
                  <FormLabel>Last Name</FormLabel>
                  <Input
                    placeholder="Doe"
                    {...register('lastName')}
                  />
                  {errors.lastName && (
                    <FormErrorMessage>{errors.lastName.message}</FormErrorMessage>
                  )}
                </FormControl>
              </Stack>

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

              <FormControl isInvalid={errors.confirmPassword}>
                <FormLabel>Confirm Password</FormLabel>
                <Input
                  type="password"
                  placeholder="••••••••"
                  {...register('confirmPassword')}
                />
                {errors.confirmPassword && (
                  <FormErrorMessage>{errors.confirmPassword.message}</FormErrorMessage>
                )}
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
              >
                Sign Up
              </Button>
            </Stack>
          </form>
        </Box>

        <Text fontSize="sm">
          Already have an account?{' '}
          <ChakraLink as={Link} to="/login" color="blue.600" fontWeight="medium">
            Sign in
          </ChakraLink>
        </Text>
      </VStack>
    </Container>
  )
}

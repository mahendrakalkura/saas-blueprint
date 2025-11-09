import { Box, Button, Container, Heading, Text, VStack } from '@chakra-ui/react'
import { useAuth } from '../contexts/AuthContext'
import { useColorMode } from '../providers'

export function DashboardPage() {
  const { user, logout } = useAuth()
  const { colorMode, toggleColorMode } = useColorMode()

  return (
    <Container maxW="container.xl" py={8}>
      <VStack spacing={8} align="stretch">
        <Box>
          <Heading size="xl">Dashboard</Heading>
          <Text mt={2} color="gray.600">
            Welcome back, {user?.first_name || user?.email}!
          </Text>
        </Box>

        <Box p={6} borderWidth={1} borderRadius="lg">
          <VStack spacing={4} align="stretch">
            <Text fontSize="lg" fontWeight="bold">
              Account Information
            </Text>
            <Text>
              <strong>Email:</strong> {user?.email}
            </Text>
            <Text>
              <strong>Email Verified:</strong> {user?.email_verified ? 'Yes' : 'No'}
            </Text>
            <Text>
              <strong>Account Status:</strong> {user?.is_active ? 'Active' : 'Inactive'}
            </Text>
          </VStack>
        </Box>

        <Box display="flex" gap={4}>
          <Button onClick={toggleColorMode} colorScheme="blue">
            Toggle {colorMode === 'light' ? 'Dark' : 'Light'} Mode
          </Button>
          <Button onClick={logout} variant="outline" colorScheme="red">
            Logout
          </Button>
        </Box>
      </VStack>
    </Container>
  )
}

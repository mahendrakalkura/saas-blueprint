import { Button, VStack, HStack, Text, Divider, Icon } from '@chakra-ui/react'
import { FaGoogle, FaGithub } from 'react-icons/fa'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'

export function SocialLogin() {
  const handleGoogleLogin = () => {
    window.location.href = `${API_URL}/auth/google`
  }

  const handleGitHubLogin = () => {
    window.location.href = `${API_URL}/auth/github`
  }

  return (
    <VStack gap={4} w="full">
      <HStack w="full" gap={2}>
        <Divider />
        <Text fontSize="sm" color="gray.500" whiteSpace="nowrap">
          Or continue with
        </Text>
        <Divider />
      </HStack>

      <HStack w="full" gap={3}>
        <Button
          flex={1}
          leftIcon={<Icon as={FaGoogle} />}
          onClick={handleGoogleLogin}
          colorScheme="red"
          variant="outline"
        >
          Google
        </Button>
        <Button
          flex={1}
          leftIcon={<Icon as={FaGithub} />}
          onClick={handleGitHubLogin}
          colorScheme="gray"
          variant="outline"
        >
          GitHub
        </Button>
      </HStack>
    </VStack>
  )
}

import { ErrorBoundary as ReactErrorBoundary } from 'react-error-boundary'
import { Box, Button, Heading, Text, VStack } from '@chakra-ui/react'

function ErrorFallback({ error, resetErrorBoundary }) {
  return (
    <Box minH="100vh" display="flex" alignItems="center" justifyContent="center" p={4}>
      <VStack spacing={4} maxW="md" textAlign="center">
        <Heading size="lg">Something went wrong</Heading>
        <Text color="gray.600">
          {error.message || 'An unexpected error occurred'}
        </Text>
        <Button onClick={resetErrorBoundary} colorScheme="blue">
          Try again
        </Button>
      </VStack>
    </Box>
  )
}

export function ErrorBoundary({ children }) {
  return (
    <ReactErrorBoundary
      FallbackComponent={ErrorFallback}
      onReset={() => window.location.reload()}
    >
      {children}
    </ReactErrorBoundary>
  )
}

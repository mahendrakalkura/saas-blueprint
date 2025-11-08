import { Box, Button, Heading, VStack, Text } from '@chakra-ui/react'
import { useColorMode } from '../providers'

export function HomePage() {
  const { colorMode, toggleColorMode } = useColorMode()

  return (
    <Box minH="100vh" display="flex" alignItems="center" justifyContent="center" p={4}>
      <VStack spacing={6} textAlign="center">
        <Heading size="2xl">Welcome to SaaS Blueprint</Heading>
        <Text fontSize="lg" color="gray.600">
          A modern full-stack SaaS application with React, Go, and PostgreSQL
        </Text>
        <Button onClick={toggleColorMode} colorScheme="blue">
          Toggle {colorMode === 'light' ? 'Dark' : 'Light'} Mode
        </Button>
      </VStack>
    </Box>
  )
}

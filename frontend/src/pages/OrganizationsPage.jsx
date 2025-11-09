import { useState } from 'react'
import {
  Container,
  Heading,
  VStack,
  HStack,
  Box,
  Text,
  Button,
  Card,
  SimpleGrid,
  Badge,
  IconButton,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  FormControl,
  FormLabel,
  Input,
  useToast,
} from '@chakra-ui/react'
import { FiPlus, FiUsers, FiSettings } from 'react-icons/fi'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

export function OrganizationsPage() {
  const [newOrgName, setNewOrgName] = useState('')
  const { isOpen, onOpen, onClose } = useDisclosure()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const toast = useToast()

  const { data: orgsData, isLoading } = useQuery({
    queryKey: ['organizations'],
    queryFn: async () => {
      const response = await api.get('/organizations')
      return response.organizations || []
    },
  })

  const createMutation = useMutation({
    mutationFn: async (name) => {
      await api.post('/organizations', { name })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] })
      toast({
        title: 'Organization created',
        status: 'success',
        duration: 3000,
      })
      setNewOrgName('')
      onClose()
    },
    onError: (error) => {
      toast({
        title: 'Failed to create organization',
        description: error.response?.data?.error || 'Please try again',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const handleCreate = () => {
    if (newOrgName.trim().length < 3) {
      toast({
        title: 'Invalid name',
        description: 'Organization name must be at least 3 characters',
        status: 'warning',
        duration: 3000,
      })
      return
    }
    createMutation.mutate(newOrgName)
  }

  return (
    <Container maxW="container.xl" py={8}>
      <VStack spacing={8} align="stretch">
        <HStack justify="space-between">
          <Heading>Organizations</Heading>
          <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={onOpen}>
            Create Organization
          </Button>
        </HStack>

        {isLoading ? (
          <Text>Loading organizations...</Text>
        ) : orgsData?.length === 0 ? (
          <Box textAlign="center" py={12}>
            <Text fontSize="lg" color="gray.500" mb={4}>
              You don't have any organizations yet
            </Text>
            <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={onOpen}>
              Create Your First Organization
            </Button>
          </Box>
        ) : (
          <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={6}>
            {orgsData?.map((org) => (
              <Card
                key={org.id}
                p={6}
                cursor="pointer"
                _hover={{ shadow: 'lg', transform: 'translateY(-2px)' }}
                transition="all 0.2s"
                onClick={() => navigate(`/organizations/${org.id}`)}
              >
                <VStack align="stretch" spacing={4}>
                  <HStack justify="space-between">
                    <Heading size="md">{org.name}</Heading>
                    <IconButton
                      icon={<FiSettings />}
                      size="sm"
                      variant="ghost"
                      aria-label="Settings"
                    />
                  </HStack>

                  <HStack spacing={4} fontSize="sm" color="gray.600">
                    <HStack>
                      <FiUsers />
                      <Text>Team</Text>
                    </HStack>
                    <Badge colorScheme={org.owner_id ? 'green' : 'gray'}>
                      {org.owner_id ? 'Active' : 'Inactive'}
                    </Badge>
                  </HStack>

                  <Text fontSize="sm" color="gray.500">
                    Created {new Date(org.created_at).toLocaleDateString()}
                  </Text>
                </VStack>
              </Card>
            ))}
          </SimpleGrid>
        )}

        <Modal isOpen={isOpen} onClose={onClose}>
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Create Organization</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <FormControl>
                <FormLabel>Organization Name</FormLabel>
                <Input
                  value={newOrgName}
                  onChange={(e) => setNewOrgName(e.target.value)}
                  placeholder="Acme Inc."
                  autoFocus
                />
              </FormControl>
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={onClose}>
                Cancel
              </Button>
              <Button
                colorScheme="blue"
                onClick={handleCreate}
                isLoading={createMutation.isPending}
              >
                Create
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </VStack>
    </Container>
  )
}

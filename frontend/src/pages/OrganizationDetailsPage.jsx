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
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
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
  Select,
  useToast,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Alert,
  AlertIcon,
} from '@chakra-ui/react'
import { FiUserPlus, FiTrash2, FiMail, FiEdit2 } from 'react-icons/fi'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

export function OrganizationDetailsPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const toast = useToast()

  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('member')
  const [editName, setEditName] = useState('')

  const { isOpen: isInviteOpen, onOpen: onInviteOpen, onClose: onInviteClose } = useDisclosure()
  const { isOpen: isEditOpen, onOpen: onEditOpen, onClose: onEditClose } = useDisclosure()

  const { data: org, isLoading: isLoadingOrg } = useQuery({
    queryKey: ['organization', id],
    queryFn: async () => {
      const response = await api.get(`/organizations/${id}`)
      return response
    },
  })

  const { data: members, isLoading: isLoadingMembers } = useQuery({
    queryKey: ['organization-members', id],
    queryFn: async () => {
      const response = await api.get(`/organizations/${id}/members`)
      return response.members || []
    },
  })

  const inviteMutation = useMutation({
    mutationFn: async ({ email, role }) => {
      await api.post(`/organizations/${id}/members/invite`, { email, role })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization-members', id] })
      toast({
        title: 'Invitation sent',
        status: 'success',
        duration: 3000,
      })
      setInviteEmail('')
      setInviteRole('member')
      onInviteClose()
    },
    onError: (error) => {
      toast({
        title: 'Failed to send invitation',
        description: error.response?.data?.error || 'Please try again',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: async (memberID) => {
      await api.delete(`/organizations/${id}/members/${memberID}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization-members', id] })
      toast({
        title: 'Member removed',
        status: 'success',
        duration: 3000,
      })
    },
    onError: (error) => {
      toast({
        title: 'Failed to remove member',
        description: error.response?.data?.error || 'Please try again',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const updateOrgMutation = useMutation({
    mutationFn: async (name) => {
      await api.put(`/organizations/${id}`, { name })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization', id] })
      queryClient.invalidateQueries({ queryKey: ['organizations'] })
      toast({
        title: 'Organization updated',
        status: 'success',
        duration: 3000,
      })
      onEditClose()
    },
    onError: (error) => {
      toast({
        title: 'Failed to update organization',
        description: error.response?.data?.error || 'Please try again',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const deleteOrgMutation = useMutation({
    mutationFn: async () => {
      await api.delete(`/organizations/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] })
      toast({
        title: 'Organization deleted',
        status: 'success',
        duration: 3000,
      })
      navigate('/organizations')
    },
    onError: (error) => {
      toast({
        title: 'Failed to delete organization',
        description: error.response?.data?.error || 'Please try again',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const handleInvite = () => {
    if (!inviteEmail || !inviteRole) {
      toast({
        title: 'Invalid input',
        description: 'Please fill in all fields',
        status: 'warning',
        duration: 3000,
      })
      return
    }
    inviteMutation.mutate({ email: inviteEmail, role: inviteRole })
  }

  const handleUpdateOrg = () => {
    if (editName.trim().length < 3) {
      toast({
        title: 'Invalid name',
        description: 'Organization name must be at least 3 characters',
        status: 'warning',
        duration: 3000,
      })
      return
    }
    updateOrgMutation.mutate(editName)
  }

  const handleDeleteOrg = () => {
    if (window.confirm('Are you sure you want to delete this organization? This action cannot be undone.')) {
      deleteOrgMutation.mutate()
    }
  }

  const handleRemoveMember = (memberID) => {
    if (window.confirm('Are you sure you want to remove this member?')) {
      removeMemberMutation.mutate(memberID)
    }
  }

  const getRoleBadgeColor = (role) => {
    switch (role) {
      case 'owner':
        return 'purple'
      case 'admin':
        return 'blue'
      default:
        return 'gray'
    }
  }

  if (isLoadingOrg) {
    return (
      <Container maxW="container.xl" py={8}>
        <Text>Loading organization...</Text>
      </Container>
    )
  }

  return (
    <Container maxW="container.xl" py={8}>
      <VStack spacing={8} align="stretch">
        <HStack justify="space-between">
          <Box>
            <Heading>{org?.name}</Heading>
            <Text color="gray.600" fontSize="sm" mt={1}>
              /{org?.slug}
            </Text>
          </Box>
          <HStack>
            <Button leftIcon={<FiEdit2 />} variant="outline" onClick={() => {
              setEditName(org?.name || '')
              onEditOpen()
            }}>
              Edit
            </Button>
            <Button leftIcon={<FiUserPlus />} colorScheme="blue" onClick={onInviteOpen}>
              Invite Member
            </Button>
          </HStack>
        </HStack>

        <Tabs colorScheme="blue">
          <TabList>
            <Tab>Members</Tab>
            <Tab>Settings</Tab>
          </TabList>

          <TabPanels>
            <TabPanel>
              <VStack spacing={4} align="stretch">
                {isLoadingMembers ? (
                  <Text>Loading members...</Text>
                ) : members?.length === 0 ? (
                  <Alert status="info">
                    <AlertIcon />
                    No members yet. Invite team members to collaborate.
                  </Alert>
                ) : (
                  <Box overflowX="auto">
                    <Table variant="simple">
                      <Thead>
                        <Tr>
                          <Th>Name</Th>
                          <Th>Email</Th>
                          <Th>Role</Th>
                          <Th>Joined</Th>
                          <Th>Actions</Th>
                        </Tr>
                      </Thead>
                      <Tbody>
                        {members?.map(({ member, user }) => (
                          <Tr key={member.id}>
                            <Td>
                              {user.first_name && user.last_name
                                ? `${user.first_name} ${user.last_name}`
                                : 'N/A'}
                            </Td>
                            <Td>{user.email}</Td>
                            <Td>
                              <Badge colorScheme={getRoleBadgeColor(member.role)}>
                                {member.role.toUpperCase()}
                              </Badge>
                            </Td>
                            <Td>
                              {member.joined_at
                                ? new Date(member.joined_at).toLocaleDateString()
                                : 'Pending'}
                            </Td>
                            <Td>
                              {member.role !== 'owner' && (
                                <IconButton
                                  icon={<FiTrash2 />}
                                  size="sm"
                                  colorScheme="red"
                                  variant="ghost"
                                  aria-label="Remove member"
                                  onClick={() => handleRemoveMember(member.id)}
                                  isLoading={removeMemberMutation.isPending}
                                />
                              )}
                            </Td>
                          </Tr>
                        ))}
                      </Tbody>
                    </Table>
                  </Box>
                )}
              </VStack>
            </TabPanel>

            <TabPanel>
              <VStack spacing={6} align="stretch">
                <Card p={6}>
                  <VStack align="stretch" spacing={4}>
                    <Heading size="md">Danger Zone</Heading>
                    <Text color="gray.600">
                      Deleting an organization is permanent and cannot be undone.
                    </Text>
                    <Button
                      colorScheme="red"
                      leftIcon={<FiTrash2 />}
                      onClick={handleDeleteOrg}
                      isLoading={deleteOrgMutation.isPending}
                    >
                      Delete Organization
                    </Button>
                  </VStack>
                </Card>
              </VStack>
            </TabPanel>
          </TabPanels>
        </Tabs>

        {/* Invite Modal */}
        <Modal isOpen={isInviteOpen} onClose={onInviteClose}>
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Invite Team Member</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4}>
                <FormControl>
                  <FormLabel>Email Address</FormLabel>
                  <Input
                    type="email"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    placeholder="colleague@example.com"
                  />
                </FormControl>
                <FormControl>
                  <FormLabel>Role</FormLabel>
                  <Select value={inviteRole} onChange={(e) => setInviteRole(e.target.value)}>
                    <option value="member">Member</option>
                    <option value="admin">Admin</option>
                  </Select>
                </FormControl>
              </VStack>
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={onInviteClose}>
                Cancel
              </Button>
              <Button
                colorScheme="blue"
                leftIcon={<FiMail />}
                onClick={handleInvite}
                isLoading={inviteMutation.isPending}
              >
                Send Invitation
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Edit Modal */}
        <Modal isOpen={isEditOpen} onClose={onEditClose}>
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Edit Organization</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <FormControl>
                <FormLabel>Organization Name</FormLabel>
                <Input
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  placeholder="Acme Inc."
                />
              </FormControl>
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={onEditClose}>
                Cancel
              </Button>
              <Button
                colorScheme="blue"
                onClick={handleUpdateOrg}
                isLoading={updateOrgMutation.isPending}
              >
                Save Changes
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </VStack>
    </Container>
  )
}

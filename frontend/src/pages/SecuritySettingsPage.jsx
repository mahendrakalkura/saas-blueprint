import { useState } from 'react'
import {
  Box,
  Button,
  Container,
  Heading,
  Text,
  VStack,
  HStack,
  Badge,
  Card,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  Input,
  useDisclosure,
  useToast,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Code,
  SimpleGrid,
} from '@chakra-ui/react'
import { FiShield, FiLock, FiKey, FiDownload, FiCopy } from 'react-icons/fi'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { MFASetup } from '../components/MFASetup'

export function SecuritySettingsPage() {
  const [disableCode, setDisableCode] = useState('')
  const [regenerateCode, setRegenerateCode] = useState('')
  const [newBackupCodes, setNewBackupCodes] = useState(null)
  const toast = useToast()
  const queryClient = useQueryClient()

  const {
    isOpen: isSetupOpen,
    onOpen: onSetupOpen,
    onClose: onSetupClose,
  } = useDisclosure()

  const {
    isOpen: isDisableOpen,
    onOpen: onDisableOpen,
    onClose: onDisableClose,
  } = useDisclosure()

  const {
    isOpen: isRegenerateOpen,
    onOpen: onRegenerateOpen,
    onClose: onRegenerateClose,
  } = useDisclosure()

  // Fetch MFA status
  const { data: mfaStatus, isLoading } = useQuery({
    queryKey: ['mfa', 'status'],
    queryFn: async () => {
      return await api.get('/mfa/status')
    },
  })

  // Disable MFA mutation
  const disableMutation = useMutation({
    mutationFn: async (code) => {
      return await api.post('/mfa/disable', { code })
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['mfa', 'status'])
      onDisableClose()
      setDisableCode('')
      toast({
        title: 'MFA Disabled',
        description: 'Two-factor authentication has been disabled',
        status: 'success',
        duration: 3000,
      })
    },
    onError: (error) => {
      toast({
        title: 'Error',
        description: error.message || 'Failed to disable MFA',
        status: 'error',
        duration: 5000,
      })
    },
  })

  // Regenerate backup codes mutation
  const regenerateMutation = useMutation({
    mutationFn: async (code) => {
      return await api.post('/mfa/backup-codes/regenerate', { code })
    },
    onSuccess: (data) => {
      setNewBackupCodes(data.backup_codes)
      setRegenerateCode('')
      toast({
        title: 'Backup Codes Regenerated',
        description: 'Your old backup codes are no longer valid',
        status: 'success',
        duration: 3000,
      })
    },
    onError: (error) => {
      toast({
        title: 'Error',
        description: error.message || 'Failed to regenerate backup codes',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const handleSetupComplete = () => {
    queryClient.invalidateQueries(['mfa', 'status'])
    onSetupClose()
  }

  const handleDisable = (e) => {
    e.preventDefault()
    if (disableCode.length === 6) {
      disableMutation.mutate(disableCode)
    }
  }

  const handleRegenerate = (e) => {
    e.preventDefault()
    if (regenerateCode.length === 6) {
      regenerateMutation.mutate(regenerateCode)
    }
  }

  const handleCopyBackupCodes = () => {
    const codes = newBackupCodes.join('\n')
    navigator.clipboard.writeText(codes)
    toast({
      title: 'Copied!',
      description: 'Backup codes copied to clipboard',
      status: 'success',
      duration: 2000,
    })
  }

  const handleDownloadBackupCodes = () => {
    const codes = newBackupCodes.join('\n')
    const blob = new Blob([codes], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'mfa-backup-codes.txt'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const handleCloseRegenerate = () => {
    setNewBackupCodes(null)
    setRegenerateCode('')
    onRegenerateClose()
  }

  if (isLoading) {
    return (
      <Container maxW="4xl" py={8}>
        <Text>Loading...</Text>
      </Container>
    )
  }

  return (
    <Container maxW="4xl" py={8}>
      <VStack gap={8} align="stretch">
        <Box>
          <Heading size="lg" mb={2}>
            Security Settings
          </Heading>
          <Text color="gray.600">
            Manage your account security and authentication methods
          </Text>
        </Box>

        {/* MFA Section */}
        <Card p={6}>
          <VStack gap={4} align="stretch">
            <HStack justify="space-between">
              <HStack gap={3}>
                <Box
                  p={3}
                  bg="blue.100"
                  borderRadius="md"
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                >
                  <FiShield size={24} color="blue" />
                </Box>
                <Box>
                  <HStack gap={2} mb={1}>
                    <Heading size="sm">Two-Factor Authentication</Heading>
                    <Badge colorScheme={mfaStatus?.mfa_enabled ? 'green' : 'gray'}>
                      {mfaStatus?.mfa_enabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </HStack>
                  <Text fontSize="sm" color="gray.600">
                    {mfaStatus?.mfa_enabled
                      ? 'Your account is protected with 2FA'
                      : 'Add an extra layer of security to your account'}
                  </Text>
                </Box>
              </HStack>
            </HStack>

            <HStack>
              {!mfaStatus?.mfa_enabled ? (
                <Button leftIcon={<FiLock />} colorScheme="blue" onClick={onSetupOpen}>
                  Enable 2FA
                </Button>
              ) : (
                <>
                  <Button
                    leftIcon={<FiKey />}
                    variant="outline"
                    onClick={onRegenerateOpen}
                  >
                    Regenerate Backup Codes
                  </Button>
                  <Button variant="outline" colorScheme="red" onClick={onDisableOpen}>
                    Disable 2FA
                  </Button>
                </>
              )}
            </HStack>
          </VStack>
        </Card>

        {/* Password Section */}
        <Card p={6}>
          <VStack gap={4} align="stretch">
            <HStack gap={3}>
              <Box
                p={3}
                bg="purple.100"
                borderRadius="md"
                display="flex"
                alignItems="center"
                justifyContent="center"
              >
                <FiLock size={24} color="purple" />
              </Box>
              <Box>
                <Heading size="sm" mb={1}>
                  Password
                </Heading>
                <Text fontSize="sm" color="gray.600">
                  Change your account password
                </Text>
              </Box>
            </HStack>

            <Button variant="outline">Change Password</Button>
          </VStack>
        </Card>
      </VStack>

      {/* MFA Setup Modal */}
      <Modal isOpen={isSetupOpen} onClose={onSetupClose} size="lg" closeOnOverlayClick={false}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Enable Two-Factor Authentication</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <MFASetup onComplete={handleSetupComplete} />
          </ModalBody>
        </ModalContent>
      </Modal>

      {/* Disable MFA Modal */}
      <Modal isOpen={isDisableOpen} onClose={onDisableClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Disable Two-Factor Authentication</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <VStack gap={4} align="stretch" as="form" onSubmit={handleDisable}>
              <Alert status="warning" borderRadius="md">
                <AlertIcon />
                <Box>
                  <AlertTitle>Are you sure?</AlertTitle>
                  <AlertDescription>
                    Disabling 2FA will make your account less secure
                  </AlertDescription>
                </Box>
              </Alert>

              <Box>
                <Text fontSize="sm" fontWeight="medium" mb={2}>
                  Enter your 6-digit authentication code
                </Text>
                <Input
                  value={disableCode}
                  onChange={(e) =>
                    setDisableCode(e.target.value.replace(/\D/g, '').slice(0, 6))
                  }
                  placeholder="000000"
                  size="lg"
                  fontSize="2xl"
                  textAlign="center"
                  letterSpacing="wide"
                  maxLength={6}
                  autoFocus
                />
              </Box>

              <HStack>
                <Button onClick={onDisableClose} flex={1}>
                  Cancel
                </Button>
                <Button
                  type="submit"
                  colorScheme="red"
                  flex={1}
                  isLoading={disableMutation.isPending}
                  isDisabled={disableCode.length !== 6}
                >
                  Disable 2FA
                </Button>
              </HStack>
            </VStack>
          </ModalBody>
        </ModalContent>
      </Modal>

      {/* Regenerate Backup Codes Modal */}
      <Modal isOpen={isRegenerateOpen} onClose={handleCloseRegenerate} size="lg">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Regenerate Backup Codes</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            {!newBackupCodes ? (
              <VStack gap={4} align="stretch" as="form" onSubmit={handleRegenerate}>
                <Alert status="warning" borderRadius="md">
                  <AlertIcon />
                  <Box>
                    <AlertTitle>Your old backup codes will be invalidated</AlertTitle>
                    <AlertDescription>
                      Any previously saved backup codes will no longer work
                    </AlertDescription>
                  </Box>
                </Alert>

                <Box>
                  <Text fontSize="sm" fontWeight="medium" mb={2}>
                    Enter your 6-digit authentication code
                  </Text>
                  <Input
                    value={regenerateCode}
                    onChange={(e) =>
                      setRegenerateCode(e.target.value.replace(/\D/g, '').slice(0, 6))
                    }
                    placeholder="000000"
                    size="lg"
                    fontSize="2xl"
                    textAlign="center"
                    letterSpacing="wide"
                    maxLength={6}
                    autoFocus
                  />
                </Box>

                <HStack>
                  <Button onClick={handleCloseRegenerate} flex={1}>
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    colorScheme="blue"
                    flex={1}
                    isLoading={regenerateMutation.isPending}
                    isDisabled={regenerateCode.length !== 6}
                  >
                    Regenerate
                  </Button>
                </HStack>
              </VStack>
            ) : (
              <VStack gap={4} align="stretch">
                <Alert status="success" borderRadius="md">
                  <AlertIcon />
                  <Box>
                    <AlertTitle>New Backup Codes Generated</AlertTitle>
                    <AlertDescription>
                      Save these codes in a safe place
                    </AlertDescription>
                  </Box>
                </Alert>

                <Box p={4} bg="gray.50" borderRadius="md" borderWidth={1}>
                  <SimpleGrid columns={2} gap={2}>
                    {newBackupCodes.map((code, index) => (
                      <Code key={index} p={2} fontSize="sm" textAlign="center">
                        {code}
                      </Code>
                    ))}
                  </SimpleGrid>
                </Box>

                <HStack>
                  <Button leftIcon={<FiCopy />} onClick={handleCopyBackupCodes} flex={1}>
                    Copy Codes
                  </Button>
                  <Button
                    leftIcon={<FiDownload />}
                    onClick={handleDownloadBackupCodes}
                    flex={1}
                  >
                    Download
                  </Button>
                </HStack>

                <Button colorScheme="blue" onClick={handleCloseRegenerate}>
                  Done
                </Button>
              </VStack>
            )}
          </ModalBody>
        </ModalContent>
      </Modal>
    </Container>
  )
}

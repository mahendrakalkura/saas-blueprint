import { useState } from 'react'
import {
  Box,
  Button,
  VStack,
  HStack,
  Text,
  Heading,
  Image,
  Input,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Code,
  SimpleGrid,
  IconButton,
  useToast,
  Spinner,
} from '@chakra-ui/react'
import { FiCopy, FiDownload, FiCheck } from 'react-icons/fi'
import { useMutation, useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'

export function MFASetup({ onComplete }) {
  const [step, setStep] = useState(1) // 1: QR code, 2: verify, 3: backup codes
  const [code, setCode] = useState('')
  const [qrData, setQrData] = useState(null)
  const toast = useToast()

  // Enable MFA mutation (get QR code)
  const enableMutation = useMutation({
    mutationFn: async () => {
      return await api.post('/mfa/enable')
    },
    onSuccess: (data) => {
      setQrData(data)
      setStep(2)
    },
    onError: (error) => {
      toast({
        title: 'Error',
        description: error.message || 'Failed to enable MFA',
        status: 'error',
        duration: 5000,
      })
    },
  })

  // Verify and activate MFA mutation
  const verifyMutation = useMutation({
    mutationFn: async (code) => {
      return await api.post('/mfa/verify', { code })
    },
    onSuccess: () => {
      setStep(3)
    },
    onError: (error) => {
      toast({
        title: 'Invalid Code',
        description: error.message || 'Please check your code and try again',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const handleEnableMFA = () => {
    enableMutation.mutate()
  }

  const handleVerify = (e) => {
    e.preventDefault()
    if (code.length === 6) {
      verifyMutation.mutate(code)
    }
  }

  const handleCopyBackupCodes = () => {
    const codes = qrData.backup_codes.join('\n')
    navigator.clipboard.writeText(codes)
    toast({
      title: 'Copied!',
      description: 'Backup codes copied to clipboard',
      status: 'success',
      duration: 2000,
    })
  }

  const handleDownloadBackupCodes = () => {
    const codes = qrData.backup_codes.join('\n')
    const blob = new Blob([codes], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'mfa-backup-codes.txt'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)

    toast({
      title: 'Downloaded!',
      description: 'Backup codes saved to file',
      status: 'success',
      duration: 2000,
    })
  }

  const handleComplete = () => {
    toast({
      title: 'MFA Enabled!',
      description: 'Two-factor authentication is now active',
      status: 'success',
      duration: 3000,
    })
    if (onComplete) {
      onComplete()
    }
  }

  if (step === 1) {
    return (
      <VStack gap={6} align="stretch">
        <Box>
          <Heading size="md" mb={2}>
            Enable Two-Factor Authentication
          </Heading>
          <Text color="gray.600">
            Add an extra layer of security to your account
          </Text>
        </Box>

        <Alert status="info" borderRadius="md">
          <AlertIcon />
          <Box>
            <AlertTitle>You'll need an authenticator app</AlertTitle>
            <AlertDescription>
              Download Google Authenticator, Authy, 1Password, or any TOTP-compatible app
            </AlertDescription>
          </Box>
        </Alert>

        <Button
          colorScheme="blue"
          onClick={handleEnableMFA}
          isLoading={enableMutation.isPending}
        >
          Get Started
        </Button>
      </VStack>
    )
  }

  if (step === 2 && qrData) {
    return (
      <VStack gap={6} align="stretch">
        <Box>
          <Heading size="md" mb={2}>
            Scan QR Code
          </Heading>
          <Text color="gray.600">
            Open your authenticator app and scan this QR code
          </Text>
        </Box>

        <Box
          p={4}
          bg="white"
          borderRadius="md"
          borderWidth={1}
          display="flex"
          justifyContent="center"
        >
          <Image
            src={`data:image/png;base64,${qrData.qr_code}`}
            alt="MFA QR Code"
            boxSize="200px"
          />
        </Box>

        <Box>
          <Text fontSize="sm" color="gray.600" mb={2}>
            Or enter this code manually:
          </Text>
          <Code p={3} borderRadius="md" fontSize="sm" display="block">
            {qrData.secret}
          </Code>
        </Box>

        <Box as="form" onSubmit={handleVerify}>
          <VStack gap={4} align="stretch">
            <Box>
              <Text fontSize="sm" fontWeight="medium" mb={2}>
                Enter 6-digit code from your app
              </Text>
              <Input
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                placeholder="000000"
                size="lg"
                fontSize="2xl"
                textAlign="center"
                letterSpacing="wide"
                maxLength={6}
                autoFocus
              />
            </Box>

            <Button
              type="submit"
              colorScheme="blue"
              isLoading={verifyMutation.isPending}
              isDisabled={code.length !== 6}
            >
              Verify and Enable
            </Button>
          </VStack>
        </Box>
      </VStack>
    )
  }

  if (step === 3 && qrData) {
    return (
      <VStack gap={6} align="stretch">
        <Box textAlign="center">
          <Box
            w={16}
            h={16}
            bg="green.100"
            borderRadius="full"
            display="flex"
            alignItems="center"
            justifyContent="center"
            mx="auto"
            mb={4}
          >
            <FiCheck size={32} color="green" />
          </Box>
          <Heading size="md" mb={2}>
            MFA Enabled Successfully!
          </Heading>
          <Text color="gray.600">
            Save these backup codes in a safe place
          </Text>
        </Box>

        <Alert status="warning" borderRadius="md">
          <AlertIcon />
          <Box>
            <AlertTitle>Important!</AlertTitle>
            <AlertDescription>
              These backup codes can be used to access your account if you lose your
              authenticator device. Each code can only be used once.
            </AlertDescription>
          </Box>
        </Alert>

        <Box
          p={4}
          bg="gray.50"
          borderRadius="md"
          borderWidth={1}
        >
          <SimpleGrid columns={2} gap={2}>
            {qrData.backup_codes.map((code, index) => (
              <Code key={index} p={2} fontSize="sm" textAlign="center">
                {code}
              </Code>
            ))}
          </SimpleGrid>
        </Box>

        <HStack>
          <Button
            leftIcon={<FiCopy />}
            onClick={handleCopyBackupCodes}
            flex={1}
          >
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

        <Button colorScheme="blue" onClick={handleComplete}>
          Done
        </Button>
      </VStack>
    )
  }

  return null
}

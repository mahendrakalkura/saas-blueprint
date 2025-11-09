import { useState, useRef } from 'react'
import { Box, Button, Text, VStack, HStack, Icon, Progress, useToast } from '@chakra-ui/react'
import { FiUpload, FiFile, FiX } from 'react-icons/fi'
import { api } from '../lib/api'

export function FileUpload({ onUploadComplete, accept, maxSize = 10 * 1024 * 1024 }) {
  const [file, setFile] = useState(null)
  const [uploading, setUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const fileInputRef = useRef(null)
  const toast = useToast()

  const handleFileSelect = (event) => {
    const selectedFile = event.target.files?.[0]
    if (!selectedFile) return

    if (selectedFile.size > maxSize) {
      toast({
        title: 'File too large',
        description: `File size must be less than ${maxSize / (1024 * 1024)}MB`,
        status: 'error',
        duration: 5000,
      })
      return
    }

    setFile(selectedFile)
  }

  const handleUpload = async () => {
    if (!file) return

    setUploading(true)
    setUploadProgress(0)

    try {
      const formData = new FormData()
      formData.append('file', file)

      const response = await api.post('/files', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          setUploadProgress(progress)
        },
      })

      toast({
        title: 'File uploaded successfully',
        status: 'success',
        duration: 3000,
      })

      setFile(null)
      setUploadProgress(0)

      if (onUploadComplete) {
        onUploadComplete(response)
      }
    } catch (error) {
      toast({
        title: 'Upload failed',
        description: error.response?.data?.error || 'Failed to upload file',
        status: 'error',
        duration: 5000,
      })
    } finally {
      setUploading(false)
    }
  }

  const handleRemoveFile = () => {
    setFile(null)
    setUploadProgress(0)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <VStack spacing={4} align="stretch">
      <Box
        borderWidth={2}
        borderStyle="dashed"
        borderColor="gray.300"
        borderRadius="md"
        p={8}
        textAlign="center"
        cursor="pointer"
        onClick={() => fileInputRef.current?.click()}
        _hover={{ borderColor: 'blue.400', bg: 'gray.50' }}
        transition="all 0.2s"
      >
        <input
          ref={fileInputRef}
          type="file"
          accept={accept}
          onChange={handleFileSelect}
          style={{ display: 'none' }}
        />

        <Icon as={FiUpload} boxSize={12} color="gray.400" mb={4} />
        <Text fontSize="lg" fontWeight="medium" mb={2}>
          Click to upload
        </Text>
        <Text fontSize="sm" color="gray.500">
          or drag and drop
        </Text>
        {maxSize && (
          <Text fontSize="xs" color="gray.400" mt={2}>
            Maximum file size: {maxSize / (1024 * 1024)}MB
          </Text>
        )}
      </Box>

      {file && (
        <Box
          p={4}
          borderWidth={1}
          borderRadius="md"
          bg="gray.50"
        >
          <HStack justify="space-between" mb={2}>
            <HStack>
              <Icon as={FiFile} color="blue.500" />
              <VStack align="start" spacing={0}>
                <Text fontSize="sm" fontWeight="medium">
                  {file.name}
                </Text>
                <Text fontSize="xs" color="gray.500">
                  {(file.size / 1024).toFixed(2)} KB
                </Text>
              </VStack>
            </HStack>

            {!uploading && (
              <Button
                size="sm"
                variant="ghost"
                onClick={handleRemoveFile}
                leftIcon={<FiX />}
              >
                Remove
              </Button>
            )}
          </HStack>

          {uploading && (
            <Progress value={uploadProgress} size="sm" colorScheme="blue" hasStripe isAnimated />
          )}

          {!uploading && (
            <Button
              onClick={handleUpload}
              colorScheme="blue"
              size="sm"
              width="full"
              mt={2}
            >
              Upload
            </Button>
          )}
        </Box>
      )}
    </VStack>
  )
}

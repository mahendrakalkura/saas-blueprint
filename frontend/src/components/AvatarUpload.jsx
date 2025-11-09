import { useState, useRef } from 'react'
import { Box, Avatar, Button, VStack, Icon, useToast, Text } from '@chakra-ui/react'
import { FiCamera, FiUpload } from 'react-icons/fi'
import { api } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'

export function AvatarUpload({ onUploadComplete }) {
  const { user, setUser } = useAuth()
  const [uploading, setUploading] = useState(false)
  const [previewUrl, setPreviewUrl] = useState(null)
  const fileInputRef = useRef(null)
  const toast = useToast()

  const handleFileSelect = async (event) => {
    const file = event.target.files?.[0]
    if (!file) return

    // Validate file type
    if (!file.type.startsWith('image/')) {
      toast({
        title: 'Invalid file type',
        description: 'Please select an image file',
        status: 'error',
        duration: 5000,
      })
      return
    }

    // Validate file size (2MB)
    if (file.size > 2 * 1024 * 1024) {
      toast({
        title: 'File too large',
        description: 'Avatar image must be less than 2MB',
        status: 'error',
        duration: 5000,
      })
      return
    }

    // Create preview
    const reader = new FileReader()
    reader.onloadend = () => {
      setPreviewUrl(reader.result)
    }
    reader.readAsDataURL(file)

    // Upload
    setUploading(true)

    try {
      const formData = new FormData()
      formData.append('avatar', file)

      const response = await api.post('/files/avatar', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })

      toast({
        title: 'Avatar uploaded successfully',
        status: 'success',
        duration: 3000,
      })

      // Update user avatar URL
      if (user) {
        setUser({
          ...user,
          avatar_url: response.url,
        })
      }

      if (onUploadComplete) {
        onUploadComplete(response)
      }
    } catch (error) {
      toast({
        title: 'Upload failed',
        description: error.response?.data?.error || 'Failed to upload avatar',
        status: 'error',
        duration: 5000,
      })
      setPreviewUrl(null)
    } finally {
      setUploading(false)
    }
  }

  const currentAvatarUrl = previewUrl || user?.avatar_url

  return (
    <VStack spacing={4}>
      <Box position="relative">
        <Avatar
          size="2xl"
          src={currentAvatarUrl}
          name={user?.email}
        />
        <Button
          position="absolute"
          bottom={0}
          right={0}
          size="sm"
          colorScheme="blue"
          borderRadius="full"
          onClick={() => fileInputRef.current?.click()}
          isLoading={uploading}
          leftIcon={<Icon as={FiCamera} />}
        >
          {currentAvatarUrl ? 'Change' : 'Upload'}
        </Button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileSelect}
          style={{ display: 'none' }}
        />
      </Box>

      <Text fontSize="sm" color="gray.500" textAlign="center">
        Upload a profile picture<br />
        Max size: 2MB • JPG, PNG, GIF, or WebP
      </Text>
    </VStack>
  )
}

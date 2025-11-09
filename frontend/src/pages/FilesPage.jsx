import { useState } from 'react'
import {
  Container,
  Heading,
  VStack,
  Box,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Button,
  Icon,
  HStack,
  Text,
  useToast,
  Badge,
} from '@chakra-ui/react'
import { FiDownload, FiTrash2, FiFile } from 'react-icons/fi'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { FileUpload } from '../components/FileUpload'
import { AvatarUpload } from '../components/AvatarUpload'
import { api } from '../lib/api'

export function FilesPage() {
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()
  const toast = useToast()

  const { data: filesData, isLoading } = useQuery({
    queryKey: ['files', page],
    queryFn: async () => {
      const response = await api.get(`/files?page=${page}&page_size=10`)
      return response
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (fileId) => {
      await api.delete(`/files/${fileId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['files'] })
      toast({
        title: 'File deleted',
        status: 'success',
        duration: 3000,
      })
    },
    onError: (error) => {
      toast({
        title: 'Delete failed',
        description: error.response?.data?.error || 'Failed to delete file',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const handleUploadComplete = () => {
    queryClient.invalidateQueries({ queryKey: ['files'] })
  }

  const handleDownload = (url, filename) => {
    window.open(url, '_blank')
  }

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  }

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <Container maxW="container.xl" py={8}>
      <VStack spacing={8} align="stretch">
        <Heading>File Management</Heading>

        <Tabs colorScheme="blue">
          <TabList>
            <Tab>Upload Files</Tab>
            <Tab>My Files</Tab>
            <Tab>Avatar</Tab>
          </TabList>

          <TabPanels>
            <TabPanel>
              <Box maxW="2xl" mx="auto">
                <FileUpload onUploadComplete={handleUploadComplete} />
              </Box>
            </TabPanel>

            <TabPanel>
              <VStack spacing={4} align="stretch">
                {isLoading ? (
                  <Text>Loading files...</Text>
                ) : filesData?.data?.length === 0 ? (
                  <Box textAlign="center" py={12}>
                    <Icon as={FiFile} boxSize={12} color="gray.300" mb={4} />
                    <Text color="gray.500">No files uploaded yet</Text>
                    <Button
                      mt={4}
                      colorScheme="blue"
                      onClick={() => setPage(1)}
                    >
                      Upload your first file
                    </Button>
                  </Box>
                ) : (
                  <>
                    <Box overflowX="auto">
                      <Table variant="simple">
                        <Thead>
                          <Tr>
                            <Th>File Name</Th>
                            <Th>Type</Th>
                            <Th>Size</Th>
                            <Th>Uploaded</Th>
                            <Th>Actions</Th>
                          </Tr>
                        </Thead>
                        <Tbody>
                          {filesData?.data?.map((file) => (
                            <Tr key={file.id}>
                              <Td>
                                <HStack>
                                  <Icon as={FiFile} color="blue.500" />
                                  <Text>{file.filename}</Text>
                                </HStack>
                              </Td>
                              <Td>
                                <Badge colorScheme="gray" size="sm">
                                  {file.mime_type}
                                </Badge>
                              </Td>
                              <Td>{formatFileSize(file.size)}</Td>
                              <Td>{formatDate(file.created_at)}</Td>
                              <Td>
                                <HStack spacing={2}>
                                  <Button
                                    size="sm"
                                    leftIcon={<FiDownload />}
                                    onClick={() => handleDownload(file.url, file.filename)}
                                  >
                                    Download
                                  </Button>
                                  <Button
                                    size="sm"
                                    colorScheme="red"
                                    variant="ghost"
                                    leftIcon={<FiTrash2 />}
                                    onClick={() => deleteMutation.mutate(file.id)}
                                    isLoading={deleteMutation.isPending}
                                  >
                                    Delete
                                  </Button>
                                </HStack>
                              </Td>
                            </Tr>
                          ))}
                        </Tbody>
                      </Table>
                    </Box>

                    {filesData?.total_pages > 1 && (
                      <HStack justify="center" spacing={2}>
                        <Button
                          size="sm"
                          onClick={() => setPage((p) => Math.max(1, p - 1))}
                          isDisabled={page === 1}
                        >
                          Previous
                        </Button>
                        <Text fontSize="sm">
                          Page {page} of {filesData.total_pages}
                        </Text>
                        <Button
                          size="sm"
                          onClick={() => setPage((p) => p + 1)}
                          isDisabled={page >= filesData.total_pages}
                        >
                          Next
                        </Button>
                      </HStack>
                    )}
                  </>
                )}
              </VStack>
            </TabPanel>

            <TabPanel>
              <Box maxW="md" mx="auto" textAlign="center">
                <AvatarUpload onUploadComplete={handleUploadComplete} />
              </Box>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>
    </Container>
  )
}

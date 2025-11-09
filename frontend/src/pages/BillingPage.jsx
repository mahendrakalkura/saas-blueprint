import { useState } from 'react'
import {
  Container,
  Heading,
  VStack,
  Box,
  Text,
  Button,
  Card,
  Badge,
  HStack,
  Divider,
  useToast,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Spinner,
} from '@chakra-ui/react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'

export function BillingPage() {
  const [loading, setLoading] = useState(false)
  const queryClient = useQueryClient()
  const toast = useToast()

  const { data: subscription, isLoading } = useQuery({
    queryKey: ['subscription'],
    queryFn: async () => {
      try {
        const response = await api.get('/billing/subscription')
        return response
      } catch (error) {
        if (error.response?.status === 404) {
          return null
        }
        throw error
      }
    },
  })

  const cancelMutation = useMutation({
    mutationFn: async (immediately) => {
      await api.post('/billing/subscription/cancel', { immediately })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscription'] })
      toast({
        title: 'Subscription canceled',
        status: 'success',
        duration: 3000,
      })
    },
    onError: (error) => {
      toast({
        title: 'Cancellation failed',
        description: error.response?.data?.error || 'Failed to cancel subscription',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const reactivateMutation = useMutation({
    mutationFn: async () => {
      await api.post('/billing/subscription/reactivate')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscription'] })
      toast({
        title: 'Subscription reactivated',
        status: 'success',
        duration: 3000,
      })
    },
    onError: (error) => {
      toast({
        title: 'Reactivation failed',
        description: error.response?.data?.error || 'Failed to reactivate subscription',
        status: 'error',
        duration: 5000,
      })
    },
  })

  const handleManageBilling = async () => {
    setLoading(true)
    try {
      const response = await api.post('/billing/portal')
      window.location.href = response.url
    } catch (error) {
      toast({
        title: 'Failed to open billing portal',
        description: error.response?.data?.error || 'Please try again',
        status: 'error',
        duration: 5000,
      })
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'active':
        return 'green'
      case 'trialing':
        return 'blue'
      case 'past_due':
        return 'orange'
      case 'canceled':
        return 'red'
      default:
        return 'gray'
    }
  }

  if (isLoading) {
    return (
      <Container maxW="container.lg" py={8}>
        <VStack spacing={8}>
          <Spinner size="xl" />
          <Text>Loading subscription...</Text>
        </VStack>
      </Container>
    )
  }

  return (
    <Container maxW="container.lg" py={8}>
      <VStack spacing={8} align="stretch">
        <Heading>Billing & Subscription</Heading>

        {!subscription ? (
          <Alert status="info">
            <AlertIcon />
            <Box>
              <AlertTitle>No active subscription</AlertTitle>
              <AlertDescription>
                You don't have an active subscription yet.
              </AlertDescription>
            </Box>
          </Alert>
        ) : (
          <Card p={6}>
            <VStack spacing={6} align="stretch">
              <HStack justify="space-between">
                <Box>
                  <Text fontSize="lg" fontWeight="bold" mb={2}>
                    Current Plan
                  </Text>
                  <Badge colorScheme={getStatusColor(subscription.status)} fontSize="md" px={3} py={1}>
                    {subscription.status.toUpperCase()}
                  </Badge>
                </Box>
                <Button
                  onClick={handleManageBilling}
                  isLoading={loading}
                  colorScheme="blue"
                >
                  Manage Billing
                </Button>
              </HStack>

              <Divider />

              <VStack spacing={4} align="stretch">
                {subscription.trial_end && new Date(subscription.trial_end) > new Date() && (
                  <Alert status="info">
                    <AlertIcon />
                    <Box>
                      <AlertTitle>Trial Period</AlertTitle>
                      <AlertDescription>
                        Your trial ends on {formatDate(subscription.trial_end)}
                      </AlertDescription>
                    </Box>
                  </Alert>
                )}

                <Box>
                  <Text fontSize="sm" color="gray.600" mb={1}>
                    Current Period
                  </Text>
                  <Text>
                    {formatDate(subscription.current_period_start)} -{' '}
                    {formatDate(subscription.current_period_end)}
                  </Text>
                </Box>

                {subscription.cancel_at_period_end && (
                  <Alert status="warning">
                    <AlertIcon />
                    <Box flex="1">
                      <AlertTitle>Subscription Ending</AlertTitle>
                      <AlertDescription>
                        Your subscription will end on {formatDate(subscription.current_period_end)}
                      </AlertDescription>
                    </Box>
                    <Button
                      size="sm"
                      onClick={() => reactivateMutation.mutate()}
                      isLoading={reactivateMutation.isPending}
                    >
                      Reactivate
                    </Button>
                  </Alert>
                )}

                {!subscription.cancel_at_period_end && subscription.status === 'active' && (
                  <Box>
                    <Text fontSize="sm" color="gray.600" mb={2}>
                      Cancel Subscription
                    </Text>
                    <HStack spacing={2}>
                      <Button
                        size="sm"
                        colorScheme="red"
                        variant="outline"
                        onClick={() => cancelMutation.mutate(false)}
                        isLoading={cancelMutation.isPending}
                      >
                        Cancel at Period End
                      </Button>
                      <Button
                        size="sm"
                        colorScheme="red"
                        onClick={() => {
                          if (window.confirm('Are you sure you want to cancel immediately?')) {
                            cancelMutation.mutate(true)
                          }
                        }}
                        isLoading={cancelMutation.isPending}
                      >
                        Cancel Immediately
                      </Button>
                    </HStack>
                  </Box>
                )}
              </VStack>
            </VStack>
          </Card>
        )}

        <Box>
          <Heading size="md" mb={4}>
            Need to upgrade or change your plan?
          </Heading>
          <Text color="gray.600" mb={4}>
            Click "Manage Billing" to update your payment method, view invoices, or change your subscription plan.
          </Text>
        </Box>
      </VStack>
    </Container>
  )
}

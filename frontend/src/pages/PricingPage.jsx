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
  List,
  ListItem,
  ListIcon,
  Badge,
  useToast,
  SimpleGrid,
} from '@chakra-ui/react'
import { FiCheck } from 'react-icons/fi'
import { api } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'

// Sample pricing plans - replace with actual Stripe Price IDs
const PRICING_PLANS = [
  {
    name: 'Free',
    price: '$0',
    period: 'forever',
    priceId: null,
    features: [
      'Up to 3 projects',
      '1 GB storage',
      'Basic support',
      'Community access',
    ],
    cta: 'Current Plan',
    highlighted: false,
  },
  {
    name: 'Pro',
    price: '$29',
    period: 'per month',
    priceId: 'price_pro_monthly', // Replace with actual Stripe Price ID
    trialDays: 14,
    features: [
      'Unlimited projects',
      '100 GB storage',
      'Priority support',
      'Advanced analytics',
      'Team collaboration',
      'API access',
    ],
    cta: 'Start 14-day Trial',
    highlighted: true,
  },
  {
    name: 'Enterprise',
    price: '$99',
    period: 'per month',
    priceId: 'price_enterprise_monthly', // Replace with actual Stripe Price ID
    features: [
      'Everything in Pro',
      'Unlimited storage',
      'Dedicated support',
      'Custom integrations',
      'SLA guarantee',
      'Advanced security',
      'On-premise option',
    ],
    cta: 'Start Free Trial',
    highlighted: false,
  },
]

export function PricingPage() {
  const [loading, setLoading] = useState(null)
  const { user } = useAuth()
  const toast = useToast()

  const handleSubscribe = async (plan) => {
    if (!user) {
      toast({
        title: 'Please log in',
        description: 'You need to be logged in to subscribe',
        status: 'warning',
        duration: 5000,
      })
      return
    }

    if (!plan.priceId) {
      toast({
        title: 'Already on free plan',
        status: 'info',
        duration: 3000,
      })
      return
    }

    setLoading(plan.name)
    try {
      const response = await api.post('/billing/checkout', {
        price_id: plan.priceId,
        trial_days: plan.trialDays || 0,
      })

      // Redirect to Stripe Checkout
      window.location.href = response.url
    } catch (error) {
      toast({
        title: 'Failed to create checkout session',
        description: error.response?.data?.error || 'Please try again',
        status: 'error',
        duration: 5000,
      })
    } finally {
      setLoading(null)
    }
  }

  return (
    <Container maxW="container.xl" py={12}>
      <VStack spacing={12}>
        <VStack spacing={4} textAlign="center">
          <Heading size="2xl">Simple, Transparent Pricing</Heading>
          <Text fontSize="lg" color="gray.600" maxW="2xl">
            Choose the plan that's right for you. All plans include a free trial.
          </Text>
        </VStack>

        <SimpleGrid columns={{ base: 1, md: 3 }} spacing={8} w="full">
          {PRICING_PLANS.map((plan) => (
            <Card
              key={plan.name}
              p={8}
              borderWidth={plan.highlighted ? 2 : 1}
              borderColor={plan.highlighted ? 'blue.500' : 'gray.200'}
              position="relative"
              _hover={{
                transform: 'translateY(-4px)',
                shadow: 'xl',
              }}
              transition="all 0.2s"
            >
              {plan.highlighted && (
                <Badge
                  position="absolute"
                  top={-3}
                  right={4}
                  colorScheme="blue"
                  fontSize="sm"
                  px={3}
                  py={1}
                >
                  POPULAR
                </Badge>
              )}

              <VStack spacing={6} align="stretch">
                <Box>
                  <Text fontSize="2xl" fontWeight="bold" mb={2}>
                    {plan.name}
                  </Text>
                  <HStack align="baseline">
                    <Text fontSize="4xl" fontWeight="bold">
                      {plan.price}
                    </Text>
                    <Text color="gray.600">/ {plan.period}</Text>
                  </HStack>
                </Box>

                <Button
                  size="lg"
                  colorScheme={plan.highlighted ? 'blue' : 'gray'}
                  onClick={() => handleSubscribe(plan)}
                  isLoading={loading === plan.name}
                  isDisabled={!plan.priceId && user}
                >
                  {plan.cta}
                </Button>

                <List spacing={4}>
                  {plan.features.map((feature, index) => (
                    <ListItem key={index} display="flex" alignItems="center">
                      <ListIcon as={FiCheck} color="green.500" fontSize="xl" />
                      <Text>{feature}</Text>
                    </ListItem>
                  ))}
                </List>
              </VStack>
            </Card>
          ))}
        </SimpleGrid>

        <Box textAlign="center" pt={8}>
          <Text color="gray.600" mb={4}>
            All plans include our core features and 24/7 support
          </Text>
          <Text fontSize="sm" color="gray.500">
            Need a custom plan? <Text as="span" color="blue.500" cursor="pointer">Contact sales</Text>
          </Text>
        </Box>
      </VStack>
    </Container>
  )
}

import { useState, useEffect, useCallback } from 'react'
import {
  Box,
  IconButton,
  Badge,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  Text,
  VStack,
  HStack,
  Spinner,
  Button,
  Icon,
} from '@chakra-ui/react'
import { FiBell, FiCheck, FiTrash2 } from 'react-icons/fi'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useWebSocket, useWebSocketSubscription } from '../lib/websocket'
import { formatDistanceToNow } from 'date-fns'

export function NotificationBell() {
  const queryClient = useQueryClient()
  const { isConnected } = useWebSocket()

  // Fetch notifications
  const { data: notificationsData, isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const response = await api.get('/notifications?limit=10')
      return response.notifications || []
    },
  })

  // Fetch unread count
  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: async () => {
      const response = await api.get('/notifications/unread-count')
      return response.count || 0
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  })

  // Handle real-time notifications via WebSocket
  const handleNewNotification = useCallback(
    (payload) => {
      // Add new notification to the list
      queryClient.setQueryData(['notifications'], (old) => {
        const notifications = old || []
        return [payload, ...notifications].slice(0, 10)
      })

      // Increment unread count
      queryClient.setQueryData(['notifications', 'unread-count'], (old) => {
        return (old || 0) + 1
      })
    },
    [queryClient]
  )

  useWebSocketSubscription('notification', handleNewNotification)

  // Mark as read mutation
  const markAsReadMutation = useMutation({
    mutationFn: async (notificationId) => {
      await api.post(`/notifications/${notificationId}/read`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications'])
      queryClient.invalidateQueries(['notifications', 'unread-count'])
    },
  })

  // Mark all as read mutation
  const markAllAsReadMutation = useMutation({
    mutationFn: async () => {
      await api.post('/notifications/mark-all-read')
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications'])
      queryClient.invalidateQueries(['notifications', 'unread-count'])
    },
  })

  // Delete notification mutation
  const deleteNotificationMutation = useMutation({
    mutationFn: async (notificationId) => {
      await api.delete(`/notifications/${notificationId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications'])
      queryClient.invalidateQueries(['notifications', 'unread-count'])
    },
  })

  const handleMarkAsRead = (e, notificationId) => {
    e.stopPropagation()
    markAsReadMutation.mutate(notificationId)
  }

  const handleDelete = (e, notificationId) => {
    e.stopPropagation()
    deleteNotificationMutation.mutate(notificationId)
  }

  const unreadCount = unreadCountData || 0
  const notifications = notificationsData || []

  return (
    <Menu>
      <MenuButton
        as={IconButton}
        icon={
          <Box position="relative">
            <Icon as={FiBell} boxSize={5} />
            {unreadCount > 0 && (
              <Badge
                position="absolute"
                top="-8px"
                right="-8px"
                colorScheme="red"
                borderRadius="full"
                fontSize="xs"
                minW="18px"
                h="18px"
                display="flex"
                alignItems="center"
                justifyContent="center"
              >
                {unreadCount > 99 ? '99+' : unreadCount}
              </Badge>
            )}
            {isConnected && (
              <Box
                position="absolute"
                bottom="0"
                right="0"
                w="8px"
                h="8px"
                bg="green.500"
                borderRadius="full"
                border="2px solid white"
              />
            )}
          </Box>
        }
        variant="ghost"
        aria-label="Notifications"
      />
      <MenuList maxW="400px" maxH="500px" overflowY="auto">
        <HStack justify="space-between" px={3} py={2}>
          <Text fontWeight="bold">Notifications</Text>
          {unreadCount > 0 && (
            <Button
              size="xs"
              variant="ghost"
              onClick={() => markAllAsReadMutation.mutate()}
              isLoading={markAllAsReadMutation.isPending}
            >
              Mark all as read
            </Button>
          )}
        </HStack>
        <MenuDivider />

        {isLoading ? (
          <Box textAlign="center" py={8}>
            <Spinner size="md" />
          </Box>
        ) : notifications.length === 0 ? (
          <Box textAlign="center" py={8} color="gray.500">
            <Icon as={FiBell} boxSize={8} mb={2} />
            <Text>No notifications yet</Text>
          </Box>
        ) : (
          <VStack gap={0} align="stretch">
            {notifications.map((notification) => (
              <MenuItem
                key={notification.id}
                onClick={() => {
                  if (!notification.is_read) {
                    markAsReadMutation.mutate(notification.id)
                  }
                }}
                bg={notification.is_read ? 'transparent' : 'blue.50'}
                _hover={{ bg: notification.is_read ? 'gray.50' : 'blue.100' }}
                py={3}
                px={4}
              >
                <VStack align="stretch" gap={1} flex={1}>
                  <HStack justify="space-between">
                    <Text fontWeight="semibold" fontSize="sm">
                      {notification.title}
                    </Text>
                    <HStack gap={1}>
                      {!notification.is_read && (
                        <IconButton
                          size="xs"
                          variant="ghost"
                          icon={<FiCheck />}
                          onClick={(e) => handleMarkAsRead(e, notification.id)}
                          aria-label="Mark as read"
                        />
                      )}
                      <IconButton
                        size="xs"
                        variant="ghost"
                        colorScheme="red"
                        icon={<FiTrash2 />}
                        onClick={(e) => handleDelete(e, notification.id)}
                        aria-label="Delete"
                      />
                    </HStack>
                  </HStack>
                  <Text fontSize="sm" color="gray.600">
                    {notification.message}
                  </Text>
                  <Text fontSize="xs" color="gray.500">
                    {formatDistanceToNow(new Date(notification.created_at), {
                      addSuffix: true,
                    })}
                  </Text>
                </VStack>
              </MenuItem>
            ))}
          </VStack>
        )}
      </MenuList>
    </Menu>
  )
}

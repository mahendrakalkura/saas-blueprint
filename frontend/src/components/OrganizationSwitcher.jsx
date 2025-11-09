import { useState } from 'react'
import {
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  Button,
  HStack,
  Text,
  Icon,
  Spinner,
} from '@chakra-ui/react'
import { FiChevronDown, FiPlus, FiBuilding } from 'react-icons/fi'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

export function OrganizationSwitcher({ currentOrgId }) {
  const navigate = useNavigate()

  const { data: orgsData, isLoading } = useQuery({
    queryKey: ['organizations'],
    queryFn: async () => {
      try {
        const response = await api.get('/organizations')
        return response.organizations || []
      } catch (error) {
        return []
      }
    },
  })

  const currentOrg = orgsData?.find((org) => org.id === currentOrgId)

  if (isLoading) {
    return <Spinner size="sm" />
  }

  return (
    <Menu>
      <MenuButton
        as={Button}
        rightIcon={<FiChevronDown />}
        leftIcon={<FiBuilding />}
        variant="outline"
        size="sm"
      >
        {currentOrg?.name || 'Select Organization'}
      </MenuButton>
      <MenuList>
        {orgsData?.length === 0 ? (
          <MenuItem onClick={() => navigate('/organizations')}>
            <HStack>
              <Icon as={FiPlus} />
              <Text>Create your first organization</Text>
            </HStack>
          </MenuItem>
        ) : (
          <>
            {orgsData?.map((org) => (
              <MenuItem
                key={org.id}
                onClick={() => navigate(`/organizations/${org.id}`)}
                bg={currentOrgId === org.id ? 'blue.50' : 'transparent'}
              >
                <HStack spacing={3}>
                  <Icon as={FiBuilding} />
                  <Text>{org.name}</Text>
                </HStack>
              </MenuItem>
            ))}
            <MenuDivider />
            <MenuItem onClick={() => navigate('/organizations')}>
              <HStack>
                <Icon as={FiPlus} />
                <Text>Create new organization</Text>
              </HStack>
            </MenuItem>
          </>
        )}
      </MenuList>
    </Menu>
  )
}

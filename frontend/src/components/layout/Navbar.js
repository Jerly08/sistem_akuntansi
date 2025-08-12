'use client';

import React from 'react';
import {
  Box,
  Flex,
  Avatar,
  HStack,
  VStack,
  Text,
  IconButton,
  Button,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  useColorModeValue,
  Stack,
} from '@chakra-ui/react';
import { FiMenu, FiChevronDown, FiX } from 'react-icons/fi';
import { useAuth } from '@/contexts/AuthContext';
import ApprovalNotifications from '@/components/notification/ApprovalNotifications';

const NavLink = ({ children }) => (
  <Box
    px={2}
    py={1}
    rounded="md"
    _hover={{
      textDecoration: 'none',
      bg: useColorModeValue('gray.200', 'gray.700'),
    }}
    href="#">
    {children}
  </Box>
);

const Navbar = ({ onMenuClick, sidebarCollapsed, onToggleSidebar, isMenuOpen = false }) => {
  const { user, logout } = useAuth();

  return (
    <>
      <Box bg={useColorModeValue('white', 'gray.900')} px={4} shadow="sm" borderBottom="1px" borderColor={useColorModeValue('gray.200', 'gray.700')}>
        <Flex h={16} alignItems="center" justifyContent="space-between">
          <IconButton
            size="md"
            icon={isMenuOpen ? <FiX /> : <FiMenu />}
            aria-label={isMenuOpen ? 'Close menu' : 'Open menu'}
            display={{ base: 'flex', md: 'none' }}
            onClick={onMenuClick}
            variant="ghost"
            colorScheme="brand"
            _hover={{
              bg: useColorModeValue('brand.50', 'gray.700'),
            }}
          />
          
          <HStack spacing={8} alignItems="center" display={{ base: 'none', md: 'flex' }}>
            <Text fontSize="lg" fontWeight="bold" color={useColorModeValue('gray.800', 'white')}>
              Accounting System
            </Text>
          </HStack>
          
          <Flex alignItems="center">
            <Stack direction="row" spacing={3}>
              {/* Approval notifications dropdown with badge */}
              <Box display={{ base: 'none', sm: 'flex' }}>
                <ApprovalNotifications />
              </Box>
              <Menu>
                <MenuButton
                  as={Button}
                  rounded="full"
                  variant="link"
                  cursor="pointer"
                  minW={0}>
                  <HStack>
                    <Avatar
                      size="sm"
                      src={user?.avatarUrl}
                    />
                    <VStack
                      display={{ base: 'none', md: 'flex' }}
                      alignItems="flex-start"
                      spacing="1px"
                      ml="2">
                      <Text fontSize="sm">{user?.name}</Text>
                      <Text fontSize="xs" color="gray.600">
                        {user?.role}
                      </Text>
                    </VStack>
                    <Box display={{ base: 'none', md: 'flex' }}>
                      <FiChevronDown />
                    </Box>
                  </HStack>
                </MenuButton>
                <MenuList>
                  <MenuItem>Profile</MenuItem>
                  <MenuItem>Settings</MenuItem>
                  <MenuDivider />
                  <MenuItem onClick={logout}>Sign out</MenuItem>
                </MenuList>
              </Menu>
            </Stack>
          </Flex>
        </Flex>
      </Box>
    </>
  );
};

export default Navbar;

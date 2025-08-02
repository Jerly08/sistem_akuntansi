'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import {
  Box,
  CloseButton,
  Flex,
  Icon,
  useColorModeValue,
  Text,
  Drawer,
  DrawerContent,
  useDisclosure,
  BoxProps,
  FlexProps,
  Collapse,
  DrawerOverlay,
  DrawerCloseButton,
} from '@chakra-ui/react';
import {
  FiTrendingUp,
  FiCompass,
  FiStar,
  FiBell,
  FiSettings,
  FiMenu,
  FiShoppingCart,
  FiUsers,
  FiDollarSign,
  FiBarChart,
  FiFileText,
  FiHome,
  FiLayers,
  FiUser,
} from 'react-icons/fi';
import { useAuth } from '@/contexts/AuthContext';

const MenuGroups = [
  {
    title: 'Dashboard',
    items: [
      { name: 'Dashboard', icon: FiHome, href: '/dashboard', roles: ['ADMIN', 'FINANCE', 'INVENTORY_MANAGER', 'DIRECTOR', 'EMPLOYEE'] },
    ]
  },
  {
    title: 'Master Data',
    items: [
      { name: 'Accounts', icon: FiFileText, href: '/accounts', roles: ['ADMIN', 'FINANCE'] },
      { name: 'Products', icon: FiLayers, href: '/products', roles: ['ADMIN', 'INVENTORY_MANAGER', 'EMPLOYEE'] },
      { name: 'Contacts', icon: FiUsers, href: '/contacts', roles: ['ADMIN', 'FINANCE', 'INVENTORY_MANAGER', 'EMPLOYEE'] },
      { name: 'Assets', icon: FiStar, href: '/assets', roles: ['ADMIN', 'FINANCE', 'DIRECTOR'] },
    ]
  },
  {
    title: 'Financial',
    items: [
      { name: 'Sales', icon: FiDollarSign, href: '/sales', roles: ['ADMIN', 'FINANCE', 'DIRECTOR', 'EMPLOYEE'] },
      { name: 'Purchases', icon: FiShoppingCart, href: '/purchases', roles: ['ADMIN', 'FINANCE', 'INVENTORY_MANAGER', 'EMPLOYEE'] },
      { name: 'Payments', icon: FiTrendingUp, href: '/payments', roles: ['ADMIN', 'FINANCE', 'DIRECTOR'] },
      { name: 'Cash & Bank', icon: FiCompass, href: '/cash-bank', roles: ['ADMIN', 'FINANCE', 'DIRECTOR'] },
    ]
  },
  {
    title: 'Reports',
    items: [
      { name: 'Reports', icon: FiBarChart, href: '/reports', roles: ['ADMIN', 'FINANCE', 'DIRECTOR'] },
    ]
  },
  {
    title: 'System',
    items: [
      { name: 'Users', icon: FiUser, href: '/users', roles: ['ADMIN'] },
      { name: 'Settings', icon: FiSettings, href: '/settings', roles: ['ADMIN'] },
    ]
  },
];

export default function Sidebar({ isOpen, onClose, display, width, collapsed, onToggleCollapse, variant, ...rest }) {
  const { user } = useAuth();
  const pathname = usePathname();
  const router = useRouter();

  // Filter menu groups based on user role
  const filteredGroups = MenuGroups.map(group => ({
    ...group,
    items: group.items.filter(item => user && item.roles.includes(user.role))
  })).filter(group => group.items.length > 0);

  const SidebarContent = ({ onClose, ...rest }) => {
    return (
      <Box
        transition="3s ease"
        bg={useColorModeValue('white', 'gray.900')}
        borderRight="1px"
        borderRightColor={useColorModeValue('gray.200', 'gray.700')}
        w={{ base: 'full', md: width || 60 }}
        pos="fixed"
        h="full"
        overflowY="auto"
        zIndex={1000}
        {...rest}>
        <Flex h="20" alignItems="center" mx="8" justifyContent="space-between">
          <Text fontSize="xl" fontFamily="Inter" fontWeight="bold" color="brand.500">
            Accounting App
          </Text>
          <CloseButton display={{ base: 'flex', md: 'none' }} onClick={onClose} />
        </Flex>
        
        {filteredGroups.map((group, index) => (
          <Box key={group.title} mb={6}>
            <Text
              fontSize="xs"
              fontWeight="semibold"
              color={useColorModeValue('gray.500', 'gray.400')}
              textTransform="uppercase"
              mx="4"
              mb="3"
              letterSpacing="wider"
            >
              {group.title}
            </Text>
            {group.items.map((link) => (
              <NavItem key={link.name} icon={link.icon} href={link.href} isActive={pathname === link.href}>
                {link.name}
              </NavItem>
            ))}
          </Box>
        ))}
      </Box>
    );
  };

  const NavItem = ({ icon, children, href, isActive, ...rest }) => {
    const handleClick = (e) => {
      e.preventDefault();
      console.log('NavItem clicked:', href);
      router.push(href);
    };

    return (
      <Flex
        onClick={handleClick}
        align="center"
        p="3"
        mx="4"
        borderRadius="lg"
        role="group"
        cursor="pointer"
        bg={isActive ? 'brand.500' : 'transparent'}
        color={isActive ? 'white' : useColorModeValue('gray.700', 'white')}
        _hover={{
          bg: isActive ? 'brand.600' : useColorModeValue('brand.50', 'gray.700'),
          color: isActive ? 'white' : useColorModeValue('brand.700', 'white'),
        }}
        transition="all 0.2s"
        fontWeight="medium"
        fontSize="sm"
        pointerEvents="auto"
        zIndex={10}
        {...rest}>
        {icon && (
          <Icon
            mr="3"
            fontSize="18"
            color={isActive ? 'white' : useColorModeValue('gray.500', 'gray.400')}
            _groupHover={{
              color: isActive ? 'white' : useColorModeValue('brand.600', 'white'),
            }}
            as={icon}
          />
        )}
        {children}
      </Flex>
    );
  };

  if (variant === 'drawer') {
    return (
      <Drawer
        autoFocus={false}
        isOpen={isOpen}
        placement="left"
        onClose={onClose}
        returnFocusOnClose={false}
        onOverlayClick={onClose}
        size="full">
        <DrawerOverlay />
        <DrawerContent>
          <SidebarContent onClose={onClose} />
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <SidebarContent
      onClose={() => onClose}
      display={display}
    />
  );
}

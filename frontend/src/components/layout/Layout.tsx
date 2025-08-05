'use client';

import React, { useState } from 'react';
import { useAuthService } from '@/hooks/useAuthService';
import {
  Box,
  Flex,
  useDisclosure,
  useBreakpointValue,
} from '@chakra-ui/react';
import Navbar from './Navbar';
import Sidebar from './SidebarNew';
import ProtectedRoute from '../auth/ProtectedRoute';
import { UserRole } from '@/contexts/AuthContext';

interface LayoutProps {
  children: React.ReactNode;
  allowedRoles?: UserRole[];
}
const Layout: React.FC<LayoutProps> = ({ children, allowedRoles = [] }) => {
  useAuthService(); // Setup unauthorized handler

  const { isOpen, onOpen, onClose } = useDisclosure();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  
  // Responsive breakpoints
  const sidebarDisplay = useBreakpointValue({ base: 'none', md: 'block' });
  const sidebarWidth = sidebarCollapsed ? '80px' : '280px';
  
  // Handle menu toggle for mobile
  const handleMenuToggle = () => {
    if (isOpen) {
      onClose();
    } else {
      onOpen();
    }
  };

  return (
    <ProtectedRoute allowedRoles={allowedRoles}>
      <Box minH="100vh" bg="gray.50">
        {/* Sidebar */}
        <Sidebar
          isOpen={isOpen}
          onClose={onClose}
          display={{ base: 'none', md: 'block' }}
          width={sidebarWidth}
          collapsed={sidebarCollapsed}
          onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        />
        
        {/* Mobile Drawer */}
        <Sidebar
          isOpen={isOpen}
          onClose={onClose}
          display={{ base: 'block', md: 'none' }}
          variant="drawer"
        />
        
        {/* Main Content Area */}
        <Flex
          direction="column"
          ml={{ base: 0, md: sidebarWidth }}
          transition="margin-left 0.3s ease"
        >
          {/* Navbar */}
          <Navbar 
            onMenuClick={handleMenuToggle}
            sidebarCollapsed={sidebarCollapsed}
            onToggleSidebar={() => setSidebarCollapsed(!sidebarCollapsed)}
            isMenuOpen={isOpen}
          />
          
          {/* Main Content */}
          <Box
            as="main"
            flex="1"
            p={6}
            pt={8}
            bg="gray.50"
            minH="calc(100vh - 64px)"
          >
            {children}
          </Box>
        </Flex>
      </Box>
    </ProtectedRoute>
  );
};

export default Layout;

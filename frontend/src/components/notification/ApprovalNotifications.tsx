'use client';

import React, { useEffect, useState } from 'react';
import {
  Box,
  IconButton,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Divider,
  Text,
  HStack,
  VStack,
  Badge,
  Spinner,
} from '@chakra-ui/react';
import { FiBell, FiCheckCircle, FiXCircle, FiClock, FiShoppingCart } from 'react-icons/fi';
import approvalService from '@/services/approvalService';

interface NotificationItem {
  id: number;
  type: string;
  title: string;
  message: string;
  priority: string;
  is_read: boolean;
  created_at: string;
  data?: string;
}

const ApprovalNotifications: React.FC = () => {
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchNotifications();
    fetchUnreadCount();
    const interval = setInterval(() => fetchUnreadCount(), 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchNotifications = async () => {
    try {
      setLoading(true);
      const response = await approvalService.getNotifications({ limit: 20, type: 'approval' });
      setNotifications(response.notifications || []);
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchUnreadCount = async () => {
    try {
      const response = await approvalService.getUnreadNotificationCount();
      setUnreadCount(response.count || 0);
    } catch (error) {
      console.error('Failed to fetch unread count:', error);
    }
  };

  const handleMarkAsRead = async (notificationId: number) => {
    try {
      await approvalService.markNotificationAsRead(notificationId);
      setNotifications(prev => prev.map(n => (n.id === notificationId ? { ...n, is_read: true } : n)));
      fetchUnreadCount();
    } catch (error) {
      console.error('Failed to mark notification as read:', error);
    }
  };

  const getIcon = (type: string) => {
    switch (type) {
      case 'approval_pending':
        return <FiClock color="#dd6b20" />; // orange.400
      case 'approval_approved':
        return <FiCheckCircle color="#38a169" />; // green.500
      case 'approval_rejected':
        return <FiXCircle color="#e53e3e" />; // red.500
      default:
        return <FiShoppingCart color="#3182ce" />; // blue.600
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);
    if (diffInHours < 1) return 'Just now';
    if (diffInHours < 24) return `${Math.floor(diffInHours)}h ago`;
    return date.toLocaleDateString('id-ID', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  const priorityColor = (priority: string) => {
    const p = (priority || '').toLowerCase();
    if (p === 'high' || p === 'urgent') return 'red.500';
    if (p === 'normal') return 'orange.400';
    if (p === 'low') return 'blue.500';
    return 'gray.500';
  };

  return (
    <Menu placement="bottom-end" autoSelect={false} onOpen={fetchNotifications}>
      <MenuButton as={IconButton} aria-label="Notifications" variant="ghost">
        <Box position="relative">
          <FiBell />
          {unreadCount > 0 && (
            <Badge colorScheme="red" borderRadius="full" position="absolute" top={-2} right={-2} fontSize="0.6em" px={2}>
              {unreadCount}
            </Badge>
          )}
        </Box>
      </MenuButton>
      <MenuList minW="380px" maxW="90vw">
        <Box px={3} pt={2} pb={1}>
          <Text fontSize="md" fontWeight="bold">Notifications</Text>
          {unreadCount > 0 && (
            <Text fontSize="xs" color="gray.500">{unreadCount} unread notification{unreadCount > 1 ? 's' : ''}</Text>
          )}
        </Box>
        <Divider />
        {loading ? (
          <MenuItem isDisabled>
            <HStack spacing={2}>
              <Spinner size="xs" />
              <Text fontSize="sm" color="gray.500">Loading notifications...</Text>
            </HStack>
          </MenuItem>
        ) : notifications.length === 0 ? (
          <MenuItem isDisabled>
            <Text fontSize="sm" color="gray.500">No notifications</Text>
          </MenuItem>
        ) : (
          <Box maxH="350px" overflowY="auto">
            {notifications.map((n) => (
              <Box key={n.id}>
                <MenuItem
                  onClick={() => {
                    if (!n.is_read) handleMarkAsRead(n.id);
                  }}
                >
                  <HStack align="start" spacing={3} w="full">
                    <Box pt={1}>{getIcon(n.type)}</Box>
                    <VStack align="start" spacing={0} flex={1}>
                      <HStack justify="space-between" w="full">
                        <Text fontSize="sm" fontWeight={n.is_read ? 'normal' : 'semibold'}>{n.title}</Text>
                        {!n.is_read && <Box w={2} h={2} bg="blue.500" borderRadius="full" />}
                      </HStack>
                      <Text fontSize="sm" color="gray.600" noOfLines={2}>{n.message}</Text>
                      <HStack spacing={2} pt={1}>
                        <Text fontSize="xs" color="gray.500">{formatDate(n.created_at)}</Text>
                        <Badge colorScheme="gray" bg={priorityColor(n.priority)} color="white">{n.priority}</Badge>
                      </HStack>
                    </VStack>
                  </HStack>
                </MenuItem>
                <Divider />
              </Box>
            ))}
          </Box>
        )}
        {notifications.length > 0 && (
          <>
            <Divider />
            <MenuItem onClick={() => { /* future navigation */ }}>
              <Text w="full" textAlign="center" fontSize="sm" color="blue.600">View All Notifications</Text>
            </MenuItem>
          </>
        )}
      </MenuList>
    </Menu>
  );
};

export default ApprovalNotifications;

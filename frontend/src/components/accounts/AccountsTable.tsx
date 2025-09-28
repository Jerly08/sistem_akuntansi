'use client';

import React from 'react';
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Box,
  Text,
  Badge,
  Button,
  HStack,
  Flex,
  useColorMode,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiEdit, FiTrash2 } from 'react-icons/fi';
import { Account } from '@/types/account';
import { useModulePermissions } from '@/hooks/usePermissions';
import { useAuth } from '@/contexts/AuthContext';
import accountService from '@/services/accountService';

interface AccountsTableProps {
  accounts: Account[];
  onEdit: (account: Account) => void;
  onDelete: (account: Account) => void;
  onAdminDelete?: (account: Account) => void; // Admin-only delete for header accounts
}

const AccountsTable: React.FC<AccountsTableProps> = ({ accounts, onEdit, onDelete, onAdminDelete }) => {
  // Permission checking
  const { canEdit, canDelete } = useModulePermissions('accounts');
  const { user } = useAuth();
  const isAdmin = user?.role?.toLowerCase() === 'admin';
  
  // Theme-aware colors
  const tableBg = useColorModeValue('white', 'var(--bg-secondary)');
  const theadBg = useColorModeValue('#F7FAFC', 'var(--bg-tertiary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.600', 'var(--text-secondary)');
  const primaryTextColor = useColorModeValue('gray.700', 'var(--text-primary)');
  const hoverBg = useColorModeValue('gray.50', 'var(--bg-tertiary)');
  const editHoverBg = useColorModeValue('blue.50', 'rgba(77, 171, 247, 0.1)');
  const deleteHoverBg = useColorModeValue('red.50', 'rgba(255, 107, 107, 0.1)');
  
  // Helper function to get balance for display
  const getDisplayBalance = (account: Account): number => {
    if (account.is_header && account.total_balance !== undefined) {
      return account.total_balance;
    }
    return account.balance;
  };

  return (
    <Box 
      bg={tableBg} 
      borderRadius="lg" 
      overflow="hidden" 
      border="1px" 
      borderColor={borderColor}
      className="table-container"
      transition="all 0.3s ease"
    >
      <Table size="md" variant="simple" className="table">
        <Thead bg={theadBg}>
          <Tr>
            <Th 
              color={textColor}
              fontWeight="500" 
              fontSize="sm" 
              textTransform="none"
              px={6} 
              py={4}
              borderColor={borderColor}
            >
              Code
            </Th>
            <Th 
              color={textColor}
              fontWeight="500" 
              fontSize="sm" 
              textTransform="none"
              px={6} 
              py={4}
              borderColor={borderColor}
            >
              Name
            </Th>
            <Th 
              color={textColor}
              fontWeight="500" 
              fontSize="sm" 
              textTransform="none"
              px={6} 
              py={4}
              borderColor={borderColor}
            >
              Type
            </Th>
            <Th 
              color={textColor}
              fontWeight="500" 
              fontSize="sm" 
              textTransform="none"
              px={6} 
              py={4}
              borderColor={borderColor}
            >
              Balance
            </Th>
            <Th 
              color={textColor}
              fontWeight="500" 
              fontSize="sm" 
              textTransform="none"
              px={6} 
              py={4}
              borderColor={borderColor}
            >
              Status
            </Th>
            <Th 
              color={textColor}
              fontWeight="500" 
              fontSize="sm" 
              textTransform="none"
              px={6} 
              py={4}
              borderColor={borderColor}
              textAlign="center"
            >
              Actions
            </Th>
          </Tr>
        </Thead>
        <Tbody>
          {accounts.map((account, index) => (
            <Tr key={account.id} _hover={{ bg: hoverBg }} transition="all 0.2s ease">
              <Td px={6} py={4} borderColor={borderColor}>
                <Text fontSize="sm" fontFamily="monospace" color={textColor}>
                  {account.code}
                </Text>
              </Td>
              <Td px={6} py={4} borderColor={borderColor}>
                <Flex align="center">
                  <Box w={`${(account.hierarchyLevel || 0) * 20}px`} />
                  <Text 
                    fontWeight={account.is_header ? '600' : '400'}
                    color={account.is_header ? primaryTextColor : textColor}
                    fontSize="sm"
                  >
                    {account.name}
                  </Text>
                </Flex>
              </Td>
              <Td px={6} py={4} borderColor={borderColor}>
                <Text fontSize="sm" color={textColor}>
                  {accountService.getAccountTypeLabel(account.type, true)}
                </Text>
              </Td>
              <Td px={6} py={4} borderColor={borderColor}>
                <Text 
                  fontSize="sm"
                  fontWeight={account.is_header ? '600' : '400'}
                  color={primaryTextColor}
                >
                  {accountService.formatBalance(getDisplayBalance(account), 'IDR', account.code, account.type)}
                </Text>
              </Td>
              <Td px={6} py={4} borderColor={borderColor}>
                <Badge 
                  colorScheme={account.is_active ? 'green' : 'gray'}
                  variant="subtle"
                  fontSize="xs"
                  px={2}
                  py={1}
                  borderRadius="md"
                >
                  {account.is_active ? 'ACTIVE' : 'INACTIVE'}
                </Badge>
              </Td>
              <Td px={6} py={4} borderColor={borderColor}>
                {/* Regular accounts - show normal edit/delete */}
                {!account.is_header && (canEdit || canDelete) && (
                  <HStack spacing={1} justify="center">
                    {canEdit && (
                      <Button
                        size="sm"
                        variant="ghost"
                        colorScheme="blue"
                        onClick={() => onEdit(account)}
                        px={2}
                        _hover={{ bg: editHoverBg }}
                        transition="all 0.2s ease"
                      >
                        <FiEdit size={16} />
                      </Button>
                    )}
                    {canDelete && (
                      <Button
                        size="sm"
                        variant="ghost"
                        colorScheme="red"
                        onClick={() => onDelete(account)}
                        isDisabled={!account.is_active}
                        px={2}
                        _hover={{ bg: deleteHoverBg }}
                        transition="all 0.2s ease"
                      >
                        <FiTrash2 size={16} />
                      </Button>
                    )}
                  </HStack>
                )}
                
                {/* Header accounts - show admin-only delete (no edit) */}
                {account.is_header && isAdmin && onAdminDelete && (
                  <HStack spacing={1} justify="center">
                    <Button
                      size="sm"
                      variant="ghost"
                      colorScheme="red"
                      onClick={() => onAdminDelete(account)}
                      px={2}
                      _hover={{ bg: deleteHoverBg }}
                      transition="all 0.2s ease"
                    >
                      <FiTrash2 size={16} />
                    </Button>
                  </HStack>
                )}
                
                {/* No actions available */}
                {(!account.is_header && !canEdit && !canDelete) && (
                  <Text fontSize="xs" color="gray.400" textAlign="center">
                    No actions available
                  </Text>
                )}
                {(account.is_header && !isAdmin) && (
                  <Text fontSize="xs" color="gray.400" textAlign="center">
                    Admin only
                  </Text>
                )}
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </Box>
  );
};

export default AccountsTable;

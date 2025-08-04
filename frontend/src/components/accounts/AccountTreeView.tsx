'use client';

import React, { useState } from 'react';
import {
  Box,
  Flex,
  Text,
  Icon,
  Badge,
  Collapse,
  useDisclosure,
  Button,
  HStack,
} from '@chakra-ui/react';
import { 
  FiChevronRight, 
  FiChevronDown, 
  FiEdit, 
  FiTrash2,
  FiFolder,
  FiFile 
} from 'react-icons/fi';
import { Account } from '@/types/account';
import accountService from '@/services/accountService';

interface AccountTreeViewProps {
  accounts: Account[];
  onEdit?: (account: Account) => void;
  onDelete?: (account: Account) => void;
  showActions?: boolean;
  showBalance?: boolean;
}

interface TreeNodeProps {
  account: Account;
  level: number;
  onEdit?: (account: Account) => void;
  onDelete?: (account: Account) => void;
  showActions?: boolean;
  showBalance?: boolean;
}

const TreeNode: React.FC<TreeNodeProps> = ({
  account,
  level,
  onEdit,
  onDelete,
  showActions = true,
  showBalance = true,
}) => {
  const { isOpen, onToggle } = useDisclosure();
  const hasChildren = account.children && account.children.length > 0;
  const indentWidth = level * 20;

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    onEdit?.(account);
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete?.(account);
  };

  return (
    <Box>
      <Flex
        align="center"
        p={2}
        pl={`${indentWidth + 8}px`}
        _hover={{ bg: 'gray.50' }}
        cursor={hasChildren ? 'pointer' : 'default'}
        onClick={hasChildren ? onToggle : undefined}
        borderRadius="md"
      >
        {/* Expand/Collapse Icon */}
        <Box w="20px" mr={2}>
          {hasChildren ? (
            <Icon
              as={isOpen ? FiChevronDown : FiChevronRight}
              color="gray.500"
            />
          ) : null}
        </Box>

        {/* Account Icon */}
        <Icon
          as={account.is_header ? FiFolder : FiFile}
          color={account.is_active ? accountService.getAccountTypeColor(account.type) : 'gray.400'}
          mr={2}
        />

        {/* Account Info */}
        <Flex flex={1} align="center" justify="space-between">
          <HStack spacing={3}>
            <Text
              fontWeight={account.is_header ? 'bold' : 'normal'}
              color={account.is_active ? 'gray.800' : 'gray.400'}
              fontSize={account.level === 1 ? 'md' : 'sm'}
            >
              {account.code} - {account.name}
            </Text>
            
            <Badge
              colorScheme={accountService.getAccountTypeColor(account.type)}
              size="sm"
              variant="subtle"
            >
              {accountService.getAccountTypeLabel(account.type)}
            </Badge>

            {!account.is_active && (
              <Badge colorScheme="gray" size="sm">
                Inactive
              </Badge>
            )}
          </HStack>

          <HStack spacing={3}>
            {showBalance && !account.is_header && (
              <Text
                fontSize="sm"
                fontWeight="medium"
                color={account.balance >= 0 ? 'green.600' : 'red.600'}
              >
                {accountService.formatBalance(account.balance)}
              </Text>
            )}

            {showActions && (
              <HStack spacing={1}>
                <Button
                  size="xs"
                  variant="ghost"
                  leftIcon={<FiEdit />}
                  onClick={handleEdit}
                  isDisabled={!account.is_active}
                >
                  Edit
                </Button>
                <Button
                  size="xs"
                  variant="ghost"
                  colorScheme="red"
                  leftIcon={<FiTrash2 />}
                  onClick={handleDelete}
                  isDisabled={!account.is_active || hasChildren}
                >
                  Delete
                </Button>
              </HStack>
            )}
          </HStack>
        </Flex>
      </Flex>

      {/* Children */}
      {hasChildren && (
        <Collapse in={isOpen}>
          <Box>
            {account.children?.map((child) => (
              <TreeNode
                key={child.id}
                account={child}
                level={level + 1}
                onEdit={onEdit}
                onDelete={onDelete}
                showActions={showActions}
                showBalance={showBalance}
              />
            ))}
          </Box>
        </Collapse>
      )}
    </Box>
  );
};

const AccountTreeView: React.FC<AccountTreeViewProps> = ({
  accounts,
  onEdit,
  onDelete,
  showActions = true,
  showBalance = true,
}) => {
  // Build hierarchical structure
  const buildTree = (accounts: Account[]): Account[] => {
    const accountMap = new Map<number, Account>();
    const rootAccounts: Account[] = [];

    // Create a map of all accounts
    accounts.forEach(account => {
      accountMap.set(account.id, { ...account, children: [] });
    });

    // Build the tree structure
    accounts.forEach(account => {
      const accountWithChildren = accountMap.get(account.id)!;
      
      if (account.parent_id) {
        const parent = accountMap.get(account.parent_id);
        if (parent) {
          if (!parent.children) parent.children = [];
          parent.children.push(accountWithChildren);
        }
      } else {
        rootAccounts.push(accountWithChildren);
      }
    });

    // Sort accounts by code
    const sortByCode = (accounts: Account[]): Account[] => {
      accounts.sort((a, b) => a.code.localeCompare(b.code));
      accounts.forEach(account => {
        if (account.children) {
          account.children = sortByCode(account.children);
        }
      });
      return accounts;
    };

    return sortByCode(rootAccounts);
  };

  const treeData = buildTree(accounts);

  if (accounts.length === 0) {
    return (
      <Box p={4} textAlign="center" color="gray.500">
        <Text>No accounts found</Text>
      </Box>
    );
  }

  return (
    <Box>
      {treeData.map((account) => (
        <TreeNode
          key={account.id}
          account={account}
          level={0}
          onEdit={onEdit}
          onDelete={onDelete}
          showActions={showActions}
          showBalance={showBalance}
        />
      ))}
    </Box>
  );
};

export default AccountTreeView;

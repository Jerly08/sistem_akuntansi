'use client';

import React from 'react';
import {
  Box,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  Text,
  Flex,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  IconButton,
  HStack,
  Spinner,
  useColorModeValue,
  TableContainer,
  Card,
  CardHeader,
  CardBody,
  Heading,
} from '@chakra-ui/react';
import {
  FiMoreVertical,
  FiEye,
  FiEdit,
  FiCheck,
  FiX,
  FiAlertCircle,
  FiClock,
  FiTrash2,
} from 'react-icons/fi';
import { Purchase } from '@/services/purchaseService';

interface PurchaseTableProps {
  purchases: Purchase[];
  loading: boolean;
  onViewDetails: (purchase: Purchase) => void;
  onEdit?: (purchase: Purchase) => void;
  onSubmitForApproval?: (purchaseId: number) => void;
  onDelete?: (purchaseId: number) => void;
  renderActions?: (purchase: Purchase) => React.ReactNode;
  title?: string;
  formatCurrency: (amount: number) => string;
  formatDate: (date: string) => string;
  canEdit?: boolean;
  canDelete?: boolean;
  userRole?: string;
}

const EnhancedPurchaseTable: React.FC<PurchaseTableProps> = ({
  purchases,
  loading,
  onViewDetails,
  onEdit,
  onSubmitForApproval,
  onDelete,
  renderActions,
  title = 'Purchase Transactions',
  formatCurrency,
  formatDate,
  canEdit = false,
  canDelete = false,
  userRole,
}) => {
  // Theme colors
  const headingColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const tableBg = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.600', 'var(--text-secondary)');
  const primaryTextColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const hoverBg = useColorModeValue('gray.50', 'var(--bg-tertiary)');
  const theadBg = useColorModeValue('gray.50', 'var(--bg-tertiary)');

  // Status color mapping for purchases
  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'approved':
      case 'completed':
        return 'green';
      case 'draft':
      case 'pending_approval':
        return 'yellow';
      case 'pending':
        return 'blue';
      case 'cancelled':
      case 'rejected':
        return 'red';
      default:
        return 'gray';
    }
  };

  // Approval status color mapping
  const getApprovalStatusColor = (approvalStatus: string) => {
    switch ((approvalStatus || '').toLowerCase()) {
      case 'approved':
        return 'green';
      case 'pending':
        return 'yellow';
      case 'rejected':
        return 'red';
      case 'not_required':
      case 'not_started':
        return 'gray';
      default:
        return 'gray';
    }
  };

  const getStatusLabel = (status: string) => {
    return status.replace('_', ' ').toUpperCase();
  };

  const getApprovalStatusLabel = (approvalStatus: string) => {
    return (approvalStatus || '').replace('_', ' ').toUpperCase();
  };

  return (
    <Card boxShadow="sm" borderRadius="lg" borderWidth="1px" borderColor={borderColor}>
      <CardHeader>
        <Flex justify="space-between" align="center">
          <Heading size="md" color={headingColor}>
            {title} ({purchases?.length || 0})
          </Heading>
        </Flex>
      </CardHeader>
      <CardBody p={0}>
        {loading ? (
          <Flex justify="center" align="center" py={10}>
            <Spinner size="lg" color="var(--accent-color)" />
            <Text ml={4} color={textColor}>Loading transactions...</Text>
          </Flex>
        ) : purchases.length === 0 ? (
          <Box p={8} textAlign="center">
            <Text color={textColor}>No purchase transactions found.</Text>
          </Box>
        ) : (
          <Box overflowX="auto">
            <Table variant="simple" size="md" className="table">
              <Thead bg={theadBg}>
                <Tr>
                  <Th color={textColor} borderColor={borderColor} fontSize="xs" fontWeight="bold">PURCHASE #</Th>
                  <Th color={textColor} borderColor={borderColor} fontSize="xs" fontWeight="bold">VENDOR</Th>
                  <Th color={textColor} borderColor={borderColor} fontSize="xs" fontWeight="bold">DATE</Th>
                  <Th color={textColor} borderColor={borderColor} fontSize="xs" fontWeight="bold">TOTAL</Th>
                  <Th color={textColor} borderColor={borderColor} fontSize="xs" fontWeight="bold">STATUS</Th>
                  <Th color={textColor} borderColor={borderColor} fontSize="xs" fontWeight="bold">APPROVAL STATUS</Th>
                  <Th color={textColor} borderColor={borderColor} fontSize="xs" fontWeight="bold" textAlign="center">ACTIONS</Th>
                </Tr>
              </Thead>
              <Tbody>
                {purchases.map((purchase, index) => (
                  <Tr 
                    key={purchase.id}
                    _hover={{ bg: hoverBg }}
                    transition="all 0.2s ease"
                    borderBottom={index === purchases.length - 1 ? 'none' : '1px solid'}
                    borderColor={borderColor}
                  >
                    <Td borderColor={borderColor} py={3}>
                      <Text fontWeight="medium" color="blue.600">
                        {purchase.code}
                      </Text>
                    </Td>
                    <Td borderColor={borderColor} py={3}>
                      <Text fontWeight="medium" color={primaryTextColor} fontSize="sm">
                        {purchase.vendor?.name || 'N/A'}
                      </Text>
                    </Td>
                    <Td borderColor={borderColor} py={3}>
                      <Text fontSize="sm" color={textColor}>
                        {formatDate(purchase.date)}
                      </Text>
                    </Td>
                    <Td borderColor={borderColor} py={3}>
                      <Text fontWeight="medium" fontSize="sm" color={primaryTextColor}>
                        {formatCurrency(purchase.total_amount)}
                      </Text>
                    </Td>
                    <Td borderColor={borderColor} py={3}>
                      <Badge 
                        colorScheme={getStatusColor(purchase.status)} 
                        variant="subtle"
                        px={2}
                        py={1}
                        borderRadius="md"
                        fontSize="xs"
                      >
                        {getStatusLabel(purchase.status)}
                      </Badge>
                    </Td>
                    <Td borderColor={borderColor} py={3}>
                      <Badge 
                        colorScheme={getApprovalStatusColor(purchase.approval_status)} 
                        variant="subtle"
                        px={2}
                        py={1}
                        borderRadius="md"
                        fontSize="xs"
                      >
                        {getApprovalStatusLabel(purchase.approval_status)}
                      </Badge>
                    </Td>
                    <Td borderColor={borderColor} py={3} textAlign="center">
                      {renderActions ? (
                        renderActions(purchase)
                      ) : (
                        <Menu>
                          <MenuButton
                            as={IconButton}
                            icon={<FiMoreVertical />}
                            variant="ghost"
                            size="sm"
                            aria-label="Options"
                          />
                          <MenuList>
                            <MenuItem 
                              icon={<FiEye />} 
                              onClick={() => onViewDetails(purchase)}
                            >
                              View Details
                            </MenuItem>
                            {purchase.status === 'DRAFT' && canEdit && onEdit && (
                              <MenuItem icon={<FiEdit />} onClick={() => onEdit(purchase)}>
                                Edit
                              </MenuItem>
                            )}
                            {purchase.status === 'DRAFT' && userRole === 'employee' && onSubmitForApproval && (
                              <MenuItem 
                                icon={<FiAlertCircle />} 
                                onClick={() => onSubmitForApproval(purchase.id)}
                              >
                                Submit for Approval
                              </MenuItem>
                            )}
                            {canDelete && onDelete && (
                              <>
                                <MenuDivider />
                                <MenuItem 
                                  icon={<FiTrash2 />} 
                                  color="red.500" 
                                  onClick={() => onDelete(purchase.id)}
                                >
                                  Delete
                                </MenuItem>
                              </>
                            )}
                          </MenuList>
                        </Menu>
                      )}
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </Box>
        )}
      </CardBody>
    </Card>
  );
};

export default EnhancedPurchaseTable;

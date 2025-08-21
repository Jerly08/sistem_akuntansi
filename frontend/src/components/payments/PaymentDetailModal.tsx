'use client';

import React from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  Box,
  Flex,
  Text,
  Badge,
  Divider,
  HStack,
  VStack,
  Table,
  Tbody,
  Tr,
  Td,
  TableContainer,
} from '@chakra-ui/react';
import { FiFilePlus } from 'react-icons/fi';
import { Payment } from '@/services/paymentService';
import paymentService from '@/services/paymentService';
import { exportPaymentDetailToPDF } from '../../utils/pdfExport';

interface PaymentDetailModalProps {
  payment: Payment | null;
  isOpen: boolean;
  onClose: () => void;
}

const PaymentDetailModal: React.FC<PaymentDetailModalProps> = ({
  payment,
  isOpen,
  onClose,
}) => {
  if (!payment) return null;

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('id-ID', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="xl">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <Flex justify="space-between" align="center">
            <Text>Payment Details</Text>
            <Badge
              colorScheme={paymentService.getStatusColorScheme(payment.status)}
              variant="subtle"
              fontSize="sm"
            >
              {payment.status}
            </Badge>
          </Flex>
        </ModalHeader>
        <ModalCloseButton />
        
        <ModalBody>
          <VStack spacing={6} align="stretch">
            {/* Basic Information */}
            <Box>
              <Text fontWeight="bold" fontSize="lg" mb={3}>
                Basic Information
              </Text>
              <TableContainer>
                <Table size="sm" variant="simple">
                  <Tbody>
                    <Tr>
                      <Td fontWeight="medium" w="30%">Payment Code:</Td>
                      <Td>{payment.code}</Td>
                    </Tr>
                    <Tr>
                      <Td fontWeight="medium">Contact:</Td>
                      <Td>
                        <Flex direction="column">
                          <Text>{payment.contact?.name || 'Unknown Contact'}</Text>
                          <Text fontSize="sm" color="gray.500">
                            {payment.contact?.type || 'N/A'}
                          </Text>
                        </Flex>
                      </Td>
                    </Tr>
                    <Tr>
                      <Td fontWeight="medium">Amount:</Td>
                      <Td>
                        <Text fontWeight="bold" fontSize="lg" color="green.600">
                          {formatCurrency(payment.amount)}
                        </Text>
                      </Td>
                    </Tr>
                    <Tr>
                      <Td fontWeight="medium">Payment Date:</Td>
                      <Td>{formatDateTime(payment.date)}</Td>
                    </Tr>
                    <Tr>
                      <Td fontWeight="medium">Method:</Td>
                      <Td>
                        <Badge variant="outline">
                          {paymentService.getMethodDisplayName(payment.method)}
                        </Badge>
                      </Td>
                    </Tr>
                  </Tbody>
                </Table>
              </TableContainer>
            </Box>

            <Divider />

            {/* Additional Details */}
            <Box>
              <Text fontWeight="bold" fontSize="lg" mb={3}>
                Additional Information
              </Text>
              <TableContainer>
                <Table size="sm" variant="simple">
                  <Tbody>
                    <Tr>
                      <Td fontWeight="medium" w="30%">Reference:</Td>
                      <Td>{payment.reference || '-'}</Td>
                    </Tr>
                    <Tr>
                      <Td fontWeight="medium">Notes:</Td>
                      <Td>
                        {payment.notes ? (
                          <Text fontSize="sm" whiteSpace="pre-wrap">
                            {payment.notes}
                          </Text>
                        ) : (
                          <Text color="gray.500" fontSize="sm">No notes</Text>
                        )}
                      </Td>
                    </Tr>
                  </Tbody>
                </Table>
              </TableContainer>
            </Box>

            <Divider />

            {/* System Information */}
            <Box>
              <Text fontWeight="bold" fontSize="lg" mb={3}>
                System Information
              </Text>
              <TableContainer>
                <Table size="sm" variant="simple">
                  <Tbody>
                    <Tr>
                      <Td fontWeight="medium" w="30%">Created:</Td>
                      <Td>{formatDateTime(payment.created_at)}</Td>
                    </Tr>
                    <Tr>
                      <Td fontWeight="medium">Last Updated:</Td>
                      <Td>{formatDateTime(payment.updated_at)}</Td>
                    </Tr>
                    <Tr>
                      <Td fontWeight="medium">User ID:</Td>
                      <Td>{payment.user_id}</Td>
                    </Tr>
                  </Tbody>
                </Table>
              </TableContainer>
            </Box>
          </VStack>
        </ModalBody>

        <ModalFooter>
          <HStack spacing={3}>
            <Button
              variant="outline"
              leftIcon={<FiFilePlus />}
              onClick={() => exportPaymentDetailToPDF(payment)}
              colorScheme="red"
            >
              Export to PDF
            </Button>
            <Button onClick={onClose}>Close</Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default PaymentDetailModal;

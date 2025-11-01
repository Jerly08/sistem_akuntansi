'use client';

import React, { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Button,
  VStack,
  HStack,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Input,
  Select,
  Textarea,
  NumberInput,
  NumberInputField,
  Box,
  Text,
  useToast,
  Alert,
  AlertIcon,
  Switch,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Card,
  CardBody,
  SimpleGrid,
  Badge,
  Tooltip,
  Icon,
  useColorModeValue,
} from '@chakra-ui/react';
import { useForm, Controller } from 'react-hook-form';
import { FiInfo, FiDollarSign } from 'react-icons/fi';
import { useAuth } from '@/contexts/AuthContext';
import cashbankService, { CashBank } from '@/services/cashbankService';

interface PPNPaymentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
  ppnType: 'INPUT' | 'OUTPUT'; // INPUT = Masukan, OUTPUT = Keluaran
}

interface PPNPaymentFormData {
  ppn_type: 'INPUT' | 'OUTPUT';
  amount: number;
  date: string;
  cash_bank_id: number;
  reference: string;
  notes: string;
}

const PPNPaymentModal: React.FC<PPNPaymentModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
  ppnType,
}) => {
  const { token } = useAuth();
  const toast = useToast();
  
  const [loading, setLoading] = useState(false);
  const [cashBanks, setCashBanks] = useState<CashBank[]>([]);
  const [ppnBalance, setPPNBalance] = useState<number | null>(null);
  const [loadingBalance, setLoadingBalance] = useState(false);

  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const mutedColor = useColorModeValue('gray.600', 'gray.400');

  const {
    control,
    handleSubmit,
    formState: { errors },
    reset,
    watch,
  } = useForm<PPNPaymentFormData>({
    defaultValues: {
      ppn_type: ppnType,
      amount: 0,
      date: new Date().toISOString().split('T')[0],
      cash_bank_id: 0,
      reference: '',
      notes: '',
    },
  });

  const amount = watch('amount');
  const selectedCashBankId = watch('cash_bank_id');

  // Load cash/bank accounts
  useEffect(() => {
    if (isOpen && token) {
      loadCashBanks();
      loadPPNBalance();
    }
  }, [isOpen, token]);

  const loadCashBanks = async () => {
    try {
      const response = await cashbankService.getAllCashBanks();
      setCashBanks(response.data || []);
    } catch (error: any) {
      console.error('Failed to load cash/bank accounts:', error);
      toast({
        title: 'Error',
        description: 'Gagal memuat akun kas/bank',
        status: 'error',
        duration: 3000,
      });
    }
  };

  const loadPPNBalance = async () => {
    if (!token) return;
    
    setLoadingBalance(true);
    try {
      // Call API to get PPN balance from tax settings
      // This would need to be implemented in your API
      const response = await fetch(`/api/v1/tax-payments/ppn/balance?type=${ppnType}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setPPNBalance(data.balance || 0);
      }
    } catch (error) {
      console.error('Failed to load PPN balance:', error);
    } finally {
      setLoadingBalance(false);
    }
  };

  const getSelectedCashBank = () => {
    return cashBanks.find(cb => cb.id === selectedCashBankId);
  };

  const onSubmit = async (data: PPNPaymentFormData) => {
    if (!token) {
      toast({
        title: 'Error',
        description: 'Anda harus login terlebih dahulu',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    if (data.amount <= 0) {
      toast({
        title: 'Error',
        description: 'Jumlah pembayaran harus lebih dari 0',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    if (!data.cash_bank_id) {
      toast({
        title: 'Error',
        description: 'Pilih akun kas/bank',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    setLoading(true);
    try {
      const response = await fetch('/api/v1/tax-payments/ppn', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          ppn_type: data.ppn_type,
          amount: data.amount,
          date: new Date(data.date).toISOString(),
          cash_bank_id: data.cash_bank_id,
          reference: data.reference,
          notes: data.notes,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Gagal membuat pembayaran PPN');
      }

      const result = await response.json();

      toast({
        title: 'Berhasil',
        description: `Pembayaran PPN ${ppnType === 'INPUT' ? 'Masukan' : 'Keluaran'} berhasil dibuat`,
        status: 'success',
        duration: 3000,
      });

      reset();
      onClose();
      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      console.error('Error creating PPN payment:', error);
      toast({
        title: 'Error',
        description: error.message || 'Gagal membuat pembayaran PPN',
        status: 'error',
        duration: 5000,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  const ppnLabel = ppnType === 'INPUT' ? 'PPN Masukan' : 'PPN Keluaran';
  const ppnDescription = ppnType === 'INPUT' 
    ? 'Pembayaran PPN Masukan ke negara (dari pembelian)'
    : 'Pembayaran PPN Keluaran ke negara (dari penjualan)';

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="xl">
      <ModalOverlay />
      <ModalContent bg={bgColor}>
        <ModalHeader>
          <VStack align="start" spacing={1}>
            <Text>Pembayaran {ppnLabel}</Text>
            <Text fontSize="sm" fontWeight="normal" color={mutedColor}>
              {ppnDescription}
            </Text>
          </VStack>
        </ModalHeader>
        <ModalCloseButton />

        <form onSubmit={handleSubmit(onSubmit)}>
          <ModalBody>
            <VStack spacing={4} align="stretch">
              {/* Info Alert */}
              <Alert status="info" borderRadius="md">
                <AlertIcon />
                <Box fontSize="sm">
                  <Text fontWeight="bold">Jurnal Otomatis:</Text>
                  <Text>
                    Debit: {ppnLabel} (saldo berkurang) | 
                    Credit: Kas/Bank (kas berkurang)
                  </Text>
                </Box>
              </Alert>

              {/* PPN Balance Info */}
              {ppnBalance !== null && (
                <Card bg={useColorModeValue('blue.50', 'blue.900')} borderColor={borderColor}>
                  <CardBody>
                    <Stat>
                      <StatLabel>Saldo {ppnLabel} Saat Ini</StatLabel>
                      <StatNumber>
                        Rp {ppnBalance.toLocaleString('id-ID')}
                      </StatNumber>
                      <StatHelpText>
                        Yang harus dibayarkan ke negara
                      </StatHelpText>
                    </Stat>
                  </CardBody>
                </Card>
              )}

              {/* Payment Date */}
              <FormControl isInvalid={!!errors.date}>
                <FormLabel>
                  Tanggal Pembayaran
                  <Tooltip label="Tanggal saat PPN dibayarkan ke negara">
                    <span>
                      <Icon as={FiInfo} ml={2} color={mutedColor} />
                    </span>
                  </Tooltip>
                </FormLabel>
                <Controller
                  name="date"
                  control={control}
                  rules={{ required: 'Tanggal pembayaran wajib diisi' }}
                  render={({ field }) => (
                    <Input type="date" {...field} />
                  )}
                />
                <FormErrorMessage>{errors.date?.message}</FormErrorMessage>
              </FormControl>

              {/* Amount */}
              <FormControl isInvalid={!!errors.amount}>
                <FormLabel>
                  Jumlah Pembayaran
                  <Tooltip label="Jumlah PPN yang dibayarkan ke negara">
                    <span>
                      <Icon as={FiInfo} ml={2} color={mutedColor} />
                    </span>
                  </Tooltip>
                </FormLabel>
                <Controller
                  name="amount"
                  control={control}
                  rules={{ 
                    required: 'Jumlah pembayaran wajib diisi',
                    min: { value: 1, message: 'Jumlah minimal 1' }
                  }}
                  render={({ field }) => (
                    <NumberInput
                      value={field.value}
                      onChange={(_, valueNumber) => field.onChange(valueNumber || 0)}
                      min={0}
                    >
                      <NumberInputField placeholder="Masukkan jumlah pembayaran" />
                    </NumberInput>
                  )}
                />
                <FormErrorMessage>{errors.amount?.message}</FormErrorMessage>
                {amount > 0 && (
                  <Text fontSize="sm" color={mutedColor} mt={1}>
                    Rp {amount.toLocaleString('id-ID')}
                  </Text>
                )}
              </FormControl>

              {/* Cash/Bank Account */}
              <FormControl isInvalid={!!errors.cash_bank_id}>
                <FormLabel>
                  Akun Kas/Bank
                  <Tooltip label="Pilih akun kas/bank yang digunakan untuk membayar PPN">
                    <span>
                      <Icon as={FiInfo} ml={2} color={mutedColor} />
                    </span>
                  </Tooltip>
                </FormLabel>
                <Controller
                  name="cash_bank_id"
                  control={control}
                  rules={{ required: 'Akun kas/bank wajib dipilih' }}
                  render={({ field }) => (
                    <Select
                      {...field}
                      placeholder="Pilih akun kas/bank"
                      onChange={(e) => field.onChange(parseInt(e.target.value))}
                    >
                      {cashBanks.map((cashBank) => (
                        <option key={cashBank.id} value={cashBank.id}>
                          {cashBank.name} - Saldo: Rp {cashBank.balance.toLocaleString('id-ID')}
                        </option>
                      ))}
                    </Select>
                  )}
                />
                <FormErrorMessage>{errors.cash_bank_id?.message}</FormErrorMessage>
              </FormControl>

              {/* Amount Validation */}
              {amount > 0 && selectedCashBankId > 0 && (() => {
                const selectedCB = getSelectedCashBank();
                if (selectedCB && amount > selectedCB.balance) {
                  return (
                    <Alert status="warning" borderRadius="md">
                      <AlertIcon />
                      <Text fontSize="sm">
                        Saldo kas/bank tidak mencukupi. Saldo: Rp {selectedCB.balance.toLocaleString('id-ID')}
                      </Text>
                    </Alert>
                  );
                }
                return null;
              })()}

              {/* Reference */}
              <FormControl>
                <FormLabel>
                  Nomor Referensi (Opsional)
                  <Tooltip label="Nomor referensi pembayaran seperti nomor SSP atau kode billing">
                    <span>
                      <Icon as={FiInfo} ml={2} color={mutedColor} />
                    </span>
                  </Tooltip>
                </FormLabel>
                <Controller
                  name="reference"
                  control={control}
                  render={({ field }) => (
                    <Input
                      {...field}
                      placeholder="Contoh: SSP-123456 atau kode billing"
                    />
                  )}
                />
              </FormControl>

              {/* Notes */}
              <FormControl>
                <FormLabel>
                  Catatan (Opsional)
                  <Tooltip label="Catatan tambahan untuk pembayaran PPN ini">
                    <span>
                      <Icon as={FiInfo} ml={2} color={mutedColor} />
                    </span>
                  </Tooltip>
                </FormLabel>
                <Controller
                  name="notes"
                  control={control}
                  render={({ field }) => (
                    <Textarea
                      {...field}
                      placeholder="Catatan tambahan..."
                      rows={3}
                    />
                  )}
                />
              </FormControl>

              {/* Summary */}
              {amount > 0 && (
                <Card bg={useColorModeValue('gray.50', 'gray.700')} borderColor={borderColor}>
                  <CardBody>
                    <VStack align="stretch" spacing={2}>
                      <Text fontWeight="bold" fontSize="sm">Ringkasan:</Text>
                      <HStack justify="space-between">
                        <Text fontSize="sm">Jumlah Pembayaran:</Text>
                        <Text fontSize="sm" fontWeight="bold">
                          Rp {amount.toLocaleString('id-ID')}
                        </Text>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm">{ppnLabel}:</Text>
                        <Badge colorScheme="red">Berkurang</Badge>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm">Kas/Bank:</Text>
                        <Badge colorScheme="red">Berkurang</Badge>
                      </HStack>
                    </VStack>
                  </CardBody>
                </Card>
              )}
            </VStack>
          </ModalBody>

          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={handleClose}>
              Batal
            </Button>
            <Button
              colorScheme="blue"
              type="submit"
              isLoading={loading}
              loadingText="Memproses..."
            >
              Bayar PPN
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default PPNPaymentModal;

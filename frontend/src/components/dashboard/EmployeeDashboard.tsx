'use client';
import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import api from '@/services/api';
import { 
  Box, 
  Heading, 
  Text, 
  Card,
  CardHeader,
  CardBody,
  Button,
  HStack,
  Icon,
  List,
  ListItem,
  ListIcon,
  Badge,
  Flex,
  Spinner
} from '@chakra-ui/react';
import {
  FiUser,
  FiPlus,
  FiFileText,
  FiActivity,
  FiBell
} from 'react-icons/fi';

interface DashboardSummary {
  statistics?: Record<string, any>;
  recent_activities?: Array<{
    id: number;
    action: string;
    table_name?: string;
    record_id?: number;
    user_id?: number;
    created_at?: string;
  }>;
  unread_notifications?: number;
  min_stock_alerts_count?: number;
}

export const EmployeeDashboard = () => {
  const router = useRouter();
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSummary = async () => {
      try {
        const res = await api.get('/dashboard/summary');
        setSummary(res.data?.data || res.data || {});
        setError(null);
      } catch (e: any) {
        setError(e?.response?.data?.error || e?.message || 'Gagal memuat ringkasan dashboard');
      } finally {
        setLoading(false);
      }
    };
    fetchSummary();
  }, []);
  
  return (
    <Box>
      <Heading as="h2" size="xl" mb={6} color="gray.800">
        Dasbor Saya
      </Heading>

      {loading ? (
        <Flex justify="center" align="center" minH="120px">
          <Spinner size="lg" color="brand.500" thickness="4px" />
        </Flex>
      ) : (
        <>
          {/* Ringkasan & Notifikasi */}
          <Box bg="white" p={6} borderRadius="lg" boxShadow="sm" mt={4}>
            <Heading as="h3" size="md" mb={3} display="flex" alignItems="center" gap={2}>
              <Icon as={FiBell} /> Ringkasan
            </Heading>
            {error ? (
              <Text color="red.500">{error}</Text>
            ) : (
              <Text>
                Anda memiliki{' '}
                <Badge colorScheme="blue" mx={1}>{summary?.unread_notifications ?? 0}</Badge>
                notifikasi belum dibaca.
              </Text>
            )}
          </Box>

          {/* Aktivitas Terbaru */}
          {summary?.recent_activities && summary.recent_activities.length > 0 && (
            <Card mt={6}>
              <CardHeader>
                <Heading size="md" display="flex" alignItems="center">
                  <Icon as={FiActivity} mr={2} color="blue.500" />
                  Aktivitas Terbaru
                </Heading>
              </CardHeader>
              <CardBody>
                <List spacing={3}>
                  {summary.recent_activities.slice(0, 8).map((act) => (
                    <ListItem key={act.id} display="flex" alignItems="center">
                      <ListIcon as={FiActivity} color="gray.500" />
                      <Box>
                        <Text fontWeight="medium">{act.action}</Text>
                        <Text fontSize="sm" color="gray.600">
                          {act.table_name ? `${act.table_name}` : ''}
                          {act.created_at ? ` • ${new Date(act.created_at).toLocaleString('id-ID')}` : ''}
                        </Text>
                      </Box>
                    </ListItem>
                  ))}
                </List>
              </CardBody>
            </Card>
          )}

          {/* Akses Cepat - sesuai role employee */}
          <Card mt={6}>
            <CardHeader>
              <Heading size="md" display="flex" alignItems="center">
                <Icon as={FiPlus} mr={2} color="blue.500" />
                Akses Cepat
              </Heading>
            </CardHeader>
            <CardBody>
              <Text mb={4} color="gray.600">
                Sebagai karyawan, Anda dapat mengakses profil dan melihat laporan yang tersedia.
              </Text>
              <HStack spacing={4} flexWrap="wrap">
                <Button
                  leftIcon={<FiUser />}
                  colorScheme="blue"
                  variant="outline"
                  onClick={() => router.push('/profile')}
                  size="md"
                >
                  Profil Saya
                </Button>
                <Button
                  leftIcon={<FiFileText />}
                  colorScheme="gray"
                  variant="outline"
                  onClick={() => router.push('/reports')}
                  size="md"
                >
                  Lihat Laporan
                </Button>
              </HStack>
            </CardBody>
          </Card>
        </>
      )}
    </Box>
  );
};

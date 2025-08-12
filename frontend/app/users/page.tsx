'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import { DataTable } from '@/components/common/DataTable';
import {
  Box,
  Heading,
  Text,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Spinner
} from '@chakra-ui/react';

interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  active: boolean;
  createdAt: string;
}

const columns = [
  { header: 'Name', accessor: 'name' },
  { header: 'Email', accessor: 'email' },
  { header: 'Role', accessor: 'role' },
  { header: 'Status', accessor: 'active' },
  { header: 'Created', accessor: 'createdAt' },
];

const UsersPage: React.FC = () => {
  const { user } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const res = await fetch('http://localhost:8080/api/v1/users', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          },
        });

        if (!res.ok) {
          throw new Error('Failed to fetch users data');
        }

        const data = await res.json();
        setUsers(data.map((user: any) => ({
          ...user,
          active: user.active ? 'Active' : 'Inactive',
          createdAt: new Date(user.createdAt).toLocaleDateString(),
        })));
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  if (loading) {
    return (
<Layout allowedRoles={['admin']}>
        <Box>
          <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
          <Text ml={4}>Loading users...</Text>
        </Box>
      </Layout>
    );
  }

  return (
<Layout allowedRoles={['admin']}>
      <Box>
        <Heading as="h1" size="xl" mb={6}>Users Management</Heading>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle>Error:</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <Box bg="white" borderRadius="lg" overflow="hidden" boxShadow="sm">
          <DataTable 
            columns={columns} 
            data={users} 
            keyField="id"
            title="Users List"
            searchable={true}
            pagination={true}
            pageSize={10}
          />
        </Box>
      </Box>
    </Layout>
  );
};

export default UsersPage;

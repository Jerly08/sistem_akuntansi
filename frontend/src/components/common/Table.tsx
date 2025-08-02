'use client';

import React from 'react';
import {
  Table as ChakraTable,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Spinner,
  Flex,
  Text,
  Badge,
  Button,
  Box,
} from '@chakra-ui/react';

interface Column<T> {
  header: string;
  accessor: string | ((item: T) => React.ReactNode);
  cell?: (item: T) => React.ReactNode;
}

interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyField: keyof T;
  title?: string;
  actions?: (item: T) => React.ReactNode;
  isLoading?: boolean;
  emptyMessage?: string;
}

function Table<T>({ columns, data, keyField, title, actions, isLoading, emptyMessage }: TableProps<T>) {
  const renderCell = (item: T, column: Column<T>) => {
    if (column.cell) {
      return column.cell(item);
    }
    
    if (typeof column.accessor === 'function') {
      return column.accessor(item);
    }
    
    return item[column.accessor as keyof T] as React.ReactNode;
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active':
        return 'green';
      case 'inactive':
        return 'red';
      case 'pending':
        return 'yellow';
      case 'approved':
        return 'blue';
      default:
        return 'gray';
    }
  };

  const renderCellContent = (content: React.ReactNode) => {
    if (typeof content === 'string') {
      // Check if it's a status-like field
      if (['active', 'inactive', 'pending', 'approved'].includes(content.toLowerCase())) {
        return (
          <Badge colorScheme={getStatusColor(content)} variant="subtle">
            {content}
          </Badge>
        );
      }
    }
    return content;
  };

  if (isLoading) {
    return (
      <Card>
        {title && (
          <CardHeader>
            <Heading size="md">{title}</Heading>
          </CardHeader>
        )}
        <CardBody>
          <Flex justify="center" align="center" py={8}>
            <Spinner size="lg" color="brand.500" />
            <Text ml={4}>Loading...</Text>
          </Flex>
        </CardBody>
      </Card>
    );
  }

  return (
    <Card>
      {title && (
        <CardHeader>
          <Heading size="md">{title}</Heading>
        </CardHeader>
      )}
      <CardBody p={0}>
        {data.length === 0 ? (
          <Flex justify="center" align="center" py={8}>
            <Text color="gray.500">{emptyMessage || 'No data available'}</Text>
          </Flex>
        ) : (
          <Box overflowX="auto">
            <ChakraTable variant="simple">
              <Thead bg="gray.50">
                <Tr>
                  {columns.map((column, index) => (
                    <Th key={index} fontWeight="bold">
                      {column.header}
                    </Th>
                  ))}
                  {actions && <Th>Actions</Th>}
                </Tr>
              </Thead>
              <Tbody>
                {data.map((item) => (
                  <Tr key={String(item[keyField])} _hover={{ bg: 'gray.50' }}>
                    {columns.map((column, index) => (
                      <Td key={index}>
                        {renderCellContent(renderCell(item, column))}
                      </Td>
                    ))}
                    {actions && (
                      <Td>
                        <Flex gap={2} justify="flex-end">
                          {actions(item)}
                        </Flex>
                      </Td>
                    )}
                  </Tr>
                ))}
              </Tbody>
            </ChakraTable>
          </Box>
        )}
      </CardBody>
    </Card>
  );
}

export default Table;

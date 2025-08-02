'use client';

import dynamic from 'next/dynamic';
import { UserRole } from '@/contexts/AuthContext';

const Layout = dynamic(() => import('./Layout'), {
  ssr: false,
  loading: () => <div>Loading...</div>
});

interface DynamicLayoutProps {
  children: React.ReactNode;
  allowedRoles?: UserRole[];
}

const DynamicLayout: React.FC<DynamicLayoutProps> = ({ children, allowedRoles = [] }) => {
  return (
    <Layout allowedRoles={allowedRoles}>
      {children}
    </Layout>
  );
};

export default DynamicLayout;

'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth, UserRole } from '@/contexts/AuthContext';
import { normalizeRole, normalizeRoles } from '@/utils/roles';

interface ProtectedRouteProps {
  children: React.ReactNode;
  allowedRoles?: (UserRole | string)[];
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ 
  children, 
  allowedRoles = [] 
}) => {
  const { isAuthenticated, isLoading, user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        router.push('/login');
        return;
      }

      if (allowedRoles.length > 0 && user) {
        const allowed = new Set(normalizeRoles(allowedRoles as string[]));
        const userRoleNorm = normalizeRole(user.role as unknown as string);
        const hasPermission = allowed.has(userRoleNorm);
        if (!hasPermission) {
          router.push('/unauthorized');
        }
      }
    }
  }, [isAuthenticated, isLoading, user, router, allowedRoles]);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return null;
  }

  if (allowedRoles.length > 0 && user) {
    const allowed = new Set(normalizeRoles(allowedRoles as string[]));
    const userRoleNorm = normalizeRole(user.role as unknown as string);
    if (!allowed.has(userRoleNorm)) {
      return null;
    }
  }

  return <>{children}</>;
};

export default ProtectedRoute; 

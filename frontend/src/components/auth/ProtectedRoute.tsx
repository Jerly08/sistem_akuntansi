'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth, UserRole } from '@/contexts/AuthContext';

interface ProtectedRouteProps {
  children: React.ReactNode;
  allowedRoles?: UserRole[];
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ 
  children, 
  allowedRoles = [] 
}) => {
  const { isAuthenticated, isLoading, user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading) {
      // If not authenticated, redirect to login
      if (!isAuthenticated) {
        router.push('/login');
        return;
      }
      
      // If roles are specified, check if user has required role
      if (allowedRoles.length > 0 && user) {
        const hasPermission = allowedRoles.includes(user.role);
        if (!hasPermission) {
          // Redirect to unauthorized page
          router.push('/unauthorized');
        }
      }
    }
  }, [isAuthenticated, isLoading, user, router, allowedRoles]);

  // Show loading state
  if (isLoading) {
    return <div>Loading...</div>;
  }

  // If not authenticated, don't render children
  if (!isAuthenticated) {
    return null;
  }

  // If roles are specified and user doesn't have required role, don't render children
  if (allowedRoles.length > 0 && user && !allowedRoles.includes(user.role)) {
    return null;
  }

  // Render children if authenticated and authorized
  return <>{children}</>;
};

export default ProtectedRoute; 
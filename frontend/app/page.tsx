'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';

const LoadingSpinner = () => (
  <div className="min-h-screen flex items-center justify-center bg-gray-50">
    <div className="text-center">
      <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
    </div>
  </div>
);

const MainContent = () => (
  <div className="min-h-screen flex items-center justify-center bg-gray-50">
    <div className="text-center">
      <h1 className="text-3xl font-bold mb-4">Accounting Application</h1>
      <p className="text-gray-600 mb-6">Please wait while we redirect you...</p>
      <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
    </div>
  </div>
);

export default function Home() {
  const { isAuthenticated, isLoading, checkAuth } = useAuth();
  const router = useRouter();

  useEffect(() => {
    const checkAndRedirect = async () => {
      const isAuth = await checkAuth();
      if (isAuth) {
        router.push('/dashboard');
      } else {
        router.push('/login');
      }
    };

    if (!isLoading) {
      checkAndRedirect();
    }
  }, [isLoading, isAuthenticated, router, checkAuth]);

  // Show loading spinner during SSR and initial client render
  if (isLoading) {
    return <LoadingSpinner />;
  }

  return <MainContent />;
}

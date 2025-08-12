import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/contexts/AuthContext";
import ClientOnly from "@/components/common/ClientOnly";

export const metadata: Metadata = {
  title: "Accounting Application",
  description: "A full-featured accounting application",
};

import { ChakraProviderWrapper } from '@/providers/ChakraProvider';

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="font-sans">
        <ClientOnly>
          <ChakraProviderWrapper>
            <AuthProvider>
              {children}
            </AuthProvider>
          </ChakraProviderWrapper>
        </ClientOnly>
      </body>
    </html>
  );
}

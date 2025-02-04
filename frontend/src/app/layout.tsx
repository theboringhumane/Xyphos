import './globals.css';
import { Providers } from './providers';
import { Navbar } from '@/components/navbar';
import { Sidebar } from '@/components/sidebar';

export const metadata = {
  title: 'Xyphos - Key Management System',
  description: 'Open-source key management system',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="min-h-screen bg-glitch-gray-50 dark:bg-glitch-gray-900 text-glitch-gray-900 dark:text-glitch-gray-50">
        <Providers>
          <div className="min-h-screen">
            <Navbar />
            <div className="flex">
              <Sidebar />
              <main className="flex-1 p-8">
                <div className="max-w-7xl mx-auto">
                  <div className="animate-fade-in">
                    {children}
                  </div>
                </div>
              </main>
            </div>
          </div>
        </Providers>
      </body>
    </html>
  );
}
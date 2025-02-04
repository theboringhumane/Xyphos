'use client';

import { useSession, signOut } from 'next-auth/react';
import Link from 'next/link';
import { useTheme } from 'next-themes';
import { Menu, Moon, Sun, User } from 'lucide-react';

export function Navbar() {
  const { data: session } = useSession();
  const { theme, setTheme } = useTheme();

  return (
    <nav className="bg-white dark:bg-gray-800 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex">
            <div className="flex-shrink-0 flex items-center">
              <Link href="/" className="text-xl font-bold text-gray-800 dark:text-white">
                🔐 Xyphos
              </Link>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <button
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              {theme === 'dark' ? (
                <Sun className="h-5 w-5 text-gray-500 dark:text-gray-400" />
              ) : (
                <Moon className="h-5 w-5 text-gray-500 dark:text-gray-400" />
              )}
            </button>

            {session ? (
              <div className="relative group">
                <button className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700">
                  <User className="h-5 w-5 text-gray-500 dark:text-gray-400" />
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    {session.user?.name}
                  </span>
                </button>
                <div className="absolute right-0 w-48 mt-2 py-2 bg-white dark:bg-gray-800 rounded-lg shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200">
                  <button
                    onClick={() => signOut()}
                    className="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    Sign out
                  </button>
                </div>
              </div>
            ) : (
              <Link
                href="/auth/signin"
                className="text-sm text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
              >
                Sign in
              </Link>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
} 
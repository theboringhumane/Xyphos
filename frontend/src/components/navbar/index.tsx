'use client';

import { useSession, signOut } from 'next-auth/react';
import Link from 'next/link';
import { useTheme } from 'next-themes';
import { Menu, Moon, Sun, User, Bell, Search } from 'lucide-react';

export function Navbar() {
  const { data: session } = useSession();
  const { theme, setTheme } = useTheme();

  return (
    <nav className="bg-background/80 backdrop-blur-sm border-b border-border sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link 
              href="/" 
              className="flex items-center space-x-2 text-xl font-bold text-foreground hover:opacity-80 transition-opacity"
            >
              <span className="text-2xl">🔐</span>
              <span className="scale-in">Xyphos</span>
            </Link>

            <div className="hidden md:flex md:ml-8">
              <div className="relative max-w-lg w-96">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Search className="h-5 w-5 text-muted" />
                </div>
                <input
                  type="text"
                  className="input-primary pl-10"
                  placeholder="Search keys, projects..."
                />
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <button
              className="p-2 rounded-md hover:bg-muted/10 transition-colors relative"
              title="Notifications"
            >
              <Bell className="h-5 w-5 text-muted" />
              <span className="absolute top-1 right-1 w-2 h-2 bg-primary rounded-full"></span>
            </button>

            <button
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="p-2 rounded-md hover:bg-muted/10 transition-colors"
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {theme === 'dark' ? (
                <Sun className="h-5 w-5 text-muted" />
              ) : (
                <Moon className="h-5 w-5 text-muted" />
              )}
            </button>

            {session ? (
              <div className="relative group">
                <button className="flex items-center space-x-2 p-2 rounded-md hover:bg-muted/10 transition-colors">
                  <div className="w-8 h-8 rounded-full bg-primary flex items-center justify-center text-white font-medium">
                    {session.user?.name?.[0].toUpperCase()}
                  </div>
                  <span className="text-sm text-foreground hidden md:inline-block">
                    {session.user?.name}
                  </span>
                </button>
                <div className="absolute right-0 w-48 mt-2 py-2 bg-background rounded-lg shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 border border-border">
                  <div className="px-4 py-2 border-b border-border">
                    <p className="text-sm font-medium text-foreground">{session.user?.name}</p>
                    <p className="text-xs text-muted truncate">{session.user?.email}</p>
                  </div>
                  <button
                    onClick={() => signOut()}
                    className="block w-full text-left px-4 py-2 text-sm text-foreground hover:bg-muted/10 transition-colors"
                  >
                    Sign out
                  </button>
                </div>
              </div>
            ) : (
              <Link
                href="/auth/signin"
                className="btn-primary"
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
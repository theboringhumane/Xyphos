'use client';

import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { Key, KeyRound, Shield, Settings, Users } from 'lucide-react';

const navigation = [
  {
    name: 'Projects',
    href: '/projects',
    icon: Shield,
    children: [
      {
        name: 'Locations',
        href: '/locations',
        icon: Users,
        children: [
          {
            name: 'Key Rings',
            href: '/keyrings',
            icon: KeyRound,
            children: [
              {
                name: 'Keys',
                href: '/keys',
                icon: Key,
              },
            ],
          },
        ],
      },
    ],
  },
  {
    name: 'Settings',
    href: '/settings',
    icon: Settings,
  },
];

function NavItem({ item, level = 0 }: { item: any; level?: number }) {
  const pathname = usePathname();
  const isActive = pathname.startsWith(item.href);
  const Icon = item.icon;

  return (
    <div>
      <Link
        href={item.href}
        className={`flex items-center space-x-2 px-4 py-2 text-sm rounded-lg ${
          isActive
            ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
        }`}
        style={{ marginLeft: `${level * 1}rem` }}
      >
        <Icon className="h-5 w-5" />
        <span>{item.name}</span>
      </Link>
      {item.children?.map((child: any) => (
        <NavItem key={child.href} item={child} level={level + 1} />
      ))}
    </div>
  );
}

export function Sidebar() {
  return (
    <div className="w-64 bg-white dark:bg-gray-800 shadow-sm h-[calc(100vh-4rem)] overflow-y-auto">
      <div className="p-4">
        <nav className="space-y-1">
          {navigation.map((item) => (
            <NavItem key={item.href} item={item} />
          ))}
        </nav>
      </div>
    </div>
  );
} 
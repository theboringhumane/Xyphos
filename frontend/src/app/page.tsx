'use client';

import { useSession } from 'next-auth/react';
import { Shield } from 'lucide-react';

export default function Home() {
  const { data: session } = useSession();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Welcome to Xyphos</h1>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Projects Overview */}
        <div className="col-span-2 space-y-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
            <div className="flex items-center space-x-3 mb-4">
              <Shield className="h-6 w-6 text-blue-500" />
              <h2 className="text-xl font-semibold">Projects Overview</h2>
            </div>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Manage your encryption keys across different projects and locations.
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
                <h3 className="font-medium mb-2">Project Structure</h3>
                <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                  <li>• Projects - Group resources by project</li>
                  <li>• Locations - Organize by geographic region</li>
                  <li>• Key Rings - Group related keys</li>
                  <li>• Keys - Individual encryption keys</li>
                </ul>
              </div>
              <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
                <h3 className="font-medium mb-2">Quick Actions</h3>
                <ul className="space-y-2 text-sm">
                  <li>
                    <button className="text-blue-500 hover:text-blue-600">
                      + Create new project
                    </button>
                  </li>
                  <li>
                    <button className="text-blue-500 hover:text-blue-600">
                      + Create key ring
                    </button>
                  </li>
                  <li>
                    <button className="text-blue-500 hover:text-blue-600">
                      + Generate new key
                    </button>
                  </li>
                </ul>
              </div>
            </div>
          </div>

          {/* Recent Activity */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
            <h2 className="text-xl font-semibold mb-4">Recent Activity</h2>
            <div className="space-y-4">
              <p className="text-gray-600 dark:text-gray-400 text-sm">
                No recent activity to display.
              </p>
            </div>
          </div>
        </div>

        {/* Stats and Info */}
        <div className="space-y-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
            <h2 className="text-xl font-semibold mb-4">Statistics</h2>
            <div className="space-y-4">
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Total Projects</p>
                <p className="text-2xl font-semibold">0</p>
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Active Keys</p>
                <p className="text-2xl font-semibold">0</p>
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Keys Rotated (24h)</p>
                <p className="text-2xl font-semibold">0</p>
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
            <h2 className="text-xl font-semibold mb-4">Quick Tips</h2>
            <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
              <li>• Use key rings to group related keys</li>
              <li>• Enable automatic key rotation</li>
              <li>• Monitor key usage in audit logs</li>
              <li>• Regularly review access policies</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
} 
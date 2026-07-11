"use client";

import { usePathname, useRouter } from 'next/navigation';
import { Moon, Sun, Bell, MessageCircle } from 'lucide-react';
import Image from 'next/image';

export const Header = () => {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = () => {
    // Clear token and redirect
    document.cookie = 'token=; Path=/; Max-Age=0';
    router.push('/login');
  };

  return (
    <header className="flex items-center justify-between px-6 py-4 bg-white border-b border-gray-200 dark:bg-gray-800 dark:border-gray-700">
      <div className="flex items-center space-x-4">
        <button
          className="lg:hidden p-2 rounded-md hover:bg-gray-100"
          aria-label="Open sidebar"
        >
          {/* Hamburger menu */}
          <span className="block h-0.5 w-5 bg-gray-600"></span>
          <span className="block h-0.5 w-5 bg-gray-600 mt-1.5"></span>
          <span className="block h-0.5 w-5 bg-gray-600 mt-1.5"></span>
        </button>
        <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
          Dashboard
        </h1>
      </div>
      <div className="flex items-center space-x-4">
        <div className="relative">
          <button
            className="relative p-2 rounded-md hover:bg-gray-100"
            aria-label="Notifications"
          >
            <Bell className="h-5 w-5 text-gray-600" />
            {/* Badge */}
            {false && (
              <div className="absolute -top-1 -right-1 flex h-2 w-2 items-center justify-center bg-red-500 rounded-full">
                <span className="text-xs font-medium text-white">
                  3
                </span>
              </div>
            )}
          </button>
          <button
            className="relative p-2 rounded-md hover:bg-gray-100"
            aria-label="Messages"
          >
            <MessageCircle className="h-5 w-5 text-gray-600" />
            {/* Badge */}
            {false && (
              <div className="absolute -top-1 -right-1 flex h-2 w-2 items-center justify-center bg-blue-500 rounded-full">
                <span className="text-xs font-medium text-white">
                  2
                </span>
              </div>
            )}
          </button>
          {/* Theme toggle */}
          <button
            className="p-2 rounded-md hover:bg-gray-100"
            aria-label="Toggle theme"
          >
            {typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches ? (
              <Sun className="h-5 w-5 text-yellow-400" />
            ) : (
              <Moon className="h-5 w-5 text-gray-600" />
            )}
          </button>
          {/* User avatar */}
          <div className="relative">
            <Image
              src="/placeholder-user.jpg"
              alt="User"
              width={40}
              height={40}
              className="rounded-full"
              priority
            />
            <div className="absolute -top-1 -right-0 flex h-2 w-2 items-center justify-center bg-green-500 rounded-full">
              <span className="text-xs font-medium text-white">
                Live
              </span>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
import React, { useState } from 'react';
import { useRouter } from 'next/router';
import Link from 'next/link';
import {
  HomeIcon,
  CubeIcon,
  ChatBubbleLeftRightIcon,
  Cog6ToothIcon,
  BellIcon,
  UserCircleIcon,
} from '@heroicons/react/24/outline';
import { useAuthStore } from '../store/auth';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const [showUserMenu, setShowUserMenu] = useState(false);

  const navigation = [
    { name: 'Dashboard', href: '/', icon: HomeIcon },
    { name: 'Kubernetes', href: '/kubernetes', icon: CubeIcon },
    { name: 'AI Chat', href: '/chat', icon: ChatBubbleLeftRightIcon },
    { name: 'Settings', href: '/settings', icon: Cog6ToothIcon },
  ];

  const handleLogout = async () => {
    try {
      // Call logout API
      const response = await fetch('/api/auth/logout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (response.ok) {
        // Clear local state
        logout();
        // Redirect to login
        router.push('/login');
      } else {
        console.error('Logout failed');
        // Still clear local state and redirect
        logout();
        router.push('/login');
      }
    } catch (error) {
      console.error('Logout error:', error);
      // Clear local state and redirect even if API call fails
      logout();
      router.push('/login');
    }
  };

  const toggleUserMenu = () => {
    setShowUserMenu(!showUserMenu);
  };

  // Close user menu when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest('.user-menu')) {
        setShowUserMenu(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Sidebar */}
      <div className="fixed inset-y-0 left-0 z-50 w-64 bg-white shadow-lg">
        <div className="flex h-16 items-center justify-center border-b border-gray-200">
          <h1 className="text-xl font-bold text-gray-900">
            K8s AI Platform
          </h1>
        </div>
        
        <nav className="mt-8 px-4">
          <ul className="space-y-2">
            {navigation.map((item) => {
              const isActive = router.pathname === item.href;
              return (
                <li key={item.name}>
                  <Link
                    href={item.href}
                    className={`sidebar-item ${
                      isActive ? 'active' : ''
                    }`}
                  >
                    <item.icon className="mr-3 h-5 w-5" />
                    {item.name}
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
      </div>

      {/* Main content */}
      <div className="pl-64">
        {/* Header */}
        <header className="bg-white shadow-sm border-b border-gray-200">
          <div className="flex h-16 items-center justify-between px-6">
            <div className="flex items-center">
              <h2 className="text-lg font-semibold text-gray-900">
                {navigation.find(item => item.href === router.pathname)?.name || 'Dashboard'}
              </h2>
            </div>
            
            <div className="flex items-center space-x-4">
              {/* Notifications */}
              <button className="p-2 text-gray-400 hover:text-gray-500">
                <BellIcon className="h-6 w-6" />
              </button>
              
              {/* User menu */}
              <div className="relative user-menu">
                <button 
                  onClick={toggleUserMenu}
                  className="flex items-center space-x-2 text-gray-700 hover:text-gray-900 p-2 rounded-md hover:bg-gray-100"
                >
                  <UserCircleIcon className="h-6 w-6" />
                  <span className="text-sm font-medium">
                    {user?.firstName} {user?.lastName}
                  </span>
                </button>
                
                {/* Dropdown menu */}
                {showUserMenu && (
                  <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-50 border border-gray-200">
                    <button
                      onClick={handleLogout}
                      className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors duration-150"
                    >
                      Sign out
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="p-6">
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout; 
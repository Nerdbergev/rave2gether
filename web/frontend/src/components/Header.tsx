// src/components/Header.tsx
import React from 'react';
import { useAuth } from '../contexts/AuthContext';

const Header: React.FC = () => {
  const { user, logout } = useAuth();

  return (
    <header className="p-4 bg-gray-800 text-white flex items-center justify-between">
      <h1 className="text-xl font-bold">Rave 2 Gether</h1>
      <div>
        {user ? (
          <>
            <span className="mr-4">
              Welcome, {user.username} ({user.role})
            </span>
            <button
              onClick={logout}
              className="bg-red-500 hover:bg-red-600 text-white py-1 px-3 rounded"
            >
              Logout
            </button>
          </>
        ) : (
          <span><a href="/login">Please sign in.</a></span>
        )}
      </div>
    </header>
  );
};

export default Header;

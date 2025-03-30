// src/components/Header.tsx
import React from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Mode } from '../types';

interface HeaderProps {
  mode: Mode;
}

const Header: React.FC<HeaderProps> = ({ mode }) => {
  const { user, logout } = useAuth();

  // Only show user info if the mode is not "simple" or "simple voting"
  const showUserInfo = mode !== Mode.Simple && mode !== Mode.Voting;

  return (
    <header className="p-4 bg-gray-800 text-white flex items-center justify-between">
      <h1 className="text-xl font-bold">Rave 2 Gether</h1>
      <div>
        {showUserInfo && (
          user ? (
            <>
              <span className="mr-4">
                Welcome, {user.username}
                {mode === Mode.UserCoin && user.coins !== undefined && ` (${user.coins} coins)`}
              </span>
              <button
                onClick={logout}
                className="bg-red-500 hover:bg-red-600 text-white py-1 px-3 rounded"
              >
                Logout
              </button>
            </>
          ) : (
            <span>
              <a href="/login" className="underline">
                Please sign in.
              </a>
            </span>
          )
        )}
      </div>
    </header>
  );
};

export default Header;

// src/contexts/AuthContext.tsx
import React, { createContext, useContext, useState, useEffect, ReactNode, FC } from 'react';
import api from "../api";

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

export interface User {
  username: string;
  role: string;
  coins: number;
  // add any additional properties that /api/self returns
}

interface AuthContextType {
  user: User | null;
  tokens: AuthTokens | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  refreshAccessToken: () => Promise<void>;
  loadUserInfo: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: FC<{ children: ReactNode }> = ({ children }) => {
  // Initialize tokens from localStorage if they exist
  const [tokens, setTokens] = useState<AuthTokens | null>(() => {
    const accessToken = localStorage.getItem('token');
    const refreshToken = localStorage.getItem('refreshToken');
    return accessToken && refreshToken ? { accessToken, refreshToken } : null;
  });
  const [user, setUser] = useState<User | null>(null);

  // Load user info from /api/self using the current access token
  const loadUserInfo = async (): Promise<void> => {
    if (!tokens) {
      setUser(null);
      return;
    }
    try {
      const response = await api.get('/self');
      const userData: User = await response.data;
      setUser(userData);
    } catch (error) {
      console.error('Error loading user info', error);
      setUser(null);
    }
  };

  // Login: send credentials to /api/login, store tokens, then fetch user info
  const login = async (username: string, password: string): Promise<void> => {
    try {
      const response = await api.post("/token", {}, {
        auth: {
          username,
          password,
        },
      });
      setTokens(response.data);
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('refreshToken', response.data.refresh_token);
      await loadUserInfo();
    } catch (error) {
      console.error('Login failed', error);
      setUser(null);
    }
  };

  // Logout: clear tokens and user info
  const logout = (): void => {
    setTokens(null);
    setUser(null);
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
  };

  // Refresh the access token using the refresh token, then update user info
  const refreshAccessToken = async (): Promise<void> => {
    if (!tokens) return;
    try {
      const response = await api.post('/refresh', { refresh_token: tokens.refreshToken });
      const newTokens: AuthTokens = response.data;
      setTokens(newTokens);
      localStorage.setItem('token', newTokens.accessToken);
      await loadUserInfo();
    }
    catch (error) {
      console.error('Token refresh failed', error);
      logout();
    }
  };

  useEffect(() => {
    if (tokens) {
      loadUserInfo();
    }
  }, [tokens])

  return (
    <AuthContext.Provider
      value={{ user, tokens, login, logout, refreshAccessToken, loadUserInfo }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import agent from '@/api/agent';
import type { User, LoginRequest, Role } from '@/types';
import i18n from '@/i18n';

interface AuthContextType {
  user: User | null;
  role: Role | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [role, setRole] = useState<Role | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const loadUser = async () => {
    const token = agent.getToken();
    if (!token) {
      setIsLoading(false);
      return;
    }

    try {
      const userData = await agent.Account.me();
      setUser(userData);
      
      // Set language from user preference
      if (userData.language) {
        i18n.changeLanguage(userData.language);
        localStorage.setItem('language', userData.language);
      }
      
      // Check if user is root
      if (userData.root) {
        setRole('root');
      } else {
        // Check enrollments to determine if admin/tutor
        try {
          const enrollments = await agent.Account.getEnrollments();
          const maxRole = Math.max(...enrollments.map(e => e.role), 0);
          const userRole: Role = maxRole >= 2 ? 'admin' : maxRole >= 1 ? 'tutor' : 'student';
          setRole(userRole);
          localStorage.setItem('user_role', userRole);
        } catch {
          setRole('student');
        }
      }
    } catch (error: any) {
      // Only clear token if it's an authentication error
      if (error?.status === 401) {
        agent.clearToken();
        window.location.href = '/login';
      }
      // For other errors, keep the user logged in
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadUser();
  }, []);

  const login = async (credentials: LoginRequest) => {
    await agent.Auth.login(credentials);
    agent.setToken(); // Mark as logged in (backend uses cookies)
    
    try {
      const userData = await agent.Account.me();
      setUser(userData);
      
      // Set language from user preference
      if (userData.language) {
        i18n.changeLanguage(userData.language);
        localStorage.setItem('language', userData.language);
      }
      
      // Check if user is root
      if (userData.root) {
        setRole('root');
        localStorage.setItem('user_role', 'root');
      } else {
        // Check enrollments to determine role
        try {
          const enrollments = await agent.Account.getEnrollments();
          const maxRole = Math.max(...enrollments.map(e => e.role), 0);
          const userRole: Role = maxRole >= 2 ? 'admin' : maxRole >= 1 ? 'tutor' : 'student';
          setRole(userRole);
          localStorage.setItem('user_role', userRole);
        } catch {
          setRole('student');
          localStorage.setItem('user_role', 'student');
        }
      }
      
      localStorage.setItem('user_email', userData.email);
    } catch (error) {
      // If fetching user data fails, clear the login state
      agent.clearToken();
      throw error;
    }
  };

  // Destroy the server-side session (DELETE /auth/sessions) before clearing
  // local state. The network call is best-effort: even if it fails (e.g. the
  // session already expired) we must still clear the client so the user is
  // logged out locally.
  const logout = async () => {
    try {
      await agent.Auth.logout();
    } catch {
      // Ignore: the local state is cleared regardless below.
    }
    agent.clearToken();
    setUser(null);
    setRole(null);
  };

  const refreshUser = async () => {
    if (!agent.getToken()) return;
    
    try {
      const userData = await agent.Account.me();
      setUser(userData);
    } catch (error) {
      logout();
    }
  };

  const value: AuthContextType = {
    user,
    role,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    refreshUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

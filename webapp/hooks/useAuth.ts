// alpha/webapp/hooks/useAuth.ts
import { useState, useEffect } from 'react';

interface AuthState {
  isAuthenticated: boolean;
  isAdmin: boolean;
  isEmployee: boolean;
  user: { id: string; name: string; role: 'Admin' | 'Employee' } | null;
  loading: boolean;
}

const mockAdmin: AuthState = {
  isAuthenticated: true,
  isAdmin: true, // Default to Admin
  isEmployee: false,
  user: { id: 'admin-123', name: 'Admin User', role: 'Admin' },
  loading: false,
}

const mockEmployee: AuthState = {
  isAuthenticated: true,
  isAdmin: false, // Default to Admin
  isEmployee: true,
  user: { id: 'admin-123', name: 'Admin User', role: 'Employee' },
  loading: false,
}

export const useAuth = (): AuthState => {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    isAdmin: false,
    isEmployee: false,
    user: null,
    loading: true    
  });

  useEffect(() => {
    // For now, we'll simulate an authenticated admin user after a short delay.
    const simulateAuth = setTimeout(() => {
      setAuthState(mockAdmin);
    }, 500); // Simulate a network request

    return () => clearTimeout(simulateAuth);
  }, []); // Empty dependency array to run once

  return authState;
};

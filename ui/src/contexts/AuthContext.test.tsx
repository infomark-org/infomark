// Exercises the AuthContext state transitions: bootstrapping without a session,
// deriving the role from enrollments on login, and destroying the server
// session on logout.
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

// Mutable handles so individual tests can tune the mocked backend responses.
const login = vi.fn().mockResolvedValue(undefined);
const logout = vi.fn().mockResolvedValue(undefined);
const me = vi.fn();
const getEnrollments = vi.fn();
const getToken = vi.fn();
const setToken = vi.fn();
const clearToken = vi.fn();

vi.mock('@/api/agent', () => ({
  default: {
    Auth: { login: (...args: unknown[]) => login(...args), logout: () => logout() },
    Account: { me: () => me(), getEnrollments: () => getEnrollments() },
    getToken: () => getToken(),
    setToken: () => setToken(),
    clearToken: () => clearToken(),
  },
}));

// Avoid initialising the real i18next instance during the unit test.
vi.mock('@/i18n', () => ({ default: { changeLanguage: vi.fn() } }));

import { AuthProvider, useAuth } from '@/contexts/AuthContext';

// Small consumer that surfaces the context state and actions to the DOM.
const Consumer: React.FC = () => {
  const { user, role, isAuthenticated, login: doLogin, logout: doLogout } = useAuth();
  return (
    <div>
      <span data-testid="auth">{isAuthenticated ? 'yes' : 'no'}</span>
      <span data-testid="email">{user?.email ?? ''}</span>
      <span data-testid="role">{role ?? ''}</span>
      <button onClick={() => doLogin({ email: 'a@b.de', plain_password: 'x' })}>login</button>
      <button onClick={() => doLogout()}>logout</button>
    </div>
  );
};

const renderProvider = () =>
  render(
    <AuthProvider>
      <Consumer />
    </AuthProvider>,
  );

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});

describe('AuthContext', () => {
  it('starts unauthenticated when there is no stored session hint', async () => {
    getToken.mockReturnValue(null);
    renderProvider();
    await waitFor(() => expect(screen.getByTestId('auth')).toHaveTextContent('no'));
    expect(me).not.toHaveBeenCalled();
  });

  it('logs in and derives an admin role from the highest enrollment', async () => {
    getToken.mockReturnValue(null);
    me.mockResolvedValue({ id: 1, first_name: 'Ada', last_name: 'Lovelace', email: 'ada@example.com', root: false });
    getEnrollments.mockResolvedValue([
      { id: 1, course_id: 1, role: 0 },
      { id: 2, course_id: 2, role: 2 },
    ]);

    renderProvider();
    await userEvent.click(screen.getByText('login'));

    await waitFor(() => expect(screen.getByTestId('email')).toHaveTextContent('ada@example.com'));
    expect(login).toHaveBeenCalledWith({ email: 'a@b.de', plain_password: 'x' });
    expect(setToken).toHaveBeenCalled();
    expect(screen.getByTestId('role')).toHaveTextContent('admin');
    expect(screen.getByTestId('auth')).toHaveTextContent('yes');
  });

  it('destroys the server session and clears local state on logout', async () => {
    getToken.mockReturnValue(null);
    me.mockResolvedValue({ id: 1, first_name: 'Ada', last_name: 'Lovelace', email: 'ada@example.com', root: false });
    getEnrollments.mockResolvedValue([{ id: 1, course_id: 1, role: 0 }]);

    renderProvider();
    await userEvent.click(screen.getByText('login'));
    await waitFor(() => expect(screen.getByTestId('auth')).toHaveTextContent('yes'));

    await userEvent.click(screen.getByText('logout'));

    await waitFor(() => expect(screen.getByTestId('auth')).toHaveTextContent('no'));
    expect(logout).toHaveBeenCalledTimes(1);
    expect(clearToken).toHaveBeenCalled();
    expect(screen.getByTestId('email')).toHaveTextContent('');
  });
});

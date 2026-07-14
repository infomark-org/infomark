// A lightweight render test that verifies the login page mounts and exposes the
// essential fields and navigation links.
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// The page only needs the `login` action from the context; stub it so we do not
// have to mount the whole provider or hit the network.
vi.mock('@/contexts/AuthContext', () => ({
  useAuth: () => ({ login: vi.fn().mockResolvedValue(undefined) }),
}));

// Initialise i18n so translated strings ("Login", "Register") resolve.
import '@/i18n';
import Login from '@/pages/Login';

describe('Login page', () => {
  it('renders the login form with a password field and auth links', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>,
    );

    // Heading and submit button both use the "Login" label.
    expect(screen.getByRole('heading', { name: /login/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /login/i })).toBeInTheDocument();

    // A masked password input must be present.
    const passwordInput = document.querySelector('input[type="password"]');
    expect(passwordInput).not.toBeNull();

    // Navigation affordances to registration and password recovery.
    expect(screen.getByRole('link', { name: /register/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /forgot password/i })).toBeInTheDocument();
  });
});

// Verify the colour-mode toggle flips state and persists the choice to
// localStorage so it survives reloads.
import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ThemeProvider, useThemeMode } from '@/contexts/ThemeContext';

const Probe = () => {
  const { mode, toggle } = useThemeMode();
  return (
    <button onClick={toggle} data-mode={mode}>
      {mode}
    </button>
  );
};

beforeEach(() => {
  localStorage.clear();
});

describe('ThemeContext', () => {
  it('defaults to light and persists a toggle to dark', async () => {
    render(
      <ThemeProvider>
        <Probe />
      </ThemeProvider>,
    );

    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('data-mode', 'light');

    await act(async () => {
      await userEvent.click(button);
    });

    expect(screen.getByRole('button')).toHaveAttribute('data-mode', 'dark');
    expect(localStorage.getItem('color_mode')).toBe('dark');
  });

  it('reads the persisted mode on mount', () => {
    localStorage.setItem('color_mode', 'dark');
    render(
      <ThemeProvider>
        <Probe />
      </ThemeProvider>,
    );
    expect(screen.getByRole('button')).toHaveAttribute('data-mode', 'dark');
  });
});

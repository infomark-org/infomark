import React, { createContext, useContext, useEffect, useMemo, useState, ReactNode } from 'react';
import { ConfigProvider, theme as antdTheme } from 'antd';
import { colors } from '@/theme/colors';

// The two visual modes the app supports. The value is persisted verbatim in
// localStorage so a returning user keeps their choice across sessions.
type ThemeMode = 'light' | 'dark';

const STORAGE_KEY = 'color_mode';

interface ThemeContextType {
  mode: ThemeMode;
  toggle: () => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

// Resolve the initial mode from localStorage, falling back to the OS
// preference so first-time dark-mode users get a sensible default. Reading
// matchMedia defensively keeps this safe under jsdom (tests) where the API is
// only a stub.
const readInitialMode = (): ThemeMode => {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === 'light' || stored === 'dark') return stored;
  const prefersDark =
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-color-scheme: dark)').matches;
  return prefersDark ? 'dark' : 'light';
};

interface ThemeProviderProps {
  children: ReactNode;
}

// ThemeProvider owns the single Ant Design ConfigProvider for the whole app so
// that switching the algorithm (default vs. dark) re-themes every component at
// once. The brand tokens stay constant; only the algorithm changes.
export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [mode, setMode] = useState<ThemeMode>(readInitialMode);

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, mode);
    // Reflect the mode on the root element so non-antd surfaces (e.g. the
    // <pre> test-log blocks) can react via a data attribute if needed.
    document.documentElement.dataset.colorMode = mode;
  }, [mode]);

  const toggle = () => setMode((prev) => (prev === 'light' ? 'dark' : 'light'));

  const value = useMemo<ThemeContextType>(() => ({ mode, toggle }), [mode]);

  const isDark = mode === 'dark';

  return (
    <ThemeContext.Provider value={value}>
      <ConfigProvider
        theme={{
          algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
          token: {
            colorPrimary: colors.primary,
            colorInfo: colors.info,
            colorError: colors.error,
            borderRadius: 8,
            fontSize: 14,
          },
          components: {
            Layout: {
              // In dark mode let the algorithm pick the header colour; in light
              // mode keep the brand purple bar.
              headerBg: isDark ? undefined : colors.primary,
              headerColor: '#ffffff',
            },
            Button: {
              primaryColor: '#ffffff',
              borderRadius: 6,
            },
            Card: {
              borderRadiusLG: 12,
            },
          },
        }}
      >
        {children}
      </ConfigProvider>
    </ThemeContext.Provider>
  );
};

export const useThemeMode = (): ThemeContextType => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useThemeMode must be used within a ThemeProvider');
  }
  return context;
};

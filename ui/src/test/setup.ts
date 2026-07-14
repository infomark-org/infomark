// Global test setup: register jest-dom matchers (toBeInTheDocument, etc.) and
// reset the DOM/mocks between tests so cases stay isolated.
import '@testing-library/jest-dom/vitest';
import { afterEach, vi } from 'vitest';
import { cleanup } from '@testing-library/react';

// Ant Design components call `window.matchMedia` for responsive behaviour, but
// jsdom does not implement it. Provide a minimal no-op stub so components render.
if (!window.matchMedia) {
  window.matchMedia = (query: string) =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }) as unknown as MediaQueryList;
}

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

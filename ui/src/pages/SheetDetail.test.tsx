// Smoke test: SheetDetail must render a student's task list with the submission
// affordance, driven entirely by a mocked API layer (no network).
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

const { get, getAll, getPoints, getResult } = vi.hoisted(() => ({
  get: vi.fn(),
  getAll: vi.fn(),
  getPoints: vi.fn(),
  getResult: vi.fn(),
}));

vi.mock('@/api/agent', () => ({
  default: {
    Account: { getEnrollments: () => Promise.resolve([{ id: 0, course_id: 1, role: 0 }]) },
    Sheets: {
      get,
      getPoints,
    },
    Tasks: {
      getAll,
      getResult,
      submitTask: vi.fn(),
      downloadOwnSubmission: vi.fn(),
    },
  },
}));

import '@/i18n';
import SheetDetail from '@/pages/SheetDetail';

beforeEach(() => {
  vi.clearAllMocks();
  // Published in the past, due in the future -> submission enabled.
  get.mockResolvedValue({
    id: 1,
    name: 'Sheet 0',
    publish_at: '2020-01-01T00:00:00Z',
    due_at: '2999-01-01T00:00:00Z',
  });
  getAll.mockResolvedValue([{ id: 1, name: 'task 0', max_points: 13 }]);
  getPoints.mockResolvedValue([{ acquired_points: 5, achievable_points: 13, max_points: 13, task_id: 1 }]);
  getResult.mockRejectedValue({ status: 404 });
});

describe('SheetDetail (student)', () => {
  it('renders the task list and an enabled submit button', async () => {
    render(
      <MemoryRouter initialEntries={['/courses/1/sheets/1']}>
        <Routes>
          <Route path="/courses/:courseId/sheets/:sheetId" element={<SheetDetail />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByText('task 0')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /submit/i })).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: /submit/i })).not.toBeDisabled();
  });
});

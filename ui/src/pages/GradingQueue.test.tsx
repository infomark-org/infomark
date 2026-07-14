// Smoke test: the tutor grading queue lists ungraded submissions and links to
// the single-grade editor. The API layer is mocked.
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

const { getMissingGrades } = vi.hoisted(() => ({ getMissingGrades: vi.fn() }));

vi.mock('@/api/agent', () => ({
  default: { Courses: { getMissingGrades } },
}));

import '@/i18n';
import GradingQueue from '@/pages/GradingQueue';

beforeEach(() => {
  vi.clearAllMocks();
  getMissingGrades.mockResolvedValue([
    {
      grade: {
        id: 8,
        submission_id: 31,
        public_test_status: 1,
        user: { id: 5, first_name: 'Erika', last_name: 'Muster', email: 'erika@uni.de' },
      },
      course_id: 1,
      sheet_id: 1,
      task_id: 2,
    },
  ]);
});

describe('GradingQueue', () => {
  it('lists an ungraded submission with the student name', async () => {
    render(
      <MemoryRouter initialEntries={['/courses/1/grading']}>
        <Routes>
          <Route path="/courses/:courseId/grading" element={<GradingQueue />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByText('Erika Muster')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /open/i })).toBeInTheDocument();
  });
});

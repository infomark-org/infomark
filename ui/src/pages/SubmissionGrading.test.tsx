// Smoke test: the single-grade editor loads a grade by its grade_id, shows the
// student, the public-test verdict and the grading form. The API layer is
// mocked, and we assert the i18n-driven labels render (no hardcoded English).
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

const mocks = vi.hoisted(() => ({
  gradesGet: vi.fn(),
  gradesUpdate: vi.fn(),
  downloadFile: vi.fn(),
}));

vi.mock('@/api/agent', () => ({
  default: {
    Grades: { get: mocks.gradesGet, update: mocks.gradesUpdate },
    Submissions: { downloadFile: mocks.downloadFile },
  },
}));

import '@/i18n';
import SubmissionGrading from '@/pages/SubmissionGrading';

beforeEach(() => {
  vi.clearAllMocks();
  // GradeResponse shape: keyed by grade id, carries submission_id, the public
  // test verdict/log and the embedded student user.
  mocks.gradesGet.mockResolvedValue({
    id: 8,
    submission_id: 31,
    acquired_points: 0,
    feedback: '',
    file_url: 'http://localhost:2020/api/v1/courses/1/submissions/31/file',
    public_test_status: 1,
    public_test_log: 'ok: 3 assertions passed',
    updated_at: '2026-02-01T10:00:00Z',
    user: { id: 5, first_name: 'Erika', last_name: 'Muster', email: 'erika@uni.de' },
  });
});

describe('SubmissionGrading', () => {
  it('renders the graded submission with translated labels', async () => {
    render(
      <MemoryRouter initialEntries={['/courses/1/grades/8']}>
        <Routes>
          <Route path="/courses/:courseId/grades/:gradeId" element={<SubmissionGrading />} />
        </Routes>
      </MemoryRouter>,
    );

    // Student identity is loaded from the mocked grade.
    expect(await screen.findByText('Erika Muster')).toBeInTheDocument();
    // The English resource for grades.gradeSubmission is "Grade submission".
    expect(screen.getByText('Grade submission')).toBeInTheDocument();
    // The download affordance uses the translated grades.downloadSubmission.
    expect(screen.getByRole('button', { name: /download submission/i })).toBeInTheDocument();
    // The save button uses grades.saveGrade.
    expect(screen.getByRole('button', { name: /save grade/i })).toBeInTheDocument();
    // The grade must be fetched by its own id, not the submission id.
    expect(mocks.gradesGet).toHaveBeenCalledWith(1, 8);
  });
});

// Smoke test: CourseDetail renders a student's course with the sheet list and
// the student-facing tabs (grades), driven by a fully mocked API layer.
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

vi.mock('@/contexts/AuthContext', () => ({
  useAuth: () => ({ user: { id: 12 }, role: 'student' }),
}));

const mocks = vi.hoisted(() => ({
  getEnrollments: vi.fn(),
  courseGet: vi.fn(),
  sheetsGetAll: vi.fn(),
  materialsGetAll: vi.fn(),
  getPoints: vi.fn(),
}));

vi.mock('@/api/agent', () => ({
  default: {
    Account: { getEnrollments: mocks.getEnrollments },
    Courses: { get: mocks.courseGet, getPoints: mocks.getPoints, getEnrollments: vi.fn() },
    Sheets: { getAll: mocks.sheetsGetAll },
    Materials: { getAll: mocks.materialsGetAll },
    Groups: { getAll: vi.fn().mockResolvedValue([]) },
    Grades: { getSummary: vi.fn() },
  },
}));

// The lazy course panels each fetch on mount; stub them so this test stays
// scoped to the CourseDetail shell.
vi.mock('@/components/course/ExamsPanel', () => ({ ExamsPanel: () => <div>exams-panel</div> }));
vi.mock('@/components/course/GroupsPanel', () => ({ GroupsPanel: () => <div>groups-panel</div> }));
vi.mock('@/components/course/EnrollmentsPanel', () => ({ EnrollmentsPanel: () => <div>enrollments-panel</div> }));

import '@/i18n';
import CourseDetail from '@/pages/CourseDetail';

beforeEach(() => {
  vi.clearAllMocks();
  mocks.getEnrollments.mockResolvedValue([{ id: 0, course_id: 1, role: 0 }]);
  mocks.courseGet.mockResolvedValue({
    id: 1,
    name: 'Info 2',
    description: 'A course',
    begins_at: '2026-01-01T00:00:00Z',
    ends_at: '2026-12-31T00:00:00Z',
    required_percentage: 50,
  });
  mocks.sheetsGetAll.mockResolvedValue([
    { id: 1, name: 'Sheet 0', publish_at: '2020-01-01T00:00:00Z', due_at: '2999-01-01T00:00:00Z' },
  ]);
  mocks.materialsGetAll.mockResolvedValue([]);
  mocks.getPoints.mockResolvedValue([{ acquired_points: 5, achievable_points: 10, max_points: 10, sheet_id: 1 }]);
});

describe('CourseDetail (student)', () => {
  it('renders the course name and the sheet list', async () => {
    render(
      <MemoryRouter initialEntries={['/courses/1']}>
        <Routes>
          <Route path="/courses/:courseId" element={<CourseDetail />} />
        </Routes>
      </MemoryRouter>,
    );

    expect((await screen.findAllByText('Info 2')).length).toBeGreaterThan(0);
    expect(await screen.findByText('Sheet 0')).toBeInTheDocument();
  });
});

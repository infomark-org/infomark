// Regression test for the root user-edit flow: the semester field must submit
// a JSON number, not a string. A plain `<Input type="number">` stores the
// value as a string in AntD, which the backend UserRequest.Semester (int)
// rejects with a 400 unmarshal error. `<InputNumber>` yields a real number.
// This test opens the editor, edits semester, saves, and asserts the payload
// type reaching agent.Users.update.
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';

vi.mock('@/contexts/AuthContext', () => ({
  useAuth: () => ({ role: 'root' }),
}));

const mocks = vi.hoisted(() => ({
  find: vi.fn(),
  update: vi.fn(),
}));

vi.mock('@/api/agent', () => ({
  default: {
    Users: { find: mocks.find, update: mocks.update },
  },
}));

import '@/i18n';
import UserSearch from '@/pages/UserSearch';

beforeEach(() => {
  vi.clearAllMocks();
  mocks.find.mockResolvedValue([
    {
      id: 5,
      first_name: 'Erika',
      last_name: 'Muster',
      email: 'erika@uni.de',
      student_number: '123456',
      semester: 3,
      subject: 'CS',
      language: 'en',
      root: false,
    },
  ]);
  mocks.update.mockResolvedValue({});
});

describe('UserSearch root edit', () => {
  it('submits semester as a number, not a string', async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <UserSearch />
      </MemoryRouter>,
    );

    // Wait for the debounced search to populate the table, then open the editor.
    await user.click(await screen.findByRole('button', { name: /edit/i }));

    // The semester field is an AntD InputNumber (role spinbutton). Replace its
    // value and save.
    const semester = await screen.findByRole('spinbutton');
    await user.clear(semester);
    await user.type(semester, '5');

    await user.click(screen.getByRole('button', { name: /save/i }));

    // The update must have been dispatched with a numeric semester.
    expect(mocks.update).toHaveBeenCalledTimes(1);
    const [userId, payload] = mocks.update.mock.calls[0];
    expect(userId).toBe(5);
    expect(typeof payload.semester).toBe('number');
    expect(payload.semester).toBe(5);
  });
});

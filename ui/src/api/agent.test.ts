// These tests pin the exact JSON payloads the API agent sends for the
// auth-adjacent flows, guarding against regressions of the contract bugs that
// broke registration, password reset, email confirmation and logout.
//
// The shapes are asserted against the backend request structs:
//   - account_requests.go   CreateUserAccountRequest  (nested user/account)
//   - auth_requests.go       UpdatePasswordRequest     (email + token + pw)
//   - auth_requests.go       ConfirmEmailRequest       (email + token)
//   - router.go              DELETE /auth/sessions      (logout)
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock the low-level HTTP client so we can inspect the exact call arguments.
// `vi.hoisted` makes the spies available inside the hoisted `vi.mock` factory.
const { post, put, del, get, patch } = vi.hoisted(() => ({
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
  get: vi.fn(),
  patch: vi.fn(),
}));

vi.mock('@/api/client', () => ({
  default: {
    post,
    put,
    delete: del,
    get,
    patch,
    upload: vi.fn(),
    download: vi.fn(),
    setToken: vi.fn(),
    clearToken: vi.fn(),
    getToken: vi.fn(),
  },
}));

import agent from '@/api/agent';

beforeEach(() => {
  vi.clearAllMocks();
  post.mockResolvedValue(undefined);
  put.mockResolvedValue(undefined);
  del.mockResolvedValue(undefined);
});

describe('Auth.register', () => {
  it('POSTs to /account with a nested { user, account } body', () => {
    agent.Auth.register({
      first_name: 'Max',
      last_name: 'Mustermensch',
      email: 'max@uni-tuebingen.de',
      student_number: '0815',
      semester: 3,
      subject: 'computer science',
      language: 'en',
      plain_password: 'supersecret',
    });

    expect(post).toHaveBeenCalledTimes(1);
    expect(post).toHaveBeenCalledWith('/account', {
      user: {
        first_name: 'Max',
        last_name: 'Mustermensch',
        email: 'max@uni-tuebingen.de',
        student_number: '0815',
        semester: 3,
        subject: 'computer science',
        language: 'en',
      },
      account: {
        email: 'max@uni-tuebingen.de',
        plain_password: 'supersecret',
      },
    });
  });

  it('sends identical email in user and account (backend enforces equality)', () => {
    agent.Auth.register({
      first_name: 'A',
      last_name: 'B',
      email: 'same@example.com',
      student_number: '1',
      semester: 1,
      subject: 'x',
      language: 'de',
      plain_password: 'password1',
    });

    const [, body] = post.mock.calls[0];
    expect(body.user.email).toBe(body.account.email);
  });
});

describe('Auth.updatePassword', () => {
  it('POSTs to /auth/update_password including the required email field', () => {
    agent.Auth.updatePassword({
      email: 'user@example.com',
      reset_password_token: 'reset-token-abc',
      plain_password: 'brandnewpw',
    });

    expect(post).toHaveBeenCalledWith('/auth/update_password', {
      email: 'user@example.com',
      reset_password_token: 'reset-token-abc',
      plain_password: 'brandnewpw',
    });
  });
});

describe('Auth.confirmEmail', () => {
  it('POSTs to /auth/confirm_email with { email, confirmation_token }', () => {
    agent.Auth.confirmEmail({
      email: 'user@example.com',
      confirmation_token: 'confirm-token-xyz',
    });

    expect(post).toHaveBeenCalledWith('/auth/confirm_email', {
      email: 'user@example.com',
      confirmation_token: 'confirm-token-xyz',
    });
  });
});

describe('Auth.logout', () => {
  it('issues DELETE /auth/sessions to destroy the server session', () => {
    agent.Auth.logout();
    expect(del).toHaveBeenCalledWith('/auth/sessions');
  });
});

describe('Courses.changeRole', () => {
  it('PUTs a flat { role } body to the enrollment (ChangeRoleInCourseRequest)', () => {
    agent.Courses.changeRole(1, 12, 1);
    expect(put).toHaveBeenCalledWith('/courses/1/enrollments/12', { role: 1 });
  });
});

describe('Users search and edit', () => {
  it('GETs /users/find with a query param', () => {
    agent.Users.find('robby');
    expect(get).toHaveBeenCalledWith('/users/find', { query: 'robby' });
  });

  it('PUTs the flat UserRequest body to /users/{id}', () => {
    const body = {
      first_name: 'Max',
      last_name: 'Mustermensch',
      email: 'max@uni-tuebingen.de',
      student_number: '0815',
      semester: 3,
      subject: 'cs',
      language: 'en',
    };
    agent.Users.update(7, body);
    expect(put).toHaveBeenCalledWith('/users/7', body);
  });
});

describe('Exams enroll/disenroll', () => {
  it('POSTs an empty body to the exam enrollment collection', () => {
    agent.Exams.enroll(1, 2);
    expect(post).toHaveBeenCalledWith('/courses/1/exams/2/enrollments', {});
  });

  it('DELETEs the exam enrollment collection to disenroll', () => {
    agent.Exams.disenroll(1, 2);
    expect(del).toHaveBeenCalledWith('/courses/1/exams/2/enrollments');
  });
});

describe('Groups.placeBid', () => {
  it('POSTs a { bid } body to the group bids endpoint (GroupBidRequest)', () => {
    agent.Groups.placeBid(1, 5, 7);
    expect(post).toHaveBeenCalledWith('/courses/1/groups/5/bids', { bid: 7 });
  });
});

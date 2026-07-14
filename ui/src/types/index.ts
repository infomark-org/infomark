// User & Authentication Types
export interface User {
  id: number;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  email: string;
  student_number?: string;
  semester?: number;
  subject?: string;
  language?: string;
  root?: boolean;
}

export type Role = 'student' | 'tutor' | 'admin' | 'root';

export interface LoginRequest {
  email: string;
  plain_password: string;
}

export interface RegisterRequest {
  first_name: string;
  last_name: string;
  email: string;
  student_number: string;
  semester: number;
  subject: string;
  language: string;
  plain_password: string;
}

export interface PasswordResetRequest {
  email: string;
}

// The backend `UpdatePasswordRequest.Bind` (auth_requests.go) requires `email`
// in addition to the token and new password, so it must be part of the payload.
export interface PasswordResetComplete {
  email: string;
  reset_password_token: string;
  plain_password: string;
}

// The backend `ConfirmEmailRequest.Bind` (auth_requests.go) requires `email`
// alongside the confirmation token.
export interface EmailConfirmation {
  email: string;
  confirmation_token: string;
}

// Course Types
export interface Course {
  id: number;
  name: string;
  description: string;
  begins_at: string;
  ends_at: string;
  required_percentage: number;
}

export interface CourseRole {
  role: number; // 0=student, 1=tutor, 2=admin
}

// Sheet Types
export interface Sheet {
  id: number;
  name: string;
  publish_at: string;
  due_at: string;
}

// Task Types
export interface Task {
  id: number;
  name: string;
  max_points: number;
  public_tests_url?: string;
  public_docker_image?: string;
  private_tests_url?: string;
  private_docker_image?: string;
}

// Material Types
//
// `kind`: 0 = slide, 1 = supplementary (material.go). `required_role` is the
// minimum course role that may download the file (0 student, 1 tutor, 2 admin);
// the backend `GetFileHandler` enforces it, so the UI hides the download for
// users below that role.
export interface Material {
  id: number;
  name: string;
  kind: number;
  publish_at: string;
  lecture_at: string;
  required_role: number;
  file_url?: string;
}

// Group Types. The tutor is embedded (group_responses.go GroupResponse).
export interface Group {
  id: number;
  course_id: number;
  tutor: User;
  description: string;
}

// `GET /courses/{id}/bids` returns the requesting student's bids across every
// group in the course (course_responses.go GroupBidsResponse).
export interface GroupBid {
  id: number;
  user_id: number;
  group_id: number;
  bid: number;
}

export interface GroupEnrollmentChange {
  user_id: number;
  group_id: number;
}

// Grade Types
export interface Grade {
  id: number;
  public_execution_state: number;
  private_execution_state: number;
  public_test_log: string;
  private_test_log: string;
  public_test_status: number;
  private_test_status: number;
  acquired_points: number;
  feedback: string;
  tutor_id: number;
  submission_id: number;
  file_url: string;
  updated_at: string;
  // Present on GradeResponse (grade_responses.go): the graded student.
  user?: {
    id: number;
    first_name: string;
    last_name: string;
    email: string;
  };
}

// `GET /courses/{id}/grades/missing` returns, for the requesting tutor, every
// grade that still lacks feedback (grade_responses.go MissingGradeResponse).
// The embedded `grade` is a full GradeResponse and carries the graded student
// under `grade.user`; the sheet/task ids locate it in the course structure.
export interface MissingGrade {
  grade: Grade;
  course_id: number;
  sheet_id: number;
  task_id: number;
}

// Submission Types
export interface Submission {
  id: number;
  user_id: number;
  task_id: number;
  file_url: string;
  created_at: string;
}

export interface TaskRatingResponse {
  rating: number;
}

// Enrollment Types. Matches the backend `EnrollmentResponse`
// (course_responses.go): the wire shape carries only the embedded `user` and
// the course `role`; there is no top-level enrollment id.
export interface Enrollment {
  user: User;
  role: number;
}

export interface UserEnrollment {
  user_id: number;
  role: number;
}

// Matches the backend `UserEnrollmentResponse` (account_responses.go):
// { id, course_id, role }. The full course object is fetched separately.
export interface AccountEnrollment {
  id: number;
  course_id: number;
  role: number;
}

// Exam Types
export interface Exam {
  id: number;
  name: string;
  description: string;
  exam_time: string;
  course_id: number;
}

// `GET /account/exams/enrollments` and the exam enroll endpoints return this
// shape (exam_responses.go ExamEnrollmentResponse). `status`: 0 = enrolled but
// not yet marked; a non-zero status means the exam has been marked and the
// student can no longer disenroll (exam.go DisenrollExamHandler).
export interface ExamEnrollment {
  status: number;
  mark: string;
  user_id: number;
  course_id: number;
  exam_id: number;
}

// Mail Types
export interface Mail {
  subject: string;
  body: string;
}

// Point Overview Types
//
// `GET /courses/{id}/points` returns one entry per sheet for the *requesting*
// user (course.go PointsHandler -> SheetPointsResponse in course_responses.go).
export interface CoursePoints {
  acquired_points: number;
  achievable_points: number;
  max_points: number;
  sheet_id: number;
}

// `GET /courses/{id}/sheets/{id}/points` returns one entry per task for the
// requesting user (sheet.go PointsHandler -> TaskPointsResponse in
// sheet_responses.go).
export interface TaskPoints {
  acquired_points: number;
  achievable_points: number;
  max_points: number;
  task_id: number;
}

// `GET /courses/{id}/grades/summary` returns the per-student point matrix for
// tutors/admins (grade.go IndexSummaryHandler -> GradeOverviewResponse).
export interface GradeOverviewSheet {
  id: number;
  name: string;
}

export interface GradeOverviewAchievement {
  user_info: {
    id: number;
    first_name: string;
    last_name: string;
    student_number: string;
    email: string;
  };
  points: number[];
}

export interface GradeOverview {
  sheets: GradeOverviewSheet[];
  achievements: GradeOverviewAchievement[];
}

// Account Types
export interface Account {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  student_number: string;
  semester: number;
  subject: string;
  language: string;
}

export interface AccountUpdate {
  first_name?: string;
  last_name?: string;
  email?: string;
  student_number?: string;
  semester?: number;
  subject?: string;
  language?: string;
  old_plain_password?: string;
  new_plain_password?: string;
}

// Admin edit of another user via PUT /users/{id} (user_requests.go UserRequest).
// The backend requires every profile field; `plain_password` is optional and
// only sent when the admin is resetting the user's password.
export interface UserUpdate {
  first_name: string;
  last_name: string;
  email: string;
  student_number: string;
  semester: number;
  subject: string;
  language: string;
  plain_password?: string;
}

// Error Types
export interface ApiError {
  message: string;
  status?: number;
}

// Utility Types
export type AsyncState<T> = {
  data: T | null;
  loading: boolean;
  error: ApiError | null;
};

export type PaginationParams = {
  limit?: number;
  offset?: number;
};

export type FilterParams = {
  search?: string;
  role?: number;
};

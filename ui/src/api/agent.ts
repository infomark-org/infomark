import apiClient from './client';
import type {
  User,
  Course,
  Sheet,
  Task,
  Material,
  Group,
  Grade,
  Submission,
  Enrollment,
  Exam,
  ExamEnrollment,
  Mail,
  LoginRequest,
  RegisterRequest,
  PasswordResetRequest,
  PasswordResetComplete,
  EmailConfirmation,
  AccountUpdate,
  UserUpdate,
  GroupBid,
  GroupEnrollmentChange,
  UserEnrollment,
  AccountEnrollment,
  MissingGrade,
  CoursePoints,
  TaskPoints,
  GradeOverview,
  TaskRatingResponse,
  PaginationParams,
  FilterParams,
} from '@/types';

const Auth = {
  // The session cookie is set via the response headers; the JSON body is not
  // consumed by the client (see AuthContext.login), so the call is typed void.
  login: (credentials: LoginRequest) =>
    apiClient.post<void>('/auth/sessions', credentials),

  // CRITICAL PATH: registration payload contract.
  // The backend registration endpoint is `POST /account`
  // (account.go CreateHandler / account_requests.go CreateUserAccountRequest).
  // It expects a nested body `{ user: {...}, account: { email, plain_password } }`
  // and rejects the request unless `user.email === account.email`.
  register: (data: RegisterRequest) =>
    apiClient.post<User>('/account', {
      user: {
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email,
        student_number: data.student_number,
        semester: data.semester,
        subject: data.subject,
        language: data.language,
      },
      account: {
        email: data.email,
        plain_password: data.plain_password,
      },
    }),

  logout: () => apiClient.delete<void>('/auth/sessions'),

  requestPasswordReset: (data: PasswordResetRequest) =>
    apiClient.post<void>('/auth/request_password_reset', data),
  
  updatePassword: (data: PasswordResetComplete) =>
    apiClient.post<void>('/auth/update_password', data),
  
  confirmEmail: (data: EmailConfirmation) =>
    apiClient.post<void>('/auth/confirm_email', data),
};

const Account = {
  me: () => apiClient.get<User>('/me'),
  
  updateMe: (data: Partial<User>) => apiClient.put<User>('/me', data),
  
  get: () => apiClient.get<User>('/account'),
  
  update: (data: AccountUpdate) => {
    const { old_plain_password, new_plain_password, ...accountFields } = data;
    const payload: any = { account: accountFields };
    
    if (new_plain_password) {
      payload.account.plain_password = new_plain_password;
    }
    
    // Only include old_plain_password if we're changing email or password
    if (accountFields.email !== undefined || new_plain_password) {
      payload.old_plain_password = old_plain_password || '';
    }
    
    return apiClient.patch<User>('/account', payload);
  },
  
  uploadAvatar: (file: File) =>
    apiClient.upload<any>('/account/avatar', file),
  
  getEnrollments: () =>
    apiClient.get<AccountEnrollment[]>('/account/enrollments'),
  
  getExamEnrollments: () =>
    apiClient.get<ExamEnrollment[]>('/account/exams/enrollments'),
};

const Users = {
  getAll: () => apiClient.get<User[]>('/users'),

  get: (id: number) => apiClient.get<User>(`/users/${id}`),

  // Root/admin fuzzy search over first name, last name and email
  // (user.go Find -> GET /users/find?query=...). Empty query returns all.
  find: (query: string) =>
    apiClient.get<User[]>('/users/find', { query }),

  // PUT /users/{id} expects the flat UserRequest (user_requests.go): every
  // profile field is required, `plain_password` is optional.
  update: (id: number, data: UserUpdate) =>
    apiClient.put<User>(`/users/${id}`, data),
};

const Courses = {
  getAll: () => apiClient.get<Course[]>('/courses'),
  
  get: (id: number) => apiClient.get<Course>(`/courses/${id}`),
  
  create: (data: Partial<Course>) =>
    apiClient.post<Course>('/courses', data),
  
  update: (id: number, data: Partial<Course>) =>
    apiClient.put<Course>(`/courses/${id}`, data),
  
  delete: (id: number) => apiClient.delete<void>(`/courses/${id}`),
  
  getEnrollments: (id: number, params?: FilterParams) =>
    apiClient.get<Enrollment[]>(`/courses/${id}/enrollments`, params),
  
  enroll: (id: number, data: UserEnrollment) =>
    apiClient.post<void>(`/courses/${id}/enrollments`, data),
  
  updateEnrollment: (courseId: number, userId: number, data: UserEnrollment) =>
    apiClient.put<void>(`/courses/${courseId}/enrollments/${userId}`, data),

  // PUT /courses/{id}/enrollments/{user_id} with the flat ChangeRoleInCourse
  // body `{ role }` (course.go ChangeRole). 0 student, 1 tutor, 2 admin.
  changeRole: (courseId: number, userId: number, role: number) =>
    apiClient.put<void>(`/courses/${courseId}/enrollments/${userId}`, { role }),
  
  deleteEnrollment: (courseId: number, userId: number) =>
    apiClient.delete<void>(`/courses/${courseId}/enrollments/${userId}`),
  
  getPoints: (id: number) =>
    apiClient.get<CoursePoints[]>(`/courses/${id}/points`),
  
  getMissingGrades: (id: number) =>
    apiClient.get<MissingGrade[]>(`/courses/${id}/grades/missing`),
  
  getMissingTasks: (id: number) =>
    apiClient.get<any[]>(`/courses/${id}/tasks/missing`),
  
  sendEmail: (id: number, mail: Mail) =>
    apiClient.post<void>(`/courses/${id}/emails`, mail),
};

const Sheets = {
  getAll: (courseId: number) =>
    apiClient.get<Sheet[]>(`/courses/${courseId}/sheets`),
  
  get: (courseId: number, sheetId: number) =>
    apiClient.get<Sheet>(`/courses/${courseId}/sheets/${sheetId}`),
  
  create: (courseId: number, data: Partial<Sheet>) =>
    apiClient.post<Sheet>(`/courses/${courseId}/sheets`, data),
  
  update: (courseId: number, sheetId: number, data: Partial<Sheet>) =>
    apiClient.put<Sheet>(`/courses/${courseId}/sheets/${sheetId}`, data),
  
  delete: (courseId: number, sheetId: number) =>
    apiClient.delete<void>(`/courses/${courseId}/sheets/${sheetId}`),
  
  uploadFile: (courseId: number, sheetId: number, file: File) =>
    apiClient.upload<void>(`/courses/${courseId}/sheets/${sheetId}/file`, file),
  
  downloadFile: (courseId: number, sheetId: number, filename: string) =>
    apiClient.download(`/courses/${courseId}/sheets/${sheetId}/file`, filename),
  
  getPoints: (courseId: number, sheetId: number) =>
    apiClient.get<TaskPoints[]>(`/courses/${courseId}/sheets/${sheetId}/points`),
};

const Tasks = {
  getAll: (courseId: number, sheetId: number) =>
    apiClient.get<Task[]>(`/courses/${courseId}/sheets/${sheetId}/tasks`),
  
  get: (courseId: number, taskId: number) =>
    apiClient.get<Task>(`/courses/${courseId}/tasks/${taskId}`),
  
  create: (courseId: number, sheetId: number, data: Partial<Task>) =>
    apiClient.post<Task>(`/courses/${courseId}/sheets/${sheetId}/tasks`, data),
  
  update: (courseId: number, taskId: number, data: Partial<Task>) =>
    apiClient.put<Task>(`/courses/${courseId}/tasks/${taskId}`, data),
  
  delete: (courseId: number, taskId: number) =>
    apiClient.delete<void>(`/courses/${courseId}/tasks/${taskId}`),
  
  uploadPublicFile: (courseId: number, taskId: number, file: File) =>
    apiClient.upload<void>(`/courses/${courseId}/tasks/${taskId}/public_file`, file),
  
  uploadPrivateFile: (courseId: number, taskId: number, file: File) =>
    apiClient.upload<void>(`/courses/${courseId}/tasks/${taskId}/private_file`, file),
  
  // A student's own submission for a task is served as the raw zip
  // (submission.go GetFileHandler); there is no JSON metadata endpoint, so the
  // existence/points/feedback all come from `getResult` (the grade) instead.
  downloadOwnSubmission: (courseId: number, taskId: number, filename: string) =>
    apiClient.download(`/courses/${courseId}/tasks/${taskId}/submission`, filename),

  submitTask: (courseId: number, taskId: number, file: File) =>
    apiClient.upload<void>(`/courses/${courseId}/tasks/${taskId}/submission`, file),
  
  getResult: (courseId: number, taskId: number) =>
    apiClient.get<Grade>(`/courses/${courseId}/tasks/${taskId}/result`),
  
  getRating: (courseId: number, taskId: number) =>
    apiClient.get<TaskRatingResponse>(`/courses/${courseId}/tasks/${taskId}/ratings`),
  
  rateTask: (courseId: number, taskId: number, rating: number) =>
    apiClient.post<void>(`/courses/${courseId}/tasks/${taskId}/ratings`, { rating }),
};

const Materials = {
  getAll: (courseId: number) =>
    apiClient.get<Material[]>(`/courses/${courseId}/materials`),
  
  get: (courseId: number, materialId: number) =>
    apiClient.get<Material>(`/courses/${courseId}/materials/${materialId}`),
  
  create: (courseId: number, data: Partial<Material>) =>
    apiClient.post<Material>(`/courses/${courseId}/materials`, data),
  
  update: (courseId: number, materialId: number, data: Partial<Material>) =>
    apiClient.put<Material>(`/courses/${courseId}/materials/${materialId}`, data),
  
  delete: (courseId: number, materialId: number) =>
    apiClient.delete<void>(`/courses/${courseId}/materials/${materialId}`),
  
  uploadFile: (courseId: number, materialId: number, file: File) =>
    apiClient.upload<void>(`/courses/${courseId}/materials/${materialId}/file`, file),
  
  downloadFile: (courseId: number, materialId: number, filename: string) =>
    apiClient.download(`/courses/${courseId}/materials/${materialId}/file`, filename),
};

const Groups = {
  getAll: (courseId: number) =>
    apiClient.get<Group[]>(`/courses/${courseId}/groups`),
  
  // GET /courses/{id}/groups/own returns a *list* (group.go GetMineHandler):
  // a student in a group gets a one-element array, otherwise it is empty.
  getOwn: (courseId: number) =>
    apiClient.get<Group[]>(`/courses/${courseId}/groups/own`),
  
  get: (courseId: number, groupId: number) =>
    apiClient.get<Group>(`/courses/${courseId}/groups/${groupId}`),
  
  create: (courseId: number, data: Partial<Group>) =>
    apiClient.post<Group>(`/courses/${courseId}/groups`, data),
  
  update: (courseId: number, groupId: number, data: Partial<Group>) =>
    apiClient.put<Group>(`/courses/${courseId}/groups/${groupId}`, data),
  
  delete: (courseId: number, groupId: number) =>
    apiClient.delete<void>(`/courses/${courseId}/groups/${groupId}`),
  
  getEnrollments: (courseId: number, groupId: number, params?: FilterParams) =>
    apiClient.get<Enrollment[]>(`/courses/${courseId}/groups/${groupId}/enrollments`, params),
  
  changeEnrollment: (courseId: number, groupId: number, data: GroupEnrollmentChange) =>
    apiClient.post<void>(`/courses/${courseId}/groups/${groupId}/enrollments`, data),
  
  getBids: (courseId: number) =>
    apiClient.get<GroupBid[]>(`/courses/${courseId}/bids`),
  
  placeBid: (courseId: number, groupId: number, bid: number) =>
    apiClient.post<void>(`/courses/${courseId}/groups/${groupId}/bids`, { bid }),
  
  sendEmail: (courseId: number, groupId: number, mail: Mail) =>
    apiClient.post<void>(`/courses/${courseId}/groups/${groupId}/emails`, mail),
  
  getSubmissionFile: (courseId: number, taskId: number, groupId: number, filename: string) =>
    apiClient.download(`/courses/${courseId}/tasks/${taskId}/groups/${groupId}/file`, filename),
};

const Grades = {
  getAll: (courseId: number, params?: PaginationParams) =>
    apiClient.get<Grade[]>(`/courses/${courseId}/grades`, params),
  
  get: (courseId: number, gradeId: number) =>
    apiClient.get<Grade>(`/courses/${courseId}/grades/${gradeId}`),
  
  update: (courseId: number, gradeId: number, data: Partial<Grade>) =>
    apiClient.put<Grade>(`/courses/${courseId}/grades/${gradeId}`, data),
  
  getSummary: (courseId: number, params?: { group_id?: number }) =>
    apiClient.get<GradeOverview>(`/courses/${courseId}/grades/summary`, params),
};

const Submissions = {
  getAll: (courseId: number, params?: PaginationParams) =>
    apiClient.get<Submission[]>(`/courses/${courseId}/submissions`, params),
  
  downloadFile: (courseId: number, submissionId: number, filename: string) =>
    apiClient.download(`/courses/${courseId}/submissions/${submissionId}/file`, filename),
};

const Exams = {
  getAll: (courseId: number) =>
    apiClient.get<Exam[]>(`/courses/${courseId}/exams`),
  
  get: (courseId: number, examId: number) =>
    apiClient.get<Exam>(`/courses/${courseId}/exams/${examId}`),
  
  create: (courseId: number, data: Partial<Exam>) =>
    apiClient.post<Exam>(`/courses/${courseId}/exams`, data),
  
  update: (courseId: number, examId: number, data: Partial<Exam>) =>
    apiClient.put<Exam>(`/courses/${courseId}/exams/${examId}`, data),
  
  delete: (courseId: number, examId: number) =>
    apiClient.delete<void>(`/courses/${courseId}/exams/${examId}`),
  
  getEnrollments: (courseId: number, examId: number) =>
    apiClient.get<ExamEnrollment[]>(`/courses/${courseId}/exams/${examId}/enrollments`),

  // Students self-enroll (201 -> the caller's full exam-enrollment list) and
  // may disenroll only while status === 0 (exam.go Enroll/Disenroll handlers).
  enroll: (courseId: number, examId: number) =>
    apiClient.post<ExamEnrollment[]>(`/courses/${courseId}/exams/${examId}/enrollments`, {}),

  disenroll: (courseId: number, examId: number) =>
    apiClient.delete<void>(`/courses/${courseId}/exams/${examId}/enrollments`),
};

const Terms = {
  get: () => apiClient.get<string>('/privacy_statement'),
};

const agent = {
  Auth,
  Account,
  Users,
  Courses,
  Sheets,
  Tasks,
  Materials,
  Groups,
  Grades,
  Submissions,
  Exams,
  Terms,
  setToken: () => apiClient.setToken(),
  clearToken: () => apiClient.clearToken(),
  getToken: () => apiClient.getToken(),
};

export default agent;

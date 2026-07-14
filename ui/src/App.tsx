import React from 'react';
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '@/contexts/AuthContext';
import { ThemeProvider } from '@/contexts/ThemeContext';
import { Layout } from '@/components/Layout';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { PageLoader } from '@/components/PageLoader';

// Lazy load pages
const Login = React.lazy(() => import('@/pages/Login'));
const Register = React.lazy(() => import('@/pages/Register'));
const Dashboard = React.lazy(() => import('@/pages/Dashboard'));
const Courses = React.lazy(() => import('@/pages/Courses'));
const CourseDetail = React.lazy(() => import('@/pages/CourseDetail'));
const CourseEditor = React.lazy(() => import('@/pages/CourseEditor'));
const SheetDetail = React.lazy(() => import('@/pages/SheetDetail'));
const SheetEditor = React.lazy(() => import('@/pages/SheetEditor'));
const MaterialEditor = React.lazy(() => import('@/pages/MaterialEditor'));
const GroupEditor = React.lazy(() => import('@/pages/GroupEditor'));
const SubmissionGrading = React.lazy(() => import('@/pages/SubmissionGrading'));
const GradingQueue = React.lazy(() => import('@/pages/GradingQueue'));
const MailEditor = React.lazy(() => import('@/pages/MailEditor'));
const ProfileEditor = React.lazy(() => import('@/pages/ProfileEditor'));
const RequestPasswordReset = React.lazy(() => import('@/pages/RequestPasswordReset'));
const PasswordReset = React.lazy(() => import('@/pages/PasswordReset'));
const MailConfirmation = React.lazy(() => import('@/pages/MailConfirmation'));
const Admin = React.lazy(() => import('@/pages/Admin'));
const UserSearch = React.lazy(() => import('@/pages/UserSearch'));
const Terms = React.lazy(() => import('@/pages/Terms'));

const App: React.FC = () => {
  return (
    <ThemeProvider>
      <HashRouter>
        <AuthProvider>
          <Layout>
            <React.Suspense fallback={<PageLoader />}>
              <Routes>
                {/* Public routes */}
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />
                <Route path="/request-password-reset" element={<RequestPasswordReset />} />
                {/* These paths mirror the links the backend emails out:
                    `{ExternalURL}/#/password_reset/{email}/{token}` and
                    `{ExternalURL}/#/confirmation/{email}/{token}`
                    (see api/app/account.go, auth.go and email/email.go). */}
                <Route path="/password_reset/:email/:token" element={<PasswordReset />} />
                <Route path="/confirmation/:email/:token" element={<MailConfirmation />} />
                <Route path="/terms" element={<Terms />} />

                {/* Protected routes */}
                <Route
                  path="/dashboard"
                  element={
                    <ProtectedRoute>
                      <Dashboard />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses"
                  element={
                    <ProtectedRoute>
                      <Courses />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId"
                  element={
                    <ProtectedRoute>
                      <CourseDetail />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/sheets/:sheetId"
                  element={
                    <ProtectedRoute>
                      <SheetDetail />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/profile"
                  element={
                    <ProtectedRoute>
                      <ProfileEditor />
                    </ProtectedRoute>
                  }
                />

                {/* Admin/Tutor routes */}
                <Route
                  path="/courses/new"
                  element={
                    <ProtectedRoute>
                      <CourseEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/edit"
                  element={
                    <ProtectedRoute>
                      <CourseEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/sheets/new"
                  element={
                    <ProtectedRoute>
                      <SheetEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/sheets/:sheetId/edit"
                  element={
                    <ProtectedRoute>
                      <SheetEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/materials/new"
                  element={
                    <ProtectedRoute>
                      <MaterialEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/materials/:materialId/edit"
                  element={
                    <ProtectedRoute>
                      <MaterialEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/groups/new"
                  element={
                    <ProtectedRoute>
                      <GroupEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/groups/:groupId/edit"
                  element={
                    <ProtectedRoute>
                      <GroupEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/grading"
                  element={
                    <ProtectedRoute>
                      <GradingQueue />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/grades/:gradeId"
                  element={
                    <ProtectedRoute>
                      <SubmissionGrading />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/courses/:courseId/mail"
                  element={
                    <ProtectedRoute>
                      <MailEditor />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/admin"
                  element={
                    <ProtectedRoute>
                      <Admin />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/admin/users"
                  element={
                    <ProtectedRoute>
                      <UserSearch />
                    </ProtectedRoute>
                  }
                />

                {/* 404 */}
                <Route path="*" element={<div>404 - Page Not Found</div>} />
              </Routes>
            </React.Suspense>
          </Layout>
        </AuthProvider>
      </HashRouter>
    </ThemeProvider>
  );
};

export default App;

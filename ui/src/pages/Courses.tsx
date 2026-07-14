import React, { useEffect, useState } from 'react';
import { Button, Modal, Spin, Empty, Tabs, message, Row, Col, Dropdown, FloatButton } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, LoginOutlined, LogoutOutlined, MoreOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import agent from '@/api/agent';
import { useAuth } from '@/contexts/AuthContext';
import { CourseCard } from '@/components/CourseCard';
import type { Course, AccountEnrollment } from '@/types';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';

const Courses: React.FC = () => {
  const { user, role } = useAuth();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [courses, setCourses] = useState<Course[]>([]);
  const [enrollments, setEnrollments] = useState<AccountEnrollment[]>([]);
  const [loading, setLoading] = useState(true);
  const [enrolling, setEnrolling] = useState<number | null>(null);
  const [disenrollModal, setDisenrollModal] = useState<{ visible: boolean; course: Course | null }>({ visible: false, course: null });
  const [deleteModal, setDeleteModal] = useState<{ visible: boolean; course: Course | null }>({ visible: false, course: null });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [coursesData, enrollmentsData] = await Promise.all([
        agent.Courses.getAll(),
        agent.Account.getEnrollments(),
      ]);
      setCourses(coursesData);
      setEnrollments(enrollmentsData);
    } catch (error) {
      message.error(t('course.loadCoursesFailed'));
    } finally {
      setLoading(false);
    }
  };

  const isEnrolled = (courseId: number) => enrollments.some(e => e.course_id === courseId);

  const handleEnroll = async (course: Course) => {
    setEnrolling(course.id);
    try {
      await agent.Courses.enroll(course.id, { user_id: user!.id, role: 0 });
      message.success(t('course.enrollSuccess'));
      fetchData();
    } catch (error) {
      message.error(t('course.enrollFailed'));
    } finally {
      setEnrolling(null);
    }
  };

  const handleDisenroll = async () => {
    if (!disenrollModal.course) return;
    try {
      await agent.Courses.deleteEnrollment(disenrollModal.course.id, user!.id);
      message.success(t('course.disenrollSuccess'));
      setDisenrollModal({ visible: false, course: null });
      fetchData();
    } catch (error: any) {
      message.error(error.message || t('course.disenrollFailed'));
    }
  };

  const handleDelete = async () => {
    if (!deleteModal.course) return;
    try {
      await agent.Courses.delete(deleteModal.course.id);
      message.success(t('course.deleted'));
      setDeleteModal({ visible: false, course: null });
      fetchData();
    } catch (error: any) {
      message.error(error.message || t('course.deleteFailed'));
    }
  };

  const now = dayjs();
  const currentCourses = courses.filter(c => dayjs(c.ends_at).isAfter(now));
  const pastCourses = courses.filter(c => dayjs(c.ends_at).isBefore(now));
  const isAdminAnywhere = role === 'root' || enrollments.some(e => e.role >= 2);

  const renderCourse = (course: Course) => {
    const enrolled = isEnrolled(course.id);
    const enrollment = enrollments.find(e => e.course_id === course.id);
    const isAdmin = enrollment && enrollment.role >= 2;
    const isTutor = enrollment && enrollment.role >= 1;

    return (
      <Col xs={24} sm={12} lg={8} key={course.id}>
        <CourseCard
          course={course}
          role={enrollment?.role}
          extra={
            (isAdmin || isTutor || role === 'root') && (
              <Dropdown
                menu={{
                  items: [
                    {
                      key: 'edit',
                      icon: <EditOutlined />,
                      label: t('common.edit'),
                      onClick: () => navigate(`/courses/${course.id}/edit`),
                    },
                    ...(isAdmin || role === 'root'
                      ? [{
                          key: 'delete',
                          icon: <DeleteOutlined />,
                          label: t('common.delete'),
                          danger: true,
                          onClick: () => setDeleteModal({ visible: true, course }),
                        }]
                      : []),
                  ],
                }}
              >
                <Button icon={<MoreOutlined />} type="text" />
              </Dropdown>
            )
          }
          actions={[
            enrolled ? (
              <Button
                type="link"
                icon={<LogoutOutlined />}
                onClick={() => setDisenrollModal({ visible: true, course })}
              >
                {t('course.disenroll')}
              </Button>
            ) : (
              <Button
                type="link"
                icon={<LoginOutlined />}
                loading={enrolling === course.id}
                onClick={() => handleEnroll(course)}
              >
                {t('course.enroll')}
              </Button>
            ),
          ]}
        />
      </Col>
    );
  };

  if (loading) return <Spin size="large" />;

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 24 }}>
        <h1 style={{ fontSize: '28px', margin: 0 }}>{t('course.courses')}</h1>
        {isAdminAnywhere && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/courses/new')} size="large">
            {t('course.newCourse')}
          </Button>
        )}
      </div>

      <Tabs
        size="large"
        items={[
          {
            key: 'current',
            label: `${t('course.currentCourses')} (${currentCourses.length})`,
            children: currentCourses.length === 0 ? (
              <Empty description={t('course.noCurrentCourses')} />
            ) : (
              <Row gutter={[16, 16]}>
                {currentCourses.map(renderCourse)}
              </Row>
            ),
          },
          {
            key: 'archive',
            label: `${t('course.archive')} (${pastCourses.length})`,
            children: pastCourses.length === 0 ? (
              <Empty description={t('course.noPastCourses')} />
            ) : (
              <Row gutter={[16, 16]}>
                {pastCourses.map(renderCourse)}
              </Row>
            ),
          },
        ]}
      />

      <Modal
        title={t('course.disenroll')}
        open={disenrollModal.visible}
        onOk={handleDisenroll}
        onCancel={() => setDisenrollModal({ visible: false, course: null })}
      >
        {t('messages.deleteConfirm')}
      </Modal>

      <Modal
        title={t('course.deleteCourse')}
        open={deleteModal.visible}
        onOk={handleDelete}
        onCancel={() => setDeleteModal({ visible: false, course: null })}
        okButtonProps={{ danger: true }}
      >
        {t('course.deleteConfirm', { name: deleteModal.course?.name })}
      </Modal>

      {(role === 'admin' || role === 'root') && (
        <FloatButton
          icon={<PlusOutlined />}
          type="primary"
          tooltip={t('course.createCourse')}
          onClick={() => navigate('/courses/new')}
        />
      )}
    </div>
  );
};

export default Courses;

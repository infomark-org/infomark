import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Typography, Spin, Empty, Statistic, FloatButton } from 'antd';
import { BookOutlined, TrophyOutlined, TeamOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import agent from '@/api/agent';
import { CourseCard } from '@/components/CourseCard';
import type { Course } from '@/types';
import { useTranslation } from 'react-i18next';

const { Title } = Typography;

interface EnrollmentWithCourse {
  id: number;
  course_id: number;
  role: number;
  course: Course;
}

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [enrollments, setEnrollments] = useState<EnrollmentWithCourse[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [enrollmentsData, coursesData] = await Promise.all([
        agent.Account.getEnrollments(),
        agent.Courses.getAll(),
      ]);

      const enriched = enrollmentsData.map(e => ({
        ...e,
        course: coursesData.find(c => c.id === e.course_id)!,
      })).filter(e => e.course);

      setEnrollments(enriched);
    } catch (error) {
      console.error('Failed to fetch enrollments:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Title level={2}>{t('dashboard.welcome')}, {user?.first_name}!</Title>
      
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic 
              title={t('course.courses')} 
              value={enrollments.length} 
              prefix={<BookOutlined />}
              valueStyle={{ color: '#A51E37' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic 
              title={t('admin.student')} 
              value={enrollments.filter(e => e.role === 0).length}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic 
              title={t('group.tutor') + '/Admin'} 
              value={enrollments.filter(e => e.role > 0).length}
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#B4A069' }}
            />
          </Card>
        </Col>
      </Row>

      <Card 
        title={t('dashboard.myCourses')} 
        extra={<Link to="/courses">{t('course.courses')}</Link>}
        style={{ marginTop: 24 }}
      >
        {loading ? (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin size="large" />
          </div>
        ) : enrollments.length === 0 ? (
          <Empty description={t('dashboard.noCourses')} />
        ) : (
          <Row gutter={[16, 16]}>
            {enrollments.map((enrollment) => (
              <Col xs={24} sm={12} lg={8} key={enrollment.course.id}>
                <CourseCard
                  course={enrollment.course}
                  role={enrollment.role}
                />
              </Col>
            ))}
          </Row>
        )}
      </Card>

      <FloatButton
        icon={<BookOutlined />}
        type="primary"
        tooltip={t('course.courses')}
        onClick={() => navigate('/courses')}
      />
    </div>
  );
};

export default Dashboard;

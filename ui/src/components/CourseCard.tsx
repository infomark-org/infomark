import React from 'react';
import { Card, Tag, Space, Button } from 'antd';
import { CalendarOutlined, TeamOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { Course } from '@/types';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';

interface CourseCardProps {
  course: Course;
  role?: number;
  actions?: React.ReactNode[];
  extra?: React.ReactNode;
}

export const CourseCard: React.FC<CourseCardProps> = ({ course, role, actions, extra }) => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  
  return (
    <Card
      title={course.name}
      extra={
        <>
          {role !== undefined && (
            <Tag color={role === 0 ? 'blue' : role === 1 ? 'green' : 'red'}>
              {role === 0 ? t('admin.student') : role === 1 ? t('group.tutor') : 'Admin'}
            </Tag>
          )}
          {extra}
        </>
      }
      actions={[
        <Button 
          type="primary" 
          icon={<EyeOutlined />}
          onClick={() => navigate(`/courses/${course.id}`)}
        >
          {t('common.view')}
        </Button>,
        ...(actions || [])
      ]}
    >
      <Card.Meta
        description={
          <div>
            <p style={{ minHeight: 60 }}>{course.description}</p>
            <Space direction="vertical" style={{ width: '100%', marginTop: 16 }}>
              <div>
                <CalendarOutlined /> {dayjs(course.begins_at).format('MMM D, YYYY')} - {dayjs(course.ends_at).format('MMM D, YYYY')}
              </div>
              <div>
                <TeamOutlined /> Required: {course.required_percentage}%
              </div>
            </Space>
          </div>
        }
      />
    </Card>
  );
};

import React from 'react';
import { Card, Row, Col, Typography, Space, Button } from 'antd';
import { UserOutlined, BookOutlined, SearchOutlined } from '@ant-design/icons';
import { Navigate, useNavigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from 'react-i18next';

const { Title, Paragraph } = Typography;

// Root-only administration hub. The heavy user directory lives on its own
// route (/admin/users); this page is the landing that routes there and to
// course administration.
const Admin: React.FC = () => {
  const { role } = useAuth();
  const navigate = useNavigate();
  const { t } = useTranslation();

  if (role && role !== 'root') {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div>
      <Title level={2}>{t('admin.admin')}</Title>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <Space>
                <UserOutlined style={{ fontSize: 22 }} />
                <Title level={4} style={{ margin: 0 }}>
                  {t('nav.userSearch')}
                </Title>
              </Space>
              <Paragraph type="secondary">{t('enrollment.searchPlaceholder')}</Paragraph>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => navigate('/admin/users')}>
                {t('nav.userSearch')}
              </Button>
            </Space>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <Space>
                <BookOutlined style={{ fontSize: 22 }} />
                <Title level={4} style={{ margin: 0 }}>
                  {t('nav.courses')}
                </Title>
              </Space>
              <Paragraph type="secondary">{t('course.createCourse')}</Paragraph>
              <Button icon={<BookOutlined />} onClick={() => navigate('/courses')}>
                {t('nav.courses')}
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Admin;

import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Alert } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import type { LoginRequest } from '@/types';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

const Login: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const onFinish = async (values: LoginRequest) => {
    setLoading(true);
    setError(null);

    try {
      await login(values);
      navigate('/dashboard');
    } catch (err: any) {
      setError(err.message || t('auth.loginFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '70vh' }}>
      <Card style={{ width: 400 }}>
        <Title level={2} style={{ textAlign: 'center' }}>
          {t('auth.login')}
        </Title>
        
        {error && (
          <Alert
            message={t('common.error')}
            description={error}
            type="error"
            closable
            onClose={() => setError(null)}
            style={{ marginBottom: 16 }}
          />
        )}

        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          layout="vertical"
        >
          <Form.Item
            name="email"
            rules={[
              { required: true, message: t('auth.email') },
              { type: 'email', message: t('auth.email') },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder={t('auth.email')}
              size="large"
            />
          </Form.Item>

          <Form.Item
            name="plain_password"
            rules={[{ required: true, message: t('auth.password') }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder={t('auth.password')}
              size="large"
            />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              {t('auth.login')}
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          <Text>
            <Link to="/request-password-reset">{t('auth.forgotPassword')}</Link>
          </Text>
          <br />
          <Text>
            {t('auth.noAccount')} <Link to="/register">{t('auth.register')}</Link>
          </Text>
        </div>
      </Card>
    </div>
  );
};

export default Login;

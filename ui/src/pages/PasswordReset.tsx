import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Alert } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import { Link, useParams, useNavigate } from 'react-router-dom';
import agent from '@/api/agent';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

const PasswordReset: React.FC = () => {
  // The backend reset email links to `/#/password_reset/{email}/{token}`
  // (auth.go + email/email.go), so both values arrive as path parameters.
  const { email, token } = useParams<{ email: string; token: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onFinish = async (values: { plain_password: string }) => {
    if (!token || !email) {
      setError(t('password.invalidLink'));
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // `UpdatePasswordRequest.Bind` (auth_requests.go) requires all three of
      // email, reset_password_token and plain_password.
      await agent.Auth.updatePassword({
        email: decodeURIComponent(email),
        reset_password_token: token,
        plain_password: values.plain_password,
      });
      navigate('/login', { state: { message: t('password.resetSuccessLogin') } });
    } catch (err: any) {
      setError(err.message || t('password.resetFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '70vh' }}>
      <Card style={{ width: 400 }}>
        <Title level={2} style={{ textAlign: 'center' }}>
          {t('password.setNewPassword')}
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

        <Form name="reset" onFinish={onFinish} layout="vertical">
          <Form.Item
            name="plain_password"
            rules={[
              { required: true, message: t('validation.newPasswordRequired') },
              { min: 7, message: t('validation.passwordMin', { min: 7 }) },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={t('password.newPassword')} size="large" />
          </Form.Item>

          <Form.Item
            name="confirm_password"
            dependencies={['plain_password']}
            rules={[
              { required: true, message: t('validation.confirmPasswordRequired') },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('plain_password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error(t('validation.passwordsNoMatch')));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={t('auth.confirmPassword')} size="large" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              {t('password.resetPassword')}
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center', marginTop: 16 }}>
          <Text>
            <Link to="/login">{t('auth.backToLogin')}</Link>
          </Text>
        </div>
      </Card>
    </div>
  );
};

export default PasswordReset;

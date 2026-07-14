import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Alert } from 'antd';
import { MailOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import agent from '@/api/agent';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

const RequestPasswordReset: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { t } = useTranslation();

  const onFinish = async (values: { email: string }) => {
    setLoading(true);
    setError(null);

    try {
      await agent.Auth.requestPasswordReset(values);
      setSuccess(true);
    } catch (err: any) {
      setError(err.message || t('password.requestFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '70vh' }}>
      <Card style={{ width: 400 }}>
        <Title level={2} style={{ textAlign: 'center' }}>
          {t('password.resetPassword')}
        </Title>

        {success ? (
          <Alert
            message={t('password.emailSentTitle')}
            description={t('password.checkEmail')}
            type="success"
            showIcon
          />
        ) : (
          <>
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
                name="email"
                rules={[
                  { required: true, message: t('validation.emailRequired') },
                  { type: 'email', message: t('validation.emailInvalid') },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder={t('auth.email')} size="large" />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading} block size="large">
                  {t('password.sendResetLink')}
                </Button>
              </Form.Item>
            </Form>
          </>
        )}

        <div style={{ textAlign: 'center', marginTop: 16 }}>
          <Text>
            <Link to="/login">{t('auth.backToLogin')}</Link>
          </Text>
        </div>
      </Card>
    </div>
  );
};

export default RequestPasswordReset;

import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Alert, InputNumber, Select } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import agent from '@/api/agent';
import type { RegisterRequest } from '@/types';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;
const { Option } = Select;

const Register: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { t } = useTranslation();

  const onFinish = async (values: RegisterRequest) => {
    setLoading(true);
    setError(null);

    try {
      await agent.Auth.register(values);
      navigate('/login', { state: { message: t('auth.registerSuccessLogin') } });
    } catch (err: any) {
      setError(err.message || t('auth.registerFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '70vh' }}>
      <Card style={{ width: 500 }}>
        <Title level={2} style={{ textAlign: 'center' }}>
          {t('auth.register')}
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
          name="register"
          onFinish={onFinish}
          autoComplete="off"
          layout="vertical"
        >
          <Form.Item
            label={t('profile.firstName')}
            name="first_name"
            rules={[{ required: true, message: t('validation.firstNameRequired') }]}
          >
            <Input prefix={<UserOutlined />} placeholder={t('profile.firstName')} />
          </Form.Item>

          <Form.Item
            label={t('profile.lastName')}
            name="last_name"
            rules={[{ required: true, message: t('validation.lastNameRequired') }]}
          >
            <Input prefix={<UserOutlined />} placeholder={t('profile.lastName')} />
          </Form.Item>

          <Form.Item
            label={t('auth.email')}
            name="email"
            rules={[
              { required: true, message: t('validation.emailRequired') },
              { type: 'email', message: t('validation.emailInvalid') },
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder={t('auth.email')} />
          </Form.Item>

          <Form.Item
            label={t('profile.studentNumber')}
            name="student_number"
            rules={[{ required: true, message: t('validation.studentNumberRequired') }]}
          >
            <Input placeholder={t('profile.studentNumber')} />
          </Form.Item>

          <Form.Item
            label={t('profile.semester')}
            name="semester"
            rules={[{ required: true, message: t('validation.semesterRequired') }]}
          >
            <InputNumber min={1} max={20} style={{ width: '100%' }} placeholder={t('profile.semester')} />
          </Form.Item>

          <Form.Item
            label={t('profile.subject')}
            name="subject"
            rules={[{ required: true, message: t('validation.subjectRequired') }]}
          >
            <Input placeholder={t('profile.subject')} />
          </Form.Item>

          <Form.Item
            label={t('profile.language')}
            name="language"
            rules={[{ required: true, message: t('validation.languageRequired') }]}
            initialValue="en"
          >
            <Select placeholder={t('validation.selectLanguage')}>
              <Option value="en">English</Option>
              <Option value="de">Deutsch</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label={t('auth.password')}
            name="plain_password"
            rules={[
              { required: true, message: t('validation.passwordRequired') },
              { min: 6, message: t('validation.passwordMin', { min: 6 }) },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={t('auth.password')} />
          </Form.Item>

          <Form.Item
            label={t('auth.confirmPassword')}
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
            <Input.Password prefix={<LockOutlined />} placeholder={t('auth.confirmPassword')} />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              {t('auth.register')}
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          <Text>
            {t('auth.hasAccount')} <Link to="/login">{t('auth.login')}</Link>
          </Text>
        </div>
      </Card>
    </div>
  );
};

export default Register;

import React, { useEffect, useState } from 'react';
import { Card, Form, Input, Button, Upload, Avatar, InputNumber, Select, message } from 'antd';
import { UploadOutlined, UserOutlined } from '@ant-design/icons';
import { useAuth } from '@/contexts/AuthContext';
import agent from '@/api/agent';
import { useTranslation } from 'react-i18next';
import i18n from '@/i18n';

const ProfileEditor: React.FC = () => {
  const { user, refreshUser } = useAuth();
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);

  useEffect(() => {
    if (user) {
      form.setFieldsValue({
        first_name: user.first_name,
        last_name: user.last_name,
        email: user.email,
        student_number: user.student_number,
        semester: user.semester,
        subject: user.subject,
        language: user.language,
      });
    }
  }, [user, form]);

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      const { old_plain_password, new_plain_password, email, language, ...userFields } = values;
      
      // Update user profile fields (first_name, last_name, etc.)
      await agent.Account.updateMe({ ...userFields, language });
      
      // Change language immediately if it was updated
      if (language && language !== i18n.language) {
        i18n.changeLanguage(language);
        localStorage.setItem('language', language);
      }
      
      // Update account fields (email/password) only if changed
      if (email !== user?.email || new_plain_password) {
        await agent.Account.update({
          email,
          old_plain_password,
          new_plain_password,
        });
      }
      
      // Upload avatar if changed
      if (avatarFile) {
        await agent.Account.uploadAvatar(avatarFile);
      }

      await refreshUser();
      setAvatarFile(null);
      setAvatarPreview(null);
      message.success(t('profile.updateSuccess'));
    } catch (error) {
      message.error(t('profile.updateFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title={t('profile.title')}>
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <Avatar
          size={100}
          src={avatarPreview || user?.avatar_url}
          icon={!avatarPreview && !user?.avatar_url && <UserOutlined />}
        />
        <div style={{ marginTop: 16 }}>
          <Upload
            beforeUpload={(file) => {
              setAvatarFile(file);
              const reader = new FileReader();
              reader.onload = (e) => setAvatarPreview(e.target?.result as string);
              reader.readAsDataURL(file);
              return false;
            }}
            showUploadList={false}
            maxCount={1}
          >
            <Button icon={<UploadOutlined />}>{t('profile.changeAvatar')}</Button>
          </Upload>
        </div>
      </div>

      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="first_name" label={t('profile.firstName')} rules={[{ required: true }]}>
          <Input />
        </Form.Item>

        <Form.Item name="last_name" label={t('profile.lastName')} rules={[{ required: true }]}>
          <Input />
        </Form.Item>

        <Form.Item name="email" label={t('profile.email')} rules={[{ required: true, type: 'email' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="student_number" label={t('profile.studentNumber')}>
          <Input />
        </Form.Item>

        <Form.Item name="semester" label={t('profile.semester')}>
          <InputNumber min={1} max={20} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="subject" label={t('profile.subject')}>
          <Input />
        </Form.Item>

        <Form.Item name="language" label={t('profile.language')}>
          <Select>
            <Select.Option value="en">English</Select.Option>
            <Select.Option value="de">Deutsch</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item name="old_plain_password" label={t('profile.currentPassword')}>
          <Input.Password />
        </Form.Item>

        <Form.Item name="new_plain_password" label={t('profile.newPassword')}>
          <Input.Password />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            {t('profile.updateProfile')}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default ProfileEditor;

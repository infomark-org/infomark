import React, { useState } from 'react';
import { Card, Form, Input, Button, message } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import agent from '@/api/agent';
import { useTranslation } from 'react-i18next';

const MailEditor: React.FC = () => {
  const { courseId } = useParams<{ courseId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      await agent.Courses.sendEmail(parseInt(courseId!), values);
      message.success(t('mail.sendSuccess'));
      navigate(`/courses/${courseId}`);
    } catch (error) {
      message.error(t('mail.sendFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title={t('mail.sendCourseEmail')}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="subject" label={t('mail.subject')} rules={[{ required: true }]}>
          <Input />
        </Form.Item>

        <Form.Item name="body" label={t('mail.message')} rules={[{ required: true }]}>
          <Input.TextArea rows={10} />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            {t('course.sendEmail')}
          </Button>
          <Button style={{ marginLeft: 8 }} onClick={() => navigate(`/courses/${courseId}`)}>
            {t('common.cancel')}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default MailEditor;

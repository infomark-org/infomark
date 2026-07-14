import React, { useEffect, useState } from 'react';
import { Form, Input, Button, Card, DatePicker, InputNumber, message } from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import agent from '@/api/agent';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';

const CourseEditor: React.FC = () => {
  const { courseId } = useParams<{ courseId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(!!courseId);

  useEffect(() => {
    if (courseId) {
      loadCourse();
    }
  }, [courseId]);

  const loadCourse = async () => {
    try {
      const course = await agent.Courses.get(parseInt(courseId!));
      form.setFieldsValue({
        ...course,
        begins_at: dayjs(course.begins_at),
        ends_at: dayjs(course.ends_at),
      });
    } catch (error) {
      message.error(t('course.loadFailed'));
    } finally {
      setInitialLoading(false);
    }
  };

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        begins_at: values.begins_at.toISOString(),
        ends_at: values.ends_at.toISOString(),
      };

      if (courseId) {
        await agent.Courses.update(parseInt(courseId), data);
        message.success(t('course.updated'));
      } else {
        await agent.Courses.create(data);
        message.success(t('course.created'));
      }
      navigate('/courses');
    } catch (error) {
      message.error(courseId ? t('course.updateFailed') : t('course.createFailed'));
    } finally {
      setLoading(false);
    }
  };

  if (initialLoading) return <Card loading />;

  return (
    <Card title={courseId ? t('course.editCourse') : t('course.createCourse')}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="name" label={t('course.courseName')} rules={[{ required: true }]}>
          <Input />
        </Form.Item>

        <Form.Item name="description" label={t('course.description')} rules={[{ required: true }]}>
          <Input.TextArea rows={4} />
        </Form.Item>

        <Form.Item name="begins_at" label={t('course.startDate')} rules={[{ required: true }]}>
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="ends_at" label={t('course.endDate')} rules={[{ required: true }]}>
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="required_percentage" label={t('course.requiredPercentage')} rules={[{ required: true }]} initialValue={50}>
          <InputNumber min={0} max={100} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            {courseId ? t('course.updateCourse') : t('course.createCourse')}
          </Button>
          <Button style={{ marginLeft: 8 }} onClick={() => navigate('/courses')}>
            {t('common.cancel')}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default CourseEditor;

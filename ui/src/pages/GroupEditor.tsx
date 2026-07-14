import React, { useEffect, useState } from 'react';
import { Form, Input, Button, Card, Select, message } from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import agent from '@/api/agent';
import type { User } from '@/types';
import { useTranslation } from 'react-i18next';

const GroupEditor: React.FC = () => {
  const { courseId, groupId } = useParams<{ courseId: string; groupId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [tutors, setTutors] = useState<User[]>([]);

  useEffect(() => {
    fetchTutors();
    if (groupId) {
      loadGroup();
    }
  }, [groupId]);

  const fetchTutors = async () => {
    try {
      const enrollments = await agent.Courses.getEnrollments(parseInt(courseId!));
      const tutorUsers = enrollments.filter(e => e.role >= 1).map(e => e.user);
      setTutors(tutorUsers);
    } catch (error) {
      message.error(t('group.loadTutorsFailed'));
    }
  };

  const loadGroup = async () => {
    try {
      const group = await agent.Groups.get(parseInt(courseId!), parseInt(groupId!));
      form.setFieldsValue({
        description: group.description,
        tutor_id: group.tutor.id,
      });
    } catch (error) {
      message.error(t('group.loadFailed'));
    }
  };

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      const cId = parseInt(courseId!);

      if (groupId) {
        await agent.Groups.update(cId, parseInt(groupId), values);
        message.success(t('group.updated'));
      } else {
        await agent.Groups.create(cId, values);
        message.success(t('group.created'));
      }

      navigate(`/courses/${courseId}`);
    } catch (error) {
      message.error(groupId ? t('group.updateFailed') : t('group.createFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title={groupId ? t('group.editGroup') : t('group.createGroup')}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="description" label={t('group.description')} rules={[{ required: true }]}>
          <Input.TextArea rows={4} />
        </Form.Item>

        <Form.Item name="tutor_id" label={t('group.tutor')} rules={[{ required: true }]}>
          <Select placeholder={t('group.selectTutor')}>
            {tutors.map(tutor => (
              <Select.Option key={tutor.id} value={tutor.id}>
                {tutor.first_name} {tutor.last_name} ({tutor.email})
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            {groupId ? t('group.updateGroup') : t('group.createGroup')}
          </Button>
          <Button style={{ marginLeft: 8 }} onClick={() => navigate(`/courses/${courseId}`)}>
            {t('common.cancel')}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default GroupEditor;

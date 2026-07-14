import React, { useCallback, useEffect, useState } from 'react';
import { Card, Table, Input, InputNumber, Tag, Button, Modal, Form, message, Space, Skeleton, Breadcrumb } from 'antd';
import { SearchOutlined, EditOutlined, HomeOutlined } from '@ant-design/icons';
import { Link, Navigate } from 'react-router-dom';
import agent from '@/api/agent';
import { useAuth } from '@/contexts/AuthContext';
import type { User, UserUpdate } from '@/types';
import { useTranslation } from 'react-i18next';

// Root-only directory search over every account (GET /users/find). Root may
// open a user and edit their profile fields via PUT /users/{id}. This is the
// system-wide counterpart to per-course enrollment role management.
const UserSearch: React.FC = () => {
  const { role } = useAuth();
  const { t } = useTranslation();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [editing, setEditing] = useState<User | null>(null);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  // Debounced fuzzy search: the backend does the matching, so we simply pass
  // the raw query (empty query returns the full list).
  const search = useCallback(async (value: string) => {
    setLoading(true);
    try {
      const data = await agent.Users.find(value.trim());
      setUsers(data);
    } catch {
      message.error(t('messages.updateFailed'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const handle = setTimeout(() => search(query), 300);
    return () => clearTimeout(handle);
  }, [query, search]);

  // Guard client-side too; the route is otherwise reachable by URL. The backend
  // still enforces authorization, so this is only a UX affordance.
  if (role && role !== 'root') {
    return <Navigate to="/dashboard" replace />;
  }

  const openEditor = (user: User) => {
    setEditing(user);
    form.setFieldsValue({
      first_name: user.first_name,
      last_name: user.last_name,
      email: user.email,
      student_number: user.student_number ?? '',
      semester: user.semester ?? 1,
      subject: user.subject ?? '',
      language: user.language ?? 'en',
    });
  };

  const handleSave = async () => {
    if (!editing) return;
    try {
      const values = (await form.validateFields()) as UserUpdate;
      setSaving(true);
      await agent.Users.update(editing.id, values);
      message.success(t('messages.updateSuccess'));
      setEditing(null);
      await search(query);
    } catch (error) {
      if ((error as { errorFields?: unknown }).errorFields) return;
      message.error(t('messages.updateFailed'));
    } finally {
      setSaving(false);
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
    {
      title: t('profile.firstName'),
      key: 'name',
      render: (_: unknown, record: User) => `${record.first_name} ${record.last_name}`,
    },
    { title: t('profile.email'), dataIndex: 'email', key: 'email' },
    {
      title: t('profile.studentNumber'),
      dataIndex: 'student_number',
      key: 'student_number',
      responsive: ['md' as const],
    },
    {
      title: t('admin.role'),
      key: 'role',
      render: (_: unknown, record: User) => (
        <Tag color={record.root ? 'red' : 'blue'}>{record.root ? t('roles.root') : t('roles.user')}</Tag>
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      render: (_: unknown, record: User) => (
        <Button icon={<EditOutlined />} onClick={() => openEditor(record)}>
          {t('common.edit')}
        </Button>
      ),
    },
  ];

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <Link to="/"><HomeOutlined /></Link> },
          { title: <Link to="/admin">{t('nav.admin')}</Link> },
          { title: t('nav.userSearch') },
        ]}
        style={{ marginBottom: 16 }}
      />
      <Card
        title={t('nav.userSearch')}
        extra={
          <Input
            allowClear
            prefix={<SearchOutlined />}
            placeholder={t('enrollment.searchPlaceholder')}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            style={{ maxWidth: 280 }}
            aria-label={t('common.search')}
          />
        }
      >
        {loading ? (
          <Skeleton active paragraph={{ rows: 6 }} />
        ) : (
          <Table
            dataSource={users}
            columns={columns}
            rowKey="id"
            pagination={{ pageSize: 15, hideOnSinglePage: true }}
            scroll={{ x: 'max-content' }}
          />
        )}
      </Card>

      <Modal
        title={`${t('common.edit')} — ${editing?.first_name ?? ''} ${editing?.last_name ?? ''}`}
        open={editing !== null}
        onOk={handleSave}
        confirmLoading={saving}
        onCancel={() => setEditing(null)}
        okText={t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Form form={form} layout="vertical">
          <Space style={{ display: 'flex' }} align="start">
            <Form.Item name="first_name" label={t('profile.firstName')} rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input />
            </Form.Item>
            <Form.Item name="last_name" label={t('profile.lastName')} rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input />
            </Form.Item>
          </Space>
          <Form.Item
            name="email"
            label={t('profile.email')}
            rules={[{ required: true, type: 'email' }]}
          >
            <Input />
          </Form.Item>
          <Space style={{ display: 'flex' }} align="start">
            <Form.Item name="student_number" label={t('profile.studentNumber')} rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input />
            </Form.Item>
            <Form.Item name="semester" label={t('profile.semester')} rules={[{ required: true }]} style={{ flex: 1 }}>
              {/* InputNumber yields a JSON number; a plain Input type="number" */}
              {/* would submit a string and the backend UserRequest.Semester */}
              {/* (int) would reject it with a 400 unmarshal error. */}
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
          </Space>
          <Form.Item name="subject" label={t('profile.subject')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="language" label={t('profile.language')} rules={[{ required: true }]}>
            <Input maxLength={2} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default UserSearch;

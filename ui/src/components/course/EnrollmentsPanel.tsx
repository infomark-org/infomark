import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Table, Input, Select, Space, Tag, Dropdown, Button, Skeleton, message, Avatar } from 'antd';
import { SearchOutlined, UserOutlined, DownOutlined } from '@ant-design/icons';
import agent from '@/api/agent';
import type { Enrollment } from '@/types';
import { useTranslation } from 'react-i18next';

interface EnrollmentsPanelProps {
  courseId: number;
  // Only admins (and root) may change a user's course role; tutors get a
  // read-only roster.
  canManage: boolean;
}

const ROLE_STUDENT = 0;
const ROLE_TUTOR = 1;
const ROLE_ADMIN = 2;

// EnrollmentsPanel is the course roster for tutors/admins: searchable, role
// filterable, and (for admins) able to promote/demote a member via
// PUT /courses/{id}/enrollments/{user_id}.
export const EnrollmentsPanel: React.FC<EnrollmentsPanelProps> = ({ courseId, canManage }) => {
  const { t } = useTranslation();
  const [enrollments, setEnrollments] = useState<Enrollment[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [roleFilter, setRoleFilter] = useState<number | 'all'>('all');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await agent.Courses.getEnrollments(courseId);
      setEnrollments(data);
    } catch {
      message.error(t('messages.updateFailed'));
    } finally {
      setLoading(false);
    }
  }, [courseId, t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const roleLabel = (role: number) =>
    role >= ROLE_ADMIN ? t('roles.admin') : role >= ROLE_TUTOR ? t('roles.tutor') : t('roles.student');
  const roleColor = (role: number) => (role >= ROLE_ADMIN ? 'red' : role >= ROLE_TUTOR ? 'green' : 'blue');

  const changeRole = async (userId: number, role: number) => {
    try {
      await agent.Courses.changeRole(courseId, userId, role);
      message.success(t('enrollment.roleChanged'));
      await fetchData();
    } catch {
      message.error(t('messages.updateFailed'));
    }
  };

  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return enrollments.filter((e) => {
      if (roleFilter !== 'all' && e.role !== roleFilter) return false;
      if (!normalized) return true;
      const haystack = `${e.user.first_name} ${e.user.last_name} ${e.user.email}`.toLowerCase();
      return haystack.includes(normalized);
    });
  }, [enrollments, query, roleFilter]);

  const columns = [
    {
      title: t('profile.firstName'),
      key: 'name',
      render: (_: unknown, record: Enrollment) => (
        <Space>
          <Avatar size="small" icon={<UserOutlined />} src={record.user.avatar_url} />
          {record.user.first_name} {record.user.last_name}
        </Space>
      ),
    },
    { title: t('profile.email'), dataIndex: ['user', 'email'], key: 'email', responsive: ['sm' as const] },
    {
      title: t('admin.role'),
      key: 'role',
      render: (_: unknown, record: Enrollment) => <Tag color={roleColor(record.role)}>{roleLabel(record.role)}</Tag>,
    },
    ...(canManage
      ? [
          {
            title: t('common.actions'),
            key: 'actions',
            render: (_: unknown, record: Enrollment) => (
              <Dropdown
                menu={{
                  items: [
                    { key: 'student', label: t('roles.student'), onClick: () => changeRole(record.user.id, ROLE_STUDENT) },
                    { key: 'tutor', label: t('roles.tutor'), onClick: () => changeRole(record.user.id, ROLE_TUTOR) },
                    { key: 'admin', label: t('roles.admin'), onClick: () => changeRole(record.user.id, ROLE_ADMIN) },
                  ],
                }}
              >
                <Button size="small">
                  {t('enrollment.changeRole')} <DownOutlined />
                </Button>
              </Dropdown>
            ),
          },
        ]
      : []),
  ];

  if (loading) return <Skeleton active paragraph={{ rows: 6 }} />;

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Space wrap>
        <Input
          allowClear
          prefix={<SearchOutlined />}
          placeholder={t('enrollment.searchPlaceholder')}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          style={{ width: 260 }}
          aria-label={t('common.search')}
        />
        <Select
          value={roleFilter}
          onChange={setRoleFilter}
          style={{ width: 160 }}
          aria-label={t('enrollment.filterRole')}
          options={[
            { value: 'all', label: t('enrollment.all') },
            { value: ROLE_STUDENT, label: t('roles.student') },
            { value: ROLE_TUTOR, label: t('roles.tutor') },
            { value: ROLE_ADMIN, label: t('roles.admin') },
          ]}
        />
      </Space>
      <Table
        dataSource={filtered}
        columns={columns}
        rowKey={(record) => record.user.id}
        pagination={{ pageSize: 15, hideOnSinglePage: true }}
        scroll={{ x: 'max-content' }}
      />
    </Space>
  );
};

export default EnrollmentsPanel;

import React, { useCallback, useEffect, useState } from 'react';
import { Card, Table, Button, Tag, Breadcrumb, Empty, Skeleton, message, Space, Input } from 'antd';
import { HomeOutlined, EditOutlined, CheckCircleOutlined, CloseCircleOutlined, SearchOutlined } from '@ant-design/icons';
import { Link, useNavigate, useParams } from 'react-router-dom';
import agent from '@/api/agent';
import type { MissingGrade } from '@/types';
import { useTranslation } from 'react-i18next';

// GradingQueue is the tutor's worklist: every submission in the course that
// still lacks feedback (GET /courses/{id}/grades/missing). Each row links to
// the single-grade editor keyed by grade_id.
const TEST_STATUS_PASSED = 1;

const GradingQueue: React.FC = () => {
  const { courseId } = useParams<{ courseId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [missing, setMissing] = useState<MissingGrade[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');

  const cId = Number(courseId);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await agent.Courses.getMissingGrades(cId);
      setMissing(data);
    } catch {
      message.error(t('messages.updateFailed'));
    } finally {
      setLoading(false);
    }
  }, [cId, t]);

  useEffect(() => {
    if (courseId) fetchData();
  }, [courseId, fetchData]);

  const normalizedQuery = query.trim().toLowerCase();
  const filtered = normalizedQuery
    ? missing.filter((m) => {
        const user = m.grade.user;
        const name = user ? `${user.first_name} ${user.last_name} ${user.email}`.toLowerCase() : '';
        return name.includes(normalizedQuery);
      })
    : missing;

  const columns = [
    {
      title: t('grades.student'),
      key: 'student',
      render: (_: unknown, record: MissingGrade) =>
        record.grade.user ? (
          <Space direction="vertical" size={0}>
            <span>
              {record.grade.user.first_name} {record.grade.user.last_name}
            </span>
            <span style={{ opacity: 0.65, fontSize: 12 }}>{record.grade.user.email}</span>
          </Space>
        ) : (
          `#${record.grade.submission_id}`
        ),
    },
    {
      title: t('grades.sheet'),
      dataIndex: 'sheet_id',
      key: 'sheet_id',
      render: (sheetId: number) => `${t('grades.sheet')} ${sheetId}`,
      responsive: ['sm' as const],
    },
    {
      title: t('grades.task'),
      dataIndex: 'task_id',
      key: 'task_id',
      render: (taskId: number) => `${t('grades.task')} ${taskId}`,
      responsive: ['sm' as const],
    },
    {
      title: t('submission.publicTests'),
      key: 'tests',
      render: (_: unknown, record: MissingGrade) =>
        record.grade.public_test_status === TEST_STATUS_PASSED ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            {t('submission.testsPassed')}
          </Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">
            {t('submission.testsFailed')}
          </Tag>
        ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      render: (_: unknown, record: MissingGrade) => (
        <Button
          type="primary"
          icon={<EditOutlined />}
          onClick={() => navigate(`/courses/${cId}/grades/${record.grade.id}`)}
        >
          {t('grades.openGrade')}
        </Button>
      ),
    },
  ];

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <Link to="/"><HomeOutlined /></Link> },
          { title: <Link to="/courses">{t('nav.courses')}</Link> },
          { title: <Link to={`/courses/${courseId}`}>{t('course.course')}</Link> },
          { title: t('grades.gradingQueue') },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card
        title={`${t('grades.missingGrades')} (${missing.length})`}
        extra={
          <Input
            allowClear
            prefix={<SearchOutlined />}
            placeholder={t('enrollment.searchPlaceholder')}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            style={{ maxWidth: 260 }}
            aria-label={t('common.search')}
          />
        }
      >
        {loading ? (
          <Skeleton active paragraph={{ rows: 6 }} />
        ) : missing.length === 0 ? (
          <Empty description={t('grades.noMissing')} />
        ) : (
          <Table
            dataSource={filtered}
            columns={columns}
            rowKey={(record) => record.grade.id}
            pagination={{ pageSize: 15, hideOnSinglePage: true }}
            scroll={{ x: 'max-content' }}
          />
        )}
      </Card>
    </div>
  );
};

export default GradingQueue;

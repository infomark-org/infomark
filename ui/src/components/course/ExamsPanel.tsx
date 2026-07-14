import React, { useCallback, useEffect, useState } from 'react';
import { List, Button, Tag, Empty, Skeleton, message, Space, Typography } from 'antd';
import { CalendarOutlined, CheckCircleOutlined } from '@ant-design/icons';
import agent from '@/api/agent';
import type { Exam, ExamEnrollment } from '@/types';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';

interface ExamsPanelProps {
  courseId: number;
  // Whether the viewer is a plain student in this course; only students may
  // self enroll/disenroll (exam.go enforces this server-side too).
  isStudent: boolean;
}

// A non-zero exam status means the exam has been marked; the backend then
// refuses disenrollment (exam.go DisenrollExamHandler).
const STATUS_UNMARKED = 0;

// ExamsPanel lists a course's exams and, for students, exposes enroll /
// disenroll. Enrollment state comes from the account-wide exam enrollment list
// filtered to this course.
export const ExamsPanel: React.FC<ExamsPanelProps> = ({ courseId, isStudent }) => {
  const { t } = useTranslation();
  const [exams, setExams] = useState<Exam[]>([]);
  const [enrollments, setEnrollments] = useState<ExamEnrollment[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyExamId, setBusyExamId] = useState<number | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const examsData = await agent.Exams.getAll(courseId);
      setExams(examsData);
      try {
        const enrollmentData = await agent.Account.getExamEnrollments();
        setEnrollments(enrollmentData.filter((e) => e.course_id === courseId));
      } catch {
        setEnrollments([]);
      }
    } catch {
      message.error(t('messages.updateFailed'));
    } finally {
      setLoading(false);
    }
  }, [courseId, t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const enrollmentFor = (examId: number) => enrollments.find((e) => e.exam_id === examId);

  const handleEnroll = async (examId: number) => {
    setBusyExamId(examId);
    try {
      await agent.Exams.enroll(courseId, examId);
      message.success(t('exam.enrollSuccess'));
      await fetchData();
    } catch (error) {
      message.error((error as { message?: string }).message || t('messages.updateFailed'));
    } finally {
      setBusyExamId(null);
    }
  };

  const handleDisenroll = async (examId: number) => {
    setBusyExamId(examId);
    try {
      await agent.Exams.disenroll(courseId, examId);
      message.success(t('exam.disenrollSuccess'));
      await fetchData();
    } catch (error) {
      message.error((error as { message?: string }).message || t('exam.markedCannotLeave'));
    } finally {
      setBusyExamId(null);
    }
  };

  if (loading) return <Skeleton active paragraph={{ rows: 4 }} />;
  if (exams.length === 0) return <Empty description={t('exam.noExams')} />;

  return (
    <List
      dataSource={exams}
      renderItem={(exam) => {
        const enrollment = enrollmentFor(exam.id);
        const enrolled = Boolean(enrollment);
        const marked = enrollment ? enrollment.status !== STATUS_UNMARKED : false;

        const actions: React.ReactNode[] = [];
        if (isStudent) {
          if (enrolled) {
            actions.push(
              <Button
                key="disenroll"
                danger
                loading={busyExamId === exam.id}
                disabled={marked}
                title={marked ? t('exam.markedCannotLeave') : undefined}
                onClick={() => handleDisenroll(exam.id)}
              >
                {t('exam.disenroll')}
              </Button>,
            );
          } else {
            actions.push(
              <Button key="enroll" type="primary" loading={busyExamId === exam.id} onClick={() => handleEnroll(exam.id)}>
                {t('exam.enroll')}
              </Button>,
            );
          }
        }

        return (
          <List.Item actions={actions}>
            <List.Item.Meta
              title={
                <Space wrap>
                  {exam.name}
                  {enrolled ? (
                    <Tag icon={<CheckCircleOutlined />} color="success">
                      {t('exam.enrolled')}
                    </Tag>
                  ) : (
                    isStudent && <Tag>{t('exam.notEnrolled')}</Tag>
                  )}
                  {marked && enrollment?.mark && <Tag color="blue">{`${t('exam.mark')}: ${enrollment.mark}`}</Tag>}
                </Space>
              }
              description={
                <Space direction="vertical" size={2} style={{ width: '100%' }}>
                  <Typography.Text type="secondary">
                    <CalendarOutlined /> {t('exam.examTime')}: {dayjs(exam.exam_time).format('MMM D, YYYY HH:mm')}
                  </Typography.Text>
                  <Typography.Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 0 }}>
                    {exam.description}
                  </Typography.Paragraph>
                </Space>
              }
            />
          </List.Item>
        );
      }}
    />
  );
};

export default ExamsPanel;

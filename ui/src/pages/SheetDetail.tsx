import React, { useCallback, useEffect, useState } from 'react';
import {
  Card,
  List,
  Button,
  Upload,
  message,
  Descriptions,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  InputNumber,
  Breadcrumb,
  Dropdown,
  Drawer,
  Divider,
  FloatButton,
  Progress,
  Empty,
  Tooltip,
} from 'antd';
import {
  UploadOutlined,
  DownloadOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  HomeOutlined,
  MoreOutlined,
  EyeOutlined,
  LoadingOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { useParams, Link } from 'react-router-dom';
import type { UploadProps } from 'antd';
import agent from '@/api/agent';
import type { Sheet, Task, Grade, TaskPoints } from '@/types';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';

// The backend caps submission uploads at 4 MB and only accepts zip archives
// (see api/app/submission.go + server config Limits.MaxRequestFile). We mirror
// both rules client-side so the user gets instant feedback instead of a 500.
const MAX_SUBMISSION_BYTES = 4 * 1024 * 1024;
const ACCEPTED_ZIP_TYPES = ['application/zip', 'application/x-zip-compressed', 'application/octet-stream'];

// Test execution/status enums as emitted by the grading worker
// (symbol.TestingResult). We only need the handful of states the UI reacts to.
const TEST_STATE_FINISHED = 1;
const TEST_STATUS_PASSED = 1;

const SheetDetail: React.FC = () => {
  const { courseId, sheetId } = useParams<{ courseId: string; sheetId: string }>();
  const { t } = useTranslation();
  const [sheet, setSheet] = useState<Sheet | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  // Per-task points for the logged-in student (acquired vs. max), keyed by task id.
  const [points, setPoints] = useState<Map<number, TaskPoints>>(new Map());
  // Per-task result grade (source of truth for submission existence + feedback).
  const [grades, setGrades] = useState<Map<number, Grade>>(new Map());
  const [loading, setLoading] = useState(true);
  const [uploadingTaskId, setUploadingTaskId] = useState<number | null>(null);
  const [taskModal, setTaskModal] = useState<{ visible: boolean; task: Task | null }>({ visible: false, task: null });
  const [feedbackDrawer, setFeedbackDrawer] = useState<{ open: boolean; grade: Grade | null; task: Task | null }>({
    open: false,
    grade: null,
    task: null,
  });
  const [form] = Form.useForm();
  const [isTutor, setIsTutor] = useState(false);

  const cId = Number(courseId);
  const sId = Number(sheetId);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      // Determine the caller's role in this course to decide between the
      // student submission UI and the tutor/admin authoring UI.
      let tutorish = false;
      try {
        const enrollments = await agent.Account.getEnrollments();
        const enrollment = enrollments.find((e) => e.course_id === cId);
        tutorish = enrollment ? enrollment.role >= 1 : false;
      } catch {
        tutorish = false;
      }
      setIsTutor(tutorish);

      const [sheetData, tasksData] = await Promise.all([agent.Sheets.get(cId, sId), agent.Tasks.getAll(cId, sId)]);
      setSheet(sheetData);
      setTasks(tasksData);

      // Per-sheet task points for the student (one round-trip, not N).
      const pointsMap = new Map<number, TaskPoints>();
      try {
        const sheetPoints = await agent.Sheets.getPoints(cId, sId);
        sheetPoints.forEach((p) => pointsMap.set(p.task_id, p));
      } catch {
        // Tutors/admins may not have per-student points; ignore.
      }
      setPoints(pointsMap);

      // Result grades in parallel (was an O(2N) serial loop before). A 404
      // simply means "no submission yet" for that task, so failures are
      // swallowed per task rather than failing the whole view.
      const gradeEntries = await Promise.all(
        tasksData.map(async (task) => {
          try {
            const grade = await agent.Tasks.getResult(cId, task.id);
            return [task.id, grade] as const;
          } catch {
            return null;
          }
        }),
      );
      const gradesMap = new Map<number, Grade>();
      gradeEntries.forEach((entry) => {
        if (entry) gradesMap.set(entry[0], entry[1]);
      });
      setGrades(gradesMap);
    } catch {
      message.error(t('messages.updateFailed'));
    } finally {
      setLoading(false);
    }
  }, [cId, sId, t]);

  useEffect(() => {
    if (courseId && sheetId) fetchData();
  }, [courseId, sheetId, fetchData]);

  // Validate a picked file against the backend's zip + size rules before the
  // network call. Returning `false` from AntD's beforeUpload stops the default
  // uploader; we drive the request ourselves so we can show upload state.
  const handleBeforeUpload = (taskId: number): UploadProps['beforeUpload'] => (file) => {
    const isZip = ACCEPTED_ZIP_TYPES.includes(file.type) || file.name.toLowerCase().endsWith('.zip');
    if (!isZip) {
      message.error(t('submission.onlyZip'));
      return Upload.LIST_IGNORE;
    }
    if (file.size > MAX_SUBMISSION_BYTES) {
      message.error(t('submission.tooLarge'));
      return Upload.LIST_IGNORE;
    }
    void handleSubmit(taskId, file as File);
    return false;
  };

  const handleSubmit = async (taskId: number, file: File) => {
    setUploadingTaskId(taskId);
    try {
      await agent.Tasks.submitTask(cId, taskId, file);
      message.success(t('messages.uploadSuccess'));
      await fetchData();
    } catch {
      message.error(t('messages.uploadFailed'));
    } finally {
      setUploadingTaskId(null);
    }
  };

  const handleDownloadOwn = async (taskId: number) => {
    try {
      await agent.Tasks.downloadOwnSubmission(cId, taskId, `task-${taskId}-submission.zip`);
    } catch {
      message.error(t('messages.downloadFailed'));
    }
  };

  const handleTaskModalOpen = (task: Task | null = null) => {
    if (task) form.setFieldsValue(task);
    else form.resetFields();
    setTaskModal({ visible: true, task });
  };

  const handleTaskSave = async () => {
    try {
      const values = await form.validateFields();
      if (taskModal.task) {
        await agent.Tasks.update(cId, taskModal.task.id, values);
        message.success(t('messages.updateSuccess'));
      } else {
        await agent.Tasks.create(cId, sId, values);
        message.success(t('messages.createSuccess'));
      }
      setTaskModal({ visible: false, task: null });
      await fetchData();
    } catch (error) {
      // validateFields rejects with the form errors; only surface real failures.
      if ((error as { errorFields?: unknown }).errorFields) return;
      message.error(t('messages.saveFailed'));
    }
  };

  const handleTaskDelete = async (taskId: number) => {
    try {
      await agent.Tasks.delete(cId, taskId);
      message.success(t('messages.deleteSuccess'));
      await fetchData();
    } catch {
      message.error(t('messages.deleteFailed'));
    }
  };

  if (loading || !sheet) return <Card loading />;

  const isPublished = dayjs().isAfter(dayjs(sheet.publish_at));
  const isPastDeadline = dayjs().isAfter(dayjs(sheet.due_at));

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <Link to="/"><HomeOutlined /></Link> },
          { title: <Link to="/courses">{t('nav.courses')}</Link> },
          { title: <Link to={`/courses/${courseId}`}>{t('course.course')}</Link> },
          { title: sheet.name },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card title={sheet.name}>
        <Descriptions column={{ xs: 1, sm: 2 }}>
          <Descriptions.Item label={t('sheet.published')}>
            {dayjs(sheet.publish_at).format('MMM D, YYYY HH:mm')}
          </Descriptions.Item>
          <Descriptions.Item label={t('sheet.due')}>
            <Space>
              {dayjs(sheet.due_at).format('MMM D, YYYY HH:mm')}
              {isPastDeadline ? (
                <Tag icon={<ClockCircleOutlined />} color="red">
                  {t('submission.deadlinePassed')}
                </Tag>
              ) : (
                <Tag color="green">{t('sheet.due')}</Tag>
              )}
            </Space>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title={t('sheet.tasks')} style={{ marginTop: 16 }}>
        {tasks.length === 0 ? (
          <Empty description={t('sheet.tasks')} />
        ) : (
          <List
            dataSource={tasks}
            renderItem={(task) => {
              const grade = grades.get(task.id);
              const taskPoints = points.get(task.id);
              const hasSubmission = Boolean(grade && grade.submission_id);
              const acquired = taskPoints?.acquired_points ?? grade?.acquired_points ?? 0;
              const percent = task.max_points > 0 ? Math.round((acquired / task.max_points) * 100) : 0;

              const testsFinished = grade?.public_execution_state === TEST_STATE_FINISHED;
              const testsPassed = grade?.public_test_status === TEST_STATUS_PASSED;

              const studentActions = [
                <Upload
                  key="upload"
                  beforeUpload={handleBeforeUpload(task.id)}
                  showUploadList={false}
                  accept=".zip"
                  disabled={!isPublished || isPastDeadline || uploadingTaskId === task.id}
                >
                  <Tooltip title={!isPublished ? t('submission.notPublished') : isPastDeadline ? t('submission.deadlinePassed') : t('submission.uploadHint')}>
                    <Button
                      type="primary"
                      icon={uploadingTaskId === task.id ? <LoadingOutlined /> : <UploadOutlined />}
                      loading={uploadingTaskId === task.id}
                      disabled={!isPublished || isPastDeadline}
                    >
                      {hasSubmission ? t('task.resubmit') : t('task.submit')}
                    </Button>
                  </Tooltip>
                </Upload>,
                hasSubmission ? (
                  <Button key="download" icon={<DownloadOutlined />} onClick={() => handleDownloadOwn(task.id)}>
                    {t('common.download')}
                  </Button>
                ) : null,
                grade ? (
                  <Button key="feedback" icon={<EyeOutlined />} onClick={() => setFeedbackDrawer({ open: true, grade, task })}>
                    {t('task.feedback')}
                  </Button>
                ) : null,
              ].filter(Boolean) as React.ReactNode[];

              const tutorActions = [
                <Dropdown
                  key="menu"
                  menu={{
                    items: [
                      { key: 'edit', icon: <EditOutlined />, label: t('common.edit'), onClick: () => handleTaskModalOpen(task) },
                      { key: 'delete', icon: <DeleteOutlined />, label: t('common.delete'), danger: true, onClick: () => handleTaskDelete(task.id) },
                    ],
                  }}
                >
                  <Button icon={<MoreOutlined />} aria-label={t('common.actions')} />
                </Dropdown>,
              ];

              return (
                <List.Item actions={isTutor ? tutorActions : studentActions}>
                  <List.Item.Meta
                    title={task.name}
                    description={
                      <Space direction="vertical" style={{ width: '100%' }} size={4}>
                        <Space wrap>
                          <span>
                            {t('sheet.maxPoints')}: {task.max_points}
                          </span>
                          {hasSubmission && (
                            <Tag color="blue">{t('task.submitted')}</Tag>
                          )}
                          {grade && (
                            <Tag
                              icon={
                                !testsFinished ? <LoadingOutlined /> : testsPassed ? <CheckCircleOutlined /> : <CloseCircleOutlined />
                              }
                              color={!testsFinished ? 'processing' : testsPassed ? 'success' : 'error'}
                            >
                              {!testsFinished
                                ? t('submission.testsRunning')
                                : testsPassed
                                  ? t('submission.testsPassed')
                                  : t('submission.testsFailed')}
                            </Tag>
                          )}
                        </Space>
                        {!isTutor && hasSubmission && (
                          <Progress
                            percent={percent}
                            size="small"
                            style={{ maxWidth: 240 }}
                            format={() => `${acquired}/${task.max_points}`}
                          />
                        )}
                        {!isTutor && !hasSubmission && <span style={{ opacity: 0.65 }}>{t('submission.noSubmission')}</span>}
                      </Space>
                    }
                  />
                </List.Item>
              );
            }}
          />
        )}
      </Card>

      <Modal
        title={taskModal.task ? t('common.edit') : t('sheet.addTask')}
        open={taskModal.visible}
        onOk={handleTaskSave}
        onCancel={() => setTaskModal({ visible: false, task: null })}
        okText={t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label={t('course.name')} rules={[{ required: true, message: t('course.name') }]}>
            <Input />
          </Form.Item>
          <Form.Item name="max_points" label={t('sheet.maxPoints')} rules={[{ required: true, message: t('sheet.maxPoints') }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        width={Math.min(640, typeof window !== 'undefined' ? window.innerWidth : 640)}
        placement="right"
        onClose={() => setFeedbackDrawer({ open: false, grade: null, task: null })}
        open={feedbackDrawer.open}
        title={`${t('task.feedback')} — ${feedbackDrawer.task?.name ?? ''}`}
      >
        {feedbackDrawer.grade && feedbackDrawer.task && (
          <>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label={t('task.grade')}>
                {feedbackDrawer.grade.acquired_points} / {feedbackDrawer.task.max_points}
              </Descriptions.Item>
              <Descriptions.Item label={t('submission.publicTests')}>
                <Tag color={feedbackDrawer.grade.public_test_status === TEST_STATUS_PASSED ? 'green' : 'red'}>
                  {feedbackDrawer.grade.public_test_status === TEST_STATUS_PASSED
                    ? t('submission.testsPassed')
                    : t('submission.testsFailed')}
                </Tag>
              </Descriptions.Item>
            </Descriptions>

            {feedbackDrawer.grade.public_test_log && (
              <>
                <Divider orientation="left">{t('submission.testLog')}</Divider>
                <pre
                  style={{
                    background: 'rgba(128,128,128,0.12)',
                    padding: 12,
                    borderRadius: 4,
                    fontSize: 12,
                    overflow: 'auto',
                    whiteSpace: 'pre-wrap',
                  }}
                >
                  {feedbackDrawer.grade.public_test_log}
                </pre>
              </>
            )}

            {feedbackDrawer.grade.feedback && (
              <>
                <Divider orientation="left">{t('task.feedback')}</Divider>
                <div style={{ padding: 12, borderRadius: 4, border: '1px solid rgba(128,128,128,0.3)' }}>
                  {feedbackDrawer.grade.feedback}
                </div>
              </>
            )}
          </>
        )}
      </Drawer>

      {isTutor && (
        <FloatButton icon={<PlusOutlined />} type="primary" tooltip={t('sheet.addTask')} onClick={() => handleTaskModalOpen()} />
      )}
    </div>
  );
};

export default SheetDetail;

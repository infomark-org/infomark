import React, { useCallback, useEffect, useState } from 'react';
import {
  Card,
  Tabs,
  Descriptions,
  Button,
  Space,
  Tag,
  message,
  Table,
  Row,
  Col,
  Statistic,
  Dropdown,
  Breadcrumb,
  List,
  Progress,
  Tooltip,
  Flex,
  FloatButton,
  Skeleton,
  Empty,
} from 'antd';
import {
  PlusOutlined,
  DownloadOutlined,
  EditOutlined,
  DeleteOutlined,
  MailOutlined,
  FileTextOutlined,
  FolderOutlined,
  TeamOutlined,
  TrophyOutlined,
  EyeOutlined,
  MoreOutlined,
  HomeOutlined,
  FileAddOutlined,
  FolderAddOutlined,
  UsergroupAddOutlined,
  LockOutlined,
  CheckSquareOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate, Link } from 'react-router-dom';
import agent from '@/api/agent';
import { useAuth } from '@/contexts/AuthContext';
import type { Course, Sheet, Material, Enrollment, CoursePoints, GradeOverview } from '@/types';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { ExamsPanel } from '@/components/course/ExamsPanel';
import { GroupsPanel } from '@/components/course/GroupsPanel';
import { EnrollmentsPanel } from '@/components/course/EnrollmentsPanel';

const ROLE_TUTOR = 1;
const ROLE_ADMIN = 2;
const MATERIAL_KIND_SLIDE = 0;

const CourseDetail: React.FC = () => {
  const { courseId } = useParams<{ courseId: string }>();
  const navigate = useNavigate();
  const { role } = useAuth();
  const { t } = useTranslation();
  const [course, setCourse] = useState<Course | null>(null);
  const [sheets, setSheets] = useState<Sheet[]>([]);
  const [materials, setMaterials] = useState<Material[]>([]);
  const [enrollmentCount, setEnrollmentCount] = useState<{ students: number; groups: number }>({ students: 0, groups: 0 });
  // Per-sheet points for the logged-in user (GET /courses/{id}/points).
  const [points, setPoints] = useState<CoursePoints[]>([]);
  // Per-student point matrix for tutors/admins (GET /courses/{id}/grades/summary).
  const [summary, setSummary] = useState<GradeOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState<number>(0);
  const [isEnrolled, setIsEnrolled] = useState(false);

  const cId = Number(courseId);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [courseData, sheetsData, materialsData] = await Promise.all([
        agent.Courses.get(cId),
        agent.Sheets.getAll(cId),
        agent.Materials.getAll(cId),
      ]);
      setCourse(courseData);
      setSheets(sheetsData);
      setMaterials(materialsData);

      // Resolve the caller's role in this course from their own enrollments.
      let myRole = 0;
      let enrolled = false;
      try {
        const myEnrollments = await agent.Account.getEnrollments();
        const mine = myEnrollments.find((e) => e.course_id === cId);
        if (mine) {
          myRole = mine.role;
          enrolled = true;
        }
      } catch {
        // ignore — treated as not enrolled
      }
      setUserRole(myRole);
      setIsEnrolled(enrolled);

      // Roster counts for the overview stats (tutor/admin can read enrollments).
      if (enrolled && myRole >= ROLE_TUTOR) {
        try {
          const enrollmentsData = await agent.Courses.getEnrollments(cId);
          setEnrollmentCount({
            students: enrollmentsData.filter((e: Enrollment) => e.role === 0).length,
            groups: (await agent.Groups.getAll(cId)).length,
          });
        } catch {
          setEnrollmentCount({ students: 0, groups: 0 });
        }
      }

      // Own per-sheet points (meaningful for every enrolled role).
      try {
        setPoints(await agent.Courses.getPoints(cId));
      } catch {
        setPoints([]);
      }

      // Tutors/admins additionally see the per-student matrix.
      if (enrolled && myRole >= ROLE_TUTOR) {
        try {
          setSummary(await agent.Grades.getSummary(cId));
        } catch {
          setSummary(null);
        }
      }
    } catch {
      message.error(t('messages.updateFailed'));
    } finally {
      setLoading(false);
    }
  }, [cId, t]);

  useEffect(() => {
    if (courseId) fetchData();
  }, [courseId, fetchData]);

  const handleDownloadSheet = async (sheetId: number, name: string) => {
    try {
      await agent.Sheets.downloadFile(cId, sheetId, `${name}.zip`);
    } catch {
      message.error(t('messages.downloadFailed'));
    }
  };

  const handleDownloadMaterial = async (materialId: number, name: string) => {
    try {
      await agent.Materials.downloadFile(cId, materialId, name);
    } catch {
      message.error(t('messages.downloadFailed'));
    }
  };

  const handleDeleteSheet = async (sheetId: number) => {
    try {
      await agent.Sheets.delete(cId, sheetId);
      message.success(t('messages.deleteSuccess'));
      await fetchData();
    } catch {
      message.error(t('messages.deleteFailed'));
    }
  };

  const handleDeleteMaterial = async (materialId: number) => {
    try {
      await agent.Materials.delete(cId, materialId);
      message.success(t('messages.deleteSuccess'));
      await fetchData();
    } catch {
      message.error(t('messages.deleteFailed'));
    }
  };

  const handleDeleteGroup = async (groupId: number) => {
    try {
      await agent.Groups.delete(cId, groupId);
      message.success(t('messages.deleteSuccess'));
      await fetchData();
    } catch {
      message.error(t('messages.deleteFailed'));
    }
  };

  if (loading || !course) return <Skeleton active paragraph={{ rows: 8 }} />;

  const isAdmin = userRole >= ROLE_ADMIN || role === 'root';
  const isTutor = userRole >= ROLE_TUTOR;
  const isStudent = isEnrolled && userRole === 0;

  // Course-wide progress from the logged-in user's own per-sheet points.
  const achievablePoints = points.reduce((sum, p) => sum + (p.max_points || 0), 0);
  const requiredPoints = Math.round((achievablePoints * course.required_percentage) / 100);
  const achievedPoints = points.reduce((sum, p) => sum + (p.acquired_points || 0), 0);
  const achievedPercent = achievablePoints > 0 ? Math.round((achievedPoints / achievablePoints) * 100) : 0;
  const requiredPercent = achievablePoints > 0 ? Math.round((requiredPoints / achievablePoints) * 100) : 0;

  // Map sheet_id -> points for the per-sheet progress bars in the sheets list.
  const pointsBySheet = new Map(points.map((p) => [p.sheet_id, p]));

  const sheetsTab = {
    key: 'sheets',
    label: (
      <span>
        <FileTextOutlined /> {t('course.sheets')} ({sheets.length})
      </span>
    ),
    children:
      sheets.length === 0 ? (
        <Empty description={t('course.sheets')} />
      ) : (
        <List
          dataSource={sheets}
          renderItem={(sheet) => {
            const sheetPoints = pointsBySheet.get(sheet.id);
            const published = dayjs().isAfter(dayjs(sheet.publish_at));
            return (
              <List.Item
                actions={[
                  <Button
                    key="view"
                    type="primary"
                    icon={<EyeOutlined />}
                    onClick={() => navigate(`/courses/${courseId}/sheets/${sheet.id}`)}
                  >
                    {t('common.view')}
                  </Button>,
                  <Button key="dl" icon={<DownloadOutlined />} onClick={() => handleDownloadSheet(sheet.id, sheet.name)}>
                    {t('common.download')}
                  </Button>,
                  isAdmin ? (
                    <Dropdown
                      key="menu"
                      menu={{
                        items: [
                          { key: 'edit', icon: <EditOutlined />, label: t('common.edit'), onClick: () => navigate(`/courses/${courseId}/sheets/${sheet.id}/edit`) },
                          { key: 'delete', icon: <DeleteOutlined />, label: t('common.delete'), danger: true, onClick: () => handleDeleteSheet(sheet.id) },
                        ],
                      }}
                    >
                      <Button icon={<MoreOutlined />} aria-label={t('common.actions')} />
                    </Dropdown>
                  ) : null,
                ].filter(Boolean)}
              >
                <List.Item.Meta
                  title={
                    <Space wrap>
                      {sheet.name}
                      {!published && <Tag>{t('submission.notPublished')}</Tag>}
                    </Space>
                  }
                  description={
                    <Space direction="vertical" size={4} style={{ width: '100%' }}>
                      <Space wrap>
                        <span>📅 {t('sheet.published')}: {dayjs(sheet.publish_at).format('MMM D, YYYY')}</span>
                        <span>⏰ {t('sheet.due')}: {dayjs(sheet.due_at).format('MMM D, YYYY HH:mm')}</span>
                      </Space>
                      {sheetPoints && sheetPoints.max_points > 0 && (
                        <Progress
                          percent={Math.round((sheetPoints.acquired_points / sheetPoints.max_points) * 100)}
                          size="small"
                          style={{ maxWidth: 260 }}
                          format={() => `${sheetPoints.acquired_points}/${sheetPoints.max_points}`}
                        />
                      )}
                    </Space>
                  }
                />
              </List.Item>
            );
          }}
        />
      ),
  };

  const materialsTab = {
    key: 'materials',
    label: (
      <span>
        <FolderOutlined /> {t('course.materials')} ({materials.length})
      </span>
    ),
    children:
      materials.length === 0 ? (
        <Empty description={t('course.materials')} />
      ) : (
        <Row gutter={[16, 16]}>
          {materials.map((material) => {
            // Honor the material's required role: a viewer below it cannot
            // download (the backend GetFileHandler enforces this too).
            const locked = userRole < material.required_role && role !== 'root';
            return (
              <Col xs={24} sm={12} lg={8} key={material.id}>
                <Card
                  title={material.name}
                  extra={
                    isAdmin && (
                      <Dropdown
                        menu={{
                          items: [
                            { key: 'edit', icon: <EditOutlined />, label: t('common.edit'), onClick: () => navigate(`/courses/${courseId}/materials/${material.id}/edit`) },
                            { key: 'delete', icon: <DeleteOutlined />, label: t('common.delete'), danger: true, onClick: () => handleDeleteMaterial(material.id) },
                          ],
                        }}
                      >
                        <Button icon={<MoreOutlined />} type="text" aria-label={t('common.actions')} />
                      </Dropdown>
                    )
                  }
                  actions={[
                    locked ? (
                      <Tooltip title={t('roles.tutor')} key="locked">
                        <Button type="text" icon={<LockOutlined />} disabled>
                          {t('material.required')}
                        </Button>
                      </Tooltip>
                    ) : (
                      <Button key="dl" type="primary" icon={<DownloadOutlined />} onClick={() => handleDownloadMaterial(material.id, material.name)}>
                        {t('common.download')}
                      </Button>
                    ),
                  ]}
                >
                  <Space direction="vertical" size={4}>
                    <Tag color="blue">{material.kind === MATERIAL_KIND_SLIDE ? t('material.material') : t('material.materials')}</Tag>
                    {material.required_role > 0 && <Tag icon={<LockOutlined />}>{t('roles.tutor')}+</Tag>}
                    <div>📅 {dayjs(material.publish_at).format('MMM D, YYYY')}</div>
                  </Space>
                </Card>
              </Col>
            );
          })}
        </Row>
      ),
  };

  const groupsTab = {
    key: 'groups',
    label: (
      <span>
        <TeamOutlined /> {t('course.groups')}
      </span>
    ),
    children: (
      <GroupsPanel
        courseId={cId}
        isStudent={isStudent}
        isTutor={isTutor}
        canManage={isAdmin}
        onEdit={(groupId) => navigate(`/courses/${courseId}/groups/${groupId}/edit`)}
        onDelete={handleDeleteGroup}
      />
    ),
  };

  const examsTab = {
    key: 'exams',
    label: (
      <span>
        <CheckSquareOutlined /> {t('course.exams')}
      </span>
    ),
    children: <ExamsPanel courseId={cId} isStudent={isStudent} />,
  };

  // Student's own per-sheet grades table.
  const gradesTab = {
    key: 'grades',
    label: (
      <span>
        <TrophyOutlined /> {t('course.grades')}
      </span>
    ),
    children: (
      <Table
        dataSource={points}
        rowKey={(p) => p.sheet_id}
        pagination={false}
        columns={[
          {
            title: t('grades.sheet'),
            key: 'sheet',
            render: (_, record) => sheets.find((s) => s.id === record.sheet_id)?.name ?? `${t('grades.sheet')} ${record.sheet_id}`,
          },
          { title: t('grades.points'), dataIndex: 'acquired_points', key: 'acquired' },
          { title: t('sheet.maxPoints'), dataIndex: 'max_points', key: 'max' },
          {
            title: '%',
            key: 'pct',
            render: (_, record) => (record.max_points > 0 ? `${Math.round((record.acquired_points / record.max_points) * 100)}%` : '—'),
          },
        ]}
        summary={(pageData) => {
          const acquired = pageData.reduce((sum, p) => sum + p.acquired_points, 0);
          const max = pageData.reduce((sum, p) => sum + p.max_points, 0);
          return (
            <Table.Summary.Row>
              <Table.Summary.Cell index={0}>
                <strong>{t('grades.totalPoints')}</strong>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={1}>
                <strong>{acquired}</strong>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={2}>
                <strong>{max}</strong>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={3}>
                <strong>{max > 0 ? `${Math.round((acquired / max) * 100)}%` : '—'}</strong>
              </Table.Summary.Cell>
            </Table.Summary.Row>
          );
        }}
      />
    ),
  };

  const enrollmentsTab = {
    key: 'enrollments',
    label: (
      <span>
        <TeamOutlined /> {t('course.enrollments')}
      </span>
    ),
    children: <EnrollmentsPanel courseId={cId} canManage={isAdmin} />,
  };

  const pointsMatrixTab = {
    key: 'points',
    label: (
      <span>
        <TrophyOutlined /> {t('course.points')}
      </span>
    ),
    children: (
      <Table
        dataSource={summary?.achievements ?? []}
        rowKey={(record) => record.user_info.id}
        scroll={{ x: 'max-content' }}
        columns={[
          {
            title: t('grades.student'),
            key: 'student',
            fixed: 'left' as const,
            render: (_, record) => `${record.user_info.first_name} ${record.user_info.last_name}`,
          },
          ...(summary?.sheets ?? []).map((sheet, index) => ({
            title: sheet.name,
            key: `sheet-${sheet.id}`,
            render: (_: unknown, record: GradeOverview['achievements'][number]) => record.points[index] ?? 0,
          })),
          {
            title: t('grades.totalPoints'),
            key: 'total',
            render: (_: unknown, record: GradeOverview['achievements'][number]) => (
              <Tag color="green">{record.points.reduce((sum, value) => sum + value, 0)}</Tag>
            ),
          },
        ]}
      />
    ),
  };

  // Compose the tab set by role. Students see grades; tutors/admins see the
  // roster and the class points matrix.
  const tabItems = [
    sheetsTab,
    materialsTab,
    groupsTab,
    examsTab,
    ...(isStudent ? [gradesTab] : []),
    ...(isTutor ? [enrollmentsTab, pointsMatrixTab] : []),
  ];

  return (
    <div style={{ overflowX: 'hidden' }}>
      <Breadcrumb
        items={[
          { title: <Link to="/"><HomeOutlined /></Link> },
          { title: <Link to="/courses">{t('nav.courses')}</Link> },
          { title: course.name },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card
        title={<span style={{ fontSize: 24 }}>{course.name}</span>}
        extra={
          <Space wrap>
            {isTutor && (
              <Button icon={<CheckSquareOutlined />} onClick={() => navigate(`/courses/${courseId}/grading`)}>
                {t('course.grading')}
              </Button>
            )}
            {isAdmin && (
              <>
                <Button icon={<EditOutlined />} onClick={() => navigate(`/courses/${courseId}/edit`)}>
                  {t('course.editCourse')}
                </Button>
                <Button icon={<MailOutlined />} onClick={() => navigate(`/courses/${courseId}/mail`)}>
                  {t('course.sendEmail')}
                </Button>
              </>
            )}
          </Space>
        }
      >
        <Flex gap="large" align="start" wrap="wrap">
          <Descriptions column={{ xs: 1, sm: 2 }} style={{ flex: 1, minWidth: 260 }}>
            <Descriptions.Item label={t('course.description')} span={2}>
              {course.description}
            </Descriptions.Item>
            <Descriptions.Item label={t('course.startDate')}>{dayjs(course.begins_at).format('MMM D, YYYY')}</Descriptions.Item>
            <Descriptions.Item label={t('course.endDate')}>{dayjs(course.ends_at).format('MMM D, YYYY')}</Descriptions.Item>
            <Descriptions.Item label={t('course.requiredPercentage')}>{course.required_percentage}%</Descriptions.Item>
          </Descriptions>
          {isEnrolled && (
            <Tooltip title={`${achievedPoints} / ${requiredPoints} / ${achievablePoints}`}>
              <Progress
                type="dashboard"
                percent={achievedPercent}
                success={{ percent: requiredPercent }}
                format={() => `${achievedPoints}/${achievablePoints}`}
              />
            </Tooltip>
          )}
        </Flex>

        <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
          <Col xs={12} sm={8}>
            <Card>
              <Statistic title={t('course.sheets')} value={sheets.length} prefix={<FileTextOutlined />} />
            </Card>
          </Col>
          <Col xs={12} sm={8}>
            <Card>
              <Statistic title={t('course.materials')} value={materials.length} prefix={<FolderOutlined />} />
            </Card>
          </Col>
          {isTutor && (
            <Col xs={12} sm={8}>
              <Card>
                <Statistic title={t('course.students')} value={enrollmentCount.students} prefix={<TeamOutlined />} />
              </Card>
            </Col>
          )}
        </Row>
      </Card>

      <Card style={{ marginTop: 16 }}>
        <Tabs items={tabItems} />
      </Card>

      {isAdmin && (
        <FloatButton.Group trigger="click" type="primary" icon={<PlusOutlined />}>
          <FloatButton icon={<FileAddOutlined />} tooltip={t('sheet.newSheet')} onClick={() => navigate(`/courses/${courseId}/sheets/new`)} />
          <FloatButton icon={<FolderAddOutlined />} tooltip={t('material.newMaterial')} onClick={() => navigate(`/courses/${courseId}/materials/new`)} />
          <FloatButton icon={<UsergroupAddOutlined />} tooltip={t('group.newGroup')} onClick={() => navigate(`/courses/${courseId}/groups/new`)} />
        </FloatButton.Group>
      )}
    </div>
  );
};

export default CourseDetail;

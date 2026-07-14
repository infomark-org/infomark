import React, { useCallback, useEffect, useState } from 'react';
import { Row, Col, Card, Tag, InputNumber, Button, Space, Skeleton, Empty, message, Avatar, List, Tooltip, Dropdown } from 'antd';
import { TeamOutlined, UserOutlined, CrownOutlined, MoreOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import agent from '@/api/agent';
import type { Group, GroupBid, Enrollment } from '@/types';
import { useTranslation } from 'react-i18next';

interface GroupsPanelProps {
  courseId: number;
  isStudent: boolean;
  isTutor: boolean;
  // Admins get per-group edit/delete affordances; the handlers live in the
  // parent so navigation/refresh stay centralised.
  canManage?: boolean;
  onEdit?: (groupId: number) => void;
  onDelete?: (groupId: number) => void;
}

// Bids are constrained to 0..10 by the backend (group_requests.go GroupBidRequest).
const BID_MIN = 0;
const BID_MAX = 10;

// GroupsPanel shows every exercise group. Students see their own group
// highlighted and can place a 0..10 preference bid on each group (used by the
// backend's auto-assignment). Tutors additionally see the students enrolled in
// each group.
export const GroupsPanel: React.FC<GroupsPanelProps> = ({ courseId, isStudent, isTutor, canManage, onEdit, onDelete }) => {
  const { t } = useTranslation();
  const [groups, setGroups] = useState<Group[]>([]);
  const [ownGroupId, setOwnGroupId] = useState<number | null>(null);
  const [bids, setBids] = useState<Map<number, number>>(new Map());
  const [members, setMembers] = useState<Map<number, Enrollment[]>>(new Map());
  const [loading, setLoading] = useState(true);
  const [savingBidFor, setSavingBidFor] = useState<number | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const groupsData = await agent.Groups.getAll(courseId);
      setGroups(groupsData);

      if (isStudent) {
        // The student's own group (a one-element list) and their bids.
        const [own, bidList] = await Promise.all([
          agent.Groups.getOwn(courseId).catch(() => [] as Group[]),
          agent.Groups.getBids(courseId).catch(() => [] as GroupBid[]),
        ]);
        setOwnGroupId(own.length > 0 ? own[0].id : null);
        const bidMap = new Map<number, number>();
        bidList.forEach((b) => bidMap.set(b.group_id, b.bid));
        setBids(bidMap);
      }

      if (isTutor) {
        // Enrolled students per group (tutor/admin only endpoint).
        const memberMap = new Map<number, Enrollment[]>();
        await Promise.all(
          groupsData.map(async (group) => {
            try {
              const enrolled = await agent.Groups.getEnrollments(courseId, group.id);
              memberMap.set(group.id, enrolled);
            } catch {
              memberMap.set(group.id, []);
            }
          }),
        );
        setMembers(memberMap);
      }
    } catch {
      message.error(t('messages.updateFailed'));
    } finally {
      setLoading(false);
    }
  }, [courseId, isStudent, isTutor, t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleBid = async (groupId: number, bid: number) => {
    setSavingBidFor(groupId);
    try {
      await agent.Groups.placeBid(courseId, groupId, bid);
      setBids((prev) => new Map(prev).set(groupId, bid));
      message.success(t('messages.saveSuccess'));
    } catch {
      message.error(t('messages.saveFailed'));
    } finally {
      setSavingBidFor(null);
    }
  };

  if (loading) return <Skeleton active paragraph={{ rows: 4 }} />;
  if (groups.length === 0) return <Empty description={t('group.groups')} />;

  return (
    <Row gutter={[16, 16]}>
      {groups.map((group) => {
        const isOwn = ownGroupId === group.id;
        const groupMembers = members.get(group.id) ?? [];
        return (
          <Col xs={24} sm={12} lg={8} key={group.id}>
            <Card
              title={
                <Space>
                  <TeamOutlined />
                  {`${t('group.group')} ${group.id}`}
                  {isOwn && (
                    <Tag icon={<CrownOutlined />} color="gold">
                      {t('course.myGroup')}
                    </Tag>
                  )}
                </Space>
              }
              style={isOwn ? { borderColor: '#faad14', borderWidth: 2 } : undefined}
              extra={
                canManage && (
                  <Dropdown
                    menu={{
                      items: [
                        { key: 'edit', icon: <EditOutlined />, label: t('common.edit'), onClick: () => onEdit?.(group.id) },
                        { key: 'delete', icon: <DeleteOutlined />, label: t('common.delete'), danger: true, onClick: () => onDelete?.(group.id) },
                      ],
                    }}
                  >
                    <Button type="text" icon={<MoreOutlined />} aria-label={t('common.actions')} />
                  </Dropdown>
                )
              }
            >
              <p style={{ minHeight: 44 }}>{group.description}</p>
              <div>
                <UserOutlined /> {t('group.tutor')}: {group.tutor.first_name} {group.tutor.last_name}
              </div>

              {isStudent && (
                <Space style={{ marginTop: 12 }} align="center">
                  <span>{t('group.group')} {t('common.actions')}:</span>
                  <Tooltip title={`${BID_MIN}–${BID_MAX}`}>
                    <InputNumber
                      min={BID_MIN}
                      max={BID_MAX}
                      defaultValue={bids.get(group.id) ?? BID_MIN}
                      aria-label={`bid-${group.id}`}
                      onPressEnter={(e) => handleBid(group.id, Number((e.target as HTMLInputElement).value))}
                      id={`bid-input-${group.id}`}
                    />
                  </Tooltip>
                  <Button
                    size="small"
                    loading={savingBidFor === group.id}
                    onClick={() => {
                      const input = document.getElementById(`bid-input-${group.id}`) as HTMLInputElement | null;
                      const value = input ? Number(input.value) : BID_MIN;
                      handleBid(group.id, Number.isNaN(value) ? BID_MIN : value);
                    }}
                  >
                    {t('common.save')}
                  </Button>
                </Space>
              )}

              {isTutor && (
                <List
                  size="small"
                  style={{ marginTop: 12 }}
                  header={<strong>{t('group.members')} ({groupMembers.filter((m) => m.role === 0).length})</strong>}
                  dataSource={groupMembers.filter((m) => m.role === 0)}
                  locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('group.members')} /> }}
                  renderItem={(member) => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={<Avatar size="small" icon={<UserOutlined />} src={member.user.avatar_url} />}
                        title={`${member.user.first_name} ${member.user.last_name}`}
                        description={member.user.email}
                      />
                    </List.Item>
                  )}
                />
              )}
            </Card>
          </Col>
        );
      })}
    </Row>
  );
};

export default GroupsPanel;

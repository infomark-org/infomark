import React, { useEffect, useState } from 'react';
import { Form, Input, Button, Card, DatePicker, Upload, message, Breadcrumb } from 'antd';
import { InboxOutlined, HomeOutlined } from '@ant-design/icons';
import { useNavigate, useParams, Link } from 'react-router-dom';
import agent from '@/api/agent';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';

const { Dragger } = Upload;

const SheetEditor: React.FC = () => {
  const { courseId, sheetId } = useParams<{ courseId: string; sheetId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [file, setFile] = useState<File | null>(null);

  useEffect(() => {
    if (sheetId) {
      loadSheet();
    }
  }, [sheetId]);

  const loadSheet = async () => {
    try {
      const sheet = await agent.Sheets.get(parseInt(courseId!), parseInt(sheetId!));
      form.setFieldsValue({
        ...sheet,
        publish_at: dayjs(sheet.publish_at),
        due_at: dayjs(sheet.due_at),
      });
    } catch (error) {
      message.error(t('sheet.loadFailed'));
    }
  };

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        publish_at: values.publish_at.toISOString(),
        due_at: values.due_at.toISOString(),
      };

      const cId = parseInt(courseId!);

      if (sheetId) {
        await agent.Sheets.update(cId, parseInt(sheetId), data);
        message.success(t('sheet.updated'));
      } else {
        const newSheet = await agent.Sheets.create(cId, data);
        if (file) {
          await agent.Sheets.uploadFile(cId, newSheet.id, file);
        }
        message.success(t('sheet.created'));
      }

      navigate(`/courses/${courseId}`);
    } catch (error) {
      message.error(sheetId ? t('sheet.updateFailed') : t('sheet.createFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Breadcrumb
        items={[
          { title: <Link to="/"><HomeOutlined /></Link> },
          { title: <Link to="/courses">{t('course.courses')}</Link> },
          { title: <Link to={`/courses/${courseId}`}>{t('course.course')}</Link> },
          { title: sheetId ? t('sheet.editSheet') : t('sheet.newSheet') },
        ]}
        style={{ marginBottom: 16 }}
      />
      
      <Card title={sheetId ? t('sheet.editSheet') : t('sheet.createSheet')}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="name" label={t('sheet.sheetName')} rules={[{ required: true }]}>
          <Input />
        </Form.Item>

        <Form.Item name="publish_at" label={t('sheet.publishDate')} rules={[{ required: true }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="due_at" label={t('sheet.dueDate')} rules={[{ required: true }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item label={t('sheet.sheetFile')}>
          <Dragger
            beforeUpload={(f) => {
              if (!f.name.endsWith('.zip')) {
                message.error(t('sheet.onlyZip'));
                return false;
              }
              setFile(f);
              return false;
            }}
            maxCount={1}
            onRemove={() => setFile(null)}
            accept=".zip"
          >
            <p className="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p className="ant-upload-text">{t('sheet.dragHint')}</p>
            <p className="ant-upload-hint">{t('sheet.uploadHint')}</p>
          </Dragger>
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            {sheetId ? t('sheet.updateSheet') : t('sheet.createSheet')}
          </Button>
          <Button style={{ marginLeft: 8 }} onClick={() => navigate(`/courses/${courseId}`)}>
            {t('common.cancel')}
          </Button>
        </Form.Item>
      </Form>
    </Card>
    </>
  );
};

export default SheetEditor;

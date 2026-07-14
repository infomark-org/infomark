import React, { useEffect, useState } from 'react';
import { Form, Input, Button, Card, DatePicker, Upload, Select, message } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import agent from '@/api/agent';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';

const MaterialEditor: React.FC = () => {
  const { courseId, materialId } = useParams<{ courseId: string; materialId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [file, setFile] = useState<File | null>(null);

  useEffect(() => {
    if (materialId) {
      loadMaterial();
    }
  }, [materialId]);

  const loadMaterial = async () => {
    try {
      const material = await agent.Materials.get(parseInt(courseId!), parseInt(materialId!));
      form.setFieldsValue({
        ...material,
        publish_at: dayjs(material.publish_at),
        lecture_at: dayjs(material.lecture_at),
      });
    } catch (error) {
      message.error(t('material.loadFailed'));
    }
  };

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        publish_at: values.publish_at.toISOString(),
        lecture_at: values.lecture_at.toISOString(),
      };

      const cId = parseInt(courseId!);

      if (materialId) {
        await agent.Materials.update(cId, parseInt(materialId), data);
        message.success(t('material.updated'));
      } else {
        const newMaterial = await agent.Materials.create(cId, data);
        if (file) {
          await agent.Materials.uploadFile(cId, newMaterial.id, file);
        }
        message.success(t('material.created'));
      }

      navigate(`/courses/${courseId}`);
    } catch (error) {
      message.error(materialId ? t('material.updateFailed') : t('material.createFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title={materialId ? t('material.editMaterial') : t('material.createMaterial')}>
      <Form form={form} layout="vertical" onFinish={onFinish} initialValues={{ kind: 0, required_role: 0 }}>
        <Form.Item name="name" label={t('material.materialName')} rules={[{ required: true }]}>
          <Input />
        </Form.Item>

        <Form.Item name="kind" label={t('material.type')} rules={[{ required: true }]}>
          <Select>
            <Select.Option value={0}>{t('material.slide')}</Select.Option>
            <Select.Option value={1}>{t('material.supplementary')}</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item name="publish_at" label={t('material.publishDate')} rules={[{ required: true }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="lecture_at" label={t('material.lectureDate')} rules={[{ required: true }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="required_role" label={t('material.requiredRole')} rules={[{ required: true }]}>
          <Select>
            <Select.Option value={0}>{t('roles.student')}</Select.Option>
            <Select.Option value={1}>{t('roles.tutor')}</Select.Option>
            <Select.Option value={2}>{t('roles.admin')}</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item label={t('material.materialFile')}>
          <Upload
            beforeUpload={(f) => {
              setFile(f);
              return false;
            }}
            maxCount={1}
          >
            <Button icon={<UploadOutlined />}>{t('material.selectFile')}</Button>
          </Upload>
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            {materialId ? t('material.updateMaterial') : t('material.createMaterial')}
          </Button>
          <Button style={{ marginLeft: 8 }} onClick={() => navigate(`/courses/${courseId}`)}>
            {t('common.cancel')}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default MaterialEditor;

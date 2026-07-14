import React, { useEffect, useState } from 'react';
import { Card, Spin } from 'antd';
import agent from '@/api/agent';
import { useTranslation } from 'react-i18next';

const Terms: React.FC = () => {
  const { t } = useTranslation();
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchTerms();
  }, []);

  const fetchTerms = async () => {
    try {
      const data = await agent.Terms.get();
      setContent(data);
    } catch (error) {
      setContent(t('terms.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <Spin size="large" />;

  return (
    <Card title={t('terms.title')}>
      <div dangerouslySetInnerHTML={{ __html: content }} />
    </Card>
  );
};

export default Terms;

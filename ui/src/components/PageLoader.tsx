import React from 'react';
import { Skeleton, Space } from 'antd';

// A route-level fallback shown while a lazily-loaded page chunk downloads.
// Using a skeleton (rather than a bare spinner) preserves layout height and
// reads as "content is coming" instead of "something is stuck".
export const PageLoader: React.FC = () => (
  <Space direction="vertical" style={{ width: '100%' }} size="large">
    <Skeleton active paragraph={{ rows: 2 }} />
    <Skeleton active paragraph={{ rows: 4 }} />
  </Space>
);

export default PageLoader;

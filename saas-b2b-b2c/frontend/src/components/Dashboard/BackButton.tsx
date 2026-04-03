import React from 'react';
import { Button } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useRouter } from 'next/router';

const BackButton: React.FC = () => {
  const router = useRouter();

  return (
    <Button
      type="text"
      icon={<ArrowLeftOutlined />}
      onClick={() => router.push('/')}
      style={{ marginBottom: '16px', padding: 0, color: '#1890ff' }}
    >
      На главную
    </Button>
  );
};

export default BackButton;
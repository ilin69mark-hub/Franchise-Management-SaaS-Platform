import React from 'react';
import { Layout, Typography } from 'antd';
import Link from 'next/link';

const { Header: AntHeader } = Layout;
const { Text } = Typography;

const Header: React.FC = () => {
  return (
    <AntHeader style={{ background: '#fff', padding: '0 24px', display: 'flex', alignItems: 'center', borderBottom: '1px solid #f0f0f0' }}>
      <Link href="/" style={{ display: 'flex', alignItems: 'center' }}>
        <Text strong style={{ fontSize: '18px', color: '#1890ff' }}>ivan.ru</Text>
      </Link>
    </AntHeader>
  );
};

export default Header;

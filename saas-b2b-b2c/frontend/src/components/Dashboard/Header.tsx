import React from 'react';
import { Layout, Typography, Space, Dropdown, Avatar } from 'antd';
import { UserOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { useDispatch } from 'react-redux';
import { logout } from '@/store/authSlice';
import NotificationBell from './NotificationBell';

const { Header: AntHeader } = Layout;
const { Text } = Typography;

const Header: React.FC = () => {
  const router = useRouter();
  const dispatch = useDispatch();

  const handleLogout = () => {
    // 1. Диспатчим экшен logout в Redux (это очистит user и isAuthenticated)
    dispatch(logout());
    
    // 2. Редиректим на логин
    router.push('/login');
  };

  const menu = {
    items: [
      { key: 'profile', icon: <UserOutlined />, label: <Link href="/profile">Профиль</Link> },
      { key: 'settings', icon: <SettingOutlined />, label: <Link href="/settings">Настройки</Link> },
      { type: 'divider' as const },
      { key: 'logout', icon: <LogoutOutlined />, label: 'Выйти', onClick: handleLogout },
    ],
  };

  return (
    <AntHeader style={{ background: '#fff', padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid #f0f0f0' }}>
      <Link href="/" style={{ display: 'flex', alignItems: 'center' }}>
        <Text strong style={{ fontSize: '18px', color: '#1890ff' }}>Franchise SaaS</Text>
      </Link>
      <Space size="middle">
        <NotificationBell />
        <Dropdown menu={menu} placement="bottomRight">
          <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
            <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1890ff' }} />
            <Text style={{ marginLeft: 8 }}>Мой кабинет</Text>
          </div>
        </Dropdown>
      </Space>
    </AntHeader>
  );
};

export default Header;

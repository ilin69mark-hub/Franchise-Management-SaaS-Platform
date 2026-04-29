import React, { useEffect, Suspense, lazy, useState } from 'react';
import { Layout, Row, Col, Card, Typography, Badge, Avatar, Dropdown, Space, Spin, Statistic, Menu, theme } from 'antd';
import {
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  BellOutlined,
  DashboardOutlined,
  ApartmentOutlined,
  HeartOutlined,
  ToolOutlined,
  DollarOutlined,
  AuditOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SunOutlined,
  MoonOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import { useDispatch } from 'react-redux';
import { logout } from '@/store/authSlice';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import { useSuperAdminStore } from '@/store/superAdminStore';

dayjs.locale('ru');

const { Header: AntHeader, Sider, Content } = Layout;
const { Title, Text } = Typography;

const MetricsSection = lazy(() => import('./sections/MetricsSection'));
const TenantsSection = lazy(() => import('./sections/TenantsSection'));
const ActivitySection = lazy(() => import('./sections/ActivitySection'));
const TechHealthSection = lazy(() => import('./sections/TechHealthSection'));
const BillingSection = lazy(() => import('./sections/BillingSection'));
const AuditSection = lazy(() => import('./sections/AuditSection'));

interface UserData {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role?: string;
}

interface SuperAdminDashboardProps {
  user: UserData;
}

const SuperAdminDashboard: React.FC<SuperAdminDashboardProps> = ({ user }) => {
  const router = useRouter();
  const dispatch = useDispatch();
  const { token } = theme.useToken();
  const {
    activeSection,
    stats,
    isLoading,
    alertCount,
    lastUpdated,
    setActiveSection,
    setStats,
    setLoading,
    setAlertCount,
    setLastUpdated,
  } = useSuperAdminStore();

  const [currentDate, setCurrentDate] = useState(dayjs().format('D MMMM YYYY, dddd'));
  const [currentTime, setCurrentTime] = useState(dayjs().format('HH:mm'));
  const [darkMode, setDarkMode] = useState(false);
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentDate(dayjs().format('D MMMM YYYY, dddd'));
      setCurrentTime(dayjs().format('HH:mm'));
    }, 60000);

    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    const fetchStats = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem('accessToken');
        const res = await fetch('/api/v1/admin/stats', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const data = await res.json();
          setStats({
            mrr: data.mrr || 0,
            arr: data.arr || 0,
            churnRate: data.churnRate || 0,
            overduePayments: data.overduePayments || 0,
            arpu: data.arpu || 0,
          });
          setAlertCount(data.alerts || 0);
        }
      } catch {
        setStats({
          mrr: 2500000,
          arr: 30000000,
          churnRate: 5.2,
          overduePayments: 3,
          arpu: 15000,
        });
      } finally {
        setLoading(false);
        setLastUpdated(new Date());
      }
    };

    fetchStats();
    const updateInterval = setInterval(fetchStats, 60 * 1000);
    return () => clearInterval(updateInterval);
  }, [setStats, setLoading, setLastUpdated]);

  const handleLogout = () => {
    dispatch(logout());
    router.push('/login');
  };

  const handleThemeToggle = () => {
    setDarkMode(!darkMode);
  };

  const userMenu = {
    items: [
      { key: 'profile', icon: <UserOutlined />, label: 'Профиль' },
      { key: 'settings', icon: <SettingOutlined />, label: 'Настройки' },
      { type: 'divider' as const },
      { key: 'logout', icon: <LogoutOutlined />, label: 'Выйти', onClick: handleLogout },
    ],
  };

  const menuItems = [
    { key: 'metrics', icon: <DashboardOutlined />, label: 'SaaS-метрики' },
    { key: 'tenants', icon: <ApartmentOutlined />, label: 'Тенанты' },
    { key: 'activity', icon: <HeartOutlined />, label: 'Активность' },
    { key: 'tech', icon: <ToolOutlined />, label: 'Тех. здоровье' },
    { key: 'billing', icon: <DollarOutlined />, label: 'Биллинг' },
    { key: 'audit', icon: <AuditOutlined />, label: 'Аудит' },
  ];

  const renderSection = () => {
    return (
      <Suspense
        fallback={
          <div style={{ textAlign: 'center', padding: 50 }}>
            <Spin size="large" />
          </div>
        }
      >
        {activeSection === 'metrics' && <MetricsSection />}
        {activeSection === 'tenants' && <TenantsSection />}
        {activeSection === 'activity' && <ActivitySection />}
        {activeSection === 'tech' && <TechHealthSection />}
        {activeSection === 'billing' && <BillingSection />}
        {activeSection === 'audit' && <AuditSection />}
      </Suspense>
    );
  };

  const layoutStyle = darkMode
    ? { minHeight: '100vh', background: '#141414' }
    : { minHeight: '100vh', background: '#f5f5f5' };

  const headerStyle = darkMode
    ? { background: '#1f1f1f', borderBottom: '1px solid #303030' }
    : { background: '#fff', borderBottom: '1px solid #f0f0f0' };

  const sidebarStyle = darkMode
    ? { background: '#1f1f1f', borderRight: '1px solid #303030' }
    : { background: '#fff', borderRight: '1px solid #f0f0f0' };

  const contentStyle = darkMode
    ? { background: '#141414', color: '#fff' }
    : { background: '#fff' };

  return (
    <Layout style={layoutStyle}>
      <AntHeader
        style={{
          ...headerStyle,
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          position: 'sticky',
          top: 0,
          zIndex: 100,
          height: 64,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
          <Title level={4} style={{ margin: 0, color: token.colorPrimary }}>
            SaaS Платформа
          </Title>
          <Badge status="processing" />
          <Text type="secondary" style={{ color: darkMode ? '#aaa' : undefined }}>
            Супер Админ
          </Text>
        </div>

        <Row gutter={[24, 8]} align="middle" style={{ flex: 1, margin: '0 48px' }}>
          <Col xs={12} sm={4} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic
                title="MRR"
                value={stats.mrr}
                precision={0}
                prefix="₽ "
                valueStyle={{ fontSize: 16, color: darkMode ? '#fff' : undefined }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={4} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic
                title="ARR"
                value={stats.arr}
                precision={0}
                prefix="₽ "
                suffix="/год"
                valueStyle={{ fontSize: 16, color: darkMode ? '#fff' : undefined }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={4} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic
                title="Churn Rate"
                value={stats.churnRate}
                precision={1}
                suffix="%"
                valueStyle={{
                  fontSize: 16,
                  color: stats.churnRate > 10 ? '#ff4d4f' : darkMode ? '#fff' : undefined,
                }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={4} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic
                title="Просроч. платежи"
                value={stats.overduePayments}
                valueStyle={{
                  fontSize: 16,
                  color: stats.overduePayments > 0 ? '#ff4d4f' : darkMode ? '#fff' : undefined,
                }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={4} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic
                title="ARPU"
                value={stats.arpu}
                precision={0}
                prefix="₽ "
                valueStyle={{ fontSize: 16, color: darkMode ? '#fff' : undefined }}
              />
            </Card>
          </Col>
        </Row>

        <Space size="middle">
          <Text type="secondary" style={{ color: darkMode ? '#aaa' : undefined }}>
            {currentDate} {currentTime}
          </Text>
          <Badge count={alertCount} offset={[-5, 5]}>
            <BellOutlined
              style={{
                fontSize: 20,
                cursor: 'pointer',
                color: darkMode ? '#fff' : undefined,
              }}
            />
          </Badge>
          <span onClick={handleThemeToggle} style={{ cursor: 'pointer', fontSize: 18 }}>
            {darkMode ? <SunOutlined /> : <MoonOutlined />}
          </span>
          <Dropdown menu={userMenu} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: token.colorPrimary }} />
            </div>
          </Dropdown>
        </Space>
      </AntHeader>

      <Layout>
        <Sider
          width={220}
          collapsedWidth={80}
          collapsed={collapsed}
          style={sidebarStyle}
          trigger={null}
        >
          <div
            onClick={() => setCollapsed(!collapsed)}
            style={{
              padding: '12px 16px',
              cursor: 'pointer',
              borderBottom: `1px solid ${darkMode ? '#303030' : '#f0f0f0'}`,
              textAlign: 'center',
            }}
          >
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
          <Menu
            mode="inline"
            selectedKeys={[activeSection]}
            onClick={({ key }) => setActiveSection(key)}
            style={{
              background: 'transparent',
              borderRight: 0,
            }}
            items={menuItems}
          />
        </Sider>

        <Content style={{ padding: 24, ...contentStyle }}>
          {renderSection()}
        </Content>
      </Layout>
    </Layout>
  );
};

export default SuperAdminDashboard;
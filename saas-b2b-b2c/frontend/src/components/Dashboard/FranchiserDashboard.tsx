import React, { useState, useEffect, Suspense, lazy } from 'react';
import { Layout, Row, Col, Card, Typography, Tabs, Badge, Avatar, Dropdown, Space, Spin, Statistic, message } from 'antd';
import { 
  UserOutlined, 
  LogoutOutlined, 
  SettingOutlined, 
  BellOutlined,
  PercentageOutlined,
  RiseOutlined,
  TeamOutlined,
  ShopOutlined,
  DesktopOutlined,
  FileTextOutlined,
  DashboardOutlined,
  HeartOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import { logout } from '@/store/authSlice';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import { useFranchiserStore, FranchiserSummary } from '@/store/franchiserStore';

dayjs.locale('ru');

const { Header: AntHeader } = Layout;
const { Content } = Layout;
const { Title, Text } = Typography;

const NetworkDashboardTab = lazy(() => import('./tabs/FranchiserNetworkTab'));
const TeamTab = lazy(() => import('./tabs/FranchiserTeamTab'));
const NetworkHealthTab = lazy(() => import('./tabs/FranchiserNetworkHealthTab'));
const ReportTab = lazy(() => import('./tabs/FranchiserReportTab'));

interface FranchiserDashboardProps {
  user: {
    id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    role?: string;
  };
  title?: string;
}

const FranchiserDashboard: React.FC<FranchiserDashboardProps> = ({ user, title }) => {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState('network');
  
  const {
    summary,
    isLoading,
    alertCount,
    fetchSummary,
  } = useFranchiserStore();

  useEffect(() => {
    fetchSummary();
    const interval = setInterval(fetchSummary, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, [fetchSummary]);

  const handleLogout = () => {
    logout();
    localStorage.removeItem('accessToken');
    router.push('/login');
  };

  const currentDate = dayjs().format('dddd, D MMMM YYYY');

  const userMenuItems = [
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: 'Настройки',
    },
    {
      key: 'divider',
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Выход',
      danger: true,
      onClick: handleLogout,
    },
  ];

  const tabsItems = [
    {
      key: 'network',
      label: (
        <span>
          <DashboardOutlined /> Пульт сети
        </span>
      ),
    },
    {
      key: 'team',
      label: (
        <span>
          <TeamOutlined /> Моя команда
        </span>
      ),
    },
    {
      key: 'health',
      label: (
        <span>
          <HeartOutlined /> Здоровье сети
        </span>
      ),
    },
    {
      key: 'report',
      label: (
        <span>
          <FileTextOutlined /> Отчёт для B2B
        </span>
      ),
    },
  ];

  const renderTabContent = () => {
    switch (activeTab) {
      case 'network':
        return <NetworkDashboardTab summary={summary} />;
      case 'team':
        return <TeamTab />;
      case 'health':
        return <NetworkHealthTab />;
      case 'report':
        return <ReportTab />;
      default:
        return null;
    }
  };

  return (
    <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
      <AntHeader style={{ 
        background: '#fff', 
        padding: '0 24px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
        position: 'sticky',
        top: 0,
        zIndex: 100,
      }}>
        <Space align="center">
          <Title level={4} style={{ margin: 0, marginRight: 24 }}>
            Руководитель отдела франчайзинга
          </Title>
          <Text type="secondary">{currentDate}</Text>
        </Space>

        <Space size={24}>
          <Row gutter={16}>
            <Col>
              <Card size="small" style={{ textAlign: 'center' }}>
                <Statistic 
                  title="Выполнение плана" 
                  value={summary.planPercent} 
                  suffix="%" 
                  valueStyle={{ fontSize: 20, color: summary.planPercent >= 100 ? '#52c41a' : '#fa8c16' }}
                />
              </Card>
            </Col>
            <Col>
              <Card size="small" style={{ textAlign: 'center' }}>
                <Statistic 
                  title="Прогноз квартала" 
                  value={summary.forecastPercent} 
                  suffix="%" 
                  valueStyle={{ fontSize: 20, color: summary.forecastPercent >= 100 ? '#52c41a' : '#fa8c16' }}
                />
              </Card>
            </Col>
            <Col>
              <Card size="small" style={{ textAlign: 'center' }}>
                <Statistic 
                  title="Активных дилеров" 
                  value={summary.activeDealers} 
                  valueStyle={{ fontSize: 20 }}
                />
              </Card>
            </Col>
            <Col>
              <Card size="small" style={{ textAlign: 'center' }}>
                <Statistic 
                  title="Конверсия сети" 
                  value={summary.avgConversion} 
                  suffix="%" 
                  valueStyle={{ fontSize: 20 }}
                />
              </Card>
            </Col>
            <Col>
              <Card size="small" style={{ textAlign: 'center' }}>
                <Statistic 
                  title="Маржинальность" 
                  value={summary.avgMargin} 
                  suffix="%" 
                  valueStyle={{ fontSize: 20, color: summary.avgMargin >= 30 ? '#52c41a' : '#fa8c16' }}
                />
              </Card>
            </Col>
          </Row>

          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Badge count={alertCount} size="small">
              <Badge dot={alertCount > 0} color="red">
                <Avatar 
                  icon={<UserOutlined />} 
                  style={{ backgroundColor: '#1890ff', cursor: 'pointer' }}
                />
              </Badge>
            </Badge>
          </Dropdown>
        </Space>
      </AntHeader>

      <Content style={{ padding: 24 }}>
        <Card>
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            items={tabsItems}
            style={{ marginBottom: 16 }}
          />
          
          <Suspense fallback={<Spin tip="Загрузка..." style={{ display: 'block', margin: '40px auto' }} />}>
            {renderTabContent()}
          </Suspense>
        </Card>
      </Content>
    </Layout>
  );
};

export default FranchiserDashboard;
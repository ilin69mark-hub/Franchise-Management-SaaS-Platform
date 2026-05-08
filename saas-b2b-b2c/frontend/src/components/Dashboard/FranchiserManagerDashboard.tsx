// src/components/Dashboard/FranchiserManagerDashboard.tsx
import React, { useEffect, Suspense, lazy, useState } from 'react';
import { useDispatch } from 'react-redux';
import { Layout, Row, Col, Card, Typography, Tabs, Badge, Avatar, Dropdown, Space, Spin, Statistic, Button, Modal, List, Divider } from 'antd';
import { 
  UserOutlined, 
  LogoutOutlined, 
  SettingOutlined, 
  BellOutlined,
  PercentageOutlined,
  RiseOutlined,
  WarningOutlined,
  AppstoreOutlined,
  LineChartOutlined,
  MessageOutlined,
  BarChartOutlined,
  MenuOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import { logout } from '@/store/authSlice';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import { useTerritoryManagerStore, TerritorySummary, DealerMetrics } from '@/store/territoryManagerStore';

dayjs.locale('ru');

const { Header: AntHeader } = Layout;
const { Content } = Layout;
const { Title, Text } = Typography;

const TerritoryMapTab = lazy(() => import('./tabs/TerritoryMapTab'));
const TerritoryFunnelTab = lazy(() => import('./tabs/TerritoryFunnelTab'));
const TerritoryPlanFactTab = lazy(() => import('./tabs/TerritoryPlanFactTab'));
const TerritoryCommunicationsTab = lazy(() => import('./tabs/TerritoryCommunicationsTab'));
const TerritoryBenchmarkTab = lazy(() => import('./tabs/TerritoryBenchmarkTab'));

interface FranchiserManagerDashboardProps {
  user: {
    id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    role?: string;
    territoryName?: string;
    region?: string;
  };
  title?: string;
}

const FranchiserManagerDashboard: React.FC<FranchiserManagerDashboardProps> = ({ user, title }) => {
  const dispatch = useDispatch();
  const router = useRouter();
  const { 
    activeTab, 
    summary, 
    isLoading: storeLoading,
    summaryModalOpen,
    setActiveTab, 
    setSummary, 
    setLoading, 
    setLastUpdated,
    setSummaryModalOpen,
    setManager,
  } = useTerritoryManagerStore();

  const [currentTime, setCurrentTime] = useState(dayjs().format('HH:mm'));
  const [currentDate, setCurrentDate] = useState(dayjs().format('D MMMM YYYY, dddd'));
  const [alerts, setAlerts] = useState(5);

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(dayjs().format('HH:mm'));
      setCurrentDate(dayjs().format('D MMMM YYYY, dddd'));
    }, 60000);

    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    if (user) {
      setManager({
        id: user.id,
        name: `${user.first_name || ''} ${user.last_name || ''}`.trim(),
        territoryName: user.territoryName || 'Москва и область',
        region: user.region || 'Центральный федеральный округ',
      });
    }
  }, [user, setManager]);

  useEffect(() => {
    const fetchSummary = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem('accessToken');
        const res = await fetch('/api/v1/territory/summary', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const data = await res.json();
          setSummary({
            planCompletionPercent: data.planCompletionPercent || 0,
            quarterForecastPercent: data.quarterForecastPercent || 0,
            redZoneDealersCount: data.redZoneDealersCount || 0,
            avgConversion: data.avgConversion || 0,
            activeAlerts: data.activeAlerts || 0,
          });
          setAlerts(data.activeAlerts || 0);
        }
      } catch (e) {
        setSummary({
          planCompletionPercent: 82,
          quarterForecastPercent: 91,
          redZoneDealersCount: 2,
          avgConversion: 4.2,
          activeAlerts: alerts,
        });
      } finally {
        setLoading(false);
        setLastUpdated(new Date());
      }
    };

    fetchSummary();
    const updateInterval = setInterval(fetchSummary, 3 * 60 * 1000);
    return () => clearInterval(updateInterval);
  }, [setSummary, setLoading, setLastUpdated, alerts]);

  const handleLogout = () => {
    dispatch(logout());
    router.push('/login');
  };

  const userMenu = {
    items: [
      { key: 'profile', icon: <UserOutlined />, label: 'Профиль' },
      { key: 'settings', icon: <SettingOutlined />, label: 'Настройки' },
      { type: 'divider' as const },
      { key: 'logout', icon: <LogoutOutlined />, label: 'Выйти', onClick: handleLogout },
    ],
  };

  const tabItems = [
    { key: 'map', label: 'Карта территории' },
    { key: 'funnel', label: 'Воронка' },
    { key: 'planfact', label: 'План-факт и Прогноз' },
    { key: 'communications', label: 'Коммуникации и Задачи' },
    { key: 'benchmark', label: 'Бенчмаркинг' },
  ];

  const renderTabContent = () => {
    return (
      <Suspense fallback={
        <div style={{ textAlign: 'center', padding: 50 }}>
          <Spin size="large" />
        </div>
      }>
        {activeTab === 'map' && <TerritoryMapTab />}
        {activeTab === 'funnel' && <TerritoryFunnelTab />}
        {activeTab === 'planfact' && <TerritoryPlanFactTab />}
        {activeTab === 'communications' && <TerritoryCommunicationsTab />}
        {activeTab === 'benchmark' && <TerritoryBenchmarkTab />}
      </Suspense>
    );
  };

  const getStatusColor = (percent: number) => {
    if (percent >= 80) return '#52c41a';
    if (percent >= 50) return '#fa8c16';
    return '#ff4d4f';
  };

  const getAlertsList = () => [
    { id: '1', title: 'Дилер "Мебель Москва" - план 68%', desc: 'Отставание от плана на 12%' },
    { id: '2', title: 'Салон "Диванит Воронеж" - конверсия 1.8%', desc: 'Ниже нормы' },
    { id: '3', title: 'Новый дилер "МебельЛига"', desc: 'Требует утверждения' },
  ];

  return (
    <Layout style={{ minHeight: '100vh', background: '#f5f5f5' }}>
      <AntHeader style={{ 
        background: '#fff', 
        padding: '0 24px', 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'space-between', 
        borderBottom: '1px solid #f0f0f0',
        position: 'sticky',
        top: 0,
        zIndex: 100,
        height: 'auto',
        lineHeight: 'normal',
        flexWrap: 'wrap',
        gap: 12,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div>
            <Title level={4} style={{ margin: 0, color: '#1890ff' }}>
              {title || `${user.first_name || ''} ${user.last_name || ''}`.trim()}
            </Title>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {user.territoryName || 'Москва и область'} • {currentDate} {currentTime}
            </Text>
          </div>
        </div>

        <Row gutter={[12, 8]} align="middle" style={{ flex: 1, margin: '0 24px', minWidth: 600 }}>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 8 }}>
              <Statistic 
                title="Выполнение плана" 
                value={summary?.planCompletionPercent || 0} 
                precision={0}
                suffix="%"
                valueStyle={{ fontSize: 16, color: getStatusColor(summary?.planCompletionPercent || 0) }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 8 }}>
              <Statistic 
                title="Прогноз квартала" 
                value={summary?.quarterForecastPercent || 0} 
                precision={0}
                suffix="%"
                prefix={<RiseOutlined />}
                valueStyle={{ fontSize: 16, color: getStatusColor(summary?.quarterForecastPercent || 0) }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 8 }}>
              <Statistic 
                title="Дилеров в красной" 
                value={summary?.redZoneDealersCount || 0} 
                valueStyle={{ fontSize: 16, color: summary?.redZoneDealersCount ? '#ff4d4f' : '#52c41a' }}
                prefix={<WarningOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 8 }}>
              <Statistic 
                title="Средняя конверсия" 
                value={summary?.avgConversion || 0} 
                precision={1}
                suffix="%"
                valueStyle={{ fontSize: 16 }}
              />
            </Card>
          </Col>
        </Row>

        <Space size="middle">
          <Button 
            icon={<MenuOutlined />} 
            onClick={() => setSummaryModalOpen(true)}
          >
            Сводка
          </Button>
          <Badge count={alerts} offset={[-5, 5]}>
            <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
          </Badge>
          <Dropdown menu={userMenu} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1890ff' }} />
            </div>
          </Dropdown>
        </Space>
      </AntHeader>

      <Content style={{ padding: 0 }}>
        <div style={{ 
          background: '#fff', 
          borderBottom: '1px solid #f0f0f0',
          padding: '0 24px',
          overflowX: 'auto',
        }}>
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            items={tabItems}
            style={{ marginBottom: -1 }}
          />
        </div>

        <div style={{ padding: 24 }}>
          {renderTabContent()}
        </div>
      </Content>

      <Modal
        title="Утренняя сводка"
        open={summaryModalOpen}
        onCancel={() => setSummaryModalOpen(false)}
        footer={null}
        width={600}
      >
        <List
          dataSource={getAlertsList()}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta
                title={item.title}
                description={item.desc}
              />
            </List.Item>
          )}
        />
        <Divider />
        <Space direction="vertical" style={{ width: '100%' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>Выполнение плана:</Text>
            <Text strong>{summary?.planCompletionPercent || 0}%</Text>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>Прогноз квартала:</Text>
            <Text strong>{summary?.quarterForecastPercent || 0}%</Text>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>Дилеров в красной зоне:</Text>
            <Text strong style={{ color: summary?.redZoneDealersCount ? '#ff4d4f' : '#52c41a' }}>
              {summary?.redZoneDealersCount || 0}
            </Text>
          </div>
        </Space>
      </Modal>
    </Layout>
  );
};

export default FranchiserManagerDashboard;
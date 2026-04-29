// src/components/Dashboard/DealerDashboardNew.tsx
import React, { useEffect, Suspense, lazy, useState } from 'react';
import { Layout, Row, Col, Card, Typography, Tabs, Badge, Avatar, Dropdown, Space, Spin, Statistic } from 'antd';
import { 
  UserOutlined, 
  LogoutOutlined, 
  SettingOutlined, 
  BellOutlined,
  DollarOutlined,
  RiseOutlined,
  PercentageOutlined,
  BankOutlined,
  MenuOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import { useDispatch } from 'react-redux';
import { logout } from '@/store/authSlice';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import { useDealerDashboardStore } from '@/store/dealerDashboardStore';

dayjs.locale('ru');

const { Header: AntHeader } = Layout;
const { Content } = Layout;
const { Title, Text } = Typography;

const ProfitTab = lazy(() => import('./tabs/ProfitTab'));
const FunnelPlanTab = lazy(() => import('./tabs/FunnelPlanTab'));
const ProductsStockTab = lazy(() => import('./tabs/ProductsStockTab'));
const CommunicationsTab = lazy(() => import('./tabs/CommunicationsTab'));

interface UserData {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role?: string;
}

interface DealerDashboardNewProps {
  user: UserData;
  title?: string;
}

const DealerDashboardNew: React.FC<DealerDashboardNewProps> = ({ user, title }) => {
  const router = useRouter();
  const dispatch = useDispatch();
  const { 
    activeTab, 
    summary, 
    dealerCenterName, 
    isLoading: storeLoading,
    setActiveTab, 
    setSummary, 
    setLoading, 
    setLastUpdated 
  } = useDealerDashboardStore();

  const [currentDate, setCurrentDate] = useState(dayjs().format('D MMMM YYYY, dddd'));
  const [alerts, setAlerts] = useState(3);

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentDate(dayjs().format('D MMMM YYYY, dddd'));
    }, 60000);

    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    const fetchSummary = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem('accessToken');
        const res = await fetch('/api/v1/dealer/summary', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const data = await res.json();
          setSummary({
            netProfit: data.netProfit || 0,
            grossRevenue: data.grossRevenue || 0,
            planCompletionPercent: data.planCompletionPercent || 0,
            marginProfit: data.marginProfit || 0,
            activeAlerts: data.activeAlerts || 0,
          });
          setAlerts(data.activeAlerts || 0);
        }
      } catch (e) {
        console.error('Summary fetch error', e);
setSummary({
            netProfit: 1250000,
            grossRevenue: 4500000,
            planCompletionPercent: 78,
            marginProfit: 890000,
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
  }, [setSummary, setLoading, setLastUpdated]);

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
    { key: 'profit', label: 'Прибыль' },
    { key: 'funnel', label: 'Воронка и План' },
    { key: 'products', label: 'Товары и Склад' },
    { key: 'communications', label: 'Коммуникации' },
  ];

  const renderTabContent = () => {
    const tabPane = tabItems.find(t => t.key === activeTab);
    if (!tabPane) return null;

    return (
      <Suspense fallback={
        <div style={{ textAlign: 'center', padding: 50 }}>
          <Spin size="large" />
        </div>
      }>
        {activeTab === 'profit' && <ProfitTab />}
        {activeTab === 'funnel' && <FunnelPlanTab />}
        {activeTab === 'products' && <ProductsStockTab />}
        {activeTab === 'communications' && <CommunicationsTab />}
      </Suspense>
    );
  };

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
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
          <Title level={4} style={{ margin: 0, color: '#1890ff' }}>
            {title || dealerCenterName}
          </Title>
          <Text type="secondary">{currentDate}</Text>
        </div>

        <Row gutter={[16, 8]} align="middle" style={{ flex: 1, margin: '0 24px' }}>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic 
                title="Чистая прибыль" 
                value={summary?.netProfit || 0} 
                precision={0}
                prefix="₽ "
                valueStyle={{ fontSize: 16 }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic 
                title="Валовый оборот" 
                value={summary?.grossRevenue || 0} 
                precision={0}
                prefix="₽ "
                valueStyle={{ fontSize: 16 }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic 
                title="% плана сети" 
                value={summary?.planCompletionPercent || 0} 
                precision={0}
                suffix="%"
                prefix={<PercentageOutlined />}
                valueStyle={{ fontSize: 16, color: (summary?.planCompletionPercent || 0) >= 80 ? '#52c41a' : '#fa8c16' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6} lg={3}>
            <Card size="small" bodyStyle={{ padding: 12 }}>
              <Statistic 
                title="Маржинальная прибыль" 
                value={summary?.marginProfit || 0} 
                precision={0}
                prefix="₽ "
                valueStyle={{ fontSize: 16 }}
              />
            </Card>
          </Col>
        </Row>

        <Space size="middle">
          <Badge count={alerts} offset={[-5, 5]}>
            <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
          </Badge>
          <Dropdown menu={userMenu} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1890ff' }} />
              <Text style={{ marginLeft: 8, display: 'none' }} className="sm:block">
                {user.first_name || user.email}
              </Text>
            </div>
          </Dropdown>
        </Space>
      </AntHeader>

      <Content style={{ padding: '0' }}>
        <div style={{ 
          background: '#fff', 
          borderBottom: '1px solid #f0f0f0',
          padding: '0 24px',
          overflowX: 'auto',
        }}>
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            items={[
              { key: 'profit', label: '💰 Прибыль' },
              { key: 'funnel', label: '📊 Воронка и План' },
              { key: 'products', label: '🛋️ Товары и Склад' },
              { key: 'communications', label: '📞 Коммуникации' },
            ]}
            style={{ marginBottom: -1 }}
          />
        </div>

        <div style={{ padding: 24 }}>
          {renderTabContent()}
        </div>
      </Content>
    </Layout>
  );
};

export default DealerDashboardNew;
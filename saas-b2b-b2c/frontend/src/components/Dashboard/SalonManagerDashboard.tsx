import React, { useState, useEffect, lazy, Suspense } from 'react';
import { Layout, Row, Col, Card, Typography, Badge, Tabs, Spin, Statistic } from 'antd';
import { ShopOutlined, DollarOutlined, PercentageOutlined, GiftOutlined, BellOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import apiClient from '@/api/axiosClient';
import NotificationBell from './NotificationBell';
import DealerDirectives from './DealerDirectives';

dayjs.locale('ru');

const { Header, Content } = Layout;
const { Title, Text } = Typography;

interface TopBarData {
  salon_name: string;
  plan_percent: number;
  avg_check: number;
  prepayments_sum: number;
  alerts_count: number;
}

interface SalonManagerDashboardProps {
  user: any;
  title?: string;
}

const MainTab = lazy(() => import('./tabs/SalonMainTab'));
const FunnelTab = lazy(() => import('./tabs/SalonFunnelTab'));
const TeamTab = lazy(() => import('./tabs/SalonTeamTab'));
const ProductsTab = lazy(() => import('./tabs/SalonProductsTab'));

const SalonManagerDashboard: React.FC<SalonManagerDashboardProps> = ({ user, title }) => {
  const [activeTab, setActiveTab] = useState('main');
  const [topBarData, setTopBarData] = useState<TopBarData | null>(null);
  const [loadingTopBar, setLoadingTopBar] = useState(true);

  const fetchTopBarData = async () => {
    try {
      const res = await apiClient.get('/salon-manager/top-bar');
      setTopBarData(res.data);
    } catch (e) {
      console.error('Error fetching top bar data', e);
      setTopBarData({
        salon_name: user.salon_name || 'Мой салон',
        plan_percent: 0,
        avg_check: 0,
        prepayments_sum: 0,
        alerts_count: 0,
      });
    } finally {
      setLoadingTopBar(false);
    }
  };

  useEffect(() => {
    fetchTopBarData();
    const interval = setInterval(fetchTopBarData, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  const getPlanColor = (percent: number) => {
    if (percent >= 80) return '#52c41a';
    if (percent >= 50) return '#faad14';
    return '#ff4d4f';
  };

  const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);

  const tabItems = [
    { key: 'main', label: 'Главная' },
    { key: 'funnel', label: 'Воронка' },
    { key: 'team', label: 'Команда' },
    { key: 'products', label: 'Товары' },
  ];

  const renderTabContent = () => {
    return (
      <Suspense fallback={<div style={{ textAlign: 'center', padding: 40 }}><Spin size="large" /></div>}>
        {activeTab === 'main' && <MainTab user={user} />}
        {activeTab === 'funnel' && <FunnelTab user={user} />}
        {activeTab === 'team' && <TeamTab user={user} />}
        {activeTab === 'products' && <ProductsTab user={user} />}
      </Suspense>
    );
  };

  return (
    <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
      <Header style={{
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
          <div>
            <Text strong style={{ fontSize: 18, color: '#1890ff' }}>
              <ShopOutlined style={{ marginRight: 8 }} />
              {topBarData?.salon_name || 'Загрузка...'}
            </Text>
          </div>
          <div style={{ borderLeft: '1px solid #d9d9d9', paddingLeft: 24 }}>
            <Text type="secondary">{dayjs().format('dddd, D MMMM YYYY')}</Text>
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <Card size="small" style={{ width: 160, textAlign: 'center', margin: 0 }}>
            <PercentageOutlined style={{ fontSize: 20, color: getPlanColor(topBarData?.plan_percent || 0) }} />
            <div style={{ marginTop: 4 }}>
              <Text strong style={{ color: getPlanColor(topBarData?.plan_percent || 0) }}>
                {topBarData?.plan_percent || 0}%
              </Text>
            </div>
            <Text type="secondary" style={{ fontSize: 12 }}>Выполнение плана</Text>
          </Card>

          <Card size="small" style={{ width: 140, textAlign: 'center', margin: 0 }}>
            <DollarOutlined style={{ fontSize: 20, color: '#1890ff' }} />
            <div style={{ marginTop: 4 }}>
              <Text strong>{formatMoney(topBarData?.avg_check || 0)}</Text>
            </div>
            <Text type="secondary" style={{ fontSize: 12 }}>Средний чек</Text>
          </Card>

          <Card size="small" style={{ width: 160, textAlign: 'center', margin: 0 }}>
            <GiftOutlined style={{ fontSize: 20, color: '#722ed1' }} />
            <div style={{ marginTop: 4 }}>
              <Text strong>{formatMoney(topBarData?.prepayments_sum || 0)}</Text>
            </div>
            <Text type="secondary" style={{ fontSize: 12 }}>Предоплата</Text>
          </Card>

          <DealerDirectives user={user} />
          <NotificationBell />
        </div>
      </Header>

      <Content style={{ padding: '16px 24px' }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ marginBottom: 16 }}
        />
        {renderTabContent()}
      </Content>
    </Layout>
  );
};

export default SalonManagerDashboard;
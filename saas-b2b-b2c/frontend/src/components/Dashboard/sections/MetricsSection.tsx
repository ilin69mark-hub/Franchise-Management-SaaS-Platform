import React from 'react';
import { Card, Row, Col, Statistic, Typography } from 'antd';
import { DollarOutlined, RiseOutlined, UserOutlined, ShopOutlined, CheckCircleOutlined, FallOutlined } from '@ant-design/icons';
import { useSuperAdminStore } from '@/store/superAdminStore';

const { Title } = Typography;

const MetricsSection: React.FC = () => {
  const { stats, isLoading, lastUpdated } = useSuperAdminStore();

  return (
    <div>
      <Title level={3}>SaaS-метрики</Title>
      <Row gutter={16}>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic
              title="MRR (Monthly Recurring Revenue)"
              value={stats.mrr}
              prefix={<DollarOutlined />}
              suffix="₽"
              precision={0}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic
              title="ARR (Annual Recurring Revenue)"
              value={stats.arr}
              prefix={<DollarOutlined />}
              suffix="₽"
              precision={0}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic
              title="Churn Rate"
              value={stats.churnRate}
              suffix="%"
              precision={1}
              valueStyle={{ color: stats.churnRate > 10 ? '#ff4d4f' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic
              title="ARPU (Average Revenue Per User)"
              value={stats.arpu}
              prefix={<UserOutlined />}
              suffix="₽"
              precision={0}
            />
          </Card>
        </Col>
      </Row>
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic
              title="Active Tenants"
              value={24}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic
              title="New This Month"
              value={5}
              prefix={<RiseOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic
              title="Churned This Month"
              value={2}
              prefix={<FallOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>
      {lastUpdated && (
        <div style={{ marginTop: 16, color: '#888' }}>
          Последнее обновление: {lastUpdated.toLocaleTimeString()}
        </div>
      )}
    </div>
  );
};

export default MetricsSection;
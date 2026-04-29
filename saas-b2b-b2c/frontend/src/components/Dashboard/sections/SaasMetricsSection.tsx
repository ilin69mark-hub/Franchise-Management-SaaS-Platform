import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Typography, Spin, message } from 'antd';
import {DollarOutlined, RiseOutlined, FallOutlined, WarningOutlined, UserOutlined} from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import { useSaasMetricsStore } from '@/store/saasMetricsStore';

const { Title, Text } = Typography;

const SaasMetricsSection: React.FC = () => {
  const {
    overview,
    mrrDynamics,
    cashflow,
    tariffs,
    isLoading,
    fetchAnalytics,
    fetchPaymentStatus,
  } = useSaasMetricsStore();

  const [localLoading, setLocalLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLocalLoading(true);
    try {
      const [analyticsRes, paymentsRes, plansRes] = await Promise.all([
        apiClient.get('/admin/analytics').catch(() => ({ data: {} })),
        apiClient.get('/admin/tenants/payments').catch(() => ({ data: [] })),
        apiClient.get('/admin/plans').catch(() => ({ data: [] })),
      ]);

      const analytics = analyticsRes.data as Record<string, unknown> || {};
      const payments = paymentsRes.data as Array<Record<string, unknown>> || [];
const plans = plansRes.data as Array<Record<string, unknown>> || [];

      if (tariffs.length === 0 && plans.length > 0) {
        const tariffData = plans.map((plan: Record<string, unknown>) => ({
          name: (plan.name as string) || '',
          count: (plan.tenant_count as number) || 0,
          mrr: (plan.price as number) || 0,
          avgCheck: (plan.price as number) || 0,
          growth: 0,
        }));
      }

      fetchAnalytics();
      fetchPaymentStatus();
    } catch {
      fetchAnalytics();
} finally {
      setLocalLoading(false);
    }
  };

  if (localLoading || isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  const renderChange = (value: number, isPositive: boolean) => {
    const color = value > 0 ? '#52c41a' : value < 0 ? '#ff4d4f' : '#888';
    const Icon = isPositive ? RiseOutlined : FallOutlined;
    const prefix = isPositive ? '+' : '';
    return (
      <span style={{ color, fontSize: 12, marginLeft: 8 }}>
        <Icon /> {prefix}{value}%
      </span>
    );
  };

  return (
    <div>
      <Title level={3}>SaaS-метрики</Title>
      
      <Title level={4}>Ключевые финансовые показатели</Title>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <div style={{ marginBottom: 8 }}>
              <Text type="secondary">MRR (ежемесячный доход)</Text>
            </div>
            <div style={{ fontSize: 24, fontWeight: 'bold' }}>
              {(overview?.mrr || 0).toLocaleString()} ₽
            </div>
            {renderChange(overview?.mrrChange || 0, (overview?.mrrChange || 0) >= 0)}
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <div style={{ marginBottom: 8 }}>
              <Text type="secondary">ARR (годовой доход)</Text>
            </div>
            <div style={{ fontSize: 24, fontWeight: 'bold' }}>
              {(overview?.arr || 0).toLocaleString()} ₽
            </div>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <div style={{ marginBottom: 8 }}>
              <Text type="secondary">Churn Rate (отток)</Text>
            </div>
            <div style={{ fontSize: 24, fontWeight: 'bold', color: (overview?.churnRate || 0) > 10 ? '#ff4d4f' : undefined }}>
              {overview?.churnRate || 0}%
            </div>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <div style={{ marginBottom: 8 }}>
              <Text type="secondary">Просроч. платежи</Text>
            </div>
            <div style={{ fontSize: 24, fontWeight: 'bold', color: (overview?.overdueCount || 0) > 0 ? '#ff4d4f' : undefined }}>
              {(overview?.overdueAmount || 0).toLocaleString()} ₽
            </div>
            <Text type="secondary"> ({overview?.overdueCount || 0} должников)</Text>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <div style={{ marginBottom: 8 }}>
              <Text type="secondary">ARPU (средний чек)</Text>
            </div>
            <div style={{ fontSize: 24, fontWeight: 'bold' }}>
              {(overview?.arpu || 0).toLocaleString()} ₽
            </div>
          </Card>
        </Col>
      </Row>
      
      <Title level={4} style={{ marginTop: 24 }}>Динамика MRR (12 месяцев)</Title>
      <Card>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={{ textAlign: 'left', padding: 8 }}>Месяц</th>
                <th style={{ textAlign: 'right', padding: 8, color: '#1890ff' }}>Новый MRR</th>
                <th style={{ textAlign: 'right', padding: 8, color: '#52c41a' }}>Expansion</th>
                <th style={{ textAlign: 'right', padding: 8, color: '#ff4d4f' }}>Churn</th>
                <th style={{ textAlign: 'right', padding: 8, fontWeight: 'bold' }}>Net MRR</th>
              </tr>
            </thead>
            <tbody>
              {mrrDynamics.slice(-12).map((item) => (
                <tr key={item.month}>
                  <td style={{ padding: 8 }}>{item.month}</td>
                  <td style={{ textAlign: 'right', padding: 8, color: '#1890ff' }}>
                    +{item.newMrr.toLocaleString()} ₽
                  </td>
                  <td style={{ textAlign: 'right', padding: 8, color: '#52c41a' }}>
                    +{item.expansionMrr.toLocaleString()} ₽
                  </td>
                  <td style={{ textAlign: 'right', padding: 8, color: '#ff4d4f' }}>
                    -{item.churnMrr.toLocaleString()} ₽
                  </td>
                  <td style={{ textAlign: 'right', padding: 8, fontWeight: 'bold' }}>
                    {item.netMrr.toLocaleString()} ₽
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div style={{ marginTop: 16, display: 'flex', gap: 16 }}>
          <div><span style={{ color: '#1890ff' }}>▄</span> Новый MRR</div>
          <div><span style={{ color: '#52c41a' }}>▄</span> Expansion</div>
          <div><span style={{ color: '#ff4d4f' }}>▄</span> Churn</div>
          <div><span style={{ color: '#000' }}>▄</span> Net</div>
        </div>
      </Card>
      
      <Title level={4} style={{ marginTop: 24 }}>Прогноз Cash Flow</Title>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="Ожидаемые автоплатежи">
            {cashflow?.expected?.map((item, idx) => (
              <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0', borderBottom: '1px solid #f0f0f0' }}>
                <span>{item.date}</span>
                <span>{item.tenant}</span>
                <span style={{ fontWeight: 'bold' }}>{item.amount.toLocaleString()} ₽</span>
              </div>
            ))}
            <div style={{ marginTop: 8, fontWeight: 'bold', borderTop: '1px solid #f0f0f0', paddingTop: 8 }}>
              Итого: {cashflow?.totalExpected?.toLocaleString() || 0} ₽
            </div>
          </Card>
        </Col>
        
        <Col xs={24} lg={12}>
          <Card title="Просроченные счета" extra={<WarningOutlined style={{ color: '#ff4d4f' }} />}>
            {cashflow?.overdue?.map((item, idx) => (
              <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0', borderBottom: '1px solid #f0f0f0', color: '#ff4d4f' }}>
                <span>{item.date} ({item.daysOverdue} дн.)</span>
                <span>{item.tenant}</span>
                <span style={{ fontWeight: 'bold' }}>{item.amount.toLocaleString()} ₽</span>
              </div>
            ))}
            <div style={{ marginTop: 8, fontWeight: 'bold', borderTop: '1px solid #f0f0f0', paddingTop: 8, color: '#ff4d4f' }}>
              Итого под угрозой: {cashflow?.atRisk?.toLocaleString() || 0} ₽
            </div>
          </Card>
        </Col>
      </Row>
      
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card title="Риски (истекающие договоры)">
            {cashflow?.risks?.length > 0 ? (
              cashflow.risks.map((item, idx) => (
                <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0', borderBottom: '1px solid #f0f0f0' }}>
                  <span>{item.date}</span>
                  <span>{item.tenant}</span>
                  <span style={{ fontWeight: 'bold' }}>{item.amount.toLocaleString()} ₽</span>
                </div>
              ))
            ) : (
              <Text type="secondary">Нет рисков</Text>
            )}
          </Card>
        </Col>
      </Row>
      
      <Title level={4} style={{ marginTop: 24 }}>Тепловая карта тарифов</Title>
      <Card>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={{ textAlign: 'left', padding: 8 }}>Тариф</th>
              <th style={{ textAlign: 'right', padding: 8 }}>Тенантов</th>
              <th style={{ textAlign: 'right', padding: 8 }}>Суммарный MRR</th>
              <th style={{ textAlign: 'right', padding: 8 }}>Средний чек</th>
              <th style={{ textAlign: 'right', padding: 8 }}>Доля</th>
              <th style={{ textAlign: 'right', padding: 8 }}>Динамика</th>
            </tr>
          </thead>
          <tbody>
            {tariffs.map((item) => {
              const totalMrr = tariffs.reduce((acc, t) => acc + t.mrr, 0);
              const share = totalMrr > 0 ? ((item.mrr / totalMrr) * 100).toFixed(1) : '0';
              const growthColor = item.growth > 0 ? '#52c41a' : item.growth < 0 ? '#ff4d4f' : '#888';
              return (
                <tr key={item.name}>
                  <td style={{ padding: 8, fontWeight: 'bold' }}>{item.name}</td>
                  <td style={{ textAlign: 'right', padding: 8 }}>{item.count}</td>
                  <td style={{ textAlign: 'right', padding: 8 }}>{item.mrr.toLocaleString()} ₽</td>
                  <td style={{ textAlign: 'right', padding: 8 }}>{item.avgCheck.toLocaleString()} ₽</td>
                  <td style={{ textAlign: 'right', padding: 8 }}>{share}%</td>
                  <td style={{ textAlign: 'right', padding: 8, color: growthColor }}>
                    {item.growth > 0 ? '+' : ''}{item.growth}%
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>
    </div>
  );
};

export default SaasMetricsSection;
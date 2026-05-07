import React, { useState, useEffect, useCallback } from 'react';
import { Row, Col, Card, Typography, Progress, Table, Tag, Spin, Alert, Tooltip } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined, WarningOutlined, CheckCircleOutlined, ClockCircleOutlined } from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

interface SalonMainTabProps {
  user: any;
}

interface DashboardMainData {
  plan: number;
  fact: number;
  plan_percent: number;
  dynamic_day: number;
  dynamic_week: number;
  forecast: number;
  avg_check: number;
  margin_percent: number;
  prepayments_sum: number;
  traffic_status: string;
  conversion_measure_status: string;
  conversion_contract_status: string;
  pending_payments: PendingPayment[];
}

interface PendingPayment {
  contract_id: string;
  client_name: string;
  amount: number;
  status: string;
  payment_date: string;
}

const SalonMainTab: React.FC<SalonMainTabProps> = ({ user }) => {
  const [data, setData] = useState<DashboardMainData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      const date = dayjs().format('YYYY-MM-DD');
      const res = await apiClient.get(`/dashboard/main?date=${date}`);
      setData(res.data);
      setError(null);
    } catch (e: any) {
      setError(e?.response?.data?.error || 'Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);

  const getDynamicColor = (value: number) => {
    if (value > 0) return '#52c41a';
    if (value < 0) return '#ff4d4f';
    return '#999';
  };

  const getDynamicIcon = (value: number) => {
    if (value > 0) return <ArrowUpOutlined />;
    if (value < 0) return <ArrowDownOutlined />;
    return null;
  };

  const getTrafficLightColor = (status: string) => {
    switch (status) {
      case 'green': return '#52c41a';
      case 'yellow': return '#faad14';
      case 'red': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  const TrafficLight: React.FC<{ status: string; label: string }> = ({ status, label }) => (
    <Tooltip title={label}>
      <div
        style={{
          width: 16,
          height: 16,
          borderRadius: '50%',
          backgroundColor: getTrafficLightColor(status),
          display: 'inline-block',
          marginRight: 8,
        }}
      />
    </Tooltip>
  );

  const KpiCard: React.FC<{
    title: string;
    value: string | number;
    suffix?: string;
    dynamic?: number;
    color?: string;
  }> = ({ title, value, suffix = '', dynamic, color }) => (
    <Card size="small" style={{ borderRadius: 8, height: '100%' }}>
      <Text type="secondary" style={{ fontSize: 12 }}>{title}</Text>
      <div style={{ marginTop: 4 }}>
        <Text strong style={{ fontSize: 20, color: color || '#000' }}>
          {typeof value === 'number' && title.includes('чек') ? formatMoney(value) : value}{suffix}
        </Text>
      </div>
      {dynamic !== undefined && dynamic !== 0 && (
        <div style={{ marginTop: 4 }}>
          <Text style={{ color: getDynamicColor(dynamic), fontSize: 12 }}>
            {getDynamicIcon(dynamic)} {Math.abs(dynamic).toFixed(1)}%
          </Text>
        </div>
      )}
    </Card>
  );

  const pendingColumns = [
    {
      title: 'Клиент',
      dataIndex: 'client_name',
      key: 'client_name',
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (val: number) => <Text strong>{formatMoney(val)} ₽</Text>,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusMap: Record<string, { color: string; label: string }> = {
          awaiting_payment: { color: 'orange', label: 'Ожидает оплаты' },
          payment_due: { color: 'red', label: 'Срок оплаты' },
        };
        const config = statusMap[status] || { color: 'default', label: status };
        return <Tag color={config.color}>{config.label}</Tag>;
      },
    },
  ];

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 40 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert
        message="Ошибка"
        description={error}
        type="error"
        showIcon
        action={
          <a onClick={fetchData}>Повторить</a>
        }
      />
    );
  }

  const planColor = data?.plan_percent !== undefined
    ? (data.plan_percent >= 80 ? '#52c41a' : data.plan_percent >= 50 ? '#faad14' : '#ff4d4f')
    : '#1890ff';

  return (
    <div>
      <Row gutter={[16, 16]}>
        {/* Основные метрики */}
        <Col xs={24} lg={16}>
          <Card title="Продажи за месяц" style={{ borderRadius: 12 }}>
            <Row gutter={16}>
              <Col xs={24} md={8}>
                <KpiCard
                  title="План"
                  value={formatMoney(data?.plan || 0)}
                  suffix=" ₽"
                />
              </Col>
              <Col xs={24} md={8}>
                <KpiCard
                  title="Факт"
                  value={formatMoney(data?.fact || 0)}
                  suffix=" ₽"
                  color="#1890ff"
                />
              </Col>
              <Col xs={24} md={8}>
                <KpiCard
                  title="Средний чек"
                  value={data?.avg_check || 0}
                  suffix=" ₽"
                />
              </Col>
            </Row>

            <div style={{ marginTop: 24, marginBottom: 16 }}>
              <Row justify="space-between" align="middle">
                <Col>
                  <Text strong>Выполнение плана: {data?.plan_percent || 0}%</Text>
                </Col>
                <Col>
                  <Text type="secondary">
                    Прогноз на конец месяца: {data?.forecast || 0}%
                  </Text>
                </Col>
              </Row>
            </div>

            <Progress
              percent={data?.plan_percent || 0}
              strokeColor={planColor}
              trailColor="#f0f0f0"
              showInfo={false}
              style={{ marginBottom: 8 }}
            />

            {/* Подсветка если факт < 70% плана */}
            {data?.plan_percent !== undefined && data.plan_percent < 70 && (
              <Alert
                message="Внимание: план выполнен менее чем на 70%"
                type="warning"
                showIcon
                icon={<WarningOutlined />}
                style={{ marginTop: 16 }}
              />
            )}
          </Card>

          {/* Динамика */}
          <Card title="Динамика" style={{ marginTop: 16, borderRadius: 12 }}>
            <Row gutter={16}>
              <Col xs={12}>
                <Card size="small" style={{ textAlign: 'center', background: '#fafafa' }}>
                  <Text type="secondary">К вчерашнему дню</Text>
                  <div style={{ marginTop: 8 }}>
                    <Text strong style={{ color: getDynamicColor(data?.dynamic_day || 0), fontSize: 24 }}>
                      {getDynamicIcon(data?.dynamic_day || 0)}
                      {Math.abs(data?.dynamic_day || 0).toFixed(1)}%
                    </Text>
                  </div>
                </Card>
              </Col>
              <Col xs={12}>
                <Card size="small" style={{ textAlign: 'center', background: '#fafafa' }}>
                  <Text type="secondary">К прошлой неделе</Text>
                  <div style={{ marginTop: 8 }}>
                    <Text strong style={{ color: getDynamicColor(data?.dynamic_week || 0), fontSize: 24 }}>
                      {getDynamicIcon(data?.dynamic_week || 0)}
                      {Math.abs(data?.dynamic_week || 0).toFixed(1)}%
                    </Text>
                  </div>
                </Card>
              </Col>
            </Row>
          </Card>

          {/* Ожидаемые оплаты */}
          <Card
            title="Сегодня должны оплатить"
            style={{ marginTop: 16, borderRadius: 12 }}
          >
            {data?.pending_payments && data.pending_payments.length > 0 ? (
              <Table
                dataSource={data.pending_payments}
                columns={pendingColumns}
                rowKey="contract_id"
                pagination={false}
                size="small"
              />
            ) : (
              <Text type="secondary">На сегодня оплат не запланировано</Text>
            )}
          </Card>
        </Col>

        {/* Боковая панель */}
        <Col xs={24} lg={8}>
          {/* Светофор */}
          <Card title="Светофор" style={{ borderRadius: 12 }}>
            <div style={{ marginBottom: 16 }}>
              <Text type="secondary">Входящий трафик</Text>
              <div style={{ marginTop: 8 }}>
                <TrafficLight status={data?.traffic_status || 'red'} label="Трафик" />
                <Text>
                  {data?.traffic_status === 'green' ? 'Норма' :
                   data?.traffic_status === 'yellow' ? 'Ниже нормы' : 'Критично'}
                </Text>
              </div>
            </div>

            <div style={{ marginBottom: 16 }}>
              <Text type="secondary">Конверсия в замеры</Text>
              <div style={{ marginTop: 8 }}>
                <TrafficLight status={data?.conversion_measure_status || 'red'} label="Замеры" />
                <Text>
                  {data?.conversion_measure_status === 'green' ? 'Хорошо' :
                   data?.conversion_measure_status === 'yellow' ? 'Средне' : 'Плохо'}
                </Text>
              </div>
            </div>

            <div>
              <Text type="secondary">Конверсия в договоры</Text>
              <div style={{ marginTop: 8 }}>
                <TrafficLight status={data?.conversion_contract_status || 'red'} label="Договоры" />
                <Text>
                  {data?.conversion_contract_status === 'green' ? 'Хорошо' :
                   data?.conversion_contract_status === 'yellow' ? 'Средне' : 'Плохо'}
                </Text>
              </div>
            </div>
          </Card>

          {/* Дополнительные метрики */}
          <Card title="Дополнительно" style={{ marginTop: 16, borderRadius: 12 }}>
            <div style={{ marginBottom: 12 }}>
              <Text type="secondary">Маржинальность</Text>
              <div>
                <Text strong style={{ fontSize: 18, color: '#52c41a' }}>
                  {(data?.margin_percent || 0).toFixed(1)}%
                </Text>
              </div>
            </div>

            <div>
              <Text type="secondary">Предоплаты за месяц</Text>
              <div>
                <Text strong style={{ fontSize: 18, color: '#722ed1' }}>
                  {formatMoney(data?.prepayments_sum || 0)} ₽
                </Text>
              </div>
            </div>
          </Card>

          {/* Прогноз */}
          <Card
            title="Прогноз выполнения"
            style={{ marginTop: 16, borderRadius: 12 }}
          >
            <Progress
              percent={data?.forecast || 0}
              strokeColor={
                (data?.forecast || 0) >= 100 ? '#52c41a' :
                (data?.forecast || 0) >= 70 ? '#1890ff' : '#faad14'
              }
              format={(percent) => `${percent}%`}
            />
            <Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
              Ожидаемое выполнение плана к концу месяца
            </Text>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default SalonMainTab;
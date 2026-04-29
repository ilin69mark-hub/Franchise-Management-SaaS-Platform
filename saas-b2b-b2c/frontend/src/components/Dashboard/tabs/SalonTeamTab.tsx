import React, { useState, useEffect, useCallback } from 'react';
import { Row, Col, Card, Typography, Table, Tag, Spin, Progress, Avatar, Button, Segmented, Modal, Tooltip } from 'antd';
import { UserOutlined, TeamOutlined, ArrowUpOutlined, ArrowDownOutlined, LineChartOutlined } from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

interface SalonTeamTabProps {
  user: any;
}

interface SalesRepMetrics {
  user_id: string;
  first_name: string;
  last_name: string;
  role: string;
  revenue: number;
  deals_count: number;
  conversion: number;
  avg_check: number;
  discount_percent: number;
  extras_sum: number;
  revenue_deviation: number;
  deals_deviation: number;
  conversion_deviation: number;
  avg_check_deviation: number;
}

interface SalesRepHistory {
  month: string;
  revenue: number;
  deals: number;
  avg_check: number;
}

interface DashboardTeamData {
  period: string;
  total_revenue: number;
  avg_revenue: number;
  avg_conversion: number;
  avg_check: number;
  sales_reps: SalesRepMetrics[];
}

const SalonTeamTab: React.FC<SalonTeamTabProps> = ({ user }) => {
  const [data, setData] = useState<DashboardTeamData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [period, setPeriod] = useState<string>('month');
  const [selectedRep, setSelectedRep] = useState<SalesRepMetrics | null>(null);
  const [history, setHistory] = useState<SalesRepHistory[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const date = dayjs().format('YYYY-MM-DD');
      const res = await apiClient.get(`/dashboard/team?period=${period}&date=${date}`);
      setData(res.data);
      setError(null);
    } catch (e: any) {
      setError(e?.response?.data?.error || 'Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  }, [period]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const fetchHistory = async (managerId: string) => {
    setHistoryLoading(true);
    try {
      const res = await apiClient.get(`/dashboard/team/${managerId}/history?months=6`);
      setHistory(res.data);
    } catch (e) {
      setHistory([]);
    } finally {
      setHistoryLoading(false);
    }
  };

  const handleRowClick = (record: SalesRepMetrics) => {
    setSelectedRep(record);
    fetchHistory(record.user_id);
  };

  const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);

  const DeviationIndicator: React.FC<{ value: number }> = ({ value }) => {
    if (value === 0) return null;
    const isPositive = value > 0;
    return (
      <Tooltip title={isPositive ? `Выше среднего на ${Math.abs(value).toFixed(1)}%` : `Ниже среднего на ${Math.abs(value).toFixed(1)}%`}>
        <span style={{ color: isPositive ? '#52c41a' : '#ff4d4f', marginLeft: 4, fontSize: 12 }}>
          {isPositive ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
          {Math.abs(value).toFixed(0)}%
        </span>
      </Tooltip>
    );
  };

  const columns = [
    {
      title: 'Продавец',
      key: 'name',
      sorter: (a: SalesRepMetrics, b: SalesRepMetrics) => a.first_name.localeCompare(b.first_name),
      render: (_: any, record: SalesRepMetrics) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <Avatar icon={<UserOutlined />} style={{ backgroundColor: record.user_id === user.id ? '#52c41a' : '#1890ff' }} />
          <div>
            <Text strong>
              {record.first_name} {record.last_name}
              {record.user_id === user.id && <Tag color="green" style={{ marginLeft: 8 }}>Вы</Tag>}
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: 'Выручка',
      dataIndex: 'revenue',
      key: 'revenue',
      sorter: (a: SalesRepMetrics, b: SalesRepMetrics) => a.revenue - b.revenue,
      render: (val: number, record: SalesRepMetrics) => (
        <Tooltip title={`Среднее по салону: ${formatMoney(data?.avg_revenue || 0)} ₽`}>
          <div>
            <Text strong style={{ fontSize: 16 }}>{formatMoney(val)} ₽</Text>
            <DeviationIndicator value={record.revenue_deviation} />
          </div>
        </Tooltip>
      ),
    },
    {
      title: 'Чеков',
      dataIndex: 'deals_count',
      key: 'deals_count',
      sorter: (a: SalesRepMetrics, b: SalesRepMetrics) => a.deals_count - b.deals_count,
      render: (val: number, record: SalesRepMetrics) => (
        <Tooltip title={`Среднее: ${Math.round((data?.total_revenue || 0) / (data?.sales_reps?.length || 1) / ((data?.avg_check) || 1))} шт`}>
          <div>
            <Text strong>{val}</Text>
            <DeviationIndicator value={record.deals_deviation} />
          </div>
        </Tooltip>
      ),
    },
    {
      title: 'Конверсия',
      dataIndex: 'conversion',
      key: 'conversion',
      sorter: (a: SalesRepMetrics, b: SalesRepMetrics) => a.conversion - b.conversion,
      render: (val: number, record: SalesRepMetrics) => (
        <Tooltip title={`Среднее по салону: ${(data?.avg_conversion || 0).toFixed(1)}%`}>
          <div>
            <Progress
              percent={Math.round(val)}
              size="small"
              strokeColor={val >= (data?.avg_conversion || 0) ? '#52c41a' : '#faad14'}
              style={{ width: 80 }}
            />
            <DeviationIndicator value={record.conversion_deviation} />
          </div>
        </Tooltip>
      ),
    },
    {
      title: 'Средний чек',
      dataIndex: 'avg_check',
      key: 'avg_check',
      sorter: (a: SalesRepMetrics, b: SalesRepMetrics) => a.avg_check - b.avg_check,
      render: (val: number, record: SalesRepMetrics) => (
        <Tooltip title={`Среднее по салону: ${formatMoney(data?.avg_check || 0)} ₽`}>
          <div>
            <Text>{formatMoney(val)} ₽</Text>
            <DeviationIndicator value={record.avg_check_deviation} />
          </div>
        </Tooltip>
      ),
    },
    {
      title: 'Скидка',
      dataIndex: 'discount_percent',
      key: 'discount_percent',
      sorter: (a: SalesRepMetrics, b: SalesRepMetrics) => a.discount_percent - b.discount_percent,
      render: (val: number) => (
        <Tag color={val > 10 ? 'error' : val > 5 ? 'warning' : 'success'}>
          {val.toFixed(1)}%
        </Tag>
      ),
    },
    {
      title: 'Допы',
      dataIndex: 'extras_sum',
      key: 'extras_sum',
      sorter: (a: SalesRepMetrics, b: SalesRepMetrics) => a.extras_sum - b.extras_sum,
      render: (val: number) => (
        <Tooltip title="Услуги и аксессуары">
          <Text style={{ color: '#722ed1' }}>{formatMoney(val)} ₽</Text>
        </Tooltip>
      ),
    },
    {
      title: 'Детали',
      key: 'action',
      render: (_: any, record: SalesRepMetrics) => (
        <Button
          size="small"
          icon={<LineChartOutlined />}
          onClick={(e) => {
            e.stopPropagation();
            handleRowClick(record);
          }}
        >
          График
        </Button>
      ),
    },
  ];

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 40 }}><Spin size="large" /></div>;
  }

  if (error) {
    return (
      <Card>
        <Text type="danger">{error}</Text>
        <Button onClick={fetchData} style={{ marginLeft: 16 }}>Повторить</Button>
      </Card>
    );
  }

  const avgPerformance = data?.sales_reps?.length
    ? Math.round(data.sales_reps.reduce((sum, r) => sum + r.conversion, 0) / data.sales_reps.length)
    : 0;

  return (
    <div>
      {/* Фильтры */}
      <Card style={{ marginBottom: 16, borderRadius: 12 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Segmented
              options={[
                { label: 'Неделя', value: 'week' },
                { label: 'Месяц', value: 'month' },
                { label: 'Квартал', value: 'quarter' },
              ]}
              value={period}
              onChange={(val) => setPeriod(val as string)}
            />
          </Col>
          <Col>
            <Text type="secondary">Сортировка по умолчанию: выручка (убывание)</Text>
          </Col>
        </Row>
      </Card>

      {/* Статистика */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} md={6}>
          <Card style={{ borderRadius: 12, textAlign: 'center' }}>
            <Title level={3} style={{ margin: 0 }}><TeamOutlined /> {data?.sales_reps?.length || 0}</Title>
            <Text type="secondary">Сотрудников</Text>
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card style={{ borderRadius: 12, textAlign: 'center' }}>
            <Title level={3} style={{ margin: 0 }}>{formatMoney(data?.total_revenue || 0)} ₽</Title>
            <Text type="secondary">Общая выручка</Text>
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card style={{ borderRadius: 12, textAlign: 'center' }}>
            <Title level={3} style={{ margin: 0 }}>{(data?.avg_conversion || 0).toFixed(1)}%</Title>
            <Text type="secondary">Средняя конверсия</Text>
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card style={{ borderRadius: 12, textAlign: 'center' }}>
            <Title level={3} style={{ margin: 0, color: avgPerformance >= 30 ? '#52c41a' : '#faad14' }}>
              {formatMoney(data?.avg_check || 0)} ₽
            </Title>
            <Text type="secondary">Средний чек</Text>
          </Card>
        </Col>
      </Row>

      {/* Таблица команды */}
      <Card title="Рейтинг продавцов" style={{ borderRadius: 12 }}>
        <Table
          dataSource={data?.sales_reps}
          columns={columns}
          rowKey="user_id"
          pagination={false}
          onRow={(record) => ({
            onClick: () => handleRowClick(record),
            style: record.user_id === user.id ? { background: '#f6ffed' } : {},
          })}
          rowClassName={() => 'cursor-pointer'}
        />
      </Card>

      {/* Модальное окно с графиком */}
      <Modal
        title={
          selectedRep
            ? `Динамика: ${selectedRep.first_name} ${selectedRep.last_name}`
            : 'Детали продавца'
        }
        open={!!selectedRep}
        onCancel={() => {
          setSelectedRep(null);
          setHistory([]);
        }}
        footer={null}
        width={700}
      >
        {selectedRep && (
          <div>
            <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
              <Col span={8}>
                <Card size="small">
                  <Text type="secondary">Выручка (6 мес)</Text>
                  <Title level={4} style={{ margin: '8px 0 0' }}>
                    {formatMoney(history.reduce((sum, h) => sum + h.revenue, 0))} ₽
                  </Title>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small">
                  <Text type="secondary">Сделок</Text>
                  <Title level={4} style={{ margin: '8px 0 0' }}>
                    {history.reduce((sum, h) => sum + h.deals, 0)}
                  </Title>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small">
                  <Text type="secondary">Ср. чек</Text>
                  <Title level={4} style={{ margin: '8px 0 0' }}>
                    {formatMoney(history.reduce((sum, h) => sum + h.avg_check, 0) / (history.length || 1))} ₽
                  </Title>
                </Card>
              </Col>
            </Row>

            {historyLoading ? (
              <Spin />
            ) : (
              <Table
                dataSource={history}
                pagination={false}
                size="small"
                columns={[
                  {
                    title: 'Месяц',
                    dataIndex: 'month',
                    key: 'month',
                    render: (m: string) => dayjs(m + '-01').format('MMMM YYYY'),
                  },
                  {
                    title: 'Выручка',
                    dataIndex: 'revenue',
                    key: 'revenue',
                    render: (v: number) => formatMoney(v) + ' ₽',
                  },
                  {
                    title: 'Сделок',
                    dataIndex: 'deals',
                    key: 'deals',
                  },
                  {
                    title: 'Ср. чек',
                    dataIndex: 'avg_check',
                    key: 'avg_check',
                    render: (v: number) => formatMoney(v) + ' ₽',
                  },
                ]}
              />
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default SalonTeamTab;
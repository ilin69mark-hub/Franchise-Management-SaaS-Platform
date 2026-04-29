import React, { useState, useEffect, useCallback } from 'react';
import { Row, Col, Card, Typography, Table, Tag, Spin, Segmented, Progress, Tooltip, Alert } from 'antd';
import { ShoppingOutlined, WarningOutlined, BarChartOutlined } from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

interface SalonProductsTabProps {
  user: any;
}

interface TopProduct {
  id: string;
  name: string;
  collection: string;
  category: string;
  revenue: number;
  quantity: number;
  share_percent: number;
  margin: number;
}

interface StockItem {
  id: string;
  name: string;
  category: string;
  showroom_qty: number;
  warehouse_qty: number;
  total_cost: number;
  turnover_days: number;
}

interface LostSale {
  reason: string;
  requests_count: number;
  lost_revenue: number;
}

interface CategoryTurnover {
  category: string;
  avg_days: number;
  is_slow_moving: boolean;
}

interface DashboardProductsData {
  top_products: TopProduct[];
  stock_items: StockItem[];
  lost_sales: LostSale[];
  category_turnover: CategoryTurnover[];
  total_revenue: number;
}

const SalonProductsTab: React.FC<SalonProductsTabProps> = ({ user }) => {
  const [data, setData] = useState<DashboardProductsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<string>('revenue');

  const fetchData = useCallback(async () => {
    try {
      const date = dayjs().format('YYYY-MM-DD');
      const res = await apiClient.get(`/dashboard/products?date=${date}`);
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

  const sortedTopProducts = React.useMemo(() => {
    if (!data?.top_products) return [];
    const sorted = [...data.top_products];
    switch (sortBy) {
      case 'revenue':
        return sorted.sort((a, b) => b.revenue - a.revenue);
      case 'quantity':
        return sorted.sort((a, b) => b.quantity - a.quantity);
      case 'margin':
        return sorted.sort((a, b) => b.margin - a.margin);
      default:
        return sorted;
    }
  }, [data?.top_products, sortBy]);

  // === КОЛОНКИ ТОП-10 ===
  const topProductsColumns = [
    {
      title: '№',
      key: 'index',
      width: 50,
      render: (_: any, __: any, index: number) => <Text strong>{index + 1}</Text>,
    },
    {
      title: 'Модель',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: 'Коллекция',
      dataIndex: 'collection',
      key: 'collection',
    },
    {
      title: 'Выручка',
      dataIndex: 'revenue',
      key: 'revenue',
      render: (val: number) => <Text strong>{formatMoney(val)} ₽</Text>,
    },
    {
      title: 'Продано',
      dataIndex: 'quantity',
      key: 'quantity',
      render: (val: number) => <Tag color="blue">{val} шт</Tag>,
    },
    {
      title: 'Доля',
      dataIndex: 'share_percent',
      key: 'share_percent',
      render: (val: number) => (
        <div style={{ width: 100 }}>
          <Progress percent={Math.round(val)} size="small" strokeColor="#1890ff" />
          <Text type="secondary" style={{ fontSize: 11 }}>{val.toFixed(1)}%</Text>
        </div>
      ),
    },
    {
      title: 'Маржа',
      dataIndex: 'margin',
      key: 'margin',
      render: (val: number) => (
        <Tag color={val >= 30 ? 'success' : val >= 20 ? 'warning' : 'error'}>
          {val.toFixed(1)}%
        </Tag>
      ),
    },
  ];

  // === КОЛОНКИ ОСТАТКОВ ===
  const stockColumns = [
    {
      title: 'Модель',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: StockItem) => (
        <div>
          <Text strong>{text}</Text>
          <br />
          <Text type="secondary" style={{ fontSize: 12 }}>{record.category}</Text>
        </div>
      ),
    },
    {
      title: 'Витрина',
      dataIndex: 'showroom_qty',
      key: 'showroom_qty',
      render: (val: number) => <Tag color="blue">{val} шт</Tag>,
    },
    {
      title: 'Склад',
      dataIndex: 'warehouse_qty',
      key: 'warehouse_qty',
      render: (val: number) => <Tag>{val} шт</Tag>,
    },
    {
      title: 'Стоимость',
      dataIndex: 'total_cost',
      key: 'total_cost',
      render: (val: number) => <Text>{formatMoney(val)} ₽</Text>,
    },
    {
      title: 'Оборачиваемость',
      dataIndex: 'turnover_days',
      key: 'turnover_days',
      render: (days: number) => (
        <Tooltip title={days > 90 ? 'Неликвидный товар (>90 дней)' : 'Норма'}>
          <Tag color={days > 90 ? 'error' : days > 60 ? 'warning' : 'success'}>
            {days} дн.
          </Tag>
        </Tooltip>
      ),
    },
  ];

  // === КОЛОНКИ УПУЩЕННЫХ ПРОДАЖ ===
  const lostSalesColumns = [
    {
      title: 'Причина',
      dataIndex: 'reason',
      key: 'reason',
      render: (reason: string, record: LostSale) => (
        <div>
          <Text strong>{reason}</Text>
          {record.lost_revenue > 100000 && (
            <WarningOutlined style={{ color: '#ff4d4f', marginLeft: 8 }} />
          )}
        </div>
      ),
    },
    {
      title: 'Запросов',
      dataIndex: 'requests_count',
      key: 'requests_count',
      render: (val: number) => <Tag color="orange">{val}</Tag>,
    },
    {
      title: 'Упущенная выручка',
      dataIndex: 'lost_revenue',
      key: 'lost_revenue',
      render: (val: number) => (
        <Tooltip title={val > 100000 ? 'Критичная сумма' : ''}>
          <Text strong style={{ color: val > 100000 ? '#ff4d4f' : undefined }}>
            {formatMoney(val)} ₽
          </Text>
        </Tooltip>
      ),
    },
  ];

  // === КОЛОНКИ ОБОРАЧИВАЕМОСТИ ===
  const turnoverColumns = [
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
      render: (cat: string, record: CategoryTurnover) => (
        <div>
          <Text strong>{cat}</Text>
          {record.is_slow_moving && (
            <Tooltip title="Неликвид - требует внимания">
              <WarningOutlined style={{ color: '#ff4d4f', marginLeft: 8 }} />
            </Tooltip>
          )}
        </div>
      ),
    },
    {
      title: 'Средняя оборачиваемость',
      dataIndex: 'avg_days',
      key: 'avg_days',
      render: (days: number, record: CategoryTurnover) => (
        <Progress
          percent={(days / 120) * 100}
          strokeColor={record.is_slow_moving ? '#ff4d4f' : days > 60 ? '#faad14' : '#52c41a'}
          format={() => `${days} дн.`}
        />
      ),
    },
  ];

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 40 }}><Spin size="large" /></div>;
  }

  if (error) {
    return (
      <Alert
        message="Ошибка"
        description={error}
        type="error"
        showIcon
        action={<a onClick={fetchData}>Повторить</a>}
      />
    );
  }

  return (
    <div>
      {/* Фильтры */}
      <Card style={{ marginBottom: 16, borderRadius: 12 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Segmented
              options={[
                { label: 'По выручке', value: 'revenue' },
                { label: 'По количеству', value: 'quantity' },
                { label: 'По маржинальности', value: 'margin' },
              ]}
              value={sortBy}
              onChange={(val) => setSortBy(val as string)}
            />
          </Col>
          <Col>
            <Text type="secondary">
              Общая выручка: <Text strong>{formatMoney(data?.total_revenue || 0)} ₽</Text>
            </Text>
          </Col>
        </Row>
      </Card>

      <Row gutter={[16, 16]}>
        {/* Топ-10 товаров */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <span>
                <ShoppingOutlined style={{ marginRight: 8 }} />
                Топ-10 продаваемых моделей
              </span>
            }
            style={{ borderRadius: 12 }}
          >
            <Table
              dataSource={sortedTopProducts}
              columns={topProductsColumns}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>

        {/* Остатки */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <span>
                <BarChartOutlined style={{ marginRight: 8 }} />
                Остатки товаров
              </span>
            }
            style={{ borderRadius: 12 }}
          >
            <Table
              dataSource={data?.stock_items}
              columns={stockColumns}
              rowKey="id"
              pagination={false}
              size="small"
              rowClassName={(record) => record.turnover_days > 90 ? 'bg-red-light' : ''}
            />
          </Card>
        </Col>

        {/* Упущенные продажи */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <span>
                <WarningOutlined style={{ marginRight: 8, color: '#ff4d4f' }} />
                Упущенные продажи
              </span>
            }
            style={{ borderRadius: 12 }}
          >
            <Table
              dataSource={data?.lost_sales}
              columns={lostSalesColumns}
              rowKey="reason"
              pagination={false}
              size="small"
              rowClassName={(record) => record.lost_revenue > 100000 ? 'bg-red-light' : ''}
            />
          </Card>
        </Col>

        {/* Оборачиваемость */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <span>
                <BarChartOutlined style={{ marginRight: 8 }} />
                Оборачиваемость по категориям
              </span>
            }
            style={{ borderRadius: 12 }}
          >
            <Table
              dataSource={data?.category_turnover}
              columns={turnoverColumns}
              rowKey="category"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default SalonProductsTab;
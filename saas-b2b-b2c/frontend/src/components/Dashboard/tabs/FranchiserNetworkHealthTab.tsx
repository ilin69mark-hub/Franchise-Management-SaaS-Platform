import React, { useState, useEffect, Suspense } from 'react';
import { Row, Col, Card, Table, Tag, Typography, Space, Progress, Statistic, Select, Empty, Tooltip, Badge, Divider, Modal, message } from 'antd';
import { 
  ArrowUpOutlined, 
  ArrowDownOutlined, 
  ShopOutlined, 
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  TeamOutlined,
  GlobalOutlined,
  DollarOutlined,
  RiseOutlined,
  FallOutlined,
  StarOutlined,
} from '@ant-design/icons';
import dynamic from 'next/dynamic';

const { Title, Text } = Typography;
const { Option } = Select;

const GeographyMap = dynamic(() => import('./GeographyMap'), { ssr: false });

interface SegmentationData {
  segment: 'A' | 'B' | 'C' | 'D';
  name: string;
  count: number;
  revenueShare: number;
  dynamics: number;
}

interface MigrationData {
  from: string;
  to: string;
  count: number;
}

interface SystemIssue {
  id: string;
  description: string;
  affectedDealers: number;
  lostRevenue: number;
  status: 'identified' | 'in_progress' | 'resolved';
  owner: string;
}

interface HealthMetric {
  name: string;
  value: number;
  unit: string;
  dynamics: number;
  target?: number;
}

interface WhiteSpot {
  city: string;
  population: number;
  potential: number;
}

interface MarketingData {
  withCentral: { avgRevenue: number; avgTraffic: number; conversion: number };
  withoutCentral: { avgRevenue: number; avgTraffic: number; conversion: number };
  roi: number;
}

const mockSegmentation: SegmentationData[] = [
  { segment: 'A', name: 'A-дилеры (лидеры)', count: 4, revenueShare: 42, dynamics: 2 },
  { segment: 'B', name: 'B-дилеры (стабильные)', count: 12, revenueShare: 38, dynamics: -1 },
  { segment: 'C', name: 'C-дилеры (отстающие)', count: 6, revenueShare: 15, dynamics: -2 },
  { segment: 'D', name: 'D-дилеры (кандидаты)', count: 2, revenueShare: 5, dynamics: 1 },
];

const mockMigration: MigrationData[] = [
  { from: 'C', to: 'B', count: 2 },
  { from: 'B', to: 'A', count: 1 },
  { from: 'B', to: 'C', count: 1 },
  { from: 'A', to: 'B', count: 0 },
];

const mockIssues: SystemIssue[] = [
  { id: '1', description: 'Падение конверсии из-за отсутствия выставочных образцов новой коллекции', affectedDealers: 5, lostRevenue: 2500000, status: 'in_progress', owner: 'Алексей Петров' },
  { id: '2', description: 'Срыв сроков поставок по кухонной группе', affectedDealers: 3, lostRevenue: 1800000, status: 'identified', owner: 'Мария Иванова' },
  { id: '3', description: 'Жалобы на качество фасадов коллекции X', affectedDealers: 7, lostRevenue: 900000, status: 'resolved', owner: 'Сергей Сидоров' },
];

const mockMetrics: HealthMetric[] = [
  { name: 'Churn Rate', value: 4.2, unit: '%', dynamics: -0.5, target: 5 },
  { name: 'NPS дилеров', value: 72, unit: 'баллов', dynamics: 5, target: 70 },
  { name: 'Срок жизни дилера', value: 38, unit: 'мес.', dynamics: 2, target: 36 },
  { name: 'Доля новых дилеров', value: 15, unit: '%', dynamics: 3, target: 20 },
  { name: 'Покрытие матрицы', value: 82, unit: '%', dynamics: -2, target: 90 },
  { name: 'Плотность покрытия', value: 0.8, unit: '/100k', dynamics: 0.1, target: 1 },
];

const mockWhiteSpots: WhiteSpot[] = [
  { city: 'Воронеж', population: 1050000, potential: 15 },
  { city: 'Краснодар', population: 918000, potential: 12 },
  { city: 'Казань', population: 1250000, potential: 18 },
  { city: 'Сочи', population: 412000, potential: 6 },
];

const mockMarketing: MarketingData = {
  withCentral: { avgRevenue: 5200000, avgTraffic: 180, conversion: 14 },
  withoutCentral: { avgRevenue: 3800000, avgTraffic: 120, conversion: 9 },
  roi: 2.4,
};

const getSegmentColor = (segment: string) => {
  const colors: Record<string, string> = { A: '#52c41a', B: '#1890ff', C: '#fa8c16', D: '#ff4d4f' };
  return colors[segment] || '#d9d9d9';
};

const FranchiserNetworkHealthTab: React.FC = () => {
  const [period, setPeriod] = useState('quarter');
  const [segmentation] = useState<SegmentationData[]>(mockSegmentation);
  const [migration] = useState<MigrationData[]>(mockMigration);
  const [issues] = useState<SystemIssue[]>(mockIssues);
  const [metrics] = useState<HealthMetric[]>(mockMetrics);
  const [whiteSpots] = useState<WhiteSpot[]>(mockWhiteSpots);
  const [marketing] = useState<MarketingData>(mockMarketing);

  const totalDealers = segmentation.reduce((s, sgm) => s + sgm.count, 0);

  const totalRevenue = segmentation.reduce((s, sgm) => s + sgm.revenueShare, 0);

  const segmentColumns = [
    {
      title: 'Сегмент',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, r: SegmentationData) => (
        <Tag color={getSegmentColor(r.segment)}>{name}</Tag>
      ),
    },
    {
      title: 'Количество',
      dataIndex: 'count',
      key: 'count',
      render: (v: number) => <Text strong>{v}</Text>,
    },
    {
      title: 'Доля выручки',
      dataIndex: 'revenueShare',
      key: 'revenueShare',
      render: (v: number) => `${v}%`,
    },
    {
      title: 'Динамика',
      dataIndex: 'dynamics',
      key: 'dynamics',
      render: (v: number) => (
        <Text type={v > 0 ? 'success' : v < 0 ? 'danger' : 'secondary'}>
          {v > 0 ? <ArrowUpOutlined /> : v < 0 ? <ArrowDownOutlined /> : '-'} {Math.abs(v)}
        </Text>
      ),
    },
    {
      title: 'Визуализация',
      key: 'chart',
      render: (_: any, r: SegmentationData) => (
        <Progress 
          percent={Math.round((r.count / totalDealers) * 100)} 
          size="small" 
          strokeColor={getSegmentColor(r.segment)}
          style={{ width: 100 }}
        />
      ),
    },
  ];

  const migrationColumns = [
    { title: 'Из', dataIndex: 'from', key: 'from', render: (s: string) => <Tag>{s}</Tag> },
    { title: 'В', dataIndex: 'to', key: 'to', render: (s: string) => <Tag>{s}</Tag> },
    { title: 'Кол-во', dataIndex: 'count', key: 'count' },
  ];

  const issueColumns = [
    {
      title: 'Проблема',
      dataIndex: 'description',
      key: 'description',
      render: (d: string) => <Text>{d}</Text>,
    },
    {
      title: 'Дилера',
      dataIndex: 'affectedDealers',
      key: 'affectedDealers',
      render: (v: number) => <Badge count={v} />,
    },
    {
      title: 'Потеряно',
      dataIndex: 'lostRevenue',
      key: 'lostRevenue',
      render: (v: number) => <Text type="danger">{v.toLocaleString()} ₽</Text>,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => (
        <Tag 
          color={s === 'resolved' ? 'green' : s === 'in_progress' ? 'processing' : 'orange'}
          icon={s === 'resolved' ? <CheckCircleOutlined /> : s === 'in_progress' ? <WarningOutlined /> : <CloseCircleOutlined />}
        >
          {s === 'identified' ? 'Выявлено' : s === 'in_progress' ? 'Решается' : 'Решено'}
        </Tag>
      ),
    },
    {
      title: 'Ответственный',
      dataIndex: 'owner',
      key: 'owner',
    },
  ];

  const whiteSpotColumns = [
    { title: 'Город', dataIndex: 'city', key: 'city' },
    { title: 'Население', dataIndex: 'population', key: 'population', render: (v: number) => v.toLocaleString() },
    { title: 'Потенциал', dataIndex: 'potential', key: 'potential', render: (v: number) => `~${v} дилеров` },
  ];

  return (
    <div>
      <Title level={4}>Здоровье сети</Title>

      <Space style={{ marginBottom: 16 }}>
        <Select value={period} onChange={setPeriod} style={{ width: 150 }}>
          <Option value="month">Месяц</Option>
          <Option value="quarter">Квартал</Option>
          <Option value="year">Год</Option>
        </Select>
      </Space>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {segmentation.map((sgm) => (
          <Col span={6} key={sgm.segment}>
            <Card size="small">
              <Statistic
                title={sgm.name}
                value={sgm.count}
                suffix={`(${sgm.revenueShare}%)`}
                valueStyle={{ color: getSegmentColor(sgm.segment), fontSize: 24 }}
                prefix={<ShopOutlined />}
              />
              <Text type={sgm.dynamics > 0 ? 'success' : sgm.dynamics < 0 ? 'danger' : 'secondary'}>
                {sgm.dynamics > 0 ? '+' : ''}{sgm.dynamics}% к прошлому периоду
              </Text>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={12}>
          <Card title="Сегментация">
            <Table
              dataSource={segmentation}
              columns={segmentColumns}
              rowKey="segment"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="Миграция дилеров (квартал)">
            <Table
              dataSource={migration}
              columns={migrationColumns}
              rowKey={r => `${r.from}-${r.to}`}
              pagination={false}
              size="small"
            />
            <Divider />
            <Text type="secondary">Матрица переходов между сегментами</Text>
          </Card>
        </Col>
      </Row>

      <Card title="Системные проблемы" style={{ marginBottom: 24 }}>
        <Table
          dataSource={issues}
          columns={issueColumns}
          rowKey="id"
          pagination={false}
          size="small"
        />
      </Card>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {metrics.map((metric) => (
          <Col span={4} key={metric.name}>
            <Card size="small">
              <Statistic
                title={metric.name}
                value={metric.value}
                suffix={metric.unit}
                valueStyle={{ fontSize: 18 }}
              />
              <Space>
                <Text type={metric.dynamics > 0 ? 'success' : metric.dynamics < 0 ? 'danger' : 'secondary'}>
                  {metric.dynamics > 0 ? <ArrowUpOutlined /> : <ArrowDownOutlined />} {Math.abs(metric.dynamics)}
                </Text>
                {metric.target && <Text type="secondary">/ целевое {metric.target}</Text>}
              </Space>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={16}>
          <Card title="География присутствия">
            <GeographyMap />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="Белые пятна">
            <Table
              dataSource={whiteSpots}
              columns={whiteSpotColumns}
              rowKey="city"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      <Card title="Маркетинговая эффективность">
        <Row gutter={24}>
          <Col span={8}>
            <Statistic
              title="С централизованным маркетингом"
              value={marketing.withCentral.avgRevenue}
              suffix="₽"
              valueStyle={{ fontSize: 18 }}
            />
            <Text>Трафик: {marketing.withCentral.avgTraffic}/мес.</Text>
            <br />
            <Text>Конверсия: {marketing.withCentral.conversion}%</Text>
          </Col>
          <Col span={8}>
            <Statistic
              title="Без централизованного маркетинга"
              value={marketing.withoutCentral.avgRevenue}
              suffix="₽"
              valueStyle={{ fontSize: 18 }}
            />
            <Text>Трафик: {marketing.withoutCentral.avgTraffic}/мес.</Text>
            <br />
            <Text>Конверсия: {marketing.withoutCentral.conversion}%</Text>
          </Col>
          <Col span={8}>
            <Card>
              <Statistic
                title="ROI маркетинга"
                value={marketing.roi}
                suffix="x"
                valueStyle={{ fontSize: 24, color: marketing.roi > 2 ? '#52c41a' : '#fa8c16' }}
              />
              <Text type="secondary">Окупаемость бюджета</Text>
            </Card>
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default FranchiserNetworkHealthTab;
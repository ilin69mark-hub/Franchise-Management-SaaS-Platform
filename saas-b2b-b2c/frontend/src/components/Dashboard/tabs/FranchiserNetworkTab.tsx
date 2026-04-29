import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Table, Progress, Tag, Typography, Select, Space, Button, Tooltip } from 'antd';
import { 
  ArrowUpOutlined, 
  ArrowDownOutlined, 
  ShopOutlined, 
  WarningOutlined,
  CheckCircleOutlined,
  TeamOutlined,
  RiseOutlined,
  FallOutlined,
} from '@ant-design/icons';
import { FranchiserSummary, DealerMetrics } from '@/store/franchiserStore';
import dynamic from 'next/dynamic';

const { Title, Text } = Typography;
const { Option } = Select;

const SalesChart = dynamic(() => import('./SalesDynamicsChart'), { ssr: false });
const GeographyMap = dynamic(() => import('./GeographyMap'), { ssr: false });

interface FranchiserNetworkTabProps {
  summary: FranchiserSummary;
}

interface TerritoryData {
  id: string;
  manager: string;
  territory: string;
  planPercent: number;
  dealersCount: number;
  redDealersPercent: number;
  sla: number;
  dealerGrowth: number;
  forecastPercent: number;
}

interface NetworkOverview {
  planAmount: number;
  planPercent: number;
  planDynamics: number;
  forecastAmount: number;
  forecastPercent: number;
  activeDealers: number;
  dealerGrowth: number;
  avgConversion: number;
  conversionBenchmark: number;
  avgMargin: number;
  redZoneDealers: number;
  redZonePercent: number;
  redZoneDynamics: number;
}

const mockTerritories: TerritoryData[] = [
  { id: '1', manager: 'Алексей Петров', territory: 'Север', planPercent: 92, dealersCount: 8, redDealersPercent: 0, sla: 98, dealerGrowth: 2, forecastPercent: 95 },
  { id: '2', manager: 'Мария Иванова', territory: 'Юг', planPercent: 78, dealersCount: 6, redDealersPercent: 17, sla: 85, dealerGrowth: 1, forecastPercent: 82 },
  { id: '3', manager: 'Сергей Сидоров', territory: 'Запад', planPercent: 65, dealersCount: 10, redDealersPercent: 30, sla: 72, dealerGrowth: -1, forecastPercent: 70 },
  { id: '4', manager: 'Елена Смирнова', territory: 'Центр', planPercent: 88, dealersCount: 5, redDealersPercent: 0, sla: 94, dealerGrowth: 1, forecastPercent: 90 },
];

const mockOverview: NetworkOverview = {
  planAmount: 42000000,
  planPercent: 78,
  planDynamics: 5,
  forecastAmount: 35700000,
  forecastPercent: 85,
  activeDealers: 24,
  dealerGrowth: 3,
  avgConversion: 12,
  conversionBenchmark: 15,
  avgMargin: 32,
  redZoneDealers: 3,
  redZonePercent: 12.5,
  redZoneDynamics: -1,
};

const mockSalesData = [
  { month: 'Янв', plan: 3500000, fact: 3200000, forecast: null },
  { month: 'Фев', plan: 3500000, fact: 2900000, forecast: null },
  { month: 'Мар', plan: 3500000, fact: 3800000, forecast: null },
  { month: 'Апр', plan: 3500000, fact: 3400000, forecast: null },
  { month: 'Май', plan: 3500000, fact: 3100000, forecast: null },
  { month: 'Июн', plan: 3500000, fact: null, forecast: 3300000 },
  { month: 'Июл', plan: 3500000, fact: null, forecast: 3500000 },
  { month: 'Авг', plan: 3500000, fact: null, forecast: 3500000 },
  { month: 'Сен', plan: 3500000, fact: null, forecast: 3500000 },
  { month: 'Окт', plan: 3500000, fact: null, forecast: 3500000 },
  { month: 'Ноя', plan: 3500000, fact: null, forecast: 3600000 },
  { month: 'Дек', plan: 3500000, fact: null, forecast: 3700000 },
];

const FranchiserNetworkTab: React.FC<FranchiserNetworkTabProps> = ({ summary }) => {
  const [period, setPeriod] = useState<'month' | 'quarter' | 'year'>('month');
  const [groupBy, setGroupBy] = useState<'manager' | 'district' | 'type'>('manager');
  const [overview] = useState<NetworkOverview>(mockOverview);
  const [territories] = useState<TerritoryData[]>(mockTerritories);
  const [selectedManager, setSelectedManager] = useState<string | null>(null);

  const getHeatmapColor = (status: 'green' | 'yellow' | 'red') => {
    return status === 'green' ? '#52c41a' : status === 'yellow' ? '#fa8c16' : '#ff4d4f';
  };

  const getTerritoryStatus = (t: TerritoryData): 'green' | 'yellow' | 'red' => {
    if (t.planPercent >= 95 && t.redDealersPercent < 10 && t.sla >= 95) return 'green';
    if (t.planPercent < 85 || t.redDealersPercent > 25 || t.sla < 80) return 'red';
    return 'yellow';
  };

  const columns = [
    {
      title: 'Менеджер',
      dataIndex: 'manager',
      key: 'manager',
      sorter: (a: TerritoryData, b: TerritoryData) => a.manager.localeCompare(b.manager),
      render: (m: string, r: TerritoryData) => (
        <Space>
          <TeamOutlined />
          <Text strong>{m}</Text>
        </Space>
      ),
    },
    {
      title: 'Территория',
      dataIndex: 'territory',
      key: 'territory',
      sorter: (a: TerritoryData, b: TerritoryData) => a.territory.localeCompare(b.territory),
    },
    {
      title: 'Выполнение плана',
      dataIndex: 'planPercent',
      key: 'planPercent',
      sorter: (a: TerritoryData, b: TerritoryData) => a.planPercent - b.planPercent,
      render: (v: number) => (
        <Tooltip title={`${v}%`}>
          <Progress 
            percent={v} 
            size="small" 
            strokeColor={v >= 95 ? '#52c41a' : v >= 85 ? '#fa8c16' : '#ff4d4f'}
            style={{ width: 80 }}
          />
        </Tooltip>
      ),
    },
    {
      title: 'Дилеров',
      dataIndex: 'dealersCount',
      key: 'dealersCount',
      sorter: (a: TerritoryData, b: TerritoryData) => a.dealersCount - b.dealersCount,
    },
    {
      title: 'В красной зоне',
      dataIndex: 'redDealersPercent',
      key: 'redDealersPercent',
      sorter: (a: TerritoryData, b: TerritoryData) => a.redDealersPercent - b.redDealersPercent,
      render: (v: number) => (
        <Tag color={v < 10 ? 'green' : v < 25 ? 'orange' : 'red'}>
          {v}%
        </Tag>
      ),
    },
    {
      title: 'SLA',
      dataIndex: 'sla',
      key: 'sla',
      sorter: (a: TerritoryData, b: TerritoryData) => a.sla - b.sla,
      render: (v: number) => (
        <Tag color={v >= 95 ? 'green' : v >= 80 ? 'orange' : 'red'}>
          {v}%
        </Tag>
      ),
    },
    {
      title: 'Прирост',
      dataIndex: 'dealerGrowth',
      key: 'dealerGrowth',
      sorter: (a: TerritoryData, b: TerritoryData) => a.dealerGrowth - b.dealerGrowth,
      render: (v: number) => (
        <Text type={v > 0 ? 'success' : v < 0 ? 'danger' : 'secondary'}>
          {v > 0 ? <ArrowUpOutlined /> : v < 0 ? <ArrowDownOutlined /> : ''} {Math.abs(v)}
        </Text>
      ),
    },
    {
      title: 'Прогноз',
      dataIndex: 'forecastPercent',
      key: 'forecastPercent',
      sorter: (a: TerritoryData, b: TerritoryData) => a.forecastPercent - b.forecastPercent,
      render: (v: number) => (
        <Tag color={v >= 95 ? 'green' : v >= 85 ? 'orange' : 'red'}>
          {v}%
        </Tag>
      ),
    },
    {
      title: 'Статус',
      key: 'status',
      render: (_: any, r: TerritoryData) => {
        const status = getTerritoryStatus(r);
        return (
          <Tag color={status} icon={<CheckCircleOutlined />}>
            {status === 'green' ? 'В норме' : status === 'yellow' ? 'Внимание' : 'Критично'}
          </Tag>
        );
      },
    },
  ];

  return (
    <div>
      <Title level={4}>Пульт сети</Title>

      <Space style={{ marginBottom: 16 }} wrap>
        <Select value={period} onChange={setPeriod} style={{ width: 120 }}>
          <Option value="month">Месяц</Option>
          <Option value="quarter">Квартал</Option>
          <Option value="year">Год</Option>
        </Select>
        <Select value={groupBy} onChange={setGroupBy} style={{ width: 180 }}>
          <Option value="manager">По менеджерам</Option>
          <Option value="district">По округам</Option>
          <Option value="type">По типу дилеров</Option>
        </Select>
      </Space>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={4}>
          <Card size="small">
            <Statistic
              title="Выполнение плана"
              value={overview.planAmount}
              suffix="₽"
              prefix={<ShopOutlined />}
              valueStyle={{ fontSize: 18 }}
            />
            <Space>
              <Text type="success">{overview.planPercent}%</Text>
              <Text type={overview.planDynamics > 0 ? 'success' : 'danger'}>
                {overview.planDynamics > 0 ? <ArrowUpOutlined /> : <ArrowDownOutlined />} {Math.abs(overview.planDynamics)}%
              </Text>
            </Space>
          </Card>
        </Col>

        <Col span={4}>
          <Card size="small">
            <Statistic
              title="Прогноз квартала"
              value={overview.forecastAmount}
              suffix="₽"
              valueStyle={{ fontSize: 18 }}
            />
            <Tag color={overview.forecastPercent > 95 ? 'green' : overview.forecastPercent >= 85 ? 'orange' : 'red'}>
              {overview.forecastPercent}%
            </Tag>
          </Card>
        </Col>

        <Col span={4}>
          <Card size="small">
            <Statistic
              title="Активных дилеров"
              value={overview.activeDealers}
              valueStyle={{ fontSize: 18 }}
            />
            <Text type={overview.dealerGrowth > 0 ? 'success' : 'danger'}>
              {overview.dealerGrowth > 0 ? '+' : ''}{overview.dealerGrowth} за период
            </Text>
          </Card>
        </Col>

        <Col span={4}>
          <Card size="small">
            <Statistic
              title="Конверсия сети"
              value={overview.avgConversion}
              suffix="%"
              valueStyle={{ fontSize: 18 }}
            />
            <Text type={overview.avgConversion >= overview.conversionBenchmark ? 'success' : 'warning'}>
              vs {overview.conversionBenchmark}% норматив
            </Text>
          </Card>
        </Col>

        <Col span={4}>
          <Card size="small">
            <Statistic
              title="Маржинальность"
              value={overview.avgMargin}
              suffix="%"
              valueStyle={{ fontSize: 18, color: overview.avgMargin >= 30 ? '#52c41a' : '#fa8c16' }}
            />
            <Text type="secondary">без OPEX</Text>
          </Card>
        </Col>

        <Col span={4}>
          <Card size="small">
            <Statistic
              title="В красной зоне"
              value={overview.redZoneDealers}
              valueStyle={{ fontSize: 18, color: '#ff4d4f' }}
              prefix={<WarningOutlined />}
            />
            <Space>
              <Text>{overview.redZonePercent}% от сети</Text>
              <Text type={overview.redZoneDynamics < 0 ? 'success' : 'danger'}>
                {overview.redZoneDynamics > 0 ? '+' : ''}{overview.redZoneDynamics}
              </Text>
            </Space>
          </Card>
        </Col>
      </Row>

      <Card title="Динамика продаж сети" style={{ marginBottom: 24 }}>
        <SalesChart data={mockSalesData} />
      </Card>

      <Card title="Теплокарта территорий">
        <Table
          dataSource={territories}
          columns={columns}
          rowKey="id"
          pagination={false}
          rowStyle={(record) => ({
            background: getHeatmapColor(getTerritoryStatus(record)) + '20',
          })}
          onRow={(record) => ({
            onClick: () => setSelectedManager(record.id),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>
    </div>
  );
};

export default FranchiserNetworkTab;
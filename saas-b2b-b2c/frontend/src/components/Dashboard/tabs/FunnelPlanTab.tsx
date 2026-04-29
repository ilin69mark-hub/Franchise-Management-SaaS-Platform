// src/components/Dashboard/tabs/FunnelPlanTab.tsx
import React, { useState, useMemo } from 'react';
import { Card, Row, Col, Progress, Typography, Table, Tag, Button, Space, Select, Radio, Collapse, Statistic, Empty, Spin } from 'antd';
import { CheckCircleOutlined, ClockCircleOutlined, SyncOutlined, TrophyOutlined, WarningOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import dayjs from 'dayjs';

const { Text, Title } = Typography;
const { Option } = Select;
const { Panel } = Collapse;

export type PeriodType = 'month' | 'quarter' | 'year';
export type GroupByType = 'salon' | 'manager';

interface SalonPlanData {
  id: string;
  name: string;
  plan: number;
  fact: number;
  percent: number;
  forecast: 'green' | 'yellow' | 'red';
  managerName: string;
  sellersCount: number;
  avgCheck: number;
}

interface FunnelStage {
  name: string;
  count: number;
  conversion: number;
}

interface NetworkFunnel {
  stages: FunnelStage[];
  traffic: number;
  consultation: number;
  measurement: number;
  kp: number;
  contract: number;
  payment: number;
}

interface BenchmarkData {
  date: string;
  dealerConversion: number;
  networkAvgConversion: number;
}

interface ManagerStats {
  id: string;
  name: string;
  salon: string;
  revenue: number;
  planPercent: number;
  conversion: number;
}

interface FunnelPlanTabProps {
  salonPlanData?: SalonPlanData[];
  networkFunnel?: NetworkFunnel;
  benchmarkData?: BenchmarkData[];
  managerStats?: ManagerStats[];
  loading?: boolean;
  period?: PeriodType;
  groupBy?: GroupByType;
}

const FunnelPlanTab: React.FC<FunnelPlanTabProps> = ({
  salonPlanData = [],
  networkFunnel,
  benchmarkData = [],
  managerStats = [],
  loading = false,
  period: initialPeriod = 'month',
  groupBy = 'salon',
}) => {
  const [period, setPeriod] = useState<PeriodType>(initialPeriod);
  const [groupByState, setGroupByState] = useState<GroupByType>(groupBy);
  const [expandedSalon, setExpandedSalon] = useState<string | null>(null);

  const totalPlan = useMemo(() => salonPlanData.reduce((sum, s) => sum + s.plan, 0), [salonPlanData]);
  const totalFact = useMemo(() => salonPlanData.reduce((sum, s) => sum + s.fact, 0), [salonPlanData]);
  const planPercent = totalPlan > 0 ? Math.round((totalFact / totalPlan) * 100) : 0;

  const stages: FunnelStage[] = useMemo(() => {
    if (!networkFunnel) return [];
    return [
      { name: 'Трафик', count: networkFunnel.traffic, conversion: 100 },
      { name: 'Консультация', count: networkFunnel.consultation, conversion: networkFunnel.traffic > 0 ? (networkFunnel.consultation / networkFunnel.traffic) * 100 : 0 },
      { name: 'Замер', count: networkFunnel.measurement, conversion: networkFunnel.consultation > 0 ? (networkFunnel.measurement / networkFunnel.consultation) * 100 : 0 },
      { name: 'КП', count: networkFunnel.kp, conversion: networkFunnel.measurement > 0 ? (networkFunnel.kp / networkFunnel.measurement) * 100 : 0 },
      { name: 'Договор', count: networkFunnel.contract, conversion: networkFunnel.kp > 0 ? (networkFunnel.kp / networkFunnel.kp) * 100 : 0 },
      { name: 'Оплата', count: networkFunnel.payment, conversion: networkFunnel.contract > 0 ? (networkFunnel.contract / networkFunnel.contract) * 100 : 0 },
    ];
  }, [networkFunnel]);

  const topManagers = useMemo(() => {
    return [...managerStats].sort((a, b) => b.revenue - a.revenue).slice(0, 3);
  }, [managerStats]);

  const antiTopManagers = useMemo(() => {
    return [...managerStats].sort((a, b) => a.conversion - b.conversion).slice(0, 3);
  }, [managerStats]);

  const getForecastColor = (forecast: 'green' | 'yellow' | 'red') => {
    const colors = { green: '#52c41a', yellow: '#faad14', red: '#ff4d4f' };
    return colors[forecast];
  };

  const getForecastIcon = (forecast: 'green' | 'yellow' | 'red') => {
    const icons = { green: '🟢', yellow: '🟡', red: '🔴' };
    return icons[forecast];
  };

  const salonColumns = [
    {
      title: 'Название салона',
      dataIndex: 'name',
      key: 'name',
      fixed: 'left' as const,
      width: 180,
    },
    {
      title: 'План (руб.)',
      dataIndex: 'plan',
      key: 'plan',
      render: (val: number) => <Text strong>{val.toLocaleString()} ₽</Text>,
    },
    {
      title: 'Факт (руб.)',
      dataIndex: 'fact',
      key: 'fact',
      render: (val: number) => <Text strong>{val.toLocaleString()} ₽</Text>,
    },
    {
      title: '% выполнения',
      dataIndex: 'percent',
      key: 'percent',
      render: (val: number) => (
        <Progress
          percent={val}
          size="small"
          strokeColor={val >= 80 ? '#52c41a' : val >= 50 ? '#faad14' : '#ff4d4f'}
        />
      ),
    },
    {
      title: 'Прогноз',
      dataIndex: 'forecast',
      key: 'forecast',
      render: (_: any, record: SalonPlanData) => (
        <span>
          {getForecastIcon(record.forecast)}{' '}
          {record.forecast === 'green' ? 'выполнит' : record.forecast === 'yellow' ? 'под вопросом' : 'не выполнит'}
        </span>
      ),
    },
  ];

  const managerColumns = [
    {
      title: 'ФИО менеджера',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: 'Салон',
      dataIndex: 'salon',
      key: 'salon',
      width: 150,
    },
    {
      title: 'Выручка',
      dataIndex: 'revenue',
      key: 'revenue',
      render: (val: number) => <Text strong>{val.toLocaleString()} ₽</Text>,
    },
    {
      title: '% плана',
      dataIndex: 'planPercent',
      key: 'planPercent',
      render: (val: number) => (
        <Tag color={val >= 80 ? 'green' : val >= 50 ? 'orange' : 'red'}>
          {val}%
        </Tag>
      ),
    },
    {
      title: 'Конверсия',
      dataIndex: 'conversion',
      key: 'conversion',
      render: (val: number) => <Text>{val.toFixed(1)}%</Text>,
    },
  ];

  const funnelStageColumns = [
    {
      title: 'Этап',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Кол-во сделок',
      dataIndex: 'count',
      key: 'count',
      render: (val: number) => <Tag color="blue">{val}</Tag>,
    },
    {
      title: 'Конверсия %',
      dataIndex: 'conversion',
      key: 'conversion',
      render: (val: number) => <Text strong>{val.toFixed(1)}%</Text>,
    },
  ];

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col xs={24} style={{ marginBottom: 16 }}>
          <Space>
            <Radio.Group
              value={period}
              onChange={(e) => setPeriod(e.target.value as PeriodType)}
              optionType="button"
              buttonStyle="solid"
            >
              <Radio.Button value="month">Месяц</Radio.Button>
              <Radio.Button value="quarter">Квартал</Radio.Button>
              <Radio.Button value="year">Год</Radio.Button>
            </Radio.Group>
            <Select
              value={groupByState}
              onChange={setGroupByState}
              style={{ width: 150 }}
            >
              <Option value="salon">По салонам</Option>
              <Option value="manager">По менеджерам</Option>
            </Select>
          </Space>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="📊 ПЛАН-ФАКТ">
            <Row gutter={[0, 16]}>
              <Col span={24}>
                <Statistic
                  title="Общий план"
                  value={totalPlan}
                  precision={0}
                  prefix="₽ "
                  valueStyle={{ fontSize: 20 }}
                />
              </Col>
              <Col span={24}>
                <Statistic
                  title="Факт"
                  value={totalFact}
                  precision={0}
                  prefix="₽ "
                  valueStyle={{ fontSize: 20 }}
                />
              </Col>
              <Col span={24}>
                <Progress
                  percent={planPercent}
                  strokeColor={planPercent >= 80 ? '#52c41a' : planPercent >= 50 ? '#faad14' : '#ff4d4f'}
                  status={planPercent >= 100 ? 'success' : undefined}
                />
                <Text>
                  {planPercent}% выполнения плана
                  {planPercent >= 100 && ' ✅'}
                </Text>
              </Col>
            </Row>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="📈 Воронка сети">
            {stages.length > 0 ? (
              <Space direction="vertical" style={{ width: '100%' }}>
                {stages.map((stage, idx) => (
                  <div key={stage.name}>
                    <Row>
                      <Col span={14}>
                        <Text type="secondary">{stage.name}</Text>
                      </Col>
                      <Col span={6}>
                        <Tag color="blue">{stage.count}</Tag>
                      </Col>
                      <Col span={4}>
                        <Text>{stage.conversion.toFixed(1)}%</Text>
                      </Col>
                    </Row>
                  </div>
                ))}
              </Space>
            ) : (
              <Empty description="Нет данных о воронке" />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card title="📋 Разбивка плана по салонам">
            {salonPlanData.length > 0 ? (
              <Collapse
                accordion
                onChange={(keys) => setExpandedSalon(keys[0] as string)}
                items={salonPlanData.map((salon) => ({
                  key: salon.id,
                  label: (
                    <Space>
                      <Text strong>{salon.name}</Text>
                      <Tag color={salon.percent >= 80 ? 'green' : salon.percent >= 50 ? 'orange' : 'red'}>
                        {salon.percent}%
                      </Tag>
                    </Space>
                  ),
                  children: (
                    <Row gutter={[16, 8]}>
                      <Col xs={24} sm={8}>
                        <Text type="secondary">Менеджер:</Text>
                        <br />
                        <Text strong>{salon.managerName}</Text>
                      </Col>
                      <Col xs={24} sm={8}>
                        <Text type="secondary">Кол-во продавцов:</Text>
                        <br />
                        <Text strong>{salon.sellersCount}</Text>
                      </Col>
                      <Col xs={24} sm={8}>
                        <Text type="secondary">Средний чек:</Text>
                        <br />
                        <Text strong>{salon.avgCheck.toLocaleString()} ₽</Text>
                      </Col>
                    </Row>
                  ),
                }))}
              />
            ) : (
              <Empty description="Нет данных по салонам" />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                <TrophyOutlined style={{ color: '#faad14' }} />
                Топ-3 менеджеров (по выручке)
              </Space>
            }
          >
            {topManagers.length > 0 ? (
              <Table
                dataSource={topManagers}
                columns={managerColumns}
                rowKey="id"
                pagination={false}
                size="small"
                locale={{ emptyText: 'Нет данных' }}
              />
            ) : (
              <Empty description="Нет данных" />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                <WarningOutlined style={{ color: '#ff4d4f' }} />
                Анти-топ-3 (по конверсии)
              </Space>
            }
          >
            {antiTopManagers.length > 0 ? (
              <Table
                dataSource={antiTopManagers}
                columns={managerColumns}
                rowKey="id"
                pagination={false}
                size="small"
                locale={{ emptyText: 'Нет данных' }}
              />
            ) : (
              <Empty description="Нет данных" />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card title="📉 Сравнение конверсии с сетью">
            {benchmarkData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={benchmarkData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="dealerConversion" stroke="#1890ff" name="Конверсия дилера" />
                  <Line type="monotone" dataKey="networkAvgConversion" stroke="#52c41a" name="Средняя по сети" />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <Empty description="Нет данных для графика" />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default FunnelPlanTab;
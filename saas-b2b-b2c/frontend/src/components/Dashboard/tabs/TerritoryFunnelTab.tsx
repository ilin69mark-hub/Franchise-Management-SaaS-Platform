// src/components/Dashboard/tabs/TerritoryFunnelTab.tsx
import React, { useState, useMemo, useEffect } from 'react';
import { Card, Row, Col, Typography, Table, Tag, Space, Statistic, Select, Button, List, Collapse, Tooltip, Segmented, Spin, Empty } from 'antd';
import { UserAddOutlined, ShopOutlined, DollarOutlined, PercentageOutlined, RiseOutlined, FallOutlined, WarningOutlined, AlertOutlined, SearchOutlined, LineChartOutlined, TableOutlined, UserOutlined, ArrowRightOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer } from 'recharts';
import { useTerritoryManagerStore, DealerMetrics } from '@/store/territoryManagerStore';

const { Text } = Typography;

interface Anomaly {
  id: string;
  dealerName: string;
  storeName?: string;
  type: 'conversion_drop' | 'avg_check' | 'traffic_drop' | 'refusal_rate';
  description: string;
  date: string;
  severity: 'warning' | 'critical';
}

interface StoreFunnel {
  storeId: string;
  storeName: string;
  traffic: number;
  consultation: number;
  measurement: number;
  kp: number;
  contract: number;
  payment: number;
  conversion: number;
}

interface ManagerFunnel {
  managerId: string;
  managerName: string;
  traffic: number;
  consultation: number;
  measurement: number;
  kp: number;
  contract: number;
  payment: number;
  conversion: number;
  history: number[];
}

interface TerritoryFunnelTabProps {
  loading?: boolean;
}

const STAGES = ['Трафик', 'Консультация', 'Замер', 'КП', 'Договор', 'Оплата'];

const TerritoryFunnelTab: React.FC<TerritoryFunnelTabProps> = ({ loading }) => {
  const { summary } = useTerritoryManagerStore();
  const [period, setPeriod] = useState('month');
  const [selectedDealers, setSelectedDealers] = useState<string[]>(['1', '2', '3']);
  const [chartMode, setChartMode] = useState<'absolute' | 'conversion_step' | 'conversion_traffic'>('conversion_traffic');
  const [anomalyFilter, setAnomalyFilter] = useState<string>('all');
  const [selectedDealer, setSelectedDealer] = useState<string | null>(null);
  const [expandedStore, setExpandedStore] = useState<string | null>(null);
  const [selectedStore, setSelectedStore] = useState<string | null>(null);

  const funnelByDealer = useMemo(() => [
    { dealerId: '1', dealerName: 'Мебель Москва', data: [450, 270, 180, 108, 54, 45] },
    { dealerId: '2', dealerName: 'Диванит Воронеж', data: [320, 192, 128, 77, 38, 32] },
    { dealerId: '3', dealerName: 'МебельЛига', data: [180, 108, 72, 43, 22, 18] },
    { dealerId: '4', dealerName: 'Евромебель', data: [250, 150, 100, 60, 30, 25] },
  ], []);

  const chartData = useMemo(() => {
    return STAGES.map((stage, i) => {
      const point: any = { name: stage };
      funnelByDealer
        .filter(d => selectedDealers.includes(d.dealerId))
        .forEach(d => {
          if (chartMode === 'absolute') {
            point[d.dealerName] = d.data[i];
          } else if (chartMode === 'conversion_step') {
            point[d.dealerName] = i > 0 ? Math.round((d.data[i] / d.data[i - 1]) * 100) : 100;
          } else {
            point[d.dealerName] = Math.round((d.data[i] / d.data[0]) * 100);
          }
        });
      return point;
    });
  }, [funnelByDealer, selectedDealers, chartMode]);

  const anomalies = useMemo((): Anomaly[] => [
    { id: '1', dealerName: 'Мебель Москва', storeName: 'Салон 1', type: 'conversion_drop', description: 'Падение конверсии Замер→КП на 15%', date: '2026-04-28', severity: 'warning' },
    { id: '2', dealerName: 'МебельЛига', type: 'avg_check', description: 'Рост среднего чека при падении кол-ва продаж (+25%, -18%)', date: '2026-04-27', severity: 'critical' },
    { id: '3', dealerName: 'Диванит Воронеж', type: 'traffic_drop', description: 'Падение трафика на 22%', date: '2026-04-26', severity: 'warning' },
    { id: '4', dealerName: 'Евромебель', storeName: 'Салон 1', type: 'refusal_rate', description: 'Рост отказов "нет в наличии" на 12%', date: '2026-04-25', severity: 'critical' },
    { id: '5', dealerName: 'Мебель Москва', type: 'traffic_drop', description: 'Падение трафика на 8%', date: '2026-04-24', severity: 'warning' },
  ], []);

  const filteredAnomalies = useMemo(() => {
    return anomalies.filter(a => {
      if (anomalyFilter === 'all') return true;
      return a.type === anomalyFilter;
    });
  }, [anomalies, anomalyFilter]);

  const storesFunnel = useMemo((): StoreFunnel[] => [
    { storeId: '1', storeName: 'Салон 1', traffic: 180, consultation: 108, measurement: 72, kp: 43, contract: 22, payment: 18, conversion: 10 },
    { storeId: '2', storeName: 'Салон 2', traffic: 150, consultation: 90, measurement: 60, kp: 36, contract: 18, payment: 15, conversion: 10 },
    { storeId: '3', storeName: 'Салон 3', traffic: 120, consultation: 72, measurement: 48, kp: 29, contract: 14, payment: 12, conversion: 10 },
  ], []);

  const managersFunnel = useMemo((): ManagerFunnel[] => [
    { managerId: '1', managerName: 'Иванов А.А.', traffic: 80, consultation: 56, measurement: 42, kp: 28, contract: 16, payment: 12, conversion: 15, history: [12, 14, 13, 15, 16, 14] },
    { managerId: '2', managerName: 'Петрова С.С.', traffic: 70, consultation: 49, measurement: 35, kp: 21, contract: 10, payment: 8, conversion: 11, history: [10, 11, 9, 12, 10, 11] },
    { managerId: '3', managerName: 'Сидоров В.В.', traffic: 30, consultation: 21, measurement: 14, kp: 8, contract: 4, payment: 3, conversion: 10, history: [8, 9, 7, 6, 5, 10] },
  ], []);

  const storeColumns = [
    { title: 'Салон', dataIndex: 'storeName', key: 'storeName', render: (name: string) => <Space><ShopOutlined />{name}</Space> },
    { title: 'Трафик', dataIndex: 'traffic', key: 'traffic' },
    { title: 'Консультация', dataIndex: 'consultation', key: 'consultation' },
    { title: 'Замер', dataIndex: 'measurement', key: 'measurement' },
    { title: 'КП', dataIndex: 'kp', key: 'kp' },
    { title: 'Договор', dataIndex: 'contract', key: 'contract' },
    { title: 'Оплата', dataIndex: 'payment', key: 'payment' },
    { title: 'Конверсия', dataIndex: 'conversion', key: 'conversion', render: (c: number) => <Tag color={c >= 10 ? 'green' : c >= 7 ? 'orange' : 'red'}>{c}%</Tag> },
  ];

  const managerColumns = [
    { title: 'Менеджер', dataIndex: 'managerName', key: 'managerName', render: (name: string) => <Space><UserOutlined />{name}</Space> },
    { title: 'Трафик', dataIndex: 'traffic', key: 'traffic' },
    { title: 'Консультация', dataIndex: 'consultation', key: 'consultation' },
    { title: 'Замер', dataIndex: 'measurement', key: 'measurement' },
    { title: 'КП', dataIndex: 'kp', key: 'kp' },
    { title: 'Договор', dataIndex: 'contract', key: 'contract' },
    { title: 'Оплата', dataIndex: 'payment', key: 'payment' },
    { title: 'Конверсия', dataIndex: 'conversion', key: 'conversion', render: (c: number, r: ManagerFunnel) => <Tag color={r.conversion >= 12 ? 'green' : r.conversion >= 8 ? 'orange' : 'red'}>{c}%</Tag> },
    { title: 'История', key: 'history', render: (_: any, r: ManagerFunnel) => (
      <div style={{ display: 'flex', alignItems: 'flex-end', height: 24, gap: 2, width: 80 }}>
        {r.history.map((v, i) => (
          <div key={i} style={{ flex: 1, background: v >= 10 ? '#52c41a' : v >= 7 ? '#fa8c16' : '#ff4d4f', height: `${v * 2}%`, borderRadius: 1 }} />
        ))}
      </div>
    ) },
  ];

  const anomalyColumns = [
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName', render: (n: string, r: Anomaly) => <Space>{r.storeName && <ShopOutlined />}{n}</Space> },
    { title: 'Тип аномалии', dataIndex: 'type', key: 'type', render: (t: string) => {
      const types: Record<string, string> = { conversion_drop: 'Падение конверсии', avg_check: 'Средний чек', traffic_drop: 'Падение трафика', refusal_rate: 'Отказы' };
      return types[t];
    }},
    { title: 'Описание', dataIndex: 'description', key: 'description' },
    { title: 'Дата', dataIndex: 'date', key: 'date' },
    { title: '', dataIndex: 'severity', key: 'severity', render: (s: string) => s === 'critical' ? <span style={{ color: '#ff4d4f' }}>🔴</span> : <span style={{ color: '#fa8c16' }}>⚠️</span> },
    { title: '', key: 'action', render: () => <Button size="small">Анализировать</Button> },
  ];

  const colors = ['#1890ff', '#52c41a', '#fa8c16', '#ff4d4f', '#722ed1'];

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col>
            <Select value={period} onChange={setPeriod} style={{ width: 120 }}>
              <Select.Option value="week">Неделя</Select.Option>
              <Select.Option value="month">Месяц</Select.Option>
              <Select.Option value="quarter">Квартал</Select.Option>
            </Select>
          </Col>
          <Col>
            <Select mode="multiple" value={selectedDealers} onChange={setSelectedDealers} style={{ width: 200 }} placeholder="Выбрать дилеров">
              {funnelByDealer.map(d => <Select.Option key={d.dealerId} value={d.dealerId}>{d.dealerName}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <Segmented value={chartMode} onChange={setChartMode} options={[
              { label: 'Абсолютные', value: 'absolute' },
              { label: 'Конверсия этапа', value: 'conversion_step' },
              { label: 'Конверсия от трафика', value: 'conversion_traffic' },
            ]} />
          </Col>
          <Col>
            <Select value={anomalyFilter} onChange={setAnomalyFilter} style={{ width: 150 }}>
              <Select.Option value="all">Все типы</Select.Option>
              <Select.Option value="conversion_drop">Конверсия</Select.Option>
              <Select.Option value="traffic_drop">Трафик</Select.Option>
              <Select.Option value="avg_check">Средний чек</Select.Option>
            </Select>
          </Col>
        </Row>
      </Card>

      <Card size="small" title="Сравнение воронок дилеров" style={{ marginBottom: 16 }}>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" />
            <YAxis />
            <RechartsTooltip />
            <Legend />
            {funnelByDealer
              .filter(d => selectedDealers.includes(d.dealerId))
              .map((d, i) => (
                <Line
                  key={d.dealerId}
                  type={d.dealerId === '3' ? 'dashed' : 'monotone'}
                  dataKey={d.dealerName}
                  stroke={colors[i % colors.length]}
                  strokeWidth={d.dealerId === '3' ? 2 : 1}
                  dot={d.dealerId === '3'}
                />
              ))}
          </LineChart>
        </ResponsiveContainer>
      </Card>

      <Collapse style={{ marginBottom: 16 }}>
        <Collapse.Panel header={`Светофор аномалий (${filteredAnomalies.length})`} key="anomalies">
          <Card size="small">
            <Table
              dataSource={filteredAnomalies}
              columns={anomalyColumns}
              rowKey="id"
              size="small"
              loading={loading}
              pagination={false}
            />
          </Card>
        </Collapse.Panel>
      </Collapse>

      <Card size="small" title="Drill-down по дилеру">
        <Row gutter={16}>
          <Col span={24}>
            <Select
              value={selectedDealer}
              onChange={setSelectedDealer}
              style={{ width: 200 }}
              placeholder="Выберите дилера"
              allowClear
            >
              {funnelByDealer.map(d => <Select.Option key={d.dealerId} value={d.dealerId}>{d.dealerName}</Select.Option>)}
            </Select>
          </Col>
        </Row>

        {selectedDealer && (
          <>
            <Row gutter={16} style={{ marginTop: 16 }}>
              <Col span={24}>
                <Collapse>
                  <Collapse.Panel header="Воронка по салонам" key="stores">
                    <Table
                      dataSource={storesFunnel}
                      columns={storeColumns}
                      rowKey="storeId"
                      size="small"
                      pagination={false}
                    />
                  </Collapse.Panel>
                </Collapse>
              </Col>
            </Row>

            <Row gutter={16} style={{ marginTop: 16 }}>
              <Col span={24}>
                <Select
                  value={selectedStore}
                  onChange={setSelectedStore}
                  style={{ width: 200 }}
                  placeholder="Выберите салон для детализации по менеджерам"
                  allowClear
                >
                  {storesFunnel.map(s => <Select.Option key={s.storeId} value={s.storeId}>{s.storeName}</Select.Option>)}
                </Select>
              </Col>
            </Row>

            {selectedStore && (
              <Row gutter={16} style={{ marginTop: 16 }}>
                <Col span={24}>
                  <Card size="small">
                    <Table
                      dataSource={managersFunnel}
                      columns={managerColumns}
                      rowKey="managerId"
                      size="small"
                      pagination={false}
                    />
                  </Card>
                </Col>
              </Row>
            )}
          </>
        )}

        {!selectedDealer && (
          <Empty description="Выберите дилера для детализации" style={{ marginTop: 24 }} />
        )}
      </Card>
    </div>
  );
};

export default TerritoryFunnelTab;
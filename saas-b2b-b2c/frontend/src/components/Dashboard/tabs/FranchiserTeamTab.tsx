import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Table, Tag, Typography, Space, Button, InputNumber, Select, Collapse, Statistic, Progress, Divider, Modal, message } from 'antd';
import { 
  ArrowUpOutlined, 
  ArrowDownOutlined, 
  TeamOutlined, 
  CheckCircleOutlined,
  WarningOutlined,
  EditOutlined,
  FilePdfOutlined,
  CopyOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import dynamic from 'next/dynamic';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { Panel } = Collapse;

const ManagerKpiChart = dynamic(() => import('./ManagerKpiChart'), { ssr: false });
const ManagerPlanFactChart = dynamic(() => import('./ManagerPlanFactChart'), { ssr: false });

interface Manager {
  id: string;
  name: string;
  territory: string;
  planPercent: number;
  redDealersPercent: number;
  sla: number;
  dealerGrowth: number;
  churnRate: number;
  forecastPercent: number;
  integralKpi: number;
  bonusForecast: number;
  kpiHistory?: { month: string; kpi: number }[];
  dealers?: { name: string; planPercent: number; status: 'green' | 'yellow' | 'red' }[];
  planFactData?: { month: string; plan: number; fact: number }[];
}

interface ManagerPlans {
  [managerId: string]: {
    salesPlan: number;
    targetDealers: number;
    targetRedDealers: number;
    targetSla: number;
  };
}

const calculateIntegralKpi = (m: Manager): number => {
  const kpiPlan = m.planPercent * 0.40;
  const kpiRed = (100 - m.redDealersPercent) * 0.25;
  const kpiSla = m.sla * 0.20;
  const kpiGrowth = Math.min((m.dealerGrowth / 2) * 0.10 * 100, 10);
  const kpiReports = 5;
  return Math.min(Math.round(kpiPlan + kpiRed + kpiSla + kpiGrowth + kpiReports), 100);
};

const calculateBonus = (kpi: number, baseBonus: number = 50000): number => {
  if (kpi >= 90) return baseBonus * 1.5;
  if (kpi >= 75) return baseBonus;
  if (kpi >= 50) return baseBonus * 0.5;
  return 0;
};

const mockManagers: Manager[] = [
  { 
    id: '1', name: 'Алексей Петров', territory: 'Север', planPercent: 92, redDealersPercent: 0, sla: 98, dealerGrowth: 2, churnRate: 5, forecastPercent: 95, integralKpi: 94, bonusForecast: 75000,
    kpiHistory: [
      { month: 'Ноя', kpi: 88 }, { month: 'Дек', kpi: 85 }, { month: 'Янв', kpi: 90 }, { month: 'Фев', kpi: 87 }, { month: 'Мар', kpi: 92 }, { month: 'Апр', kpi: 94 }
    ],
    dealers: [
      { name: 'Дилер 1С', planPercent: 95, status: 'green' },
      { name: 'Дилер Север', planPercent: 88, status: 'green' },
    ],
    planFactData: [
      { month: 'Янв', plan: 800000, fact: 750000 }, { month: 'Фев', plan: 800000, fact: 820000 }, { month: 'Мар', plan: 800000, fact: 780000 }, { month: 'Апр', plan: 800000, fact: 850000 }
    ]
  },
  { 
    id: '2', name: 'Мария Иванова', territory: 'Юг', planPercent: 78, redDealersPercent: 17, sla: 85, dealerGrowth: 1, churnRate: 8, forecastPercent: 82, integralKpi: 78, bonusForecast: 50000,
    kpiHistory: [
      { month: 'Ноя', kpi: 72 }, { month: 'Дек', kpi: 75 }, { month: 'Янв', kpi: 70 }, { month: 'Фев', kpi: 74 }, { month: 'Мар', kpi: 76 }, { month: 'Апр', kpi: 78 }
    ]
  },
  { 
    id: '3', name: 'Сергей Сидоров', territory: 'Запад', planPercent: 65, redDealersPercent: 30, sla: 72, dealerGrowth: -1, churnRate: 15, forecastPercent: 70, integralKpi: 58, bonusForecast: 0,
    kpiHistory: [
      { month: 'Ноя', kpi: 65 }, { month: 'Дек', kpi: 62 }, { month: 'Янв', kpi: 60 }, { month: 'Фев', kpi: 58 }, { month: 'Мар', kpi: 55 }, { month: 'Апр', kpi: 58 }
    ]
  },
  { 
    id: '4', name: 'Елена Смирнова', territory: 'Центр', planPercent: 88, redDealersPercent: 0, sla: 94, dealerGrowth: 1, churnRate: 3, forecastPercent: 90, integralKpi: 89, bonusForecast: 50000,
    kpiHistory: [
      { month: 'Ноя', kpi: 85 }, { month: 'Дек', kpi: 82 }, { month: 'Янв', kpi: 88 }, { month: 'Фев', kpi: 86 }, { month: 'Мар', kpi: 87 }, { month: 'Апр', kpi: 89 }
    ]
  },
];

const defaultPlans: ManagerPlans = {
  '1': { salesPlan: 5000000, targetDealers: 10, targetRedDealers: 10, targetSla: 95 },
  '2': { salesPlan: 3500000, targetDealers: 7, targetRedDealers: 15, targetSla: 85 },
  '3': { salesPlan: 4000000, targetDealers: 12, targetRedDealers: 20, targetSla: 80 },
  '4': { salesPlan: 3000000, targetDealers: 6, targetRedDealers: 10, targetSla: 90 },
};

const FranchiserTeamTab: React.FC = () => {
  const [managers] = useState<Manager[]>(mockManagers);
  const [selectedManager, setSelectedManager] = useState<string | null>(null);
  const [expandedKeys, setExpandedKeys] = useState<string[]>([]);
  const [plansModalOpen, setPlansModalOpen] = useState(false);
  const [selectedQuarter, setSelectedQuarter] = useState('2026-Q2');
  const [plans, setPlans] = useState<ManagerPlans>(defaultPlans);

  const getRowColor = (kpi: number) => {
    if (kpi >= 90) return '#f6ffed';
    if (kpi >= 75) return '#fffbe6';
    return '#fff1f0';
  };

  const columns = [
    {
      title: 'Менеджер',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <TeamOutlined />
          <Text strong>{name}</Text>
        </Space>
      ),
    },
    {
      title: 'Территория',
      dataIndex: 'territory',
      key: 'territory',
    },
    {
      title: '% плана',
      dataIndex: 'planPercent',
      key: 'planPercent',
      sorter: (a: Manager, b: Manager) => a.planPercent - b.planPercent,
      render: (v: number) => (
        <Progress percent={v} size="small" strokeColor={v >= 95 ? '#52c41a' : v >= 85 ? '#fa8c16' : '#ff4d4f'} />
      ),
    },
    {
      title: 'В красной зоне',
      dataIndex: 'redDealersPercent',
      key: 'redDealersPercent',
      sorter: (a: Manager, b: Manager) => a.redDealersPercent - b.redDealersPercent,
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
      sorter: (a: Manager, b: Manager) => a.sla - b.sla,
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
      sorter: (a: Manager, b: Manager) => a.dealerGrowth - b.dealerGrowth,
      render: (v: number) => (
        <Text type={v > 0 ? 'success' : v < 0 ? 'danger' : 'secondary'}>
          {v > 0 ? <ArrowUpOutlined /> : v < 0 ? <ArrowDownOutlined /> : '-'} {Math.abs(v)}
        </Text>
      ),
    },
    {
      title: 'Churn',
      dataIndex: 'churnRate',
      key: 'churnRate',
      render: (v: number) => (
        <Tag color={v < 5 ? 'green' : v < 10 ? 'orange' : 'red'}>
          {v}%
        </Tag>
      ),
    },
    {
      title: 'Прогноз',
      dataIndex: 'forecastPercent',
      key: 'forecastPercent',
      render: (v: number) => (
        <Tag color={v >= 95 ? 'green' : v >= 85 ? 'orange' : 'red'}>
          {v}%
        </Tag>
      ),
    },
    {
      title: 'KPI',
      dataIndex: 'integralKpi',
      key: 'integralKpi',
      sorter: (a: Manager, b: Manager) => a.integralKpi - b.integralKpi,
      render: (v: number) => (
        <Progress 
          percent={v} 
          size="small" 
          strokeColor={v >= 90 ? '#52c41a' : v >= 75 ? '#fa8c16' : '#ff4d4f'}
          format={(p) => `${p}%`}
        />
      ),
    },
    {
      title: 'Бонус',
      dataIndex: 'bonusForecast',
      key: 'bonusForecast',
      sorter: (a: Manager, b: Manager) => a.bonusForecast - b.bonusForecast,
      render: (v: number) => (
        <Text strong type={v > 0 ? 'success' : 'secondary'}>
          {v.toLocaleString()} ₽
        </Text>
      ),
    },
  ];

  const handlePlanChange = (managerId: string, field: keyof ManagerPlans[string], value: number) => {
    setPlans(prev => ({
      ...prev,
      [managerId]: { ...prev[managerId], [field]: value }
    }));
  };

  const handleSavePlans = () => {
    message.success('Планы сохранены');
    setPlansModalOpen(false);
  };

  const handleCopyFromPrev = () => {
    message.info('Скопировано из прошлого квартала');
  };

  const renderDetailPanel = (managerId: string) => {
    const manager = managers.find(m => m.id === managerId);
    if (!manager) return null;

    return (
      <Card style={{ marginTop: 16, marginLeft: 48, background: '#fafafa' }}>
        <Row gutter={24}>
          <Col span={12}>
            <Card title="KPI за 6 месяцев" size="small">
              <ManagerKpiChart data={manager.kpiHistory || []} />
            </Card>
          </Col>
          <Col span={12}>
            <Card title="План vs Факт" size="small">
              <ManagerPlanFactChart data={manager.planFactData || []} />
            </Card>
          </Col>
        </Row>
        <Divider />
        <Card title="Дилеры менеджера" size="small">
          <Table
            dataSource={manager.dealers || []}
            columns={[
              { title: 'Дилер', dataIndex: 'name', key: 'name' },
              { title: '% плана', dataIndex: 'planPercent', key: 'planPercent', render: (v: number) => `${v}%` },
              { title: 'Статус', dataIndex: 'status', key: 'status', render: (s: string) => <Tag color={s}>{s}</Tag> },
            ]}
            rowKey="name"
            pagination={false}
            size="small"
          />
        </Card>
        <Button type="link" icon={<FilePdfOutlined />}>Детальный отчёт (PDF)</Button>
      </Card>
    );
  };

  return (
    <div>
      <Title level={4}>Моя команда</Title>

      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<EditOutlined />} onClick={() => setPlansModalOpen(true)}>
          Назначить планы
        </Button>
        <Text type="secondary">Квартал: {selectedQuarter}</Text>
      </Space>

      <Card>
        <Table
          dataSource={managers}
          columns={columns}
          rowKey="id"
          expandable={{
            expandedRowRender: (record) => renderDetailPanel(record.id),
            rowExpandable: (record) => record.id !== undefined,
          }}
          pagination={false}
          rowStyle={(record) => ({
            background: getRowColor(record.integralKpi),
          })}
          onRow={(record) => ({
            onClick: () => setSelectedManager(record.id),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>

      <Modal
        title="Назначение планов менеджерам"
        open={plansModalOpen}
        onCancel={() => setPlansModalOpen(false)}
        footer={null}
        width={800}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Select value={selectedQuarter} onChange={setSelectedQuarter} style={{ width: 200 }}>
            <Option value="2026-Q1">Q1 2026</Option>
            <Option value="2026-Q2">Q2 2026</Option>
            <Option value="2026-Q3">Q3 2026</Option>
            <Option value="2026-Q4">Q4 2026</Option>
          </Select>

          <Table
            dataSource={managers}
            columns={[
              { title: 'Менеджер', dataIndex: 'name', key: 'name' },
              { title: 'Территория', dataIndex: 'territory', key: 'territory' },
              { 
                title: 'План продаж', 
                key: 'salesPlan',
                render: (_: any, r: Manager) => (
                  <InputNumber 
                    value={plans[r.id]?.salesPlan} 
                    onChange={(v) => handlePlanChange(r.id, 'salesPlan', v || 0)}
                    formatter={(v) => `${v}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                    parser={(v) => parseInt(v?.replace(/ /g, '') || '0')}
                    style={{ width: 120 }}
                  />
                )
              },
              { 
                title: 'Цель дилеров', 
                key: 'targetDealers',
                render: (_: any, r: Manager) => (
                  <InputNumber 
                    value={plans[r.id]?.targetDealers}
                    onChange={(v) => handlePlanChange(r.id, 'targetDealers', v || 0)}
                    min={0}
                    style={{ width: 80 }}
                  />
                )
              },
              { 
                title: 'Макс красных %', 
                key: 'targetRedDealers',
                render: (_: any, r: Manager) => (
                  <InputNumber 
                    value={plans[r.id]?.targetRedDealers}
                    onChange={(v) => handlePlanChange(r.id, 'targetRedDealers', v || 0)}
                    min={0}
                    max={100}
                    formatter={(v) => `${v}%`}
                    parser={(v) => parseInt(v?.replace('%', '') || '0')}
                    style={{ width: 80 }}
                  />
                )
              },
              { 
                title: 'Цель SLA', 
                key: 'targetSla',
                render: (_: any, r: Manager) => (
                  <InputNumber 
                    value={plans[r.id]?.targetSla}
                    onChange={(v) => handlePlanChange(r.id, 'targetSla', v || 0)}
                    min={0}
                    max={100}
                    formatter={(v) => `${v}%`}
                    parser={(v) => parseInt(v?.replace('%', '') || '0')}
                    style={{ width: 80 }}
                  />
                )
              },
            ]}
            rowKey="id"
            pagination={false}
            size="small"
          />

          <Space>
            <Button icon={<CopyOutlined />} onClick={handleCopyFromPrev}>Скопировать из прошлого квартала</Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSavePlans}>Сохранить планы</Button>
          </Space>
        </Space>
      </Modal>
    </div>
  );
};

export default FranchiserTeamTab;
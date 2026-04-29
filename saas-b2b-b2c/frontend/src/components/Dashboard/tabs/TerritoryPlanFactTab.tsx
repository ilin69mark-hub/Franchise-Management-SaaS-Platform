// src/components/Dashboard/tabs/TerritoryPlanFactTab.tsx
import React, { useState, useMemo } from 'react';
import { Card, Row, Col, Typography, Table, Tag, Space, Statistic, Select, Button, Modal, Form, Input, Upload, Empty, Segmented } from 'antd';
import { ShopOutlined, DollarOutlined, PercentageOutlined, RiseOutlined, CheckCircleOutlined, ExclamationCircleOutlined, DownloadOutlined, FilePdfOutlined, WarningOutlined, EditOutlined, MessageOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer, BarChart, Bar, ReferenceLine, ComposedChart, Area } from 'recharts';
import { useTerritoryManagerStore, DealerMetrics } from '@/store/territoryManagerStore';

const { Text } = Typography;

interface DeviationRow {
  dealerId: string;
  dealerName: string;
  plan: number;
  fact: number;
  deviation: number;
  deviationPercent: number;
  reason?: string;
  forecast: number;
  actions?: string;
}

interface ForecastScenario {
  type: 'optimistic' | 'realistic' | 'pessimistic';
  amount: number;
  percent: number;
  color: string;
}

interface ContributionData {
  dealerName: string;
  plan: number;
  fact: number;
}

interface TerritoryPlanFactTabProps {
  loading?: boolean;
}

const DEVIATION_REASONS = [
  { value: 'traffic_drop', label: 'Падение трафика' },
  { value: 'low_conversion', label: 'Низкая конверсия' },
  { value: 'goods_deficit', label: 'Дефицит товара' },
  { value: 'staff_issue', label: 'Кадровая проблема' },
  { value: 'seasonal', label: 'Сезонный спад' },
  { value: 'other', label: 'Прочее' },
];

const TerritoryPlanFactTab: React.FC<TerritoryPlanFactTabProps> = ({ loading }) => {
  const [period, setPeriod] = useState('quarter');
  const [reportModalOpen, setReportModalOpen] = useState(false);
  const [deviationReasons, setDeviationReasons] = useState<Record<string, string>>({});
  const [deviationActions, setDeviationActions] = useState<Record<string, string>>({});
  const [reportProposals, setReportProposals] = useState('');

  const dealersData = useMemo(() => [
    { dealerId: '1', dealerName: 'Мебель Москва', plan: 15000000, fact: 13800000, forecast: 16500000 },
    { dealerId: '2', dealerName: 'Диванит Воронеж', plan: 8000000, fact: 6240000, forecast: 8800000 },
    { dealerId: '3', dealerName: 'МебельЛига', plan: 5000000, fact: 2250000, forecast: 2750000 },
    { dealerId: '4', dealerName: 'Салон мебели Казань', plan: 12000000, fact: 13200000, forecast: 13200000 },
    { dealerId: '5', dealerName: 'Евромебель', plan: 7000000, fact: 4690000, forecast: 5250000 },
  ], []);

  const totalPlan = useMemo(() => dealersData.reduce((s, d) => s + d.plan, 0), [dealersData]);
  const totalFact = useMemo(() => dealersData.reduce((s, d) => s + d.fact, 0), [dealersData]);
  const totalForecast = useMemo(() => dealersData.reduce((s, d) => s + d.forecast, 0), [dealersData]);

  const optimisticScenario = useMemo((): ForecastScenario => {
    const leaders = dealersData.filter(d => d.fact / d.plan >= 1.1);
    const problems = dealersData.filter(d => d.fact / d.plan < 0.7);
    const total = leaders.reduce((s, d) => s + d.forecast, 0) + problems.reduce((s, d) => s + d.fact * 1.1, 0);
    return { type: 'optimistic', amount: total, percent: (total / totalPlan) * 100, color: '#52c41a' };
  }, [dealersData, totalPlan]);

  const realisticScenario = useMemo((): ForecastScenario => ({
    type: 'realistic', amount: totalForecast, percent: (totalForecast / totalPlan) * 100, color: '#1890ff'
  }), [totalForecast, totalPlan]);

  const pessimisticScenario = useMemo((): ForecastScenario => {
    const problems = dealersData.filter(d => d.fact / d.plan < 0.7);
    const total = dealersData.reduce((s, d) => s + (d.fact / d.plan < 0.7 ? d.fact : d.fact * 0.9), 0);
    return { type: 'pessimistic', amount: total, percent: (total / totalPlan) * 100, color: '#ff4d4f' };
  }, [dealersData, totalPlan]);

  const contributionData = useMemo((): ContributionData[] => {
    return [...dealersData]
      .sort((a, b) => b.plan - a.plan)
      .map(d => ({ dealerName: d.dealerName, plan: d.plan, fact: d.fact }));
  }, [dealersData]);

  const dynamicsData = useMemo(() => [
    { month: 'Янв', plan: 12000000, fact: 10800000, forecast: null },
    { month: 'Фев', plan: 13000000, fact: 11700000, forecast: null },
    { month: 'Мар', plan: 14000000, fact: 12600000, forecast: null },
    { month: 'Апр', plan: 15000000, fact: 13800000, forecast: null },
    { month: 'Май', plan: 15000000, fact: null, forecast: 14500000 },
    { month: 'Июн', plan: 15000000, fact: null, forecast: 15000000 },
  ], []);

  const top3Dealers = useMemo(() => {
    const sorted = [...dealersData].sort((a, b) => b.fact - a.fact);
    const top3 = sorted.slice(0, 3);
    return top3.reduce((s, d) => s + d.fact, 0);
  }, [dealersData]);

  const deviationData = useMemo((): DeviationRow[] => {
    return dealersData.map(d => ({
      dealerId: d.dealerId,
      dealerName: d.dealerName,
      plan: d.plan,
      fact: d.fact,
      deviation: d.fact - d.plan,
      deviationPercent: ((d.fact - d.plan) / d.plan) * 100,
      forecast: d.forecast,
    }));
  }, [dealersData]);

  const deviationText = useMemo(() => {
    const deviations = dealersData.filter(d => d.fact < d.plan).sort((a, b) => (a.fact - a.plan) - (b.fact - b.plan));
    if (deviations.length === 0) return 'Все дилеры выполняют план';
    const totalGap = deviations.reduce((s, d) => s + d.plan - d.fact, 0);
    const list = deviations.slice(0, 3).map(d => `${d.dealerName} (-${((d.plan - d.fact) / 1000000).toFixed(1)} млн)`).join(', ');
    return `Отставание от плана ${(totalGap / 1000000).toFixed(1)} млн руб. сформировано из-за ${list}`;
  }, [dealersData]);

  const handleReasonChange = (dealerId: string, reason: string) => {
    setDeviationReasons(prev => ({ ...prev, [dealerId]: reason }));
  };

  const handleActionsChange = (dealerId: string, actions: string) => {
    setDeviationActions(prev => ({ ...prev, [dealerId]: actions }));
  };

  const deviationColumns = [
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName', render: (n: string, r: DeviationRow) => <Space><ShopOutlined style={{ color: r.deviation >= 0 ? '#52c41a' : '#ff4d4f' }} />{n}</Space> },
    { title: 'План', dataIndex: 'plan', key: 'plan', render: (v: number) => `${(v / 1000000).toFixed(1)} млн ₽` },
    { title: 'Факт', dataIndex: 'fact', key: 'fact', render: (v: number) => `${(v / 1000000).toFixed(1)} млн ₽` },
    { title: 'Отклонение', key: 'deviation', render: (_: any, r: DeviationRow) => (
      <Tag color={r.deviation >= 0 ? 'green' : 'red'}>
        {(r.deviation / 1000000).toFixed(1)} млн ({r.deviationPercent.toFixed(0)}%)
      </Tag>
    )},
    { title: 'Причина', key: 'reason', render: (_: any, r: DeviationRow) => (
      <Select
        value={deviationReasons[r.dealerId]}
        onChange={(v) => handleReasonChange(r.dealerId, v)}
        style={{ width: 150 }}
        placeholder="Выберите"
      >
        {DEVIATION_REASONS.map(reason => (
          <Select.Option key={reason.value} value={reason.value}>{reason.label}</Select.Option>
        ))}
      </Select>
    )},
    { title: 'Прогноз', dataIndex: 'forecast', key: 'forecast', render: (v: number) => `${(v / 1000000).toFixed(1)} млн ₽` },
    { title: 'Действия', key: 'actions', render: (_: any, r: DeviationRow) => (
      <Input.TextArea
        value={deviationActions[r.dealerId]}
        onChange={(e) => handleActionsChange(r.dealerId, e.target.value)}
        placeholder="План мероприятий"
        rows={1}
      />
    )},
  ];

  const generatePDF = async () => {
    const token = localStorage.getItem('accessToken');
    try {
      const res = await fetch('/api/franchiser/reports/generate', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          period,
          proposals: reportProposals,
          deviations: deviationReasons,
          actions: deviationActions,
        }),
      });
      if (res.ok) {
        const blob = await res.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `отчёт-${period}.pdf`;
        a.click();
      }
    } catch (e) {
      setReportModalOpen(true);
    }
  };

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col>
            <Select value={period} onChange={setPeriod} style={{ width: 120 }}>
              <Select.Option value="month">Месяц</Select.Option>
              <Select.Option value="quarter">Квартал</Select.Option>
            </Select>
          </Col>
          <Col>
            <Button type="primary" icon={<FilePdfOutlined />} onClick={() => setReportModalOpen(true)}>
              Сформировать PDF-отчёт
            </Button>
          </Col>
        </Row>
      </Card>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={8}>
          <Card size="small" style={{ background: optimisticScenario.color + '15', borderColor: optimisticScenario.color }}>
            <Statistic
              title="Оптимистичный"
              value={optimisticScenario.amount / 1000000}
              suffix="млн ₽"
              precision={1}
              valueStyle={{ color: optimisticScenario.color }}
            />
            <Text style={{ color: optimisticScenario.color }}>{optimisticScenario.percent.toFixed(0)}% плана</Text>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card size="small" style={{ background: realisticScenario.color + '15', borderColor: realisticScenario.color }}>
            <Statistic
              title="Реалистичный"
              value={realisticScenario.amount / 1000000}
              suffix="млн ₽"
              precision={1}
              valueStyle={{ color: realisticScenario.color }}
            />
            <Text style={{ color: realisticScenario.color }}>{realisticScenario.percent.toFixed(0)}% плана</Text>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card size="small" style={{ background: pessimisticScenario.color + '15', borderColor: pessimisticScenario.color }}>
            <Statistic
              title="Пессимистичный"
              value={pessimisticScenario.amount / 1000000}
              suffix="млн ₽"
              precision={1}
              valueStyle={{ color: pessimisticScenario.color }}
            />
            <Text style={{ color: pessimisticScenario.color }}>{pessimisticScenario.percent.toFixed(0)}% плана</Text>
          </Card>
        </Col>
      </Row>

      <Text>{deviationText}</Text>

      <Card size="small" title="Вклад дилеров в план" style={{ marginTop: 16, marginBottom: 16 }}>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={contributionData} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis type="number" />
            <YAxis type="category" dataKey="dealerName" width={120} />
            <RechartsTooltip />
            <Legend />
            <Bar dataKey="plan" fill="#d9d9d9" name="План" />
            <Bar dataKey="fact" name="Факт" fill="#52c41a" />
          </BarChart>
        </ResponsiveContainer>
        {
          top3Dealers > totalFact * 0.6 && (
            <Tag color="gold" style={{ marginTop: 8 }}>Топ-3 дилера дают {((top3Dealers / totalFact) * 100).toFixed(0)}% результата</Tag>
          )
        }
      </Card>

      <Card size="small" title="План-факт динамика" style={{ marginBottom: 16 }}>
        <ResponsiveContainer width="100%" height={250}>
          <ComposedChart data={dynamicsData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="month" />
            <YAxis />
            <RechartsTooltip />
            <Legend />
            <Area type="monotone" dataKey="fact" fill="#ff4d4f" fillOpacity={0.2} stroke="transparent" />
            <Line type="dashed" dataKey="plan" stroke="#d9d9d9" strokeDasharray="5 5" name="План" />
            <Line type="solid" dataKey="fact" stroke="#52c41a" strokeWidth={2} name="Факт" />
            <Line type="dashed" dataKey="forecast" stroke="#1890ff" strokeDasharray="3 3" name="Прогноз" />
          </ComposedChart>
        </ResponsiveContainer>
      </Card>

      <Card size="small" title="Детализация отклонений">
        <Table
          dataSource={deviationData}
          columns={deviationColumns}
          rowKey="dealerId"
          size="small"
          loading={loading}
          pagination={false}
        />
      </Card>

      <Modal
        title="Предпросмотр отчёта"
        open={reportModalOpen}
        onCancel={() => setReportModalOpen(false)}
        width={800}
        footer={[
          <Button key="cancel" onClick={() => setReportModalOpen(false)}>Закрыть</Button>,
          <Button key="download" type="primary" icon={<DownloadOutlined />} onClick={generatePDF}>Скачать PDF</Button>,
        ]}
      >
        <Card size="small" title="Сводка">
          <Space direction="vertical">
            <Text>Период: {period === 'quarter' ? 'Квартал' : 'Месяц'}</Text>
            <Text>План: {(totalPlan / 1000000).toFixed(1)} млн ₽</Text>
            <Text>Факт: {(totalFact / 1000000).toFixed(1)} млн ₽</Text>
            <Text>Прогноз: {(totalForecast / 1000000).toFixed(1)} млн ₽</Text>
          </Space>
        </Card>
        <Form.Item label="Предложения">
          <Input.TextArea value={reportProposals} onChange={e => setReportProposals(e.target.value)} rows={4} placeholder="Введите предложения для руководителя" />
        </Form.Item>
      </Modal>
    </div>
  );
};

export default TerritoryPlanFactTab;
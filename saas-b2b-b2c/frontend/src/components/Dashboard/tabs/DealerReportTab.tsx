// src/components/Dashboard/tabs/DealerReportTab.tsx
import React, { useState, useMemo } from 'react';
import { Card, Row, Col, Table, Tag, Button, Space, Modal, Form, Input, Select, Typography, Empty, Spin, Progress, message, DatePicker, List, Avatar, Divider } from 'antd';
import { 
  FilePdfOutlined, 
  DownloadOutlined, 
  SendOutlined, 
  EyeOutlined, 
  CalendarOutlined, 
  HistoryOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  DollarOutlined,
  ShopOutlined,
  UndoOutlined,
  BarChartOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Text, Title } = Typography;
const { TextArea } = Input;
const { MonthPicker } = DatePicker;

export type ReportPeriod = 'month' | 'quarter' | 'year';

interface PlanFactData {
  plan: number;
  fact: number;
  percent: number;
}

interface FunnelData {
  traffic: number;
  consultation: number;
  measurement: number;
  kp: number;
  contract: number;
  payment: number;
}

interface InventoryItem {
  id: string;
  name: string;
  stock: number;
  value: number;
}

interface ReturnsData {
  count: number;
  amount: number;
}

interface LostSalesData {
  total: number;
  topReasons: { reason: string; amount: number }[];
}

interface MarketingBudgetData {
  total: number;
  used: number;
  remaining: number;
}

interface ReportData {
  dealerName: string;
  period: ReportPeriod;
  startDate: string;
  endDate: string;
  generatedAt: string;
  planFact: PlanFactData;
  funnel: FunnelData;
  inventory: InventoryItem[];
  returns: ReturnsData;
  lostSales: LostSalesData;
  marketingBudget: MarketingBudgetData;
  comment?: string;
}

interface Report {
  id: string;
  period: string;
  generatedAt: string;
  sentAt?: string;
  status: 'draft' | 'sent';
}

interface DealerReportTabProps {
  reportData?: ReportData;
  reportHistory?: Report[];
  loading?: boolean;
  onSend?: (data: ReportData) => Promise<void>;
}

const DealerReportTab: React.FC<DealerReportTabProps> = ({
  reportData,
  reportHistory = [],
  loading = false,
  onSend,
}) => {
  const [previewOpen, setPreviewOpen] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [selectedPeriod, setSelectedPeriod] = useState<ReportPeriod>('month');
  const [selectedDate, setSelectedDate] = useState(dayjs());
  const [comment, setComment] = useState('');
  const [sending, setSending] = useState(false);

  const filteredHistory = useMemo(() => {
    return reportHistory.slice(0, 12);
  }, [reportHistory]);

  const handleGenerate = async () => {
    setGenerating(true);
    await new Promise(resolve => setTimeout(resolve, 1500));
    setGenerating(false);
    setPreviewOpen(true);
  };

  const handleDownload = async () => {
    message.success('PDF скачивается...');
  };

  const handleSend = async () => {
    if (!reportData) return;
    setSending(true);
    try {
      await onSend?.(reportData);
      message.success('Отчёт отправлен');
      setPreviewOpen(false);
    } catch (error) {
      message.error('Ошибка отправки');
    } finally {
      setSending(false);
    }
  };

  const historyColumns = [
    {
      title: 'Период',
      dataIndex: 'period',
      key: 'period',
      width: 120,
    },
    {
      title: 'Сформирован',
      dataIndex: 'generatedAt',
      key: 'generatedAt',
      width: 150,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (val: string) => (
        <Tag color={val === 'sent' ? 'green' : 'orange'}>
          {val === 'sent' ? 'От��равлен' : 'Черновик'}
        </Tag>
      ),
    },
    {
      title: 'Отправлен',
      dataIndex: 'sentAt',
      key: 'sentAt',
      width: 150,
      render: (val?: string) => val || '-',
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
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24}>
          <Card title="📊 Генерация отчёта для бренда">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Row gutter={16}>
                <Col xs={24} sm={8}>
                  <Text type="secondary">Период отчёта:</Text>
                  <Select 
                    value={selectedPeriod} 
                    onChange={setSelectedPeriod} 
                    style={{ width: '100%', marginTop: 4 }}
                  >
                    <Select.Option value="month">Месяц</Select.Option>
                    <Select.Option value="quarter">Квартал</Select.Option>
                    <Select.Option value="year">Год</Select.Option>
                  </Select>
                </Col>
                <Col xs={24} sm={8}>
                  <Text type="secondary">Дата:</Text>
                  <MonthPicker 
                    value={selectedDate} 
                    onChange={setSelectedDate}
                    style={{ width: '100%', marginTop: 4 }}
                    format="MMMM YYYY"
                  />
                </Col>
                <Col xs={24} sm={8}>
                  <Text type="secondary">Комментарий дилера:</Text>
                  <TextArea 
                    value={comment}
                    onChange={e => setComment(e.target.value)}
                    placeholder="Добавьте комментарий..."
                    rows={2}
                    style={{ marginTop: 4 }}
                  />
                </Col>
              </Row>
              <Row>
                <Col>
                  <Button 
                    type="primary" 
                    icon={<FilePdfOutlined />} 
                    loading={generating}
                    onClick={handleGenerate}
                  >
                    Сформировать отчёт
                  </Button>
                </Col>
              </Row>
            </Space>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="📋 Предпросмотр отчёта" extra={
            <Space>
              <Button icon={<DownloadOutlined />} onClick={handleDownload}>
                Скачать PDF
              </Button>
              <Button type="primary" icon={<SendOutlined />} onClick={handleSend} loading={sending}>
                Отправить менеджеру
              </Button>
            </Space>
          }>
            {reportData ? (
              <div style={{ maxHeight: 600, overflow: 'auto' }}>
                <div style={{ textAlign: 'center', marginBottom: 24 }}>
                  <Title level={3}>ОТЧЁТ ДИЛЕРА</Title>
                  <Text type="secondary">{reportData.dealerName}</Text>
                  <br />
                  <Text>Период: {reportData.startDate} - {reportData.endDate}</Text>
                  <br />
                  <Text type="secondary">Сформирован: {reportData.generatedAt}</Text>
                </div>

                <Divider>📈 Блок 1: План-факт</Divider>
                <Row gutter={16}>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic title="План" value={reportData.planFact.plan} prefix="₽ " />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic title="Факт" value={reportData.planFact.fact} prefix="₽ " />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic 
                        title="Выполнение" 
                        value={reportData.planFact.percent} 
                        suffix="%"
                        valueStyle={{ color: reportData.planFact.percent >= 80 ? '#52c41a' : '#fa8c16' }}
                      />
                    </Card>
                  </Col>
                </Row>

                <Divider>🔽 Блок 2: Воронка продаж</Divider>
                <Row gutter={[0, 8]}>
                  <Col span={24}><Text strong>Трафик:</Text> {reportData.funnel.traffic}</Col>
                  <Col span={24}><Text strong>Консультация:</Text> {reportData.funnel.consultation}</Col>
                  <Col span={24}><Text strong>Замер:</Text> {reportData.funnel.measurement}</Col>
                  <Col span={24}><Text strong>КП:</Text> {reportData.funnel.kp}</Col>
                  <Col span={24}><Text strong>Договор:</Text> {reportData.funnel.contract}</Col>
                  <Col span={24}><Text strong>Оплата:</Text> {reportData.funnel.payment}</Col>
                </Row>

                <Divider>📦 Блок 3: Товарные остатки (топ-10)</Divider>
                <Table 
                  dataSource={reportData.inventory.slice(0, 10)} 
                  columns={[
                    { title: 'Название', dataIndex: 'name', key: 'name' },
                    { title: 'Остаток', dataIndex: 'stock', key: 'stock', render: (v: number) => `${v} шт` },
                    { title: 'Стоимость', dataIndex: 'value', key: 'value', render: (v: number) => `${v.toLocaleString()} ₽` },
                  ]}
                  rowKey="id"
                  pagination={false}
                  size="small"
                />

                <Divider>↩️ Блок 4: Возвраты и рекламации</Divider>
                <Row gutter={16}>
                  <Col span={12}>
                    <Statistic title="Количество" value={reportData.returns.count} />
                  </Col>
                  <Col span={12}>
                    <Statistic title="Сумма" value={reportData.returns.amount} prefix="₽ " />
                  </Col>
                </Row>

                <Divider>💸 Блок 5: Упущенная прибыль</Divider>
                <Row gutter={16}>
                  <Col>
                    <Statistic title="Общая сумма" value={reportData.lostSales.total} prefix="₽ " valueStyle={{ color: '#ff4d4f' }} />
                  </Col>
                </Row>
                <Text type="secondary">Топ-3 причины:</Text>
                <List
                  size="small"
                  dataSource={reportData.lostSales.topReasons}
                  renderItem={(item, idx) => (
                    <List.Item>
                      <Text>{idx + 1}. {item.reason}</Text>
                      <Text strong>{item.amount.toLocaleString()} ₽</Text>
                    </List.Item>
                  )}
                />

                <Divider>💰 Блок 6: Маркетинговый бюджет</Divider>
                <Row gutter={16}>
                  <Col span={8}>
                    <Statistic title="Выделено" value={reportData.marketingBudget.total} prefix="₽ " />
                  </Col>
                  <Col span={8}>
                    <Statistic title="Использовано" value={reportData.marketingBudget.used} prefix="₽ " />
                  </Col>
                  <Col span={8}>
                    <Statistic title="Остаток" value={reportData.marketingBudget.remaining} prefix="₽ " valueStyle={{ color: '#52c41a' }} />
                  </Col>
                </Row>
                <Progress 
                  percent={Math.round((reportData.marketingBudget.used / reportData.marketingBudget.total) * 100)} 
                  status="normal"
                />

                <Divider>💬 Блок 7: Комментарий дилера</Divider>
                <Text>{reportData.comment || 'Нет комментария'}</Text>
              </div>
            ) : (
              <Empty description="Нажмите 'Сформировать отчёт' для предпросмотра" />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title={<><HistoryOutlined /> История отчётов</>}>
            {filteredHistory.length > 0 ? (
              <List
                itemLayout="horizontal"
                dataSource={filteredHistory}
                renderItem={(item) => (
                  <List.Item
                    actions={[
                      <Button key="view" size="small" icon={<EyeOutlined />} />,
                    ]}
                  >
                    <List.Item.Meta
                      avatar={<Avatar icon={<FilePdfOutlined />} style={{ backgroundColor: item.status === 'sent' ? '#52c41a' : '#fa8c16' }} />}
                      title={item.period}
                      description={
                        <Space direction="vertical" size={0}>
                          <Text type="secondary">Сформирован: {item.generatedAt}</Text>
                          {item.sentAt && <Text type="secondary">Отправлен: {item.sentAt}</Text>}
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="История пуста" />
            )}
          </Card>
        </Col>
      </Row>

      <Modal
        title="Предпросмотр отчёта"
        open={previewOpen}
        onCancel={() => setPreviewOpen(false)}
        width={800}
        footer={[
          <Button key="download" icon={<DownloadOutlined />} onClick={handleDownload}>
            Скачать PDF
          </Button>,
          <Button key="send" type="primary" icon={<SendOutlined />} onClick={handleSend} loading={sending}>
            Отправить менеджеру
          </Button>,
        ]}
      >
        <Spin spinning={generating}>
          <div style={{ maxHeight: 500, overflow: 'auto' }}>
            {reportData && (
              <div style={{ padding: 16 }}>
                <Title level={4} style={{ textAlign: 'center' }}>ОТЧЁТ ДИЛЕРА</Title>
                <Text style={{ display: 'block', textAlign: 'center' }}>{reportData.dealerName}</Text>
                <Text type="secondary" style={{ display: 'block', textAlign: 'center' }}>
                  {reportData.startDate} - {reportData.endDate}
                </Text>
              </div>
            )}
          </div>
        </Spin>
      </Modal>
    </div>
  );
};

export default DealerReportTab;
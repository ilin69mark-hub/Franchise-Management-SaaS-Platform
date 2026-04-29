import React, { Suspense, lazy } from 'react';
import { Row, Col, Card, Table, Tag, Typography, Space, Button, Select, Checkbox, Input, InputNumber, Modal, message, Divider, CheckboxChangeEvent, Statistic } from 'antd';
import { 
  ArrowUpOutlined, 
  ArrowDownOutlined, 
  ShopOutlined, 
  WarningOutlined,
  CheckCircleOutlined,
  DownloadOutlined,
  MailOutlined,
  SaveOutlined,
  HistoryOutlined,
  FilePdfOutlined,
  SettingOutlined,
  DollarOutlined,
  TeamOutlined,
  GlobalOutlined,
  RiseOutlined,
} from '@ant-design/icons';
import dynamic from 'next/dynamic';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const ReportPlanFactChart = lazy(() => import('./ReportPlanFactChart'));

interface ReportBlock {
  key: string;
  title: string;
  enabled: boolean;
}

interface ExecutiveSummary {
  planAmount: number;
  factAmount: number;
  forecastAmount: number;
  forecastPercent: number;
  conclusion: string;
  trend: 'up' | 'stable' | 'down';
}

interface NetworkGrowthData {
  dealersStart: number;
  dealersEnd: number;
  newDealers: number;
  churnedDealers: number;
  avgRevenuePerDealer: number;
  revenueDynamics: number;
}

interface SalesStructureData {
  category: string;
  share: number;
  dynamics: number;
}

interface TerritoryRatingData {
  manager: string;
  planPercent: number;
  dealersCount: number;
  growth: number;
  forecast: number;
}

interface RiskData {
  issue: string;
  affectedDealers: number;
  impact: number;
}

interface ReportDraft {
  id?: string;
  period: string;
  date: string;
  blocks: ReportBlock[];
  executiveSummary: ExecutiveSummary;
  networkGrowth: NetworkGrowthData;
  salesStructure: SalesStructureData[];
  territoryRating: TerritoryRatingData[];
  risks: RiskData[];
  proposals: string;
}

const defaultBlocks: ReportBlock[] = [
  { key: 'executive', title: 'Executive Summary', enabled: true },
  { key: 'planFact', title: 'График План-Факт', enabled: true },
  { key: 'growth', title: 'Анализ роста сети', enabled: true },
  { key: 'structure', title: 'Структура продаж', enabled: true },
  { key: 'territories', title: 'Рейтинг территорий', enabled: true },
  { key: 'risks', title: 'Ключевые риски', enabled: true },
  { key: 'proposals', title: 'Предложения', enabled: true },
];

const mockExecutive: ExecutiveSummary = {
  planAmount: 42000000,
  factAmount: 37800000,
  forecastAmount: 41160000,
  forecastPercent: 98,
  conclusion: 'Отставание 5% обусловлено срывом поставок коллекции Весна. Риски устранены. Прогноз — 98%.',
  trend: 'up',
};

const mockGrowth: NetworkGrowthData = {
  dealersStart: 21,
  dealersEnd: 24,
  newDealers: 4,
  churnedDealers: 1,
  avgRevenuePerDealer: 1575000,
  revenueDynamics: 8,
};

const mockStructure: SalesStructureData[] = [
  { category: 'Кухни', share: 45, dynamics: 5 },
  { category: 'Мягкая мебель', share: 22, dynamics: -2 },
  { category: 'Корпусная', share: 18, dynamics: 3 },
  { category: 'Матрасы', share: 10, dynamics: 1 },
  { category: 'Аксессуары', share: 5, dynamics: 0 },
];

const mockTerritories: TerritoryRatingData[] = [
  { manager: 'Алексей Петров', planPercent: 92, dealersCount: 8, growth: 2, forecast: 95 },
  { manager: 'Елена Смирнова', planPercent: 88, dealersCount: 5, growth: 1, forecast: 90 },
  { manager: 'Мария Иванова', planPercent: 78, dealersCount: 6, growth: 1, forecast: 82 },
  { manager: 'Сергей Сидоров', planPercent: 65, dealersCount: 10, growth: -1, forecast: 70 },
];

const mockRisks: RiskData[] = [
  { issue: 'Падение конверсии из-за отсутствия выставочных образцов', affectedDealers: 5, impact: 2500000 },
  { issue: 'Срыв сроков поставок по кухонной группе', affectedDealers: 3, impact: 1800000 },
  { issue: 'Жалобы на качество фасадов', affectedDealers: 7, impact: 900000 },
];

const FranchiserReportTab: React.FC = () => {
  const [period, setPeriod] = useState('quarter');
  const [date, setDate] = useState('2026-Q2');
  const [blocks, setBlocks] = useState<ReportBlock[]>(defaultBlocks);
  const [executive] = useState<ExecutiveSummary>(mockExecutive);
  const [growth] = useState<NetworkGrowthData>(mockGrowth);
  const [structure] = useState<SalesStructureData[]>(mockStructure);
  const [territories] = useState<TerritoryRatingData[]>(mockTerritories);
  const [risks] = useState<RiskData[]>(mockRisks);
  const [proposals, setProposals] = useState('');
  const [historyModalOpen, setHistoryModalOpen] = useState(false);
  const [commentModalOpen, setCommentModalOpen] = useState(false);
  const [selectedBlock, setSelectedBlock] = useState<string | null>(null);
  const [comment, setComment] = useState('');

  const handleBlockToggle = (key: string) => (e: CheckboxChangeEvent) => {
    setBlocks(blocks.map(b => b.key === key ? { ...b, enabled: e.target.checked } : b));
  };

  const handleAddComment = (blockKey: string) => {
    setSelectedBlock(blockKey);
    setCommentModalOpen(true);
  };

  const handleSaveDraft = () => {
    message.success('Черновик сохранён');
  };

  const handleGeneratePdf = () => {
    message.loading('Генерация PDF...', 2);
    setTimeout(() => message.success('PDF сгенерирован'), 2000);
  };

  const handleSend = () => {
    message.loading('Отправка...', 2);
    setTimeout(() => message.success('Отчёт отправлен'), 2000);
  };

  const enabledBlocks = blocks.filter(b => b.enabled);

  const territoryColumns = [
    { title: 'Менеджер', dataIndex: 'manager', key: 'manager' },
    { title: '% плана', dataIndex: 'planPercent', key: 'planPercent', render: (v: number) => <Tag color={v >= 90 ? 'green' : v >= 70 ? 'orange' : 'red'}>{v}%</Tag> },
    { title: 'Дилеров', dataIndex: 'dealersCount', key: 'dealersCount' },
    { title: 'Прирост', dataIndex: 'growth', key: 'growth', render: (v: number) => <Text type={v > 0 ? 'success' : 'danger'}>{v > 0 ? '+' : ''}{v}</Text> },
    { title: 'Прогноз', dataIndex: 'forecast', key: 'forecast', render: (v: number) => <Tag color={v >= 90 ? 'green' : v >= 70 ? 'orange' : 'red'}>{v}%</Tag> },
  ];

  const renderBlock = (blockKey: string) => {
    switch (blockKey) {
      case 'executive':
        return (
          <Card size="small" title="Executive Summary" extra={<Button size="small" type="link" onClick={() => handleAddComment('executive')}>Комментарий</Button>}>
            <Row gutter={16}>
              <Col span={6}>
                <Statistic title="План сети" value={executive.planAmount} suffix="₽" valueStyle={{ fontSize: 18 }} />
              </Col>
              <Col span={6}>
                <Statistic title="Факт" value={executive.factAmount} suffix="₽" valueStyle={{ fontSize: 18 }} />
                <Text type="secondary">{Math.round(executive.factAmount / executive.planAmount * 100)}%</Text>
              </Col>
              <Col span={6}>
                <Statistic title="Прогноз" value={executive.forecastAmount} suffix="₽" valueStyle={{ fontSize: 18 }} />
                <Text>{executive.forecastPercent}%</Text>
              </Col>
              <Col span={6}>
                <Space>
                  <Text strong>Тренд:</Text>
                  <Text mark={executive.trend === 'up' ? '🟢' : executive.trend === 'down' ? '🔴' : '🟡'}>
                    {executive.trend === 'up' ? 'Рост' : executive.trend === 'down' ? 'Падение' : 'Стабильно'}
                  </Text>
                </Space>
              </Col>
            </Row>
            <Divider />
            <Paragraph>
              <Text strong>Вывод:</Text> {executive.conclusion}
            </Paragraph>
          </Card>
        );
      case 'planFact':
        return (
          <Card size="small" title="График План-Факт динамика" extra={<Button size="small" type="link" onClick={() => handleAddComment('planFact')}>Комментарий</Button>}>
            <Suspense fallback={<Text>Загрузка...</Text>}>
              <ReportPlanFactChart />
            </Suspense>
          </Card>
        );
      case 'growth':
        return (
          <Card size="small" title="Анализ роста сети" extra={<Button size="small" type="link" onClick={() => handleAddComment('growth')}>Комментарий</Button>}>
            <Row gutter={16}>
              <Col span={4}>
                <Statistic title="Дилеров на начало" value={growth.dealersStart} />
              </Col>
              <Col span={4}>
                <Statistic title="Привлечено новых" value={growth.newDealers} valueStyle={{ color: '#52c41a' }} />
              </Col>
              <Col span={4}>
                <Statistic title="Выбыло" value={growth.churnedDealers} valueStyle={{ color: '#ff4d4f' }} />
              </Col>
              <Col span={4}>
                <Statistic title="Дилеров на конец" value={growth.dealersEnd} />
              </Col>
              <Col span={4}>
                <Statistic title="Средняя выручка" value={growth.avgRevenuePerDealer} suffix="₽" valueStyle={{ fontSize: 18 }} />
                <Text type={growth.revenueDynamics > 0 ? 'success' : 'danger'}>
                  {growth.revenueDynamics > 0 ? '+' : ''}{growth.revenueDynamics}%
                </Text>
              </Col>
              <Col span={4}>
                <Statistic title="Прирост сети" value={growth.dealersEnd - growth.dealersStart} suffix="дилеров" valueStyle={{ fontSize: 18 }} />
              </Col>
            </Row>
          </Card>
        );
      case 'structure':
        return (
          <Card size="small" title="Структура продаж" extra={<Button size="small" type="link" onClick={() => handleAddComment('structure')}>Комментарий</Button>}>
            <Table
              dataSource={structure}
              columns={[
                { title: 'Категория', dataIndex: 'category', key: 'category' },
                { title: 'Доля', dataIndex: 'share', key: 'share', render: (v: number) => `${v}%` },
                { title: 'Динамика', dataIndex: 'dynamics', key: 'dynamics', render: (v: number) => <Text type={v > 0 ? 'success' : v < 0 ? 'danger' : 'secondary'}>{v > 0 ? '+' : ''}{v}%</Text> },
              ]}
              rowKey="category"
              pagination={false}
              size="small"
            />
            <Paragraph>
              <Text type="secondary">Вывод: Растёт доля кухонь — соответствует стратегическому приоритету</Text>
            </Paragraph>
          </Card>
        );
      case 'territories':
        return (
          <Card size="small" title="Рейтинг территорий" extra={<Button size="small" type="link" onClick={() => handleAddComment('territories')}>К��мментарий</Button>}>
            <Table
              dataSource={territories.sort((a, b) => b.planPercent - a.planPercent)}
              columns={territoryColumns}
              rowKey="manager"
              pagination={false}
              size="small"
            />
            <Divider />
            <Space>
              <Tag color="green">Топ-3</Tag>: {territories.slice(0, 3).map(t => t.manager).join(', ')}
            </Space>
          </Card>
        );
      case 'risks':
        return (
          <Card size="small" title="Ключевые риски" extra={<Button size="small" type="link" onClick={() => handleAddComment('risks')}>Комментарий</Button>}>
            <Table
              dataSource={risks}
              columns={[
                { title: 'Проблема', dataIndex: 'issue', key: 'issue' },
                { title: 'Дилера', dataIndex: 'affectedDealers', key: 'affectedDealers' },
                { title: 'Влияние', dataIndex: 'impact', key: 'impact', render: (v: number) => <Text type="danger">{v.toLocaleString()} ₽</Text> },
              ]}
              rowKey="issue"
              pagination={false}
              size="small"
            />
          </Card>
        );
      case 'proposals':
        return (
          <Card size="small" title="Предложения и запрос ресурсов" extra={<Button size="small" type="link" onClick={() => handleAddComment('proposals')}>Комментарий</Button>}>
            <TextArea 
              value={proposals} 
              onChange={(e) => setProposals(e.target.value)} 
              placeholder="Что нужно от руководства (бюджет, изменение условий, кадровые решения)..."
              rows={4}
              style={{ width: '100%' }}
            />
          </Card>
        );
      default:
        return null;
    }
  };

  return (
    <div>
      <Title level={4}>Отчёт для B2B</Title>

      <Space style={{ marginBottom: 16 }} wrap>
        <Select value={period} onChange={setPeriod} style={{ width: 120 }}>
          <Option value="month">Месяц</Option>
          <Option value="quarter">Квартал</Option>
          <Option value="year">Год</Option>
        </Select>
        <Select value={date} onChange={setDate} style={{ width: 120 }}>
          <Option value="2026-Q1">Q1 2026</Option>
          <Option value="2026-Q2">Q2 2026</Option>
          <Option value="2026-Q3">Q3 2026</Option>
          <Option value="2026-Q4">Q4 2026</Option>
        </Select>
        <Button icon={<HistoryOutlined />} onClick={() => setHistoryModalOpen(true)}>История</Button>
      </Space>

      <Card style={{ marginBottom: 16 }}>
        <Text strong>Блоки отчёта:</Text>
        <Checkbox.Group style={{ marginLeft: 16 }}>
          {blocks.map(block => (
            <Checkbox key={block.key} checked={block.enabled} onChange={handleBlockToggle(block.key)}>
              {block.title}
            </Checkbox>
          ))}
        </Checkbox.Group>
      </Card>

      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<FilePdfOutlined />} onClick={handleGeneratePdf}>Сформировать PDF</Button>
        <Button icon={<MailOutlined />} onClick={handleSend}>Отправить руководителю</Button>
        <Button icon={<SaveOutlined />} onClick={handleSaveDraft}>Сохранить черновик</Button>
      </Space>

      <Divider />

      {enabledBlocks.map(block => (
        <div key={block.key} style={{ marginBottom: 16 }}>
          {renderBlock(block.key)}
        </div>
      ))}

      <Modal
        title="История отчётов"
        open={historyModalOpen}
        onCancel={() => setHistoryModalOpen(false)}
        footer={null}
        width={600}
      >
        <Table
          dataSource={[
            { id: '1', date: '2026-Q1', createdAt: '2026-03-31', sent: true },
            { id: '2', date: '2025-Q4', createdAt: '2025-12-31', sent: true },
            { id: '3', date: '2025-Q3', createdAt: '2025-09-30', sent: true },
          ]}
          columns={[
            { title: 'Период', dataIndex: 'date', key: 'date' },
            { title: 'Создан', dataIndex: 'createdAt', key: 'createdAt' },
            { title: 'Статус', key: 'status', render: () => <Tag color="green">Отправлен</Tag> },
            { title: 'Действие', key: 'action', render: () => <Button size="small">Загрузить</Button> },
          ]}
          rowKey="id"
          pagination={false}
          size="small"
        />
      </Modal>

      <Modal
        title="Добавить комментарий"
        open={commentModalOpen}
        onCancel={() => setCommentModalOpen(false)}
        onOk={() => { message.success('Комментарий сохранён'); setCommentModalOpen(false); }}
      >
        <TextArea 
          value={comment} 
          onChange={(e) => setComment(e.target.value)} 
          placeholder="Комментарий к блоку..."
          rows={3}
          style={{ width: '100%' }}
        />
      </Modal>
    </div>
  );
};

export default FranchiserReportTab;
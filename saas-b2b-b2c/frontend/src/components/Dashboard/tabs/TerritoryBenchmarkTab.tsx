// src/components/Dashboard/tabs/TerritoryBenchmarkTab.tsx
import React, { useMemo, useState } from 'react';
import { Card, Row, Col, Typography, Table, Tag, Space, Statistic, Select, Segmented, Tooltip, Spin } from 'antd';
import { ShopOutlined, TrophyOutlined, StarOutlined, ArrowUpOutlined, ArrowDownOutlined, WarningOutlined, CheckCircleOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer, BarChart, Bar, Cell, ReferenceLine } from 'recharts';
import { useTerritoryManagerStore } from '@/store/territoryManagerStore';

const { Text } = Typography;

interface DealerScatter {
  id: string;
  name: string;
  revenue: number;
  margin: number;
  salons: number;
  segment: 'A' | 'B' | 'C' | 'D';
  conversion: number;
  discount: number;
}

interface SalesStructure {
  dealerName: string;
  kitchen: number;
  soft: number;
  корпусная: number;
  mattress: number;
  accessories: number;
}

interface WhiteSpot {
  dealerName: string;
  category: string;
  dealerPercent: number;
  networkPercent: number;
  recommendation: string;
}

interface InventoryRisk {
  dealerName: string;
  product: string;
  collection: string;
  stock: number;
  cost: number;
  days: number;
  risk: 'high' | 'medium' | 'low';
}

interface DealerDeficit {
  dealerName: string;
  product: string;
  currentStock: number;
  lostRevenue: number;
  recommendation: string;
}

interface DealerIndex {
  dealerId: string;
  dealerName: string;
  planPercent: number;
  conversion: number;
  margin: number;
  debtScore: number;
  reportScore: number;
  totalIndex: number;
  rank: number;
  rankChange: number;
}

interface TerritoryBenchmarkTabProps {
  loading?: boolean;
}

const SEGMENT_COLORS = { A: '#52c41a', B: '#1890ff', C: '#fa8c16', D: '#ff4d4f' };

const TerritoryBenchmarkTab: React.FC<TerritoryBenchmarkTabProps> = ({ loading }) => {
  const [period, setPeriod] = useState('month');
  const [categoryFilter, setCategoryFilter] = useState('all');
  const [segmentFilter, setSegmentFilter] = useState('all');
  const [activeBlock, setActiveBlock] = useState<'margin' | 'structure' | 'risks' | 'index'>('margin');

  const scatterData = useMemo((): DealerScatter[] => [
    { id: '1', name: 'Мебель Москва', revenue: 13800000, margin: 28, salons: 3, segment: 'A', conversion: 4.8, discount: 5 },
    { id: '2', name: 'Диванит Воронеж', revenue: 6240000, margin: 22, salons: 2, segment: 'B', conversion: 3.2, discount: 12 },
    { id: '3', name: 'МебельЛига', revenue: 2250000, margin: 15, salons: 1, segment: 'D', conversion: 1.8, discount: 18 },
    { id: '4', name: 'Салон мебели Казань', revenue: 13200000, margin: 26, salons: 2, segment: 'A', conversion: 5.2, discount: 4 },
    { id: '5', name: 'Евромебель', revenue: 4690000, margin: 19, salons: 1, segment: 'C', conversion: 2.9, discount: 15 },
  ], []);

  const structureData = useMemo((): SalesStructure[] => [
    { dealerName: 'Мебель Москва', kitchen: 35, soft: 25, корпусная: 20, mattress: 15, accessories: 5 },
    { dealerName: 'Диванит Воронеж', kitchen: 30, soft: 28, корпусная: 25, mattress: 12, accessories: 5 },
    { dealerName: 'МебельЛига', kitchen: 15, soft: 35, корпусная: 30, mattress: 10, accessories: 10 },
    { dealerName: 'Салон мебели Казань', kitchen: 38, soft: 22, корпусная: 18, mattress: 18, accessories: 4 },
    { dealerName: 'Евромебель', kitchen: 25, soft: 30, корпусная: 25, mattress: 15, accessories: 5 },
  ], []);

  const whiteSpots = useMemo((): WhiteSpot[] => [
    { dealerName: 'МебельЛига', category: 'Кухни', dealerPercent: 15, networkPercent: 30, recommendation: 'Предложить обучение по кухням' },
    { dealerName: 'МебельЛига', category: 'Матрасы', dealerPercent: 10, networkPercent: 15, recommendation: 'Проверить выставочные образцы' },
    { dealerName: 'Евромебель', category: 'Корпусная', dealerPercent: 25, networkPercent: 22, recommendation: 'ОК - в норме' },
  ], []);

  const inventoryRisks = useMemo((): InventoryRisk[] => [
    { dealerName: 'Мебель Москва', product: 'Диван Бостон', collection: 'Бостон', stock: 15, cost: 1200000, days: 180, risk: 'high' },
    { dealerName: 'Диванит Воронеж', product: 'Кровать Прима', collection: 'Прима', stock: 8, cost: 480000, days: 120, risk: 'medium' },
    { dealerName: 'МебельЛига', product: 'Кухня Лайт', collection: 'Лайт', stock: 3, cost: 180000, days: 45, risk: 'low' },
  ], []);

  const deficits = useMemo((): DealerDeficit[] => [
    { dealerName: 'Евромебель', product: 'Диван Комфорт', currentStock: 0, lostRevenue: 350000, recommendation: 'Срочная дозаказка' },
    { dealerName: 'МебельЛига', product: 'Кровать Эталон', currentStock: 1, lostRevenue: 180000, recommendation: 'Проверить поставку' },
  ], []);

  const dealerIndexData = useMemo((): DealerIndex[] => {
    const data = scatterData.map(d => {
      const planPercent = (d.revenue / 10000000) * 100;
      const index = Math.min(100, 
        (planPercent * 0.4) + 
        (d.conversion * 5 * 0.25) + 
        (d.margin * 1.5 * 0.2) + 
        (d.discount < 10 ? 10 : 0) * 0.1 + 
        (d.salons > 1 ? 5 : 0) * 0.05
      );
      return {
        dealerId: d.id,
        dealerName: d.name,
        planPercent,
        conversion: d.conversion,
        margin: d.margin,
        debtScore: d.discount < 10 ? 10 : 5,
        reportScore: 5,
        totalIndex: Math.round(index),
      };
    });
    return data.sort((a, b) => b.totalIndex - a.totalIndex).map((d, i) => ({ ...d, rank: i + 1, rankChange: i === 0 ? 0 : Math.floor(Math.random() * 3) - 1 }));
  }, [scatterData]);

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <Card size="small">
          <Space direction="vertical">
            <Text strong>{data.name}</Text>
            <Text>Выручка: {(data.revenue / 1000000).toFixed(1)} млн ₽</Text>
            <Text>Маржа: {data.margin}%</Text>
            <Text>Скидка: {data.discount}%</Text>
            <Text>Конверсия: {data.conversion}%</Text>
            <Text>Салонов: {data.salons}</Text>
          </Space>
        </Card>
      );
    }
    return null;
  };

  const marginScatterColors = scatterData.map(d => SEGMENT_COLORS[d.segment]);

  const riskColumns = [
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName' },
    { title: 'Позиция', dataIndex: 'product', key: 'product' },
    { title: 'Коллекция', dataIndex: 'collection', key: 'collection' },
    { title: 'Остаток', dataIndex: 'stock', key: 'stock' },
    { title: 'Стоимость', dataIndex: 'cost', key: 'cost', render: (v: number) => `${(v / 1000).toFixed(0)}k ₽` },
    { title: 'Дней', dataIndex: 'days', key: 'days' },
    { title: 'Риск', dataIndex: 'risk', key: 'risk', render: (r: string) => <Tag color={r === 'high' ? 'red' : r === 'medium' ? 'orange' : 'green'}>{r === 'high' ? 'Высокий' : r === 'medium' ? 'Средний' : 'Низкий'}</Tag> },
  ];

  const deficitColumns = [
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName' },
    { title: 'Позиция', dataIndex: 'product', key: 'product' },
    { title: 'Остаток', dataIndex: 'currentStock', key: 'currentStock' },
    { title: 'Упущенная выручка', dataIndex: 'lostRevenue', key: 'lostRevenue', render: (v: number) => `${(v / 1000).toFixed(0)}k ₽` },
    { title: '', dataIndex: 'recommendation', key: 'recommendation', render: (r: string) => <Tag color="red">{r}</Tag> },
  ];

  const whiteSpotColumns = [
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName' },
    { title: 'Категория', dataIndex: 'category', key: 'category' },
    { title: 'Доля дилера', dataIndex: 'dealerPercent', key: 'dealerPercent', render: (v: number) => `${v}%` },
    { title: 'Доля сети', dataIndex: 'networkPercent', key: 'networkPercent', render: (v: number) => `${v}%` },
    { title: 'Рекомендация', dataIndex: 'recommendation', key: 'recommendation', render: (r: string) => <Text type="secondary">{r}</Text> },
  ];

  const indexColumns = [
    { title: '#', dataIndex: 'rank', key: 'rank', render: (r: number) => r <= 3 ? <TrophyOutlined style={{ color: r === 1 ? '#ffd700' : r === 2 ? '#c0c0c0' : '#cd7f32' }} /> : r },
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName' },
    { title: '% плана', dataIndex: 'planPercent', key: 'planPercent', render: (v: number) => <Tag color={v >= 90 ? 'green' : v >= 70 ? 'orange' : 'red'}>{v.toFixed(0)}%</Tag> },
    { title: 'Конверсия', dataIndex: 'conversion', key: 'conversion', render: (v: number) => `${v}%` },
    { title: 'Маржа', dataIndex: 'margin', key: 'margin', render: (v: number) => `${v}%` },
    { title: 'Дебиторка', dataIndex: 'debtScore', key: 'debtScore', render: (v: number) => <Tag color={v >= 8 ? 'green' : 'orange'}>{v}</Tag> },
    { title: 'Отчёты', dataIndex: 'reportScore', key: 'reportScore' },
    { title: 'Индекс', dataIndex: 'totalIndex', key: 'totalIndex', render: (v: number) => <Text strong style={{ color: v >= 80 ? '#52c41a' : v >= 60 ? '#1890ff' : '#fa8c16' }}>{v}</Text> },
    { title: 'Динамика', key: 'rankChange', render: (_: any, r: DealerIndex) => r.rankChange > 0 ? <ArrowUpOutlined style={{ color: '#52c41a' }} /> : r.rankChange < 0 ? <ArrowDownOutlined style={{ color: '#ff4d4f' }} /> : <Text>-</Text> },
  ];

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col>
            <Select value={period} onChange={setPeriod} style={{ width: 120 }}>
              <Select.Option value="month">Месяц</Select.Option>
              <Select.Option value="quarter">Квартал</Select.Option>
              <Select.Option value="year">Год</Select.Option>
            </Select>
          </Col>
          <Col>
            <Select value={categoryFilter} onChange={setCategoryFilter} style={{ width: 140 }}>
              <Select.Option value="all">Все категории</Select.Option>
              <Select.Option value="kitchen">Кухни</Select.Option>
              <Select.Option value="soft">Мягкая мебель</Select.Option>
              <Select.Option value="mattress">Матрасы</Select.Option>
            </Select>
          </Col>
          <Col>
            <Select value={segmentFilter} onChange={setSegmentFilter} style={{ width: 120 }}>
              <Select.Option value="all">Все сегменты</Select.Option>
              <Select.Option value="A">A-дилеры</Select.Option>
              <Select.Option value="B">B-дилеры</Select.Option>
              <Select.Option value="C">C-дилеры</Select.Option>
              <Select.Option value="D">D-дилеры</Select.Option>
            </Select>
          </Col>
          <Col>
            <Segmented value={activeBlock} onChange={(v) => setActiveBlock(v as any)} options={[
              { label: 'Маржинальность', value: 'margin' },
              { label: 'Структура', value: 'structure' },
              { label: 'Риски', value: 'risks' },
              { label: 'Рейтинг', value: 'index' },
            ]} />
          </Col>
        </Row>
      </Card>

      {activeBlock === 'margin' && (
        <Card size="small" title="Сравнение дилеров по маржинальности">
          <ResponsiveContainer width="100%" height={400}>
            <ScatterChart>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis type="number" dataKey="revenue" name="Выручка" unit="₽" tickFormatter={(v) => `${(v / 1000000).toFixed(1)}M`} />
              <YAxis type="number" dataKey="margin" name="Маржа" unit="%" tickFormatter={(v) => `${v}%`} />
              <RechartsTooltip content={<CustomTooltip />} />
              <Legend />
              {Object.entries(SEGMENT_COLORS).map(([segment, color]) => (
                <Scatter
                  key={segment}
                  name={`${segment}-дилеры`}
                  data={scatterData.filter(d => d.segment === segment)}
                  fill={color}
                  shape={(props: any) => {
                    const { cx, cy, payload } = props;
                    const size = 20 + (payload.salons * 15);
                    return <circle cx={cx} cy={cy} r={size / 2} fill={color} opacity={0.7} />;
                  }}
                />
              ))}
              <ReferenceLine x={7000000} stroke="#d9d9d9" strokeDasharray="3 3" label={{ value: 'Средняя выручка', position: 'top' }} />
              <ReferenceLine y={22} stroke="#d9d9d9" strokeDasharray="3 3" label={{ value: 'Средняя маржа', position: 'right' }} />
            </ScatterChart>
          </ResponsiveContainer>
          <Row gutter={16} style={{ marginTop: 16 }}>
            <Col><Space><Cell fill="#52c41a" />A-дилеры</Space></Col>
            <Col><Space><Cell fill="#1890ff" />B-дилеры</Space></Col>
            <Col><Space><Cell fill="#fa8c16" />C-дилеры</Space></Col>
            <Col><Space><Cell fill="#ff4d4f" />D-дилеры</Space></Col>
          </Row>
        </Card>
      )}

      {activeBlock === 'structure' && (
        <>
          <Card size="small" title="Структура продаж по категориям" style={{ marginBottom: 16 }}>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={structureData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="dealerName" />
                <YAxis />
                <RechartsTooltip />
                <Legend />
                <Bar dataKey="kitchen" stackId="a" fill="#1890ff" name="Кухни" />
                <Bar dataKey="soft" stackId="a" fill="#52c41a" name="Мягкая" />
                <Bar dataKey="корпусная" stackId="a" fill="#722ed1" name="Корпусная" />
                <Bar dataKey="mattress" stackId="a" fill="#fa8c16" name="Матрасы" />
                <Bar dataKey="accessories" stackId="a" fill="#d9d9d9" name="Аксессуары" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
          <Card size="small" title="Белые пятна">
            <Table
              dataSource={whiteSpots}
              columns={whiteSpotColumns}
              rowKey={(r) => `${r.dealerName}-${r.category}`}
              size="small"
              pagination={false}
            />
          </Card>
        </>
      )}

      {activeBlock === 'risks' && (
        <>
          <Card size="small" title="Затоваренность" style={{ marginBottom: 16 }}>
            <Table
              dataSource={inventoryRisks}
              columns={riskColumns}
              rowKey={(r) => `${r.dealerName}-${r.product}`}
              size="small"
              pagination={false}
            />
          </Card>
          <Card size="small" title="Дефицит">
            <Table
              dataSource={deficits}
              columns={deficitColumns}
              rowKey={(r) => `${r.dealerName}-${r.product}`}
              size="small"
              pagination={false}
            />
          </Card>
        </>
      )}

      {activeBlock === 'index' && (
        <Card size="small" title="Рейтинг дилеров по комплексному показателю">
          <Table
            dataSource={dealerIndexData}
            columns={indexColumns}
            rowKey="dealerId"
            size="small"
            pagination={false}
            summary={() => (
              <Table.Summary>
                <Table.Summary.Row>
                  <Table.Summary.Cell index={0}><Text strong>Формула индекса:</Text></Table.Summary.Cell>
                  <Table.Summary.Cell colSpan={7}>
                    <Text type="secondary">
                      % плана (40%) + Конверсия (25%) + Маржа (20%) + Дебиторка (10%) + Отчёты (5%)
                    </Text>
                  </Table.Summary.Cell>
                </Table.Summary.Row>
              </Table.Summary>
            )}
          />
        </Card>
      )}
    </div>
  );
};

export default TerritoryBenchmarkTab;
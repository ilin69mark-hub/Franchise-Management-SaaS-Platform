// src/components/Dashboard/tabs/TerritoryMapTab.tsx
import React, { useEffect, useMemo, useState } from 'react';
import { Card, Row, Col, Typography, Table, Tag, Space, Statistic, Input, Select, Collapse, List, Avatar, Tooltip, Progress, Segmented, Spin, Button } from 'antd';
import { ShopOutlined, WarningOutlined, CheckCircleOutlined, ClockCircleOutlined, SearchOutlined, ArrowUpOutlined, ArrowDownOutlined, DollarOutlined, PercentageOutlined, RiseOutlined, UserOutlined, AlertOutlined, LinkOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { useTerritoryManagerStore, DealerMetrics } from '@/store/territoryManagerStore';

const { Text } = Typography;
const { Search } = Input;

interface DealerDetail {
  dealerId: string;
  dealerName: string;
  salons: { name: string; sales: number }[];
  salesHistory: number[];
  recentAlerts: { title: string; date: string }[];
}

interface TerritoryMapTabProps {
  dealers?: DealerMetrics[];
  loading?: boolean;
}

const TerritoryMapTab: React.FC<TerritoryMapTabProps> = ({ dealers: initialDealers, loading: initialLoading }) => {
  const { dealers: storeDealers, setDealers, summary } = useTerritoryManagerStore();
  const dealers = initialDealers || storeDealers;
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [expandedDealer, setExpandedDealer] = useState<string | null>(null);
  const [detailData, setDetailData] = useState<DealerDetail | null>(null);
  const [sortField, setSortField] = useState<string>('planPercent');
  const [sortOrder, setSortOrder] = useState<'ascend' | 'descend'>('descend');

  useEffect(() => {
    if (initialDealers) {
      setDealers(initialDealers);
    }
  }, [initialDealers, setDealers]);

  const filteredDealers = useMemo(() => {
    return dealers.filter(d => {
      const matchesSearch = d.dealerName.toLowerCase().includes(searchText.toLowerCase());
      const matchesStatus = statusFilter === 'all' || d.status === statusFilter;
      const matchesLeader = statusFilter === 'leaders' ? d.planPercent >= 100 : true;
      const matchesProblem = statusFilter === 'problem' ? d.planPercent < 70 : true;
      return matchesSearch && matchesStatus && (statusFilter !== 'leaders' || matchesLeader) && (statusFilter !== 'problem' || matchesProblem);
    });
  }, [dealers, searchText, statusFilter]);

  const sortedDealers = useMemo(() => {
    return [...filteredDealers].sort((a, b) => {
      const aVal = a[sortField as keyof DealerMetrics] as number;
      const bVal = b[sortField as keyof DealerMetrics] as number;
      return sortOrder === 'ascend' ? aVal - bVal : bVal - aVal;
    });
  }, [filteredDealers, sortField, sortOrder]);

  const territoryTotals = useMemo(() => {
    const totalPlan = dealers.reduce((s, d) => s + (d.plan || 0), 0);
    const totalFact = dealers.reduce((s, d) => s + (d.fact || 0), 0);
    const avgConversion = dealers.length ? dealers.reduce((s, d) => s + d.conversion, 0) / dealers.length : 0;
    const avgMargin = dealers.length ? dealers.reduce((s, d) => s + d.margin, 0) / dealers.length : 0;
    const redZoneCount = dealers.filter(d => d.status === 'red').length;
    const totalDebt = dealers.reduce((s, d) => s + d.debt, 0);
    return { totalPlan, totalFact, avgConversion, avgMargin, redZoneCount, totalDebt };
  }, [dealers]);

  const planCompletionPercent = territoryTotals.totalPlan > 0 ? (territoryTotals.totalFact / territoryTotals.totalPlan) * 100 : 0;
  const forecastPercent = 92;

  const handleSort = (field: string) => {
    if (field === sortField) {
      setSortOrder(sortOrder === 'ascend' ? 'descend' : 'ascend');
    } else {
      setSortField(field);
      setSortOrder('descend');
    }
  };

  const fetchDealerDetails = async (dealerId: string) => {
    try {
      const token = localStorage.getItem('accessToken');
      const res = await fetch(`/api/franchiser/dealers/${dealerId}/details`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();
        setDetailData(data);
      }
    } catch (e) {
      setDetailData({
        dealerId,
        dealerName: dealers.find(d => d.dealerId === dealerId)?.dealerName || '',
        salons: [
          { name: 'Салон 1', sales: 4200000 },
          { name: 'Салон 2', sales: 2800000 },
        ],
        salesHistory: [3200, 4100, 3800, 4500, 4200, 4800],
        recentAlerts: [
          { title: 'Падение конверсии', date: '2026-04-28' },
          { title: 'Низкий трафик', date: '2026-04-25' },
        ],
      });
    }
  };

  const handleExpand = (dealerId: string) => {
    if (expandedDealer === dealerId) {
      setExpandedDealer(null);
      setDetailData(null);
    } else {
      setExpandedDealer(dealerId);
      fetchDealerDetails(dealerId);
    }
  };

  const getStatusColor = (dealer: DealerMetrics) => {
    if (dealer.planPercent >= 90 && dealer.conversion >= 3) return '#52c41a';
    if (dealer.planPercent >= 70 && dealer.conversion >= 2) return '#fa8c16';
    return '#ff4d4f';
  };

  const getCellColor = (field: string, value: number) => {
    switch (field) {
      case 'planPercent':
        return value >= 90 ? '#f6ffed' : value >= 70 ? '#fffbe6' : '#fff1f0';
      case 'conversion':
        return value >= 3 ? '#f6ffed' : value >= 2 ? '#fffbe6' : '#fff1f0';
      case 'margin':
        return value >= 25 ? '#f6ffed' : value >= 20 ? '#fffbe6' : '#fff1f0';
      default:
        return 'transparent';
    }
  };

  const columns = [
    {
      title: 'Дилер',
      dataIndex: 'dealerName',
      key: 'dealerName',
      sorter: true,
      render: (name: string, record: DealerMetrics) => (
        <Space>
          <ShopOutlined style={{ color: getStatusColor(record) }} />
          <Text strong>{name}</Text>
          {record.status === 'red' && <WarningOutlined style={{ color: '#ff4d4f' }} />}
        </Space>
      ),
    },
    {
      title: () => (
        <Text onClick={() => handleSort('planPercent')}>
          План % {sortField === 'planPercent' && (sortOrder === 'ascend' ? <ArrowUpOutlined /> : <ArrowDownOutlined />)}
        </Text>
      ),
      dataIndex: 'planPercent',
      key: 'planPercent',
      sorter: true,
      render: (percent: number) => (
        <div style={{ background: getCellColor('planPercent', percent), padding: '4px 8px', borderRadius: 4 }}>
          <Tag color={percent >= 100 ? 'green' : percent >= 70 ? 'orange' : 'red'}>
            {percent}%
          </Tag>
        </div>
      ),
    },
    {
      title: () => (
        <Text onClick={() => handleSort('conversion')}>
          Конверсия {sortField === 'conversion' && (sortOrder === 'ascend' ? <ArrowUpOutlined /> : <ArrowDownOutlined />)}
        </Text>
      ),
      dataIndex: 'conversion',
      key: 'conversion',
      sorter: true,
      render: (conv: number) => (
        <div style={{ background: getCellColor('conversion', conv), padding: '4px 8px', borderRadius: 4 }}>
          <Text>{conv.toFixed(1)}%</Text>
        </div>
      ),
    },
    {
      title: 'Ср. чек',
      dataIndex: 'avgCheck',
      key: 'avgCheck',
      render: (v: number) => <Text>{(v / 1000).toFixed(0)}k ₽</Text>,
    },
    {
      title: () => (
        <Text onClick={() => handleSort('margin')}>
          Маржа {sortField === 'margin' && (sortOrder === 'ascend' ? <ArrowUpOutlined /> : <ArrowDownOutlined />)}
        </Text>
      ),
      dataIndex: 'margin',
      key: 'margin',
      sorter: true,
      render: (v: number) => (
        <div style={{ background: getCellColor('margin', v), padding: '4px 8px', borderRadius: 4 }}>
          <Text>{v.toFixed(1)}%</Text>
        </div>
      ),
    },
    {
      title: 'Дебиторка',
      dataIndex: 'debt',
      key: 'debt',
      render: (v: number) => (
        <Tooltip title={v > 100000 ? 'Превышен порог!' : ''}>
          <Text style={{ color: v > 100000 ? '#ff4d4f' : undefined }}>
            {v > 0 ? `${(v / 1000).toFixed(0)}k ₽` : '-'}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: 'Задачи',
      dataIndex: 'taskCount',
      key: 'taskCount',
      render: (count: number) => count > 0 ? <Tag color="blue">{count}</Tag> : <Text type="secondary">-</Text>,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: 'green' | 'yellow' | 'red') => {
        const icons = { green: <CheckCircleOutlined style={{ color: '#52c41a' }} />, yellow: <ClockCircleOutlined style={{ color: '#fa8c16' }} />, red: <WarningOutlined style={{ color: '#ff4d4f' }} /> };
        const labels = { green: 'Норма', yellow: 'Внимание', red: 'Проблема' };
        return <Space>{icons[status]} {labels[status]}</Space>;
      },
    },
  ];

  const redZoneDealers = dealers.filter(d => d.status === 'red');

  const renderDetailPanel = () => {
    if (!detailData) return <Spin />;
    return (
      <Card size="small" title={`${detailData.dealerName} - детализация`} style={{ marginTop: 16 }}>
        <Row gutter={16}>
          <Col span={12}>
            <Text strong>Продажи по салонам:</Text>
            <List
              size="small"
              dataSource={detailData.salons}
              renderItem={item => (
                <List.Item>
                  <Text>{item.name}</Text>
                  <Text>{(item.sales / 1000000).toFixed(1)} млн ₽</Text>
                </List.Item>
              )}
            />
          </Col>
          <Col span={12}>
            <Text strong>Динамика (6 мес):</Text>
            <div style={{ display: 'flex', alignItems: 'flex-end', height: 60, gap: 4 }}>
              {detailData.salesHistory.map((v, i) => (
                <div key={i} style={{ flex: 1, background: '#1890ff', height: `${(v / 5000)}%`, borderRadius: 2 }} />
              ))}
            </div>
            <Text type="secondary" style={{ fontSize: 11 }}>млн ₽</Text>
          </Col>
        </Row>
        <Row style={{ marginTop: 16 }}>
          <Col span={24}>
            <Text strong>Последние алерты:</Text>
            <List
              size="small"
              dataSource={detailData.recentAlerts}
              renderItem={item => (
                <List.Item>
                  <Space>
                    <AlertOutlined style={{ color: '#ff4d4f' }} />
                    <Text>{item.title}</Text>
                    <Text type="secondary">{item.date}</Text>
                  </Space>
                </List.Item>
              )}
            />
          </Col>
        </Row>
        <Space style={{ marginTop: 16 }}>
          <a href={`/dealer/${detailData.dealerId}`}><Button icon={<LinkOutlined />}>Перейти к дилеру</Button></a>
        </Space>
      </Card>
    );
  };

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col flex="auto">
            <Search placeholder="Поиск дилера..." allowClear onChange={e => setSearchText(e.target.value)} />
          </Col>
          <Col>
            <Segmented value={statusFilter} onChange={setStatusFilter} options={[
              { label: 'Все', value: 'all' },
              { label: 'Проблемные', value: 'problem' },
              { label: 'Лидеры', value: 'leaders' },
            ]} />
          </Col>
        </Row>
      </Card>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8}>
          <Card size="small">
            <Statistic 
              title="Выполнение плана" 
              value={territoryTotals.totalFact / 1000000} 
              prefix="₽ "
              suffix="млн"
              valueStyle={{ fontSize: 18 }}
            />
            <Text type="secondary">
              {planCompletionPercent.toFixed(0)}% <ArrowUpOutlined style={{ color: '#52c41a' }} /> +5% к прошлому месяцу
            </Text>
          </Card>
        </Col>
        <Col xs={12} sm={8}>
          <Card size="small">
            <Statistic 
              title="Прогноз квартала" 
              value={forecastPercent} 
              suffix="%"
              prefix={<RiseOutlined />}
              valueStyle={{ fontSize: 18 }}
            />
            <Text type="secondary">Run Rate</Text>
          </Card>
        </Col>
        <Col xs={12} sm={8}>
          <Card size="small" style={{ cursor: 'pointer' }} onClick={() => setStatusFilter('problem')}>
            <Statistic 
              title="В красной зоне" 
              value={territoryTotals.redZoneCount} 
              valueStyle={{ fontSize: 18, color: territoryTotals.redZoneCount > 0 ? '#ff4d4f' : '#52c41a' }}
              prefix={<WarningOutlined />}
            />
            <Text type="secondary">&lt; 70% плана</Text>
          </Card>
        </Col>
        <Col xs={12} sm={8}>
          <Card size="small">
            <Statistic 
              title="Дебиторская задолженность" 
              value={territoryTotals.totalDebt / 1000000} 
              prefix="₽ "
              suffix="млн"
              valueStyle={{ fontSize: 18, color: territoryTotals.totalDebt > 500000 ? '#ff4d4f' : undefined }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8}>
          <Card size="small">
            <Statistic 
              title="Конверсия средняя" 
              value={territoryTotals.avgConversion} 
              suffix="%"
              precision={1}
              valueStyle={{ fontSize: 18 }}
            />
            <Text type="secondary">норма: 3%</Text>
          </Card>
        </Col>
        <Col xs={12} sm={8}>
          <Card size="small">
            <Statistic 
              title="Маржинальность" 
              value={territoryTotals.avgMargin} 
              suffix="%"
              precision={1}
              valueStyle={{ fontSize: 18 }}
            />
          </Card>
        </Col>
      </Row>

      {redZoneDealers.length > 0 && expandedDealer === null && (
        <Collapse style={{ marginBottom: 16 }}>
          <Collapse.Panel header={`⚠️ Дилеры в красной зоне (${redZoneDealers.length})`} key="red">
            <List
              size="small"
              dataSource={redZoneDealers}
              renderItem={d => (
                <List.Item>
                  <Space>
                    <WarningOutlined style={{ color: '#ff4d4f' }} />
                    <Text>{d.dealerName}</Text>
                    <Tag color="red">{d.planPercent}% плана</Tag>
                    <Button size="small" onClick={() => handleExpand(d.dealerId)}>Детали</Button>
                  </Space>
                </List.Item>
              )}
            />
          </Collapse.Panel>
        </Collapse>
      )}

      <Card size="small" title="Теплокарта дилер��в">
        <Table
          dataSource={sortedDealers}
          columns={columns}
          rowKey="dealerId"
          size="small"
          loading={initialLoading}
          pagination={{ pageSize: 10 }}
          expandable={{
            expandedRowRender: () => renderDetailPanel(),
            rowExpandable: () => true,
          }}
          expandedRowKeys={[expandedDealer]}
          onExpand={(expanded, record) => handleExpand(record.dealerId)}
        />
      </Card>
    </div>
  );
};

export default TerritoryMapTab;
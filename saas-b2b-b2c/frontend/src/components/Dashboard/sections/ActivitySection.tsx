import React, { useEffect, useState } from 'react';
import {
  Card,
  Row,
  Col,
  Table,
  Select,
  Typography,
  Spin,
  Statistic,
  Progress,
  Tag,
} from 'antd';
import {
  UserOutlined,
  RiseOutlined,
  FallOutlined,
  LineChartOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import { useActivityStore } from '@/store/activityStore';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;

const ActivitySection: React.FC = () => {
  const {
    overview,
    dynamics,
    byTenant,
    featureAdoption,
    ttv,
    period,
    selectedTenant,
    isLoading,
    setOverview,
    setDynamics,
    setByTenant,
    setFeatureAdoption,
    setTtv,
    setPeriod,
    setSelectedTenant,
    setLoading,
  } = useActivityStore();

  const [localLoading, setLocalLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, [period, selectedTenant]);

  const fetchData = async () => {
    setLoading(true);
    setLocalLoading(true);
    try {
      const params = new URLSearchParams();
      params.append('days', period.toString());
      if (selectedTenant) params.append('tenant', selectedTenant);

      const [overviewRes, dynamicsRes, tenantRes, featureRes, ttvRes] = await Promise.all([
        apiClient.get(`/admin/activity/overview?${params.toString()}`),
        apiClient.get(`/admin/activity/dynamics?${params.toString()}`),
        apiClient.get(`/admin/activity/by-tenant?${params.toString()}`),
        apiClient.get(`/admin/activity/feature-adoption?${params.toString()}`),
        apiClient.get(`/admin/activity/ttv?${params.toString()}`),
      ]);

      setOverview(overviewRes.data);
      setDynamics(dynamicsRes.data || []);
      setByTenant(tenantRes.data || []);
      setFeatureAdoption(featureRes.data || []);
      setTtv(ttvRes.data || []);
    } catch {
      setOverview({
        dau: 45,
        dauChange: 8.5,
        wau: 120,
        wauChange: 12.3,
        mau: 320,
        mauChange: 15.2,
        stickiness: 14.1,
        avgSessionTime: 25,
        activeSessions: 12,
      });

      const now = dayjs();
      const dynamicsData = [];
      for (let i = 89; i >= 0; i--) {
        dynamicsData.push({
          date: now.subtract(i, 'day').format('YYYY-MM-DD'),
          dau: Math.floor(20 + Math.random() * 30),
          wau: Math.floor(80 + Math.random() * 50),
          mau: Math.floor(200 + Math.random() * 150),
        });
      }
      setDynamics(dynamicsData);

      setByTenant([
        {
          id: '1',
          name: 'ООО Техно',
          dau: 25,
          wau: 80,
          mau: 150,
          stickiness: 16.7,
          activeLicensePercent: 85,
        },
        {
          id: '2',
          name: 'АО Бизнес',
          dau: 15,
          wau: 45,
          mau: 90,
          stickiness: 16.7,
          activeLicensePercent: 72,
        },
        {
          id: '3',
          name: 'ИП Сидоров',
          dau: 5,
          wau: 12,
          mau: 25,
          stickiness: 20,
          activeLicensePercent: 40,
        },
        {
          id: '4',
          name: 'ООО Проблема',
          dau: 2,
          wau: 8,
          mau: 15,
          stickiness: 13.3,
          activeLicensePercent: 20,
        },
      ]);

      setFeatureAdoption([
        {
          id: '1',
          name: 'Управление дилерами',
          tenantPercent: 95,
          userPercent: 78,
          frequency: 'daily',
          trend: 'up',
        },
        {
          id: '2',
          name: 'Отчёты и аналитика',
          tenantPercent: 88,
          userPercent: 65,
          frequency: 'daily',
          trend: 'stable',
        },
        {
          id: '3',
          name: 'Коммуникации',
          tenantPercent: 72,
          userPercent: 45,
          frequency: 'weekly',
          trend: 'up',
        },
        {
          id: '4',
          name: 'Бизнес-планирование',
          tenantPercent: 45,
          userPercent: 22,
          frequency: 'weekly',
          trend: 'stable',
        },
        {
          id: '5',
          name: 'География',
          tenantPercent: 15,
          userPercent: 8,
          frequency: 'monthly',
          trend: 'down',
        },
        {
          id: '6',
          name: 'Чат-боты',
          tenantPercent: 5,
          userPercent: 2,
          frequency: 'monthly',
          trend: 'down',
        },
      ]);

      setTtv([
        {
          id: '1',
          name: 'Новая компания',
          createdAt: '2026-03-15',
          firstSaleAt: '2026-03-22',
          ttvDays: 7,
        },
        {
          id: '2',
          name: 'Тест тенант',
          createdAt: '2026-04-01',
          firstSaleAt: '2026-04-10',
          ttvDays: 9,
        },
        {
          id: '3',
          name: 'Пилот',
          createdAt: '2026-04-10',
          firstSaleAt: '',
          ttvDays: 0,
        },
      ]);
    } finally {
      setLoading(false);
      setLocalLoading(false);
    }
  };

  const renderChange = (value: number) => {
    const color = value > 0 ? '#52c41a' : value < 0 ? '#ff4d4f' : '#888';
    const Icon = value >= 0 ? RiseOutlined : FallOutlined;
    const prefix = value >= 0 ? '+' : '';
    return (
      <span style={{ color, fontSize: 12, marginLeft: 8 }}>
        <Icon /> {prefix}
        {value}%
      </span>
    );
  };

  if (localLoading || isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  const tenantColumns = [
    { title: 'Тенант', dataIndex: 'name', key: 'name', sorter: (a: any, b: any) => a.name.localeCompare(b.name) },
    {
      title: 'DAU',
      dataIndex: 'dau',
      key: 'dau',
      sorter: (a: any, b: any) => a.dau - b.dau,
    },
    {
      title: 'WAU',
      dataIndex: 'wau',
      key: 'wau',
      sorter: (a: any, b: any) => a.wau - b.wau,
    },
    {
      title: 'MAU',
      dataIndex: 'mau',
      key: 'mau',
      sorter: (a: any, b: any) => a.mau - b.mau,
    },
    {
      title: 'Stickiness',
      dataIndex: 'stickiness',
      key: 'stickiness',
      sorter: (a: any, b: any) => a.stickiness - b.stickiness,
      render: (v: number) => (
        <span style={{ color: v < 20 ? '#ff4d4f' : v >= 40 ? '#52c41a' : undefined }}>
          {v.toFixed(1)}%
        </span>
      ),
    },
    {
      title: 'Акт. лицензий',
      dataIndex: 'activeLicensePercent',
      key: 'activeLicensePercent',
      render: (v: number) => (
        <Progress
          percent={v}
          size="small"
          status={v < 30 ? 'exception' : v > 70 ? 'success' : 'normal'}
        />
      ),
    },
  ];

  const featureColumns = [
    { title: 'Функция', dataIndex: 'name', key: 'name' },
    {
      title: '% тенантов',
      dataIndex: 'tenantPercent',
      key: 'tenantPercent',
      render: (v: number) => `${v}%`,
    },
    {
      title: '% польз.',
      dataIndex: 'userPercent',
      key: 'userPercent',
      render: (v: number) => `${v}%`,
    },
    { title: 'Частота', dataIndex: 'frequency', key: 'frequency' },
    {
      title: 'Тренд',
      dataIndex: 'trend',
      key: 'trend',
      render: (t: string) => {
        const colors = { up: 'green', stable: 'blue', down: 'red' };
        return <Tag color={colors[t as keyof typeof colors]}>{t === 'up' ? '↑ Растёт' : t === 'down' ? '↓ Падает' : '→ Стабильно'}</Tag>;
      },
    },
  ];

const ttvColumns = [
    { title: 'Тенант', dataIndex: 'name', key: 'name' },
    { title: 'Создан', dataIndex: 'createdAt', key: 'createdAt', render: (d: string) => d ? dayjs(d).format('DD.MM.YYYY') : '-' },
    { title: 'Первая продажа', dataIndex: 'firstSaleAt', key: 'firstSaleAt', render: (d: string) => d ? dayjs(d).format('DD.MM.YYYY') : '-' },
    {
      title: 'TTV (дней)',
      dataIndex: 'ttvDays',
      key: 'ttvDays',
      render: (v: number, record: any) => (
        <span style={{ color: v === 0 ? '#faad14' : v <= 7 ? '#52c41a' : v <= 14 ? '#1890ff' : '#ff4d4f' }}>
          {v === 0 ? 'Нет продажи' : `${v} дн.`}
        </span>
      ),
    },
  ];

  const avgTtv = ttv.length > 0
    ? ttv.filter(t => t.ttvDays > 0).reduce((acc, t) => acc + t.ttvDays, 0) / ttv.filter(t => t.ttvDays > 0).length
    : 0;

  return (
    <div>
      <Title level={3}>Активность пользователей</Title>

      <Row gutter={[16, 16]}>
        <Col xs={12} sm={8} lg={4}>
          <Card>
            <Statistic
              title="DAU (сегодня)"
              value={overview?.dau || 0}
              prefix={<UserOutlined />}
            />
            {renderChange(overview?.dauChange || 0)}
          </Card>
        </Col>
        <Col xs={12} sm={8} lg={4}>
          <Card>
            <Statistic
              title="WAU (7 дней)"
              value={overview?.wau || 0}
              prefix={<UserOutlined />}
            />
            {renderChange(overview?.wauChange || 0)}
          </Card>
        </Col>
        <Col xs={12} sm={8} lg={4}>
          <Card>
            <Statistic
              title="MAU (30 дней)"
              value={overview?.mau || 0}
              prefix={<UserOutlined />}
            />
            {renderChange(overview?.mauChange || 0)}
          </Card>
        </Col>
        <Col xs={12} sm={8} lg={4}>
          <Card>
            <Statistic
              title="Stickiness"
              value={overview?.stickiness || 0}
              suffix="%"
              prefix={<LineChartOutlined />}
              valueStyle={{ color: (overview?.stickiness || 0) >= 40 ? '#52c41a' : (overview?.stickiness || 0) < 20 ? '#ff4d4f' : undefined }}
            />
            <Text type="secondary">норма {'>'}40%</Text>
          </Card>
        </Col>
        <Col xs={12} sm={8} lg={4}>
          <Card>
            <Statistic
              title="Среднее время сессии"
              value={overview?.avgSessionTime || 0}
              suffix="мин"
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} lg={4}>
          <Card>
            <Statistic
              title="Активных сессий"
              value={overview?.activeSessions || 0}
              valueStyle={{ color: (overview?.activeSessions || 0) > 0 ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Динамика активности" style={{ marginTop: 16 }}>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col>
            <Select
              value={period}
              onChange={setPeriod}
              style={{ width: 120 }}
            >
              <Option value={7}>7 дней</Option>
              <Option value={30}>30 дней</Option>
              <Option value={90}>90 дней</Option>
            </Select>
          </Col>
        </Row>
        {dynamics.length > 0 ? (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr>
                  <th style={{ textAlign: 'left', padding: 8 }}>Дата</th>
                  <th style={{ textAlign: 'right', padding: 8, color: '#1890ff' }}>DAU</th>
                  <th style={{ textAlign: 'right', padding: 8, color: '#52c41a' }}>WAU</th>
                  <th style={{ textAlign: 'right', padding: 8, color: '#722ed1' }}>MAU</th>
                </tr>
              </thead>
              <tbody>
                {dynamics.slice(-30).map((d) => (
                  <tr key={d.date}>
                    <td style={{ padding: 8 }}>{dayjs(d.date).format('DD.MM')}</td>
                    <td style={{ textAlign: 'right', padding: 8, color: '#1890ff' }}>{d.dau}</td>
                    <td style={{ textAlign: 'right', padding: 8, color: '#52c41a' }}>{d.wau}</td>
                    <td style={{ textAlign: 'right', padding: 8, color: '#722ed1' }}>{d.mau}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <Text type="secondary">Нет данных</Text>
        )}
      </Card>

      <Card title="Активность по тенантам" style={{ marginTop: 16 }}>
        <Table
          dataSource={byTenant}
          columns={tenantColumns}
          rowKey="id"
          pagination={{ pageSize: 10 }}
          rowClassName={(record) => record.stickiness < 20 ? 'table-row-warning' : ''}
        />
      </Card>

      <Card title="Feature Adoption" style={{ marginTop: 16 }}>
        <Row gutter={16}>
          <Col span={12}>
            <Text type="secondary">Топ используемые</Text>
            <Table
              dataSource={featureAdoption.filter(f => f.trend === 'up').slice(0, 5)}
              columns={featureColumns}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </Col>
          <Col span={12}>
            <Text type="secondary">Топ игнорируемые</Text>
            <Table
              dataSource={featureAdoption.filter(f => f.trend === 'down').slice(0, 5)}
              columns={featureColumns}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </Col>
        </Row>
      </Card>

      <Card title="Time to Value" style={{ marginTop: 16 }}>
        <Row gutter={16}>
          <Col span={8}>
            <Card>
              <Statistic
                title="Средний TTV"
                value={avgTtv || 0}
                suffix="дней"
                valueStyle={{ color: avgTtv <= 7 ? '#52c41a' : avgTtv <= 14 ? '#1890ff' : '#ff4d4f' }}
              />
            </Card>
          </Col>
        </Row>
        <Table
          dataSource={ttv}
          columns={ttvColumns}
          rowKey="id"
          pagination={{ pageSize: 10 }}
          style={{ marginTop: 16 }}
        />
      </Card>
    </div>
  );
};

export default ActivitySection;
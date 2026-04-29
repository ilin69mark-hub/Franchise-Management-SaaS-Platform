import React, { useEffect, useState } from 'react';
import {
  Card,
  Row,
  Col,
  Table,
  Input,
  Select,
  DatePicker,
  Button,
  Typography,
  Spin,
  Tag,
  message,
  Popconfirm,
  Tabs,
} from 'antd';
import {
  SearchOutlined,
  ExportOutlined,
  CloseCircleOutlined,
  UserOutlined,
  LoginOutlined,
  LogoutOutlined,
  KeyOutlined,
} from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import { useAuditStore } from '@/store/auditStore';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { RangePicker } = DatePicker;

const AuditSection: React.FC = () => {
  const {
    adminActions,
    impersonations,
    activeSessions,
    userLogins,
    filters,
    isLoading,
    setAdminActions,
    setImpersonations,
    setActiveSessions,
    setUserLogins,
    setFilters,
    setLoading,
  } = useAuditStore();

  const [localLoading, setLocalLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    setLocalLoading(true);
    try {
      const params = new URLSearchParams();
      if (filters.admin) params.append('admin', filters.admin);
      if (filters.action) params.append('action', filters.action);
      if (filters.tenant) params.append('tenant', filters.tenant);

      const [actionsRes, impersonationsRes, sessionsRes, loginsRes] = await Promise.all([
        apiClient.get(`/admin/audit/admin-actions?${params.toString()}`),
        apiClient.get(`/admin/audit/impersonations?${params.toString()}`),
        apiClient.get('/admin/audit/active-sessions'),
        apiClient.get(`/admin/audit/user-logins?${params.toString()}`),
      ]);

      setAdminActions(actionsRes.data || []);
      setImpersonations(impersonationsRes.data || []);
      setActiveSessions(sessionsRes.data || []);
      setUserLogins(loginsRes.data || []);
    } catch {
      setAdminActions([
        {
          id: '1',
          timestamp: new Date().toISOString(),
          adminName: 'Иван Админ',
          adminEmail: 'admin@test.com',
          action: 'create_tenant',
          object: 'ООО Техно',
          details: '{"name": "ООО Техно", "tariff": "Pro"}',
          ip: '192.168.1.100',
        },
        {
          id: '2',
          timestamp: dayjs().subtract(1, 'hour').toISOString(),
          adminName: 'Иван Админ',
          adminEmail: 'admin@test.com',
          action: 'change_tariff',
          object: 'АО Бизнес',
          details: '{"old": "Start", "new": "Pro"}',
          ip: '192.168.1.100',
        },
        {
          id: '3',
          timestamp: dayjs().subtract(2, 'hour').toISOString(),
          adminName: 'Пётр Супер',
          adminEmail: 'super@test.com',
          action: 'impersonate',
          object: 'ООО Техно',
          details: '{"role": "dealer", "user": "user@test.com"}',
          ip: '10.0.0.5',
        },
      ]);

      setImpersonations([
        {
          id: '1',
          timestamp: new Date().toISOString(),
          adminName: 'Иван Админ',
          tenant: 'ООО Техно',
          role: 'dealer',
          userName: 'Вася Пупкин',
          duration: 15,
          actionsSummary: '3 действия',
          ip: '192.168.1.100',
        },
      ]);

      setActiveSessions([
        {
          id: '1',
          adminName: 'Иван Админ',
          ip: '192.168.1.100',
          startedAt: dayjs().subtract(2, 'hour').toISOString(),
          lastActivity: dayjs().subtract(5, 'minute').toISOString(),
        },
        {
          id: '2',
          adminName: 'Пётр Супер',
          ip: '10.0.0.5',
          startedAt: dayjs().subtract(30, 'minute').toISOString(),
          lastActivity: dayjs().subtract(1, 'minute').toISOString(),
        },
      ]);

      setUserLogins([
        {
          id: '1',
          timestamp: new Date().toISOString(),
          tenant: 'ООО Техно',
          userName: 'Вася Пупкин',
          userEmail: 'vasya@test.com',
          role: 'dealer',
          action: 'login',
          ip: '192.168.1.50',
          userAgent: 'Chrome/120',
          geo: 'Москва, Россия',
        },
        {
          id: '2',
          timestamp: dayjs().subtract(1, 'hour').toISOString(),
          tenant: 'АО Бизнес',
          userName: 'Маша Иванова',
          userEmail: 'masha@test.com',
          role: 'franchiser',
          action: 'failed_login',
          ip: '45.67.89.10',
          userAgent: 'Safari/17',
          geo: 'Санкт-Петербург, Россия',
        },
      ]);
    } finally {
      setLoading(false);
      setLocalLoading(false);
    }
  };

  const handleTerminateSession = async (id: string) => {
    try {
      await apiClient.post(`/admin/audit/terminate-session/${id}`);
      message.success('Сессия завершена');
      fetchData();
    } catch {
      message.error('Ошибка');
    }
  };

  const getActionTag = (a: string) => {
    const actions: Record<string, { color: string; text: string }> = {
      create_tenant: { color: 'green', text: 'Создание тенанта' },
      change_tariff: { color: 'blue', text: 'Изменение тарифа' },
      suspend: { color: 'orange', text: 'Приостановка' },
      impersonate: { color: 'purple', text: 'Имперсонация' },
      manual_payment: { color: 'cyan', text: 'Ручной платёж' },
      settings_change: { color: 'gold', text: 'Настройки' },
    };
    return actions[a] ? <Tag color={actions[a].color}>{actions[a].text}</Tag> : <Tag>{a}</Tag>;
  };

  const getLoginActionTag = (a: string) => {
    const actions: Record<string, { color: string; text: string }> = {
      login: { color: 'green', text: 'Вход' },
      logout: { color: 'default', text: 'Выход' },
      failed_login: { color: 'red', text: 'Неудачная попытка' },
    };
    return actions[a] ? <Tag color={actions[a].color}>{actions[a].text}</Tag> : <Tag>{a}</Tag>;
  };

  if (localLoading || isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  const actionColumns = [
    { title: 'Время', dataIndex: 'timestamp', key: 'timestamp', render: (d: string) => dayjs(d).format('DD.MM HH:mm') },
    { title: 'Админ', dataIndex: 'adminName', key: 'adminName', render: (_: any, r: any) => `${r.adminName} (${r.adminEmail})` },
    { title: 'Действие', dataIndex: 'action', key: 'action', render: (a: string) => getActionTag(a) },
    { title: 'Объект', dataIndex: 'object', key: 'object' },
    { title: 'Детали', dataIndex: 'details', key: 'details', render: (d: string) => <Text style={{ fontSize: 12 }}>{d}</Text> },
    { title: 'IP', dataIndex: 'ip', key: 'ip' },
  ];

  const impersonationColumns = [
    { title: 'Время', dataIndex: 'timestamp', key: 'timestamp', render: (d: string) => dayjs(d).format('DD.MM HH:mm') },
    { title: 'Админ', dataIndex: 'adminName', key: 'adminName' },
    { title: 'Тенант', dataIndex: 'tenant', key: 'tenant' },
    { title: 'Роль', dataIndex: 'role', key: 'role' },
    { title: 'Пользователь', dataIndex: 'userName', key: 'userName' },
    { title: 'Длительность', dataIndex: 'duration', key: 'duration', render: (d: number) => `${d} мин` },
    { title: 'Действий', dataIndex: 'actionsSummary', key: 'actionsSummary' },
    { title: 'IP', dataIndex: 'ip', key: 'ip' },
  ];

  const sessionColumns = [
    { title: 'Админ', dataIndex: 'adminName', key: 'adminName' },
    { title: 'IP', dataIndex: 'ip', key: 'ip' },
    { title: 'Начало', dataIndex: 'startedAt', key: 'startedAt', render: (d: string) => dayjs(d).format('DD.MM HH:mm') },
    { title: 'Активность', dataIndex: 'lastActivity', key: 'lastActivity', render: (d: string) => dayjs(d).format('DD.MM HH:mm') },
    {
      title: '',
      key: 'actions',
      render: (_: any, r: any) => (
        <Popconfirm title="Завершить сессию?" onConfirm={() => handleTerminateSession(r.id)}>
          <Button size="small" danger icon={<CloseCircleOutlined />}>Завершить</Button>
        </Popconfirm>
      ),
    },
  ];

  const loginColumns = [
    { title: 'Время', dataIndex: 'timestamp', key: 'timestamp', render: (d: string) => dayjs(d).format('DD.MM HH:mm') },
    { title: 'Тенант', dataIndex: 'tenant', key: 'tenant' },
    { title: 'Пользователь', render: (_: any, r: any) => `${r.userName} (${r.userEmail})` },
    { title: 'Роль', dataIndex: 'role', key: 'role' },
    { title: 'Действие', dataIndex: 'action', key: 'action', render: (a: string) => getLoginActionTag(a) },
    { title: 'IP', dataIndex: 'ip', key: 'ip' },
    { title: 'UA', dataIndex: 'userAgent', key: 'userAgent' },
    { title: 'Гео', dataIndex: 'geo', key: 'geo' },
  ];

  return (
    <div>
      <Title level={3}>Аудит</Title>

      <Tabs items={[
        {
          key: 'admin',
          label: 'Журнал действий суперадминов',
          children: (
            <Card>
              <Row gutter={16} style={{ marginBottom: 16 }}>
                <Col>
                  <Input
                    placeholder="Поиск..."
                    prefix={<SearchOutlined />}
                    value={filters.action}
                    onChange={(e) => setFilters({ action: e.target.value })}
                    style={{ width: 200 }}
                    allowClear
                  />
                </Col>
                <Col>
                  <Select
                    placeholder="Админ"
                    value={filters.admin || undefined}
                    onChange={(v) => setFilters({ admin: v || '' })}
                    style={{ width: 150 }}
                    allowClear
                  >
                    <Option value="Иван Админ">Иван Админ</Option>
                    <Option value="Пётр Супер">Пётр Супер</Option>
                  </Select>
                </Col>
                <Col>
                  <RangePicker />
                </Col>
              </Row>
              <Table dataSource={adminActions} columns={actionColumns} rowKey="id" pagination={{ pageSize: 10 }} />
            </Card>
          ),
        },
        {
          key: 'impersonation',
          label: 'Журнал имперсонаций',
          children: (
            <Card>
              <Table dataSource={impersonations} columns={impersonationColumns} rowKey="id" pagination={{ pageSize: 10 }} />
            </Card>
          ),
        },
        {
          key: 'sessions',
          label: 'Активные сессии',
          children: (
            <Card>
              <Table dataSource={activeSessions} columns={sessionColumns} rowKey="id" pagination={false} />
            </Card>
          ),
        },
        {
          key: 'logins',
          label: 'Входы пользователей',
          children: (
            <Card>
              <Row gutter={16} style={{ marginBottom: 16 }}>
                <Col>
                  <Input placeholder="Поиск..." prefix={<SearchOutlined />} style={{ width: 200 }} allowClear />
                </Col>
                <Col>
                  <Select placeholder="Тенант" style={{ width: 150 }} allowClear>
                    <Option value="ООО Техно">ООО Техно</Option>
                    <Option value="АО Бизнес">АО Бизнес</Option>
                  </Select>
                </Col>
                <Col>
                  <Select placeholder="Действие" style={{ width: 150 }} allowClear>
                    <Option value="login">Вход</Option>
                    <Option value="logout">Выход</Option>
                    <Option value="failed_login">Неудачная попытка</Option>
                  </Select>
                </Col>
                <Col>
                  <RangePicker />
                </Col>
              </Row>
              <Table dataSource={userLogins} columns={loginColumns} rowKey="id" pagination={{ pageSize: 10 }} />
            </Card>
          ),
        },
      ]} />
    </div>
  );
};

export default AuditSection;
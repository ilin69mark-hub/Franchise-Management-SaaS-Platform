import React, { useEffect, useState } from 'react';
import {
  Card,
  Row,
  Col,
  Table,
  Input,
  Select,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  Input as InputForm,
  DatePicker,
  Progress,
  Tabs,
  Typography,
  Spin,
  message,
  Drawer,
  Checkbox,
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  StopOutlined,
  DeleteOutlined,
  UserOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ExportOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import { useTenantsStore } from '@/store/tenantsStore';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;

const TenantsSection: React.FC = () => {
  const {
    tenants,
    tenantsLoading,
    selectedTenant,
    showCard,
    filters,
    showCreateModal,
    showSuspendModal,
    showTerminateModal,
    showImpersonateModal,
    selectedTenantForAction,
    onboarding,
    setTenants,
    setSelectedTenant,
    setTenantsLoading,
    setShowCreateModal,
    setShowSuspendModal,
    setShowTerminateModal,
    setShowImpersonateModal,
    setSelectedTenantForAction,
    setShowCard,
    setFilters,
    setOnboarding,
  } = useTenantsStore();

  const [createForm] = Form.useForm();
  const [suspendReason, setSuspendReason] = useState('');
  const [terminateReason, setTerminateReason] = useState('');
  const [exportData, setExportData] = useState(false);
  const [deleteDelay, setDeleteDelay] = useState('30');
  const [selectedRole, setSelectedRole] = useState('');

  useEffect(() => {
    fetchTenants();
    fetchOnboarding();
  }, []);

  const fetchTenants = async () => {
    setTenantsLoading(true);
    try {
      const params = new URLSearchParams();
      if (filters.status) params.append('status', filters.status);
      if (filters.tariff) params.append('tariff', filters.tariff);
      if (filters.hasDebt) params.append('hasDebt', 'true');
      if (filters.search) params.append('search', filters.search);

      const res = await apiClient.get(`/admin/tenants?${params.toString()}`);
      setTenants(res.data || []);
    } catch {
      setTenants([
        {
          id: '1',
          name: 'ООО Техно',
          legalEntity: 'ООО Техно',
          inn: '1234567890',
          tariff: 'Pro',
          dealers: 10,
          users: 25,
          licenseLimit: 30,
          mrr: 40000,
          paymentStatus: 'paid',
          nextPaymentDate: '2026-05-01',
          status: 'active',
        },
        {
          id: '2',
          name: 'АО Бизнес',
          legalEntity: 'АО Бизнес',
          inn: '0987654321',
          tariff: 'Enterprise',
          dealers: 25,
          users: 80,
          licenseLimit: 100,
          mrr: 150000,
          paymentStatus: 'pending',
          nextPaymentDate: '2026-05-15',
          status: 'active',
        },
        {
          id: '3',
          name: 'ИП Сидоров',
          tariff: 'Start',
          dealers: 2,
          users: 5,
          licenseLimit: 10,
          mrr: 5000,
          paymentStatus: 'overdue',
          nextPaymentDate: '2026-04-01',
          status: 'suspended',
        },
      ]);
    } finally {
      setTenantsLoading(false);
    }
  };

  const fetchOnboarding = async () => {
    try {
      const res = await apiClient.get('/admin/tenants/onboarding');
      setOnboarding(res.data || []);
    } catch {
      setOnboarding([
        {
          id: 'new1',
          name: 'Новая компания',
          createdAt: '2026-04-20',
          step: 'account_created',
          progress: 25,
          daysToTtv: 14,
        },
      ]);
    }
  };

  const handleCreateTenant = async (values: any) => {
    try {
      await apiClient.post('/admin/tenants', values);
      message.success('Тенант создан');
      setShowCreateModal(false);
      createForm.resetFields();
      fetchTenants();
    } catch {
      message.error('Ошибка создания');
    }
  };

  const handleSuspend = async () => {
    if (!selectedTenantForAction) return;
    try {
      await apiClient.post(`/admin/tenants/${selectedTenantForAction.id}/suspend`, {
        reason: suspendReason,
      });
      message.success('Тенант приостановлен');
      setShowSuspendModal(false);
      setSuspendReason('');
      fetchTenants();
    } catch {
      message.error('Ошибка');
    }
  };

  const handleTerminate = async () => {
    if (!selectedTenantForAction) return;
    try {
      await apiClient.post(`/admin/tenants/${selectedTenantForAction.id}/terminate`, {
        reason: terminateReason,
        exportData,
        deleteDelay: parseInt(deleteDelay),
      });
      message.success('Тенант расторгнут');
      setShowTerminateModal(false);
      fetchTenants();
    } catch {
      message.error('Ошибка');
    }
  };

  const handleImpersonate = async () => {
    if (!selectedTenantForAction) return;
    try {
      const res = await apiClient.post(`/admin/tenants/${selectedTenantForAction.id}/impersonate`, {
        role: selectedRole,
      });
      window.open(`/auth?token=${res.data.token}`, '_blank');
      message.success('Сессия открыта');
      setShowImpersonateModal(false);
    } catch {
      message.error('Ошибка');
    }
  };

  const getPaymentStatusTag = (status: string) => {
    const config = {
      paid: { color: 'green', icon: <CheckCircleOutlined />, text: 'Оплачен' },
      pending: { color: 'gold', icon: <ClockCircleOutlined />, text: 'Ожидается' },
      overdue: { color: 'red', icon: <CloseCircleOutlined />, text: 'Просрочен' },
    };
    const c = config[status as keyof typeof config] || config.pending;
    return (
      <Tag icon={c.icon} color={c.color}>
        {c.text}
      </Tag>
    );
  };

  const getTenantStatusTag = (status: string) => {
    const config = {
      active: { color: 'green', text: 'Активен' },
      suspended: { color: 'gold', text: 'Приостановлен' },
      terminated: { color: 'default', text: 'Отключён' },
    };
    const c = config[status as keyof typeof config] || config.active;
    return <Tag color={c.color}>{c.text}</Tag>;
  };

  const columns = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: any, b: any) => a.name.localeCompare(b.name),
    },
    {
      title: 'Юрлицо',
      dataIndex: 'legalEntity',
      key: 'legalEntity',
    },
    {
      title: 'Тариф',
      dataIndex: 'tariff',
      key: 'tariff',
      filters: [
        { text: 'Start', value: 'Start' },
        { text: 'Pro', value: 'Pro' },
        { text: 'Enterprise', value: 'Enterprise' },
      ],
      onFilter: (value: any, record: any) => record.tariff === value,
    },
    {
      title: 'Дилеров',
      dataIndex: 'dealers',
      key: 'dealers',
      sorter: (a: any, b: any) => a.dealers - b.dealers,
    },
    {
      title: 'Пользователей',
      dataIndex: 'users',
      key: 'users',
      render: (users: number, record: any) =>
        `${users}/${record.licenseLimit}`,
      sorter: (a: any, b: any) => a.users - b.users,
    },
    {
      title: 'MRR',
      dataIndex: 'mrr',
      key: 'mrr',
      render: (v: number) => `${v.toLocaleString()} ₽`,
      sorter: (a: any, b: any) => a.mrr - b.mrr,
    },
    {
      title: 'Оплата',
      dataIndex: 'paymentStatus',
      key: 'paymentStatus',
      render: (v: string) => getPaymentStatusTag(v),
    },
    {
      title: 'След. платёж',
      dataIndex: 'nextPaymentDate',
      key: 'nextPaymentDate',
      render: (d: string) => (d ? dayjs(d).format('DD.MM.YYYY') : '-'),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (v: string) => getTenantStatusTag(v),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, record: any) => (
        <Space>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              setSelectedTenant(record);
              setShowCard(true);
            }}
          />
          {record.status === 'active' && (
            <Button
              size="small"
              icon={<PauseCircleOutlined />}
              onClick={() => {
                setSelectedTenantForAction(record);
                setShowSuspendModal(true);
              }}
            />
          )}
          {record.status === 'suspended' && (
            <Button
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleResume(record.id)}
            />
          )}
          {record.status !== 'terminated' && (
            <Button
              size="small"
              danger
              icon={<StopOutlined />}
              onClick={() => {
                setSelectedTenantForAction(record);
                setShowTerminateModal(true);
              }}
            />
          )}
        </Space>
      ),
    },
  ];

  const handleResume = async (id: string) => {
    try {
      await apiClient.post(`/admin/tenants/${id}/resume`);
      message.success('Возобновлён');
      fetchTenants();
    } catch {
      message.error('Ошибка');
    }
  };

  const handleChangeTariff = async (id: string, tariff: string) => {
    try {
      await apiClient.put(`/admin/tenants/${id}`, { tariff });
      message.success('Тариф изменён');
      fetchTenants();
    } catch {
      message.error('Ошибка');
    }
  };

  const handleChangeLicenses = async (id: string, licenses: number) => {
    try {
      await apiClient.put(`/admin/tenants/${id}`, { licenseLimit: licenses });
      message.success('Лицензии обновлены');
      fetchTenants();
    } catch {
      message.error('Ошибка');
    }
  };

  const filteredTenants = tenants.filter((t) => {
    if (filters.status && t.status !== filters.status) return false;
    if (filters.tariff && t.tariff !== filters.tariff) return false;
    if (filters.hasDebt && t.paymentStatus !== 'overdue') return false;
    if (filters.search && !t.name.toLowerCase().includes(filters.search.toLowerCase()))
      return false;
    return true;
  });

  const getOnboardingStepTag = (step: string) => {
    const steps = {
      account_created: 'Аккаунт создан',
      dealers_added: 'Добавлены дилеры',
      users_added: 'Добавлены пользователи',
      first_sale: 'Первая продажа',
    };
    return steps[step as keyof typeof steps] || step;
  };

  return (
    <div>
      <Title level={3}>Тенанты</Title>

      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="Поиск по названию..."
            prefix={<SearchOutlined />}
            value={filters.search}
            onChange={(e) => setFilters({ search: e.target.value })}
            style={{ width: 200 }}
            allowClear
          />
          <Select
            placeholder="Статус"
            value={filters.status || undefined}
            onChange={(v) => setFilters({ status: v || '' })}
            style={{ width: 150 }}
            allowClear
          >
            <Option value="active">Активен</Option>
            <Option value="suspended">Приостановлен</Option>
            <Option value="terminated">Отключён</Option>
          </Select>
          <Select
            placeholder="Тариф"
            value={filters.tariff || undefined}
            onChange={(v) => setFilters({ tariff: v || '' })}
            style={{ width: 150 }}
            allowClear
          >
            <Option value="Start">Start</Option>
            <Option value="Pro">Pro</Option>
            <Option value="Enterprise">Enterprise</Option>
          </Select>
          <Checkbox
            checked={filters.hasDebt}
            onChange={(e) => setFilters({ hasDebt: e.target.checked })}
          >
            Только с долгами
          </Checkbox>
        </Space>
      </Card>

      <Card
        title={
          <Space>
            <span>Список тенантов</span>
            <Text type="secondary">({filteredTenants.length})</Text>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setShowCreateModal(true)}
          >
            Добавить тенанта
          </Button>
        }
      >
        <Table
          dataSource={filteredTenants}
          columns={columns}
          rowKey="id"
          loading={tenantsLoading}
          pagination={{ pageSize: 20 }}
        />
      </Card>

      {onboarding.length > 0 && (
        <Card title="Онбординг (новые тенанты)" style={{ marginTop: 16 }}>
          <Table
            dataSource={onboarding}
            columns={[
              { title: 'Название', dataIndex: 'name', key: 'name' },
              { title: 'Создан', dataIndex: 'createdAt', key: 'createdAt', render: (d: string) => dayjs(d).format('DD.MM.YYYY') },
              { title: 'Этап', dataIndex: 'step', key: 'step', render: (s: string) => getOnboardingStepTag(s) },
              {
                title: 'Прогресс',
                dataIndex: 'progress',
                key: 'progress',
                render: (p: number) => <Progress percent={p} size="small" />,
              },
              { title: 'Дней до TTV', dataIndex: 'daysToTtv', key: 'daysToTtv' },
            ]}
            rowKey="id"
            pagination={false}
          />
        </Card>
      )}

      <Drawer
        title={selectedTenant?.name}
        open={showCard}
        onClose={() => setShowCard(false)}
        width={600}
      >
        {selectedTenant && (
          <Tabs
            items={[
              {
                key: 'main',
                label: 'Основное',
                children: (
                  <div>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Text type="secondary">Юрлицо</Text>
                        <div>{selectedTenant.legalEntity || '-'}</div>
                      </Col>
                      <Col span={12}>
                        <Text type="secondary">ИНН</Text>
                        <div>{selectedTenant.inn || '-'}</div>
                      </Col>
                    </Row>
                    <Row gutter={16} style={{ marginTop: 16 }}>
                      <Col span={12}>
                        <Text type="secondary">Тариф</Text>
                        <div>
                          <Select
                            defaultValue={selectedTenant.tariff}
                            onChange={(v) => handleChangeTariff(selectedTenant.id, v)}
                            style={{ width: '100%' }}
                          >
                            <Option value="Start">Start</Option>
                            <Option value="Pro">Pro</Option>
                            <Option value="Enterprise">Enterprise</Option>
                          </Select>
                        </div>
                      </Col>
                      <Col span={12}>
                        <Text type="secondary">Лицензии</Text>
                        <div>
                          <InputForm
                            type="number"
                            defaultValue={selectedTenant.licenseLimit}
                            onPressEnter={(e: any) =>
                              handleChangeLicenses(
                                selectedTenant.id,
                                parseInt(e.target.value)
                              )
                            }
                          />
                        </div>
                      </Col>
                    </Row>
                    <Row gutter={16} style={{ marginTop: 16 }}>
                      <Col span={24}>
                        <Text type="secondary">Использование лицензий</Text>
                        <Progress
                          percent={Math.round(
                            (selectedTenant.users / selectedTenant.licenseLimit) * 100
                          )}
                          status={
                            selectedTenant.users / selectedTenant.licenseLimit > 0.9
                              ? 'exception'
                              : 'normal'
                          }
                        />
                      </Col>
                    </Row>
                  </div>
                ),
              },
              {
                key: 'payments',
                label: 'Платежи',
                children: (
                  <Table
                    dataSource={[]}
                    columns={[
                      { title: 'Дата', dataIndex: 'date' },
                      { title: 'Счёт', dataIndex: 'invoiceNumber' },
                      { title: 'Сумма', dataIndex: 'amount' },
                      { title: 'Статус', dataIndex: 'status' },
                    ]}
                    locale={{ emptyText: 'Нет данных' }}
                  />
                ),
              },
              {
                key: 'users',
                label: 'Пользователи',
                children: (
                  <Table
                    dataSource={[]}
                    columns={[
                      { title: 'ФИО', render: () => '-' },
                      { title: 'Роль', dataIndex: 'role' },
                      { title: 'Email', dataIndex: 'email' },
                      { title: 'Последний вход', dataIndex: 'lastLogin' },
                    ]}
                    locale={{ emptyText: 'Нет данных' }}
                  />
                ),
              },
              {
                key: 'activity',
                label: 'Активность',
                children: (
                  <div>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Card>
                          <Statistic title="DAU" value={0} />
                        </Card>
                      </Col>
                      <Col span={12}>
                        <Card>
                          <Statistic title="MAU" value={0} />
                        </Card>
                      </Col>
                    </Row>
                  </div>
                ),
              },
            ]}
          />
        )}
      </Drawer>

      <Modal
        title="Создать тенанта"
        open={showCreateModal}
        onCancel={() => setShowCreateModal(false)}
        onOk={() => createForm.submit()}
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreateTenant}>
          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true }]}
          >
            <InputForm />
          </Form.Item>
          <Form.Item name="legalEntity" label="Юрлицо">
            <InputForm />
          </Form.Item>
          <Form.Item name="inn" label="ИНН">
            <InputForm />
          </Form.Item>
          <Form.Item name="tariff" label="Тариф" rules={[{ required: true }]}>
            <Select>
              <Option value="Start">Start</Option>
              <Option value="Pro">Pro</Option>
              <Option value="Enterprise">Enterprise</Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="licenseLimit"
            label="Количество лицензий"
            rules={[{ required: true }]}
          >
            <InputForm type="number" />
          </Form.Item>
          <Form.Item name="contactName" label="Контактное лицо">
            <InputForm />
          </Form.Item>
          <Form.Item name="contactEmail" label="Email">
            <InputForm type="email" />
          </Form.Item>
          <Form.Item name="contractStartDate" label="Дата начала договора">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="subdomain" label="Поддомен">
            <InputForm addonAfter=".example.com" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Приостановить тенанта"
        open={showSuspendModal}
        onCancel={() => setShowSuspendModal(false)}
        onOk={handleSuspend}
        okText="Приостановить"
        okButtonProps={{ danger: true }}
      >
        <p>Все пользователи тенанта потеряют доступ к системе.</p>
        <InputForm
          placeholder="Причина приостановки"
          value={suspendReason}
          onChange={(e) => setSuspendReason(e.target.value)}
          style={{ marginTop: 16 }}
        />
      </Modal>

      <Modal
        title="Расторгнуть договор"
        open={showTerminateModal}
        onCancel={() => setShowTerminateModal(false)}
        onOk={handleTerminate}
        okText="Расторгнуть"
        okButtonProps={{ danger: true }}
      >
        <p style={{ color: '#ff4d4f' }}>
          Внимание! Это действие необратимо.
        </p>
        <Form.Item label="Причина расторжения">
          <InputForm
            value={terminateReason}
            onChange={(e) => setTerminateReason(e.target.value)}
          />
        </Form.Item>
        <Checkbox
          checked={exportData}
          onChange={(e) => setExportData(e.target.checked)}
        >
          Экспортировать данные перед удалением
        </Checkbox>
        <div style={{ marginTop: 8 }}>
          <Text>Удалить данные через: </Text>
          <Select
            value={deleteDelay}
            onChange={setDeleteDelay}
            style={{ width: 100 }}
          >
            <Option value="7">7 дней</Option>
            <Option value="30">30 дней</Option>
            <Option value="90">90 дней</Option>
          </Select>
        </div>
      </Modal>

      <Modal
        title="Войти как"
        open={showImpersonateModal}
        onCancel={() => setShowImpersonateModal(false)}
        onOk={handleImpersonate}
        okText="Войти"
      >
        <p>Выберите роль для имперсонации:</p>
        <Select
          value={selectedRole}
          onChange={setSelectedRole}
          style={{ width: '100%' }}
        >
          <Option value="franchiser">Франчайзер</Option>
          <Option value="franchiser_manager">Менеджер франчайзера</Option>
          <Option value="dealer">Дилер</Option>
        </Select>
        <p style={{ marginTop: 8, fontSize: 12, color: '#888' }}>
          Каждое использование будет залогировано в аудит.
        </p>
      </Modal>
    </div>
  );
};

export default TenantsSection;
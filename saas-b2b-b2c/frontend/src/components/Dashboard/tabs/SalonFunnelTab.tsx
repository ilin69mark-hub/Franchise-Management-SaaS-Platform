import React, { useState, useEffect, useCallback } from 'react';
import { Row, Col, Card, Typography, Button, Modal, Form, Input, Select, Table, Tag, Spin, message, Alert } from 'antd';
import { PlusOutlined, WarningOutlined, ExclamationCircleOutlined, UserAddOutlined } from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import { useCreateLeadMutation, useGetLeadsQuery, useUpdateLeadStatusMutation } from '@/services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;

interface SalonFunnelTabProps {
  user: any;
}

interface FunnelStage {
  stage: string;
  label: string;
  count: number;
  conversion: number;
  sum: number;
}

interface HotDeal {
  id: string;
  client_name: string;
  phone: string;
  amount: number;
  created_at: string;
  days_stalled: number;
  manager_id: string;
  manager_name: string;
}

interface FreshLead {
  id: string;
  source: string;
  client_name: string;
  phone: string;
  created_at: string;
  status: string;
  assigned_to: string | null;
  manager_name: string;
}

interface DashboardFunnelData {
  stages: FunnelStage[];
  hot_deals: HotDeal[];
  fresh_leads: FreshLead[];
}

const stageConfig: Record<string, { color: string; icon: string }> = {
  traffic: { color: '#1890ff', icon: '👥' },
  consultation: { color: '#722ed1', icon: '📞' },
  measurement: { color: '#13c2c2', icon: '📏' },
  kp: { color: '#faad14', icon: '📄' },
  contract: { color: '#52c41a', icon: '✍️' },
  payment: { color: '#eb2f96', icon: '💰' },
};

const SalonFunnelTab: React.FC<SalonFunnelTabProps> = ({ user }) => {
  const { data: leads, isLoading: isLeadsLoading, refetch: refetchLeads } = useGetLeadsQuery();
  const [createLead] = useCreateLeadMutation();
  const [updateLeadStatus] = useUpdateLeadStatusMutation();

  const [data, setData] = useState<DashboardFunnelData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedStage, setSelectedStage] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedLead, setSelectedLead] = useState<any>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    try {
      const date = dayjs().format('YYYY-MM-DD');
      const res = await apiClient.get(`/dashboard/funnel?date=${date}`);
      setData(res.data);
      setError(null);
    } catch (e: any) {
      setError(e?.response?.data?.error || 'Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);

  const handleCreateLead = async (values: any) => {
    try {
      await createLead({
        ...values,
        budget: values.budget ? Number(values.budget) : undefined,
      }).unwrap();
      message.success('Лид добавлен');
      setIsModalOpen(false);
      form.resetFields();
      refetchLeads();
      fetchData();
    } catch (e: any) {
      message.error(e?.data?.error || 'Ошибка создания лида');
    }
  };

  const handleStatusChange = async (leadId: string, newStatus: string) => {
    try {
      await updateLeadStatus({ id: leadId, status: newStatus }).unwrap();
      message.success('Статус обновлён');
      refetchLeads();
      fetchData();
    } catch {
      message.error('Ошибка обновления статуса');
    }
  };

  const handleAssignLead = async (leadId: string, managerId: string) => {
    try {
      await apiClient.patch(`/leads/${leadId}/assign`, { manager_id: managerId });
      message.success('Ответственный назначен');
      fetchData();
    } catch (e: any) {
      message.error(e?.response?.data?.error || 'Ошибка назначения');
    }
  };

  // Воронка
  const funnelChart = (
    <Row gutter={[8, 8]} style={{ marginBottom: 24 }}>
      {data?.stages.map((stage) => (
        <Col xs={8} md={4} key={stage.stage}>
          <Card
            size="small"
            style={{
              borderRadius: 8,
              borderTop: `3px solid ${stageConfig[stage.stage]?.color || '#999'}`,
              cursor: 'pointer',
              textAlign: 'center',
            }}
            onClick={() => setSelectedStage(stage.stage)}
            hoverable
          >
            <div style={{ fontSize: 20 }}>{stageConfig[stage.stage]?.icon}</div>
            <Title level={4} style={{ margin: '8px 0', color: stageConfig[stage.stage]?.color }}>
              {stage.count}
            </Title>
            <Text type="secondary" style={{ fontSize: 12 }}>{stage.label}</Text>
            <div style={{ marginTop: 4 }}>
              <Text style={{ fontSize: 11, color: stage.conversion > 50 ? '#52c41a' : '#faad14' }}>
                {stage.conversion}%
              </Text>
            </div>
          </Card>
        </Col>
      ))}
    </Row>
  );

  // Таблица горячих сделок
  const hotDealsColumns = [
    {
      title: 'Клиент',
      dataIndex: 'client_name',
      key: 'client_name',
      render: (text: string, record: HotDeal) => (
        <a onClick={() => {
          setSelectedLead(record);
          setIsModalOpen(true);
        }}>{text}</a>
      ),
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (val: number) => <Text strong>{formatMoney(val)} ₽</Text>,
    },
    {
      title: 'Дата КП',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: 'Дней без движения',
      dataIndex: 'days_stalled',
      key: 'days_stalled',
      render: (days: number) => (
        <Tag color={days > 7 ? 'error' : days > 5 ? 'warning' : 'default'}>
          {days} дн.
        </Tag>
      ),
    },
    {
      title: 'Ответственный',
      dataIndex: 'manager_name',
      key: 'manager_name',
      render: (name: string) => name || '-',
    },
  ];

  // Таблица свежих лидов
  const freshLeadsColumns = [
    {
      title: 'Источник',
      dataIndex: 'source',
      key: 'source',
    },
    {
      title: 'Клиент',
      dataIndex: 'client_name',
      key: 'client_name',
    },
    {
      title: 'Телефон',
      dataIndex: 'phone',
      key: 'phone',
    },
    {
      title: 'Время',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record: FreshLead) => {
        const createdTime = dayjs(record.created_at, 'YYYY-MM-DD HH:mm');
        const minutesAgo = dayjs().diff(createdTime, 'minute');

        if (status === 'unassigned') {
          return (
            <Tag color={minutesAgo > 30 ? 'error' : 'warning'} icon={<ExclamationCircleOutlined />}>
              {minutesAgo > 30 ? 'Без ответа > 30 мин' : 'Не назначен'}
            </Tag>
          );
        }
        if (status === 'in_progress') {
          return <Tag color="processing">В работе</Tag>;
        }
        return <Tag color="success">Обработан</Tag>;
      },
    },
    {
      title: 'Действие',
      key: 'action',
      render: (_: any, record: FreshLead) => (
        record.status === 'unassigned' ? (
          <Button
            size="small"
            icon={<UserAddOutlined />}
            onClick={() => handleAssignLead(record.id, user.id)}
          >
            Взять
          </Button>
        ) : null
      ),
    },
  ];

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 40 }}><Spin size="large" /></div>;
  }

  if (error) {
    return (
      <Alert
        message="Ошибка"
        description={error}
        type="error"
        showIcon
        action={<a onClick={fetchData}>Повторить</a>}
      />
    );
  }

  // Фильтр сделок по этапу
  const stageLeads = selectedStage ? leads?.filter(l => {
    const stageMap: Record<string, string[]> = {
      traffic: ['new'],
      consultation: ['contact'],
      measurement: ['meeting'],
      kp: ['wait'],
      contract: ['sale'],
      payment: ['paid'],
    };
    return stageMap[selectedStage]?.includes(l.status);
  }) : [];

  return (
    <div>
      {/* Воронка */}
      <Card title="Воронка продаж" style={{ marginBottom: 16, borderRadius: 12 }}>
        {funnelChart}
      </Card>

      {/* Горячие сделки */}
      <Card
        title={
          <span>
            <WarningOutlined style={{ color: '#ff4d4f', marginRight: 8 }} />
            Горячие сделки (КП &gt; 5 дней)
          </span>
        }
        style={{ marginBottom: 16, borderRadius: 12 }}
      >
        {data?.hot_deals && data.hot_deals.length > 0 ? (
          <Table
            dataSource={data.hot_deals}
            columns={hotDealsColumns}
            rowKey="id"
            pagination={false}
            size="small"
            rowClassName={(record) => record.days_stalled > 7 ? 'bg-red-light' : ''}
          />
        ) : (
          <Text type="secondary">Нет горячих сделок</Text>
        )}
      </Card>

      {/* Свежие лиды */}
      <Card
        title={
          <span>
            <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: 8 }} />
            Свежие лиды (за сегодня)
          </span>
        }
        style={{ marginBottom: 16, borderRadius: 12 }}
      >
        {data?.fresh_leads && data.fresh_leads.length > 0 ? (
          <Table
            dataSource={data.fresh_leads}
            columns={freshLeadsColumns}
            rowKey="id"
            pagination={{ pageSize: 5 }}
            size="small"
          />
        ) : (
          <Text type="secondary">Новых лидов за сегодня нет</Text>
        )}
      </Card>

      {/* Все лиды */}
      <Card
        title="Все лиды"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalOpen(true)}>
            Новый лид
          </Button>
        }
        style={{ borderRadius: 12 }}
      >
        <Table
          dataSource={leads}
          rowKey="id"
          pagination={{ pageSize: 10 }}
          columns={[
            {
              title: 'Клиент',
              dataIndex: 'full_name',
              key: 'full_name',
              render: (text: string, record: any) => (
                <a onClick={() => {
                  setSelectedLead(record);
                  setIsModalOpen(true);
                }}>{text}</a>
              ),
            },
            {
              title: 'Телефон',
              dataIndex: 'phone',
              key: 'phone',
            },
            {
              title: 'Бюджет',
              dataIndex: 'budget',
              key: 'budget',
              render: (val: number) => val ? `${formatMoney(val)} ₽` : '-',
            },
            {
              title: 'Статус',
              dataIndex: 'status',
              key: 'status',
              render: (status: string) => {
                const statusMap: Record<string, { color: string; label: string }> = {
                  new: { color: 'blue', label: 'Новый' },
                  contact: { color: 'orange', label: 'Контакт' },
                  meeting: { color: 'cyan', label: 'Замер' },
                  wait: { color: 'purple', label: 'КП' },
                  sale: { color: 'green', label: 'Договор' },
                  paid: { color: 'gold', label: 'Оплачен' },
                };
                return <Tag color={statusMap[status]?.color || 'default'}>{statusMap[status]?.label || status}</Tag>;
              },
            },
          ]}
        />
      </Card>

      {/* Модальное окно создания лида */}
      <Modal
        title="Добавить клиента"
        open={isModalOpen && !selectedLead}
        onCancel={() => {
          setIsModalOpen(false);
          setSelectedLead(null);
        }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateLead}>
          <Form.Item name="full_name" label="Имя клиента" rules={[{ required: true }]}>
            <Input placeholder="Иван Иванов" />
          </Form.Item>
          <Form.Item name="phone" label="Телефон">
            <Input placeholder="+7 (999) 123-45-67" />
          </Form.Item>
          <Form.Item name="interest_product" label="Интерес (Товар)">
            <Input placeholder="Диван" />
          </Form.Item>
          <Form.Item name="budget" label="Бюджет">
            <Input type="number" placeholder="50000" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Модальное окно детальной информации */}
      <Modal
        title="Детали лида"
        open={!!selectedLead}
        onCancel={() => setSelectedLead(null)}
        footer={selectedLead ? [
          <Select
            key="status"
            value={selectedLead.status}
            style={{ width: 150 }}
            onChange={(val) => selectedLead && handleStatusChange(selectedLead.id, val)}
          >
            <Option value="new">Новый</Option>
            <Option value="contact">Контакт</Option>
            <Option value="meeting">Замер</Option>
            <Option value="wait">КП</Option>
            <Option value="sale">Договор</Option>
            <Option value="paid">Оплачен</Option>
          </Select>,
        ] : undefined}
        width={500}
      >
        {selectedLead && (
          <div>
            <p><strong>Клиент:</strong> {selectedLead.full_name}</p>
            <p><strong>Телефон:</strong> {selectedLead.phone}</p>
            <p><strong>Интерес:</strong> {selectedLead.interest_product || 'Не указан'}</p>
            <p><strong>Бюджет:</strong> {selectedLead.budget ? `${formatMoney(selectedLead.budget)} ₽` : 'Не указан'}</p>
            <p><strong>Статус:</strong> {
              ['new', 'contact', 'meeting', 'wait', 'sale', 'paid'].includes(selectedLead.status)
                ? ['Новый', 'Контакт', 'Замер', 'КП', 'Договор', 'Оплачен'][
                    ['new', 'contact', 'meeting', 'wait', 'sale', 'paid'].indexOf(selectedLead.status)
                  ]
                : selectedLead.status
            }</p>
            <p><strong>Дата создания:</strong> {new Date(selectedLead.created_at).toLocaleString('ru-RU')}</p>
          </div>
        )}
      </Modal>

      {/* Модальное окно списка сделок этапа */}
      <Modal
        title={`Сделки: ${stageConfig[selectedStage || '']?.icon} ${selectedStage}`}
        open={!!selectedStage && !selectedLead}
        onCancel={() => setSelectedStage(null)}
        footer={null}
        width={700}
      >
        {selectedStage && (
          <Table
            dataSource={stageLeads}
            rowKey="id"
            pagination={{ pageSize: 10 }}
            columns={[
              { title: 'Клиент', dataIndex: 'full_name', key: 'full_name' },
              { title: 'Телефон', dataIndex: 'phone', key: 'phone' },
              {
                title: 'Бюджет',
                dataIndex: 'budget',
                key: 'budget',
                render: (val: number) => val ? `${formatMoney(val)} ₽` : '-',
              },
              {
                title: 'Статус',
                dataIndex: 'status',
                key: 'status',
                render: (s: string) => <Tag>{s}</Tag>,
              },
            ]}
          />
        )}
      </Modal>
    </div>
  );
};

export default SalonFunnelTab;
// src/components/Dashboard/tabs/TerritoryCommunicationsTab.tsx
import React, { useState, useMemo } from 'react';
import { Card, Row, Col, Typography, Table, Tag, Space, Input, Button, Select, Modal, Form, DatePicker, Badge, Tabs, Segmented, TextArea, Upload, message, Timeline, Popconfirm } from 'antd';
import { MessageOutlined, TeamOutlined, CheckCircleOutlined, ClockCircleOutlined, UserOutlined, PlusOutlined, SearchOutlined, FilterOutlined, SendOutlined, PhoneOutlined, MailOutlined, FileTextOutlined, DollarOutlined, WarningOutlined, ExclamationCircleOutlined, TagOutlined, CalendarOutlined, UploadOutlined } from '@ant-design/icons';
import { useTerritoryManagerStore } from '@/store/territoryManagerStore';
import dayjs from 'dayjs';

const { Text } = Typography;

interface Request {
  id: string;
  type: 'discount' | 'return' | 'marketing' | 'assortment' | 'document' | 'other';
  dealerName: string;
  description: string;
  amount?: number;
  createdAt: string;
  slaHours: number;
  status: 'new' | 'in_progress' | 'resolved' | 'escalated';
}

interface Task {
  id: string;
  title: string;
  dealerName: string;
  assignedAt: string;
  dueDate: string;
  status: 'sent' | 'accepted' | 'in_progress' | 'done' | 'overdue';
  priority: 'high' | 'medium' | 'low';
}

interface Interaction {
  id: string;
  date: string;
  type: 'call' | 'meeting' | 'email' | 'task' | 'discount';
  description: string;
  result?: string;
  files?: string[];
}

interface TerritoryCommunicationsTabProps {
  loading?: boolean;
}

const REQUEST_TYPES = [
  { value: 'discount', label: 'Согласование скидки', icon: <DollarOutlined /> },
  { value: 'return', label: 'Возврат / брак', icon: <WarningOutlined /> },
  { value: 'marketing', label: 'Маркетинговая поддержка', icon: <MessageOutlined /> },
  { value: 'assortment', label: 'Вопрос по ассортименту', icon: <TagOutlined /> },
  { value: 'document', label: 'Документооборот', icon: <FileTextOutlined /> },
  { value: 'other', label: 'Другое', icon: <MessageOutlined /> },
];

const TASK_TEMPLATES = [
  { value: 'display', label: 'Оформить витрину по новому планшету' },
  { value: 'pricelist', label: 'Обновить ценники на коллекцию' },
  { value: 'training', label: 'Провести обучение продавцов' },
  { value: 'report', label: 'Предоставить отчёт по остаткам' },
  { value: 'custom', label: 'Своя задача' },
];

const INTERACTION_TYPES = [
  { value: 'call', label: 'Звонок', icon: <PhoneOutlined /> },
  { value: 'meeting', label: 'Встреча', icon: <TeamOutlined /> },
  { value: 'email', label: 'Письмо', icon: <MailOutlined /> },
  { value: 'task', label: 'Задача', icon: <FileTextOutlined /> },
  { value: 'discount', label: 'Согласование', icon: <DollarOutlined /> },
];

const TerritoryCommunicationsTab: React.FC<TerritoryCommunicationsTabProps> = ({ loading }) => {
  const [activeBlock, setActiveBlock] = useState<'requests' | 'tasks' | 'history'>('requests');
  const [requestFilter, setRequestFilter] = useState<string>('all');
  const [requestTypeFilter, setRequestTypeFilter] = useState<string>('all');
  const [taskFilter, setTaskFilter] = useState<string>('all');
  const [selectedDealerHistory, setSelectedDealerHistory] = useState<string | null>(null);
  const [taskModalOpen, setTaskModalOpen] = useState(false);
  const [interactionModalOpen, setInteractionModalOpen] = useState(false);
  const [newRequestCount, setNewRequestCount] = useState(3);

  const requests = useMemo((): Request[] => [
    { id: '1', type: 'discount', dealerName: 'Мебель Москва', description: 'Согласовать скидку 15% на диван Бостон', amount: 120000, createdAt: '2026-04-28T10:00:00', slaHours: 26, status: 'new' },
    { id: '2', type: 'return', dealerName: 'Диванит Воронеж', description: 'Возврат бракованного кресла', amount: 25000, createdAt: '2026-04-28T09:00:00', slaHours: 2, status: 'new' },
    { id: '3', type: 'marketing', dealerName: 'МебельЛига', description: 'Запрос баннеров к юбилею', amount: 0, createdAt: '2026-04-27T14:00:00', slaHours: 48, status: 'in_progress' },
    { id: '4', type: 'assortment', dealerName: 'Евромебель', description: 'Новая коллекция когда?', amount: 0, createdAt: '2026-04-26T11:00:00', slaHours: -5, status: 'escalated' },
    { id: '5', type: 'document', dealerName: 'Салон мебели Казань', description: 'Договор на поставку', amount: 500000, createdAt: '2026-04-28T08:00:00', slaHours: 30, status: 'resolved' },
  ], []);

  const tasks = useMemo((): Task[] => [
    { id: '1', title: 'Оформить витрину по новой коллекции', dealerName: 'Мебель Москва', assignedAt: '2026-04-25', dueDate: '2026-04-30', status: 'in_progress', priority: 'high' },
    { id: '2', title: 'Обновить ценники', dealerName: 'Диванит Воронеж', assignedAt: '2026-04-26', dueDate: '2026-05-01', status: 'accepted', priority: 'medium' },
    { id: '3', title: 'Отчёт по остаткам', dealerName: 'МебельЛига', assignedAt: '2026-04-27', dueDate: '2026-04-28', status: 'done', priority: 'low' },
    { id: '4', title: 'Провести обучение', dealerName: 'Евромебель', assignedAt: '2026-04-20', dueDate: '2026-04-25', status: 'overdue', priority: 'high' },
  ], []);

  const interactions = useMemo((): Interaction[] => [
    { id: '1', date: '2026-04-28T10:00:00', type: 'call', description: 'Звонок по скидке', result: 'Согласовано 10%' },
    { id: '2', date: '2026-04-27T15:00:00', type: 'meeting', description: 'Встреча в салоне', result: 'Договорились о партнёрстве' },
    { id: '3', date: '2026-04-26T09:00:00', type: 'email', description: 'Отправил коммерческое предложение', result: 'Ожидает ответа' },
    { id: '4', date: '2026-04-25T11:00:00', type: 'task', description: 'Поставил задачу на витрину', result: 'В работе' },
  ], []);

  const dealers = ['Мебель Москва', 'Диванит Воронеж', 'МебельЛига', 'Салон мебели Казань', 'Евромебель'];

  const filteredRequests = useMemo(() => {
    return requests.filter(r => {
      if (requestFilter === 'overdue' && r.slaHours >= 0) return false;
      if (requestFilter === 'new' && r.status !== 'new') return false;
      if (requestFilter === 'escalated' && r.status !== 'escalated') return false;
      if (requestTypeFilter !== 'all' && r.type !== requestTypeFilter) return false;
      return true;
    }).sort((a, b) => a.slaHours - b.slaHours);
  }, [requests, requestFilter, requestTypeFilter]);

  const filteredTasks = useMemo(() => {
    return tasks.filter(t => {
      if (taskFilter === 'all') return true;
      if (taskFilter === 'overdue' && t.status !== 'overdue') return false;
      if (taskFilter === 'active' && (t.status === 'done' || t.status === 'overdue')) return false;
      return true;
    });
  }, [tasks, taskFilter]);

  const getSlaColor = (hours: number) => {
    if (hours < 0) return '#ff4d4f';
    if (hours < 2) return '#ff4d4f';
    if (hours < 24) return '#fa8c16';
    return '#52c41a';
  };

  const getSlaLabel = (hours: number) => {
    if (hours < 0) return `${Math.abs(hours)}ч просрочено`;
    if (hours < 24) return `${hours}ч`;
    return `${Math.round(hours / 24)}дн`;
  };

  const requestColumns = [
    { title: 'Тип', dataIndex: 'type', key: 'type', render: (t: string) => {
      const type = REQUEST_TYPES.find(x => x.value === t);
      return <Space>{type?.icon} {type?.label}</Space>;
    }},
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName' },
    { title: 'Суть', dataIndex: 'description', key: 'description' },
    { title: 'Сумма', dataIndex: 'amount', key: 'amount', render: (v: number) => v > 0 ? `${(v / 1000).toFixed(0)}k ₽` : '-' },
    { title: 'Дата', dataIndex: 'createdAt', key: 'createdAt', render: (d: string) => dayjs(d).format('DD.MM HH:mm') },
    { title: 'SLA', key: 'sla', render: (_: any, r: Request) => (
      <Tag color={getSlaColor(r.slaHours)}>{getSlaLabel(r.slaHours)}</Tag>
    )},
    { title: 'Статус', dataIndex: 'status', key: 'status', render: (s: string) => {
      const colors: Record<string, string> = { new: 'red', in_progress: 'blue', resolved: 'green', escalated: 'orange' };
      const labels: Record<string, string> = { new: 'Новый', in_progress: 'В работе', resolved: 'Решён', escalated: 'Эскалирован' };
      return <Tag color={colors[s]}>{labels[s]}</Tag>;
    }},
    { title: '', key: 'actions', render: (_: any, r: Request) => (
      <Space>
        {r.status === 'new' && <Button size="small" onClick={() => message.success('Взять в работу')}>Взять</Button>}
        <Button size="small">Ответить</Button>
        {r.status !== 'escalated' && <Button size="small" danger>Эскалировать</Button>}
      </Space>
    )},
  ];

  const taskColumns = [
    { title: 'Задача', dataIndex: 'title', key: 'title' },
    { title: 'Дилер', dataIndex: 'dealerName', key: 'dealerName' },
    { title: 'Назначена', dataIndex: 'assignedAt', key: 'assignedAt' },
    { title: 'Срок', dataIndex: 'dueDate', key: 'dueDate' },
    { title: 'Статус', dataIndex: 'status', key: 'status', render: (s: string) => {
      const colors: Record<string, string> = { sent: 'default', accepted: 'blue', in_progress: 'processing', done: 'green', overdue: 'red' };
      const labels: Record<string, string> = { sent: 'Отправлено', accepted: 'Принято', in_progress: 'В работе', done: 'Готово', overdue: 'Просрочено' };
      return <Tag color={colors[s]}>{labels[s]}</Tag>;
    }},
    { title: 'Приоритет', dataIndex: 'priority', key: 'priority', render: (p: string) => {
      const colors: Record<string, string> = { high: 'red', medium: 'orange', low: 'default' };
      return <Tag color={colors[p]}>{p === 'high' ? 'Высокий' : p === 'medium' ? 'Средний' : 'Низкий'}</Tag>;
    }},
  ];

  const [taskForm] = Form.useForm();

  const handleCreateTask = async () => {
    try {
      const values = await taskForm.validateFields();
      message.success('Задача создана');
      setTaskModalOpen(false);
      taskForm.resetFields();
    } catch (e) {}
  };

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col>
            <Segmented value={activeBlock} onChange={(v) => setActiveBlock(v as any)} options={[
              { label: <Badge offset={[5, 0]}><span>Входящие запросы</span></Badge>, value: 'requests' },
              { label: 'Мои задачи', value: 'tasks' },
              { label: 'История', value: 'history' },
            ]} />
          </Col>
          <Col>
            <Select value={requestFilter} onChange={setRequestFilter} style={{ width: 140 }}>
              <Select.Option value="all">Все</Select.Option>
              <Select.Option value="new">Новые</Select.Option>
              <Select.Option value="overdue">Просроченные</Select.Option>
              <Select.Option value="escalated">Эскалированные</Select.Option>
            </Select>
          </Col>
          <Col>
            <Select value={requestTypeFilter} onChange={setRequestTypeFilter} style={{ width: 160 }}>
              <Select.Option value="all">Все типы</Select.Option>
              {REQUEST_TYPES.map(t => <Select.Option key={t.value} value={t.value}>{t.label}</Select.Option>)}
            </Select>
          </Col>
        </Row>
      </Card>

      {activeBlock === 'requests' && (
        <>
          <Card size="small" title="Входящие запросы от дилеров">
            <Table
              dataSource={filteredRequests}
              columns={requestColumns}
              rowKey="id"
              size="small"
              loading={loading}
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </>
      )}

      {activeBlock === 'tasks' && (
        <>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Space>
              <Select value={taskFilter} onChange={setTaskFilter} style={{ width: 140 }}>
                <Select.Option value="all">Все задачи</Select.Option>
                <Select.Option value="active">Активные</Select.Option>
                <Select.Option value="overdue">Просроченные</Select.Option>
              </Select>
              <Button type="primary" icon={<PlusOutlined />} onClick={() => setTaskModalOpen(true)}>
                Поставить задачу
              </Button>
            </Space>
          </Card>
          <Card size="small" title="Мои задачи дилерам">
            <Table
              dataSource={filteredTasks}
              columns={taskColumns}
              rowKey="id"
              size="small"
              loading={loading}
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </>
      )}

      {activeBlock === 'history' && (
        <>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Space>
              <Select
                value={selectedDealerHistory}
                onChange={setSelectedDealerHistory}
                style={{ width: 200 }}
                placeholder="Выберите дилера"
                allowClear
              >
                {dealers.map(d => <Select.Option key={d} value={d}>{d}</Select.Option>)}
              </Select>
              <Button icon={<PlusOutlined />} onClick={() => setInteractionModalOpen(true)}>
                Добавить контакт
              </Button>
            </Space>
          </Card>
          <Card size="small" title={`История взаимодействий${selectedDealerHistory ? ` - ${selectedDealerHistory}` : ''}`}>
            <Timeline
              items={interactions.map(i => {
                const type = INTERACTION_TYPES.find(t => t.value === i.type);
                return {
                  color: type?.value === 'call' ? 'blue' : type?.value === 'meeting' ? 'green' : 'gray',
                  children: (
                    <Space direction="vertical" size={0}>
                      <Text type="secondary" style={{ fontSize: 12 }}>{dayjs(i.date).format('DD.MM.YYYY HH:mm')}</Text>
                      <Space>{type?.icon} {i.description}</Space>
                      {i.result && <Text type="secondary">{i.result}</Text>}
                    </Space>
                  ),
                };
              })}
            />
          </Card>
        </>
      )}

      <Modal
        title="Поставить задачу"
        open={taskModalOpen}
        onCancel={() => setTaskModalOpen(false)}
        onOk={handleCreateTask}
      >
        <Form form={taskForm} layout="vertical">
          <Form.Item name="dealer" label="Дилер" rules={[{ required: true }]}>
            <Select mode="multiple" placeholder="Выберите дилера">
              {dealers.map(d => <Select.Option key={d} value={d}>{d}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item name="template" label="Тип задачи" rules={[{ required: true }]}>
            <Select placeholder="Выберите шаблон">
              {TASK_TEMPLATES.map(t => <Select.Option key={t.value} value={t.value}>{t.label}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item name="description" label="Описание" rules={[{ required: true }]}>
            <TextArea rows={3} placeholder="Описание задачи" />
          </Form.Item>
          <Form.Item name="dueDate" label="Срок выполнения" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="priority" label="Приоритет">
            <Select>
              <Select.Option value="high">Высокий</Select.Option>
              <Select.Option value="medium">Средний</Select.Option>
              <Select.Option value="low">Низкий</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="files" label="Прикрепить файл">
            <Upload>
              <Button icon={<UploadOutlined />}>Загрузить</Button>
            </Upload>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Добавить контакт"
        open={interactionModalOpen}
        onCancel={() => setInteractionModalOpen(false)}
        footer={null}
      >
        <Form layout="vertical">
          <Form.Item label="Тип контакта">
            <Select placeholder="Выберите тип">
              {INTERACTION_TYPES.map(t => <Select.Option key={t.value} value={t.value}>{t.icon} {t.label}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item label="Дата и время">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="Описание">
            <TextArea rows={2} placeholder="Краткое описание" />
          </Form.Item>
          <Form.Item label="Результат">
            <TextArea rows={2} placeholder="Договорённости" />
          </Form.Item>
          <Form.Item label="Прикрепить файл">
            <Upload>
              <Button icon={<UploadOutlined />}>Загрузить</Button>
            </Upload>
          </Form.Item>
          <Button type="primary" block icon={<SendOutlined />}>Сохранить</Button>
        </Form>
      </Modal>
    </div>
  );
};

export default TerritoryCommunicationsTab;
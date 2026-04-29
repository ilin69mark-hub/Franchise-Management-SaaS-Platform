// src/components/Dashboard/tabs/CommunicationsTab.tsx
import React, { useState, useMemo, useEffect, useRef } from 'react';
import { Card, Row, Col, Table, Tag, Button, Space, Modal, Form, Input, Select, InputNumber, Progress, Timeline, Typography, Empty, Spin, DatePicker, message, Statistic, notification, Badge } from 'antd';
import { 
  CheckCircleOutlined, 
  ClockCircleOutlined, 
  ExclamationCircleOutlined, 
  PlusOutlined, 
  CommentOutlined,
  DollarOutlined,
  SendOutlined,
  FileTextOutlined,
  AlertOutlined,
  BellOutlined,
  WifiOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Text, Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;
const { RangePicker } = DatePicker;

export type TaskStatus = 'new' | 'in_progress' | 'done' | 'overdue';
export type TaskPriority = 'high' | 'medium' | 'low';
export type RequestStatus = 'pending' | 'approved' | 'rejected' | 'needs_clarification';
export type RequestType = 'discount' | 'return' | 'marketing' | 'other';

interface Task {
  id: string;
  title: string;
  description: string;
  dueDate: string;
  status: TaskStatus;
  priority: TaskPriority;
  createdAt: string;
  comments?: { text: string; author: string; date: string }[];
}

interface Request {
  id: string;
  type: RequestType;
  description: string;
  sentDate: string;
  status: RequestStatus;
  manager: string;
  amount?: number;
}

interface BudgetItem {
  id: string;
  date: string;
  purpose: string;
  amount: number;
  status: 'pending' | 'approved' | 'rejected';
}

interface Interaction {
  id: string;
  date: string;
  type: 'call' | 'meeting' | 'email' | 'approval';
  summary: string;
  result: string;
  managerName: string;
}

interface MarketingBudget {
  total: number;
  used: number;
  remaining: number;
  items: BudgetItem[];
}

interface CommunicationsTabProps {
  tasks?: Task[];
  requests?: Request[];
  marketingBudget?: MarketingBudget;
  interactions?: Interaction[];
  loading?: boolean;
}

const CommunicationsTab: React.FC<CommunicationsTabProps> = ({
  tasks: initialTasks = [],
  requests: initialRequests = [],
  marketingBudget,
  interactions = [],
  loading = false,
}) => {
  const [activeTab, setActiveTab] = useState<'tasks' | 'requests' | 'budget' | 'history'>('tasks');
  const [requestModalOpen, setRequestModalOpen] = useState(false);
  const [commentModalOpen, setCommentModalOpen] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [filterPriority, setFilterPriority] = useState<string>('all');
  const [form] = Form.useForm();
  const [commentForm] = Form.useForm();
  const wsRef = useRef<WebSocket | null>(null);
  const [wsConnected, setWsConnected] = useState(false);
  const [tasks, setTasks] = useState<Task[]>(initialTasks);
  const [requests, setRequests] = useState<Request[]>(initialRequests);

  useEffect(() => {
    setTasks(initialTasks);
  }, [initialTasks]);

  useEffect(() => {
    setRequests(initialRequests);
  }, [initialRequests]);

  useEffect(() => {
    const connectWebSocket = () => {
      const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';
      try {
        wsRef.current = new WebSocket(wsUrl);
        
        wsRef.current.onopen = () => {
          setWsConnected(true);
          console.log('WebSocket connected');
        };
        
        wsRef.current.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
          } catch (e) {
            console.error('WebSocket message parse error', e);
          }
        };
        
        wsRef.current.onclose = () => {
          setWsConnected(false);
          setTimeout(connectWebSocket, 5000);
        };
        
        wsRef.current.onerror = (error) => {
          console.error('WebSocket error', error);
        };
      } catch (e) {
        console.error('WebSocket connection error', e);
      }
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const handleWebSocketMessage = (data: { type: string; payload: any }) => {
    switch (data.type) {
      case 'TASK_UPDATED':
        setTasks(prev => prev.map(t => t.id === data.payload.id ? { ...t, ...data.payload } : t));
        notification.info({
          message: 'Задача обновлена',
          description: data.payload.title,
          icon: <BellOutlined />,
          duration: 5,
        });
        break;
      case 'TASK_CREATED':
        setTasks(prev => [data.payload, ...prev]);
        notification.info({
          message: 'Новая задача от бренда',
          description: data.payload.title,
          icon: <BellOutlined />,
          duration: 5,
        });
        break;
      case 'REQUEST_STATUS_CHANGED':
        setRequests(prev => prev.map(r => r.id === data.payload.id ? { ...r, ...data.payload } : r));
        const statusText: Record<string, string> = {
          pending: 'На рассмотрении',
          approved: 'Согласовано',
          rejected: 'Отклонено',
          needs_clarification: 'Требуется уточнение',
        };
        notification.info({
          message: 'Статус запроса изменён',
          description: statusText[data.payload.status] || data.payload.status,
          icon: <BellOutlined />,
          duration: 5,
        });
        break;
      default:
        console.log('Unknown WebSocket message type:', data.type);
    }
  };

  const handleTaskStatusChange = async (taskId: string, newStatus: TaskStatus) => {
    setTasks(prev => prev.map(t => t.id === taskId ? { ...t, status: newStatus } : t));
  };

  const filteredTasks = useMemo(() => {
    return tasks.filter(task => {
      if (filterStatus !== 'all' && task.status !== filterStatus) return false;
      if (filterPriority !== 'all' && task.priority !== filterPriority) return false;
      return true;
    });
  }, [tasks, filterStatus, filterPriority]);

  const tasksTableColumns = [
    {
      title: 'Задача',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      render: (val: string, record: Task) => (
        <Space direction="vertical" size={0}>
          <Text strong>{val}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>{record.description}</Text>
        </Space>
      ),
    },
    {
      title: 'Срок',
      dataIndex: 'dueDate',
      key: 'dueDate',
      width: 120,
      render: (val: string) => {
        const due = dayjs(val);
        const now = dayjs();
        const daysDiff = due.diff(now, 'day');
        let color = '#52c41a';
        if (daysDiff < 0) color = '#ff4d4f';
        else if (daysDiff < 3) color = '#fa8c16';
        return <Tag color={color}>{due.format('DD.MM.YYYY')}</Tag>;
      },
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (val: TaskStatus) => {
        const statusMap: Record<TaskStatus, { color: string; text: string }> = {
          new: { color: 'blue', text: 'Новая' },
          in_progress: { color: 'orange', text: 'В работе' },
          done: { color: 'green', text: 'Готово' },
          overdue: { color: 'red', text: 'Просрочено' },
        };
        return <Tag color={statusMap[val].color}>{statusMap[val].text}</Tag>;
      },
    },
    {
      title: 'Приоритет',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      render: (val: TaskPriority) => {
        const priorityMap: Record<TaskPriority, { color: string; text: string }> = {
          high: { color: 'red', text: 'Высокий' },
          medium: { color: 'orange', text: 'Средний' },
          low: { color: 'default', text: 'Низкий' },
        };
        return <Tag color={priorityMap[val].color}>{priorityMap[val].text}</Tag>;
      },
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 150,
      render: (_: any, record: Task) => (
        <Space>
          {record.status === 'new' && (
            <Button size="small" type="primary" onClick={() => handleTaskStatusChange(record.id, 'in_progress')}>
              Принять
            </Button>
          )}
          {(record.status === 'new' || record.status === 'in_progress') && (
            <Button size="small" type="primary" ghost onClick={() => handleTaskStatusChange(record.id, 'done')}>
              Готово
            </Button>
          )}
          <Button size="small" icon={<CommentOutlined />} onClick={() => { setSelectedTask(record); setCommentModalOpen(true); }}>
            Комментарий
          </Button>
        </Space>
      ),
    },
  ];

  const requestsTableColumns = [
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      width: 150,
      render: (val: RequestType) => {
        const typeMap: Record<RequestType, string> = {
          discount: 'Согласование скидки',
          return: 'Возврат/брак',
          marketing: 'Маркетинговая поддержка',
          other: 'Другое',
        };
        return <Tag>{typeMap[val]}</Tag>;
      },
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      width: 200,
    },
    {
      title: 'Дата',
      dataIndex: 'sentDate',
      key: 'sentDate',
      width: 100,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 150,
      render: (val: RequestStatus) => {
        const statusMap: Record<RequestStatus, { color: string; text: string }> = {
          pending: { color: 'blue', text: 'На рассмотрении' },
          approved: { color: 'green', text: 'Согласовано' },
          rejected: { color: 'red', text: 'Отклонено' },
          needs_clarification: { color: 'orange', text: 'Требуется уточнение' },
        };
        return <Tag color={statusMap[val].color}>{statusMap[val].text}</Tag>;
      },
    },
    {
      title: 'Менеджер',
      dataIndex: 'manager',
      key: 'manager',
      width: 150,
    },
  ];

  const budgetPercent = marketingBudget ? Math.round((marketingBudget.used / marketingBudget.total) * 100) : 0;

  const budgetTableColumns = [
    { title: 'Дата', dataIndex: 'date', key: 'date', width: 100 },
    { title: 'Назначение', dataIndex: 'purpose', key: 'purpose', width: 200 },
    { title: 'Сум��а', dataIndex: 'amount', key: 'amount', render: (val: number) => <Text>{val.toLocaleString()} ₽</Text> },
    { title: 'Статус', dataIndex: 'status', key: 'status', width: 120, render: (val: string) => <Tag color={val === 'approved' ? 'green' : val === 'pending' ? 'orange' : 'red'}>{val}</Tag> },
  ];

  const interactionColumns = [
    { title: 'Дата', dataIndex: 'date', key: 'date', width: 100 },
    { title: 'Тип', dataIndex: 'type', key: 'type', width: 100, render: (val: string) => <Tag>{val}</Tag> },
    { title: 'Содержание', dataIndex: 'summary', key: 'summary', width: 200 },
    { title: 'Результат', dataIndex: 'result', key: 'result', width: 150 },
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
          <Space>
            <Button type={activeTab === 'tasks' ? 'primary' : 'default'} onClick={() => setActiveTab('tasks')}>
              Задачи от бренда
              {wsConnected ? <Badge status="success" /> : <Badge status="error" />}
            </Button>
            <Button type={activeTab === 'requests' ? 'primary' : 'default'} onClick={() => setActiveTab('requests')}>
              Мои запросы
            </Button>
            <Button type={activeTab === 'budget' ? 'primary' : 'default'} onClick={() => setActiveTab('budget')}>
              Маркетинговый бюджет
            </Button>
            <Button type={activeTab === 'history' ? 'primary' : 'default'} onClick={() => setActiveTab('history')}>
              История взаимодействий
            </Button>
          </Space>
        </Col>
      </Row>

      {activeTab === 'tasks' && (
        <>
          <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
            <Col xs={24}>
              <Space>
                <Select value={filterStatus} onChange={setFilterStatus} style={{ width: 150 }}>
                  <Option value="all">Все статусы</Option>
                  <Option value="new">Новые</Option>
                  <Option value="in_progress">В работе</Option>
                  <Option value="done">Готово</Option>
                  <Option value="overdue">Просроченные</Option>
                </Select>
                <Select value={filterPriority} onChange={setFilterPriority} style={{ width: 150 }}>
                  <Option value="all">Все приоритеты</Option>
                  <Option value="high">Высокий</Option>
                  <Option value="medium">Средний</Option>
                  <Option value="low">Низкий</Option>
                </Select>
              </Space>
            </Col>
          </Row>
          <Card title="📋 Задачи от франчайзера">
            {filteredTasks.length > 0 ? (
              <Table dataSource={filteredTasks} columns={tasksTableColumns} rowKey="id" pagination={{ pageSize: 10 }} scroll={{ x: 800 }} />
            ) : (
              <Empty description="Нет задач" />
            )}
          </Card>
        </>
      )}

      {activeTab === 'requests' && (
        <>
          <Card 
            title="📨 Мои запросы к бренду"
            extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => setRequestModalOpen(true)}>Создать запрос</Button>}
          >
            {requests.length > 0 ? (
              <Table dataSource={requests} columns={requestsTableColumns} rowKey="id" pagination={{ pageSize: 10 }} scroll={{ x: 800 }} />
            ) : (
              <Empty description="Нет запросов" />
            )}
          </Card>

          <Modal
            title="Создать запрос"
            open={requestModalOpen}
            onCancel={() => setRequestModalOpen(false)}
            onOk={() => { form.validateFields().then(() => { message.success('Запрос отправлен'); setRequestModalOpen(false); }); }}
          >
            <Form form={form} layout="vertical">
              <Form.Item name="type" label="Тип запроса" rules={[{ required: true }]}>
                <Select>
                  <Option value="discount">Согласование скидки</Option>
                  <Option value="return">Возврат/брак</Option>
                  <Option value="marketing">Маркетинговая поддержка</Option>
                  <Option value="other">Другое</Option>
                </Select>
              </Form.Item>
              <Form.Item name="amount" label="Сумма (для скидки/возврата)">
                <InputNumber style={{ width: '100%' }} min={0} placeholder="0 ₽" />
              </Form.Item>
              <Form.Item name="description" label="Описание" rules={[{ required: true }]}>
                <TextArea rows={4} placeholder="Опишите ваш запрос..." />
              </Form.Item>
              <Form.Item name="file" label="Прикрепить файл">
                <Button icon={<FileTextOutlined />}>Выбрать файл</Button>
              </Form.Item>
            </Form>
          </Modal>
        </>
      )}

      {activeTab === 'budget' && (
        <>
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <Card title="💰 Маркетинговый бюджет">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Row gutter={16}>
                    <Col span={8}>
                      <Statistic title="Выделено" value={marketingBudget?.total || 0} prefix="₽ " />
                    </Col>
                    <Col span={8}>
                      <Statistic title="Использовано" value={marketingBudget?.used || 0} prefix="₽ " valueStyle={{ color: '#fa8c16' }} />
                    </Col>
                    <Col span={8}>
                      <Statistic title="Остаток" value={marketingBudget?.remaining || 0} prefix="₽ " valueStyle={{ color: '#52c41a' }} />
                    </Col>
                  </Row>
                  <Progress 
                    percent={budgetPercent} 
                    status={budgetPercent > 90 ? 'exception' : budgetPercent === 100 ? 'success' : undefined}
                    strokeColor={budgetPercent > 90 ? '#ff4d4f' : '#1890ff'}
                  />
                </Space>
              </Card>
            </Col>
          </Row>
          <Card title="Детализация трат" style={{ marginTop: 16 }}>
            {marketingBudget?.items.length ? (
              <Table dataSource={marketingBudget.items} columns={budgetTableColumns} rowKey="id" pagination={{ pageSize: 5 }} />
            ) : (
              <Empty description="Нет данных о тратах" />
            )}
          </Card>
        </>
      )}

      {activeTab === 'history' && (
        <>
          <Card title="📜 История взаимодействий" extra={<Button icon={<PlusOutlined />}>Добавить запись</Button>}>
            {interactions.length > 0 ? (
              <Table dataSource={interactions} columns={interactionColumns} rowKey="id" pagination={{ pageSize: 10 }} />
            ) : (
              <Empty description="Нет взаимодействий" />
            )}
          </Card>
        </>
      )}

      <Modal
        title={`Комментарий к задаче: ${selectedTask?.title}`}
        open={commentModalOpen}
        onCancel={() => { setCommentModalOpen(false); setSelectedTask(null); commentForm.resetFields(); }}
        onOk={() => { commentForm.validateFields().then(() => { message.success('Комментарий добавлен'); setCommentModalOpen(false); }); }}
      >
        <Form form={commentForm} layout="vertical">
          <Form.Item name="comment" label="Ваш комментарий" rules={[{ required: true }]}>
            <TextArea rows={3} placeholder="Введите комментарий..." />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CommunicationsTab;
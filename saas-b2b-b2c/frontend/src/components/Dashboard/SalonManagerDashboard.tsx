// src/components/Dashboard/SalonManagerDashboard.tsx
import React, { useState, useMemo, useEffect } from 'react';
import {
  Row,
  Col,
  Card,
  Button,
  Modal,
  Form,
  Input,
  Typography,
  message,
  Badge,
  Tabs,
  Select,
  Tag,
  Progress,
  List,
  DatePicker,
} from 'antd';
import {
  PlusOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  UserOutlined,
  PhoneOutlined,
  TeamOutlined,
} from '@ant-design/icons';

import { User, Lead, Checklist } from '@/types';
import dayjs from 'dayjs';

/* ← Вставка GoalCard */
import GoalCard from '@/components/Dashboard/GoalCard';

/* ← Импорты API‑хуков (все нужные хуки) */
import {
  useGetChecklistsQuery,
  useGetLeadsQuery,
  useCreateLeadMutation,
  useUpdateLeadStatusMutation,          // <-- добавлен импорт
  useUpdateChecklistMutation,
} from '@/services/api';

const { Title, Text } = Typography;
const { TabPane } = Tabs;
const { Option } = Select;

interface SalonManagerDashboardProps {
  user: User;
  title?: string;
}

/* ----------------------- KPI‑карточка ----------------------- */
const StatCard: React.FC<{
  title: string;
  icon: React.ReactNode;
  plan: number;
  fact: number;
  suffix?: string;
  type: 'money' | 'count';
}> = ({ title, icon, plan, fact, suffix = '', type }) => {
  const percent = plan > 0 ? Math.min(100, Math.round((fact / plan) * 100)) : 0;
  const color = percent >= 80 ? '#52c41a' : percent >= 50 ? '#faad14' : '#ff4d4f';

  const formatVal = (val: number) =>
    type === 'money' ? new Intl.NumberFormat('ru-RU').format(val) : val;

  return (
    <Card size="small" style={{ marginBottom: 12, borderRadius: 8 }}>
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
        <span style={{ fontSize: 18, marginRight: 8, color: '#1890ff' }}>{icon}</span>
        <Text strong>{title}</Text>
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <Text>{formatVal(fact)}{suffix}</Text>
        <Text type="secondary">План: {formatVal(plan)}{suffix}</Text>
      </div>
      <Progress percent={percent} showInfo={false} strokeColor={color} size="small" style={{ marginTop: 8 }} />
      <div style={{ textAlign: 'right' }}>
        <Text style={{ color, fontWeight: 'bold' }}>{percent}%</Text>
      </div>
    </Card>
  );
};

/* ----------------------- Задача / Checklist ----------------------- */
const TaskCard: React.FC<{ task: Checklist }> = ({ task }) => {
  const [updateChecklist, { isLoading: isUpdating }] = useUpdateChecklistMutation();

  const isOverdue =
    task.status === 'pending' && task.end_date && new Date(task.end_date) < new Date();

  const priorityColor = {
    urgent: '#ff4d4f',
    important: '#faad14',
    normal: '#d9d9d9',
  }[task.priority || 'normal'];

  const handleChange = async (newStatus: string) => {
    try {
      await updateChecklist({ id: task.id, status: newStatus }).unwrap();
      message.success('Статус обновлен');
    } catch {
      message.error('Ошибка обновления статуса');
    }
  };

  return (
    <Card
      size="small"
      style={{
        marginBottom: 12,
        borderLeft: `4px solid ${priorityColor}`,
        opacity: isOverdue ? 0.8 : 1,
        background: isOverdue ? '#fff1f0' : '#fff',
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div style={{ flex: 1, marginRight: 12 }}>
          <Text strong delete={task.status === 'completed'}>
            {task.title}
          </Text>
          <br />
          {task.end_date && (
            <Text type="secondary" style={{ fontSize: 12 }}>
              Срок: {new Date(task.end_date).toLocaleDateString()}
            </Text>
          )}
        </div>

        <div style={{ minWidth: 130, textAlign: 'right' }}>
          {isOverdue ? (
            <Tag color="error" icon={<ClockCircleOutlined />}>
              Просрочено
            </Tag>
          ) : task.status === 'completed' ? (
            <Tag color="success" icon={<CheckCircleOutlined />}>
              Выполнено
            </Tag>
          ) : (
            <Select
              value={task.status || 'pending'}
              style={{ width: 130 }}
              size="small"
              loading={isUpdating}
              onChange={handleChange}
            >
              <Option value="pending">В работе</Option>
              <Option value="completed">Выполнено</Option>
            </Select>
          )}
        </div>
      </div>
    </Card>
  );
};

/* ----------------------- Kanban‑колонка ----------------------- */
const KanbanColumn: React.FC<{
  title: string;
  targetStatus: string;
  leads: Lead[] | undefined;
  onDrop: (id: string, status: string) => void;
}> = ({ title, targetStatus, leads = [], onDrop }) => {
  const [isOver, setIsOver] = useState(false);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsOver(true);
  };
  const handleDragLeave = () => setIsOver(false);
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsOver(false);
    const leadId = e.dataTransfer.getData('leadId');
    if (leadId) onDrop(leadId, targetStatus);
  };

  const getStatusColor = (s: string) => {
    const colors: Record<string, string> = {
      new: '#1890ff',
      contact: '#faad14',
      meeting: '#722ed1',
      sale: '#52c41a',
    };
    return colors[s] || '#d9d9d9';
  };

  return (
    <div
      style={{
        background: isOver ? '#e6f7ff' : '#fff',
        padding: '12px',
        borderRadius: '8px',
        minHeight: '300px',
        border: `1px solid ${isOver ? '#1890ff' : '#f0f0f0'}`,
      }}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <div
        style={{
          marginBottom: 12,
          borderBottom: '1px solid #f0f0f0',
          paddingBottom: 8,
          display: 'flex',
          justifyContent: 'space-between',
        }}
      >
        <Text strong>{title}</Text>
        <Badge count={leads.length} style={{ backgroundColor: '#999' }} />
      </div>

      {leads.map((lead) => (
        <Card
          key={lead.id}
          size="small"
          style={{
            marginBottom: '8px',
            cursor: 'grab',
            borderLeft: `4px solid ${getStatusColor(lead.status)}`,
          }}
          draggable
          onDragStart={(e) => {
            e.dataTransfer.setData('leadId', lead.id);
          }}
        >
          <Text strong style={{ display: 'block' }}>{lead.full_name}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {lead.interest_product}
          </Text>
          <br />
          {lead.budget ? (
            <Text type="success" style={{ fontSize: 12 }}>
              {lead.budget} ₽
            </Text>
          ) : null}
        </Card>
      ))}
    </div>
  );
};

/* ----------------------- Основной компонент ----------------------- */
const SalonManagerDashboard: React.FC<SalonManagerDashboardProps> = ({
  user,
  title,
}) => {
  const { data: allChecklists, isLoading: isChecklistLoading } =
    useGetChecklistsQuery();
  const { data: leads, isLoading: isLeadsLoading } = useGetLeadsQuery();

  const [createLead] = useCreateLeadMutation();
  const [updateLeadStatus] = useUpdateLeadStatusMutation();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState('active');

  /* ---------- KPI + расписание ---------- */
  const [stats, setStats] = useState<any>(null);
  const [schedule, setSchedule] = useState<any[]>([]);
  const [loadingStats, setLoadingStats] = useState(true);

  useEffect(() => {
    const fetchDashboardData = async () => {
      setLoadingStats(true);
      try {
        const token = localStorage.getItem('token');
        const headers = { Authorization: `Bearer ${token}` };

        const today = new Date().toISOString().split('T')[0];

        // KPI
        const statsRes = await fetch('/api/v1/stats/my', { headers });
        if (statsRes.ok) setStats(await statsRes.json());

        // Расписание
        const schedRes = await fetch(`/api/v1/schedule?date=${today}`, {
          headers,
        });
        if (schedRes.ok) setSchedule(await schedRes.json());
      } catch (e) {
        console.error('Dashboard fetch error', e);
      } finally {
        setLoadingStats(false);
      }
    };

    fetchDashboardData();
  }, []);

  /* ---------- Канбан: сортировка и фильтрация задач ---------- */
  const { activeTasks, archiveTasks } = useMemo(() => {
    if (!allChecklists) return { activeTasks: [], archiveTasks: [] };

    const priorityOrder: Record<string, number> = {
      urgent: 1,
      important: 2,
      normal: 3,
    };
    const sorted = [...allChecklists].sort(
      (a, b) =>
        (priorityOrder[a.priority || 'normal'] || 3) -
        (priorityOrder[b.priority || 'normal'] || 3),
    );

    const now = new Date();

    const active = sorted.filter((t) => {
      const overdue =
        t.end_date && new Date(t.end_date) < now && t.status === 'pending';
      return t.status === 'pending' && !overdue;
    });

    const archive = sorted.filter((t) => {
      const overdue =
        t.end_date && new Date(t.end_date) < now && t.status === 'pending';
      return t.status === 'completed' || overdue;
    });

    return { activeTasks: active, archiveTasks: archive };
  }, [allChecklists]);

  /* ---------- Обработчики ---------- */
  const handleCreateLead = async (values: any) => {
    try {
      const payload = {
        ...values,
        budget: values.budget ? Number(values.budget) : undefined,
      };
      await createLead(payload).unwrap();
      message.success('Лид добавлен');
      setIsModalOpen(false);
      form.resetFields();
    } catch (e: any) {
      console.error('Lead creation error:', e);
      message.error(e?.data?.error || 'Ошибка создания лида');
    }
  };

  const handleStatusChange = async (leadId: string, newStatus: string) => {
    try {
      await updateLeadStatus({ id: leadId, status: newStatus }).unwrap();
      message.info('Статус лида обновлён');
    } catch {
      message.error('Ошибка обновления статуса лида');
    }
  };

  const handleScheduleStatus = async (eventId: string, status: string) => {
    try {
      const token = localStorage.getItem('token');
      await fetch(`/api/v1/schedule/${eventId}/status`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ status }),
      });
      message.success('Статус задачи обновлен');
      setSchedule((prev) =>
        prev.map((s) => (s.id === eventId ? { ...s, status } : s)),
      );
    } catch {
      message.error('Ошибка обновления статуса задачи');
    }
  };

  /* ---------- Показ загрузки ---------- */
  if (isChecklistLoading || isLeadsLoading) {
    return <div style={{ padding: 24 }}>Загрузка...</div>;
  }

  /* ---------- UI ---------- */
  return (
    <div style={{ padding: 24, background: '#f5f5f5', minHeight: '100vh' }}>
      {/* Шапка */}
      <Row justify="space-between" align="middle" style={{ marginBottom: 24 }}>
        <Col>
          <Title level={3} style={{ margin: 0 }}>
            {title || 'Рабочий стол менеджера'}
          </Title>
          <Text type="secondary">
            Добро пожаловать, {user.first_name || user.email}!
          </Text>
        </Col>
        <Col>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setIsModalOpen(true)}
          >
            Новый лид
          </Button>
        </Col>
      </Row>

      {/* KPI + GoalCard + Расписание */}
      <Row gutter={24} style={{ marginBottom: 24 }}>
        {/* KPI */}
        <Col xs={24} md={12} lg={10}>
          <Card
            title="📊 Мои показатели на сегодня"
            loading={loadingStats}
            style={{ borderRadius: 12 }}
          >
            {stats && (
              <>
                <StatCard
                  title="Продажи"
                  icon={<DollarOutlined />}
                  plan={stats.sales.plan}
                  fact={stats.sales.fact}
                  suffix=" ₽"
                  type="money"
                />
                <StatCard
                  title="Лиды"
                  icon={<UserOutlined />}
                  plan={stats.leads.plan}
                  fact={stats.leads.fact}
                  type="count"
                />
                <StatCard
                  title="Звонки"
                  icon={<PhoneOutlined />}
                  plan={stats.calls.plan}
                  fact={stats.calls.fact}
                  type="count"
                />
                <StatCard
                  title="Встречи"
                  icon={<TeamOutlined />}
                  plan={stats.meetings.plan}
                  fact={stats.meetings.fact}
                  type="count"
                />
              </>
            )}
          </Card>
        </Col>

        {/* Личный план (GoalCard) */}
        <GoalCard date={dayjs().format('YYYY-MM-DD')} />

        {/* Расписание */}
        <Col xs={24} md={12} lg={14}>
          <Card title="📅 Расписание на сегодня" style={{ borderRadius: 12 }}>
            <List
              itemLayout="horizontal"
              dataSource={schedule}
              renderItem={(item) => (
                <List.Item
                  actions={[
                    <Select
                      key="status"
                      value={item.status}
                      style={{ width: 120 }}
                      size="small"
                      onChange={(val) => handleScheduleStatus(item.id, val)}
                    >
                      <Option value="planned">
                        <ClockCircleOutlined /> План
                      </Option>
                      <Option value="completed">
                        <CheckCircleOutlined /> Готово
                      </Option>
                      <Option value="postponed">Перенос</Option>
                    </Select>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <Tag
                        color={item.priority === 'urgent' ? 'red' : 'blue'}
                      >
                        {item.type === 'meeting' ? '🤝' : '📌'}
                      </Tag>
                    }
                    title={`${item.start_time?.split('T')[1]?.substring(0, 5) ||
                      ''} – ${item.title}`}
                    description={item.description}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* Воронка продаж (Kanban) */}
      <Card title="Воронка продаж" style={{ marginBottom: 24 }}>
        <Row gutter={16}>
          <Col span={6}>
            <KanbanColumn
              title="Новые"
              targetStatus="new"
              leads={leads?.filter((l) => l.status === 'new')}
              onDrop={handleStatusChange}
            />
          </Col>
          <Col span={6}>
            <KanbanColumn
              title="Контакт"
              targetStatus="contact"
              leads={leads?.filter((l) => l.status === 'contact')}
              onDrop={handleStatusChange}
            />
          </Col>
          <Col span={6}>
            <KanbanColumn
              title="Встреча"
              targetStatus="meeting"
              leads={leads?.filter((l) => l.status === 'meeting')}
              onDrop={handleStatusChange}
            />
          </Col>
          <Col span={6}>
            <KanbanColumn
              title="Продажа"
              targetStatus="sale"
              leads={leads?.filter((l) => l.status === 'sale')}
              onDrop={handleStatusChange}
            />
          </Col>
        </Row>
      </Card>

      {/* Задачи / Checklist */}
      <Card title="Задачи">
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="В работе" key="active">
            {activeTasks.length === 0 ? (
              <Text type="secondary">Нет активных задач</Text>
            ) : (
              activeTasks.map((task) => (
                <TaskCard key={task.id} task={task} />
              ))
            )}
          </TabPane>
          <TabPane tab="Архив" key="archive">
            {archiveTasks.length === 0 ? (
              <Text type="secondary">Архив пуст</Text>
            ) : (
              archiveTasks.map((task) => (
                <TaskCard key={task.id} task={task} />
              ))
            )}
          </TabPane>
        </Tabs>
      </Card>

      {/* Модальное окно добавления лида */}
      <Modal
        title="Добавить клиента"
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateLead}>
          <Form.Item
            name="full_name"
            label="Имя клиента"
            rules={[{ required: true }]}
          >
            <Input placeholder="Иван Иванов" />
          </Form.Item>
          <Form.Item name="phone" label="Телефон">
            <Input placeholder="+7 (999) 123‑45‑67" />
          </Form.Item>
          <Form.Item
            name="interest_product"
            label="Интерес (Товар)"
          >
            <Input placeholder="Диван «Комфорт»" />
          </Form.Item>
          <Form.Item name="budget" label="Бюджет">
            <Input type="number" placeholder="50000" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SalonManagerDashboard;

import React, { useState, useMemo, useEffect } from 'react';
import { Card, Row, Col, Statistic, Typography, Spin, Tabs, Table, Button, Modal, Form, Input, Select, message, Space, Popconfirm, Tag, Radio, Progress, DatePicker } from 'antd';
import { 
  CheckCircleOutlined, UserOutlined, UserAddOutlined, HomeOutlined, EditOutlined, DeleteOutlined, 
  AimOutlined, ScheduleOutlined, LineChartOutlined, PlusOutlined, PhoneOutlined, TeamOutlined, DollarOutlined
} from '@ant-design/icons';
import { User, Employee } from '@/types';
import { useGetChecklistsQuery, useGetEmployeesQuery, useCreateEmployeeMutation, useUpdateEmployeeMutation, useDeleteEmployeeMutation } from '@/services/api';
import ChecklistBoard from './ChecklistBoard';
import SalonManagerWidget from './SalonManagerWidget';
import dayjs from 'dayjs';
import isoWeek from 'dayjs/plugin/isoWeek';
dayjs.extend(isoWeek);

const { Title, Text } = Typography;
const { Option } = Select;

interface DealerDashboardProps {
  user: User;
  title?: string;
}

// --- Вспомогательный компонент для ячейки KPI ---
const KpiCell: React.FC<{ plan: number; fact: number; suffix?: string }> = ({ plan, fact, suffix = '' }) => {
  const percent = plan > 0 ? Math.min(100, Math.round((fact / plan) * 100)) : 0;
  const color = percent >= 80 ? '#52c41a' : percent >= 50 ? '#faad14' : '#ff4d4f';
  
  return (
    <Space direction="vertical" size={0} style={{ width: '100%' }}>
      <Text style={{ fontSize: 12 }}>{fact}{suffix} / {plan}{suffix}</Text>
      <Progress percent={percent} size="small" showInfo={false} strokeColor={color} />
      <Text type="secondary" style={{ fontSize: 10 }}>{percent}%</Text>
    </Space>
  );
};

const DealerDashboard: React.FC<DealerDashboardProps> = ({ user, title }) => {
  // --- State для новых функций ---
  const [period, setPeriod] = useState<'day' | 'week' | 'month'>('day');
  const [analytics, setAnalytics] = useState<any[]>([]);
  const [schedule, setSchedule] = useState<any[]>([]);
  const [loadingStats, setLoadingStats] = useState(false);
  
  const [isGoalModal, setIsGoalModal] = useState(false);
  const [isSchedModal, setIsSchedModal] = useState(false);
  const [schedForm] = Form.useForm();
  const [goalForm] = Form.useForm();

  // --- Существующий State ---
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingManager, setEditingManager] = useState<Employee | null>(null);
  const [form] = Form.useForm();

  // --- RTK Query Hooks ---
  const { data: checklistData, isLoading: isChecklistLoading } = useGetChecklistsQuery();
  const { data: allEmployees, isLoading: isEmployeesLoading } = useGetEmployeesQuery();
  
  const [createEmployee, { isLoading: isCreating }] = useCreateEmployeeMutation();
  const [updateEmployee] = useUpdateEmployeeMutation();
  const [deleteEmployee] = useDeleteEmployeeMutation();

  // Фильтруем сотрудников
  const salonManagers = useMemo(() => {
    if (!allEmployees) return [];
    return allEmployees.filter((emp: Employee) =>
      emp.role === 'salon_manager' && emp.managed_by === user.id
    );
  }, [allEmployees, user.id]);

  // --- НОВОЕ: Загрузка аналитики и расписания ---
  useEffect(() => {
    fetchAnalytics();
    fetchSchedule();
  }, [period, salonManagers]);

  const fetchAnalytics = async () => {
    if (salonManagers.length === 0) return;
    setLoadingStats(true);
    try {
      const token = localStorage.getItem('token');
      const res = await fetch(`/api/v1/stats/team/analytics?period=${period}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (res.ok) {
        setAnalytics(await res.json());
      }
    } catch (e) {
      console.error('Analytics error', e);
    } finally {
      setLoadingStats(false);
    }
  };

  const fetchSchedule = async () => {
    const today = dayjs().format('YYYY-MM-DD');
    try {
      const token = localStorage.getItem('token');
      const res = await fetch(`/api/v1/schedule/all?date=${today}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (res.ok) setSchedule(await res.json());
    } catch (e) { console.error(e); }
  };

  // --- Обработчики менеджеров (старые) ---
  const showCreateModal = () => { setEditingManager(null); form.resetFields(); setIsModalVisible(true); };
  const showEditModal = (record: Employee) => { setEditingManager(record); form.setFieldsValue(record); setIsModalVisible(true); };
  
  const handleFinish = async (values: any) => {
    try {
      if (editingManager) {
        await updateEmployee({ id: editingManager.id, ...values }).unwrap();
        message.success('Менеджер обновлен');
      } else {
        const payload = { ...values, role: 'salon_manager', managed_by: user.id };
        await createEmployee(payload).unwrap();
        message.success('Менеджер салона создан');
      }
      setIsModalVisible(false);
      form.resetFields();
    } catch (e: any) {
      message.error(e?.data?.error || 'Ошибка при сохранении');
    }
  };

  const handleDelete = async (id: string) => {
    try { await deleteEmployee(id).unwrap(); message.success('Менеджер удален'); } 
    catch (e) { message.error('Ошибка при удалении'); }
  };

  // --- НОВОЕ: Обработчики плана и расписания ---
  const handleCreateSched = async (values: any) => {
    try {
      const token = localStorage.getItem('token');
      const payload = {
          ...values,
          start_time: values.start_time.toISOString()
      };
      await fetch('/api/v1/schedule/manager', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify(payload)
      });
      message.success('Задача поставлена');
      setIsSchedModal(false);
      schedForm.resetFields();
      fetchSchedule();
    } catch (e) { message.error('Ошибка'); }
  };

  const handleDeleteSched = async (id: string) => {
    try {
      const token = localStorage.getItem('token');
      await fetch(`/api/v1/schedule/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('Удалено');
      fetchSchedule();
    } catch (e) { message.error('Ошибка'); }
  };

  const handleSetGoal = async (values: any) => {
    try {
      const token = localStorage.getItem('token');
      // Форматируем дату в строку YYYY-MM-DD
      const payload = {
        ...values,
        target_date: values.target_date ? values.target_date.format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD')
      };
      
      await fetch('/api/v1/goals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify(payload)
      });
      message.success('План сохранен');
      setIsGoalModal(false);
      goalForm.resetFields();
      fetchAnalytics(); // Обновляем таблицу
    } catch (e) { message.error('Ошибка'); }
  };

  // --- Колонки таблиц ---

  // НОВАЯ: Колонки для Аналитики
  const analyticsColumns = [
    { title: 'Менеджер', dataIndex: 'name', fixed: 'left' as const, width: 150 },
    { 
      title: '💰 Продажи', 
      children: [
        { title: 'План / Факт', key: 'sales', width: 150, render: (r: any) => <KpiCell plan={r.kpi?.sales?.plan || 0} fact={r.kpi?.sales?.fact || 0} suffix="₽" /> }
      ]
    },
    { 
      title: '👤 Лиды (Ввод)', 
      children: [
        { title: 'План / Факт', key: 'leads', width: 130, render: (r: any) => <KpiCell plan={r.kpi?.leads?.plan || 0} fact={r.kpi?.leads?.fact || 0} /> }
      ]
    },
    { 
      title: '📞 Звонки', 
      children: [
        { title: 'План / Факт', key: 'calls', width: 130, render: (r: any) => <KpiCell plan={r.kpi?.calls?.plan || 0} fact={r.kpi?.calls?.fact || 0} /> }
      ]
    },
    { 
      title: '🤝 Встречи', 
      children: [
        { title: 'План / Факт', key: 'meetings', width: 130, render: (r: any) => <KpiCell plan={r.kpi?.meetings?.plan || 0} fact={r.kpi?.meetings?.fact || 0} /> }
      ]
    },
    {
      title: '📊 Воронка Лидов (Факты)',
      children: [
        { title: 'Занесли', key: 'entered', render: (r: any) => <Tag color="blue">{r.funnel?.entered || 0}</Tag> },
        { title: 'В работе', key: 'in_work', render: (r: any) => <Tag color="orange">{r.funnel?.in_work || 0}</Tag> },
        { title: 'Думают', key: 'waiting', render: (r: any) => <Tag color="purple">{r.funnel?.waiting || 0}</Tag> },
        { title: 'Продажа', key: 'sale', render: (r: any) => <Tag color="green">{r.funnel?.sale || 0}</Tag> },
      ]
    }
  ];

  // НОВАЯ: Колонки для Расписания
  const scheduleColumns = [
    { title: 'Время', dataIndex: 'start_time', width: 80, render: (d: string) => dayjs(d).format('HH:mm') },
    { title: 'Менеджер', dataIndex: 'user_name' },
    { title: 'Задача', dataIndex: 'title' },
    { title: 'Статус', dataIndex: 'status', render: (s: string) => <Tag color={s === 'completed' ? 'green' : 'blue'}>{s}</Tag> },
    { 
      title: 'Действия', 
      render: (r: any) => (
        <Popconfirm title="Удалить задачу?" onConfirm={() => handleDeleteSched(r.id)}>
          <Button danger size="small" icon={<DeleteOutlined />} />
        </Popconfirm>
      )
    }
  ];

  // СТАРАЯ: Колонки сотрудников
  const empColumns = [
    { title: 'Имя', render: (r: Employee) => `${r.first_name} ${r.last_name}` },
    { title: 'Email', dataIndex: 'email' },
    { title: 'Телефон', dataIndex: 'phone' },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, record: Employee) => (
        <Space>
          <Button icon={<EditOutlined />} size="small" onClick={() => showEditModal(record)} />
          <Popconfirm title="Удалить менеджера?" onConfirm={() => handleDelete(record.id)}>
            <Button danger icon={<DeleteOutlined />} size="small" />
          </Popconfirm>
        </Space>
      )
    },
  ];

  if (isChecklistLoading) return <Spin style={{margin: 50}} />;

  return (
    <div style={{ padding: 24 }}>
      <Row justify="space-between" align="middle" style={{ marginBottom: 20 }}>
        <div>
            <Title level={2} style={{margin: 0}}>{title || 'Личный кабинет дилера'}</Title>
            <Text type="secondary">Добро пожаловать, {user.first_name || user.email}!</Text>
        </div>
        <Space>
            <Button icon={<AimOutlined />} onClick={() => setIsGoalModal(true)}>Выставить план</Button>
            {/* Убрали кнопку "Добавить задачу" отсюда, так как это дублирует вкладку Задачи */}
            <Button type="primary" icon={<UserAddOutlined />} onClick={showCreateModal}>Новый менеджер</Button>
        </Space>
      </Row>

      {/* НОВОЕ: Сводная таблица */}
      <Card style={{ marginBottom: 20 }}>
        <Space style={{ marginBottom: 16 }}>
            <Text strong>Период:</Text>
            <Radio.Group value={period} onChange={e => setPeriod(e.target.value)}>
                <Radio.Button value="day">День</Radio.Button>
                <Radio.Button value="week">Неделя</Radio.Button>
                <Radio.Button value="month">Месяц</Radio.Button>
            </Radio.Group>
        </Space>

        <Table 
            dataSource={analytics}
            columns={analyticsColumns}
            rowKey="id"
            loading={loadingStats}
            bordered
            size="small"
            scroll={{ x: 1300 }}
            locale={{ emptyText: 'Нет данных по команде' }}
        />
      </Card>

      <Tabs defaultActiveKey="schedule">
        {/* НОВАЯ ВКЛАДКА: Расписание */}
        <Tabs.TabPane tab={<span><ScheduleOutlined /> Расписание команды</span>} key="schedule">
            <Card 
                title="Список задач на сегодня"
                extra={
                    <Button 
                        type="primary" 
                        icon={<PlusOutlined />} 
                        onClick={() => setIsSchedModal(true)}
                    >
                        Поставить задачу
                    </Button>
                }
            >
                <Table 
                    dataSource={schedule}
                    columns={scheduleColumns}
                    rowKey="id"
                    locale={{ emptyText: 'Нет задач на сегодня' }}
                />
            </Card>
        </Tabs.TabPane>

        {/* СТАРАЯ ВКЛАДКА: Задачи */}
        <Tabs.TabPane tab="Задачи" key="tasks">
            <Card>
                <ChecklistBoard canCreate={true} employees={salonManagers} />
            </Card>
        </Tabs.TabPane>
        
        {/* СТАРАЯ ВКЛАДКА: Менеджеры */}
        <Tabs.TabPane tab="Менеджеры салонов" key="managers">
             <Card>
                <Table
                    dataSource={salonManagers}
                    columns={empColumns}
                    rowKey="id"
                    loading={isEmployeesLoading}
                    locale={{ emptyText: 'Вы еще не добавили менеджеров салонов' }}
                />
             </Card>
        </Tabs.TabPane>

        {/* СТАРАЯ ВКЛАДКА: Салоны */}
        <Tabs.TabPane tab={<span><HomeOutlined />Салоны</span>} key="salons">
             <SalonManagerWidget />
        </Tabs.TabPane>
      </Tabs>

      {/* СТАРАЯ МОДАЛКА: Менеджер */}
      <Modal
        title={editingManager ? "Редактировать менеджера" : "Новый менеджер салона"}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        onOk={() => form.submit()}
        confirmLoading={isCreating}
      >
        <Form form={form} layout="vertical" onFinish={handleFinish}>
          <Form.Item name="first_name" label="Имя" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="last_name" label="Фамилия">
            <Input />
          </Form.Item>
          <Form.Item name="email" label="Email" rules={[{ required: true, type: 'email' }]}>
            <Input disabled={!!editingManager} />
          </Form.Item>
          {!editingManager && (
            <Form.Item name="password" label="Пароль" rules={[{ required: true, min: 6 }]}>
              <Input.Password />
            </Form.Item>
          )}
          <Form.Item name="phone" label="Телефон">
            <Input />
          </Form.Item>
        </Form>
      </Modal>

      {/* НОВАЯ МОДАЛКА: Поставить задачу (только для Расписания) */}
      <Modal title="Поставить задачу менеджеру" visible={isSchedModal} onCancel={() => setIsSchedModal(false)} onOk={() => schedForm.submit()}>
        <Form form={schedForm} layout="vertical" onFinish={handleCreateSched}>
          <Form.Item name="user_id" label="Менеджер" rules={[{ required: true }]}>
            <Select placeholder="Выберите менеджера">
              {salonManagers.map((m: Employee) => <Option key={m.id} value={m.id}>{m.first_name} {m.last_name}</Option>)}
            </Select>
          </Form.Item>
          <Form.Item name="title" label="Задача" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="start_time" label="Время начала" rules={[{ required: true }]}>
            <DatePicker showTime format="YYYY-MM-DD HH:mm" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="type" label="Тип" initialValue="task">
            <Select>
                <Option value="task">Задача</Option>
                <Option value="meeting">Встреча</Option>
                <Option value="call">Звонок</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* ОБНОВЛЕННАЯ МОДАЛКА: Выставить план */}
      <Modal title="Выставить план" visible={isGoalModal} onCancel={() => setIsGoalModal(false)} onOk={() => goalForm.submit()}>
        <Form form={goalForm} layout="vertical" onFinish={handleSetGoal}>
          <Form.Item name="user_id" label="Менеджер">
            <Select allowClear placeholder="Все или конкретный">
              {salonManagers.map((m: Employee) => <Option key={m.id} value={m.id}>{m.first_name} {m.last_name}</Option>)}
            </Select>
          </Form.Item>
          
          <Form.Item name="target_date" label="Дата плана" initialValue={dayjs()}>
            <DatePicker style={{ width: '100%' }} format="YYYY-MM-DD" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="sales_plan" label="План продаж (₽)">
                <Input type="number" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="leads_plan" label="План лидов">
                <Input type="number" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="calls_plan" label="План звонков">
                <Input type="number" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="meetings_plan" label="План встреч">
                <Input type="number" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default DealerDashboard;
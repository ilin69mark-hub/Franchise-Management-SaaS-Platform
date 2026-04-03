import { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Table, Typography, Tag, Button, Space, Modal, Form, Input, Select, Tabs, message, Popconfirm, DatePicker, Progress } from 'antd';
import { ShopOutlined, UserOutlined, DollarOutlined, DeleteOutlined, StopOutlined, PlusOutlined, EditOutlined, RiseOutlined, FallOutlined, CheckCircleOutlined, AimOutlined, ThunderboltOutlined, WarningOutlined, HeartOutlined, LineChartOutlined } from '@ant-design/icons';
import Head from 'next/head';
import apiClient from '@/api/axiosClient';
import dayjs from 'dayjs';
import { useGetEmployeesQuery, useCreateEmployeeMutation, useUpdateEmployeeMutation, useDeleteEmployeeMutation } from '@/services/api';

const { Title } = Typography;
const { Option } = Select;

const AdminDashboard = () => {
  // State
  const [stats, setStats] = useState({ total_tenants: 0, active_tenants: 0, churned_tenants: 0, total_users: 0, mrr: 0, arpu: 0, new_this_month: 0, plan_stats: [] as any[] });
  const [analytics, setAnalytics] = useState<any>({ dau: 0, mau: 0, sticky_factor: 0, funnel: {} });
  const [product, setProduct] = useState<any>({ kpi: {}, trials: {}, product: {} });
  const [risks, setRisks] = useState<any[]>([]); // Этап 4
  const [economics, setEconomics] = useState<any>({}); // Этап 5
  const [tenants, setTenants] = useState<any[]>([]);
  const [plans, setPlans] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  // Modals State
  const [isTenantModalVisible, setIsTenantModalVisible] = useState(false);
  const [isPlanModalVisible, setIsPlanModalVisible] = useState(false);
  const [isEditModalVisible, setIsEditModalVisible] = useState(false);
  const [isUserModalVisible, setIsUserModalVisible] = useState(false);
  const [isEditPlanModalVisible, setIsEditPlanModalVisible] = useState(false);

  // Forms
  const [tenantForm] = Form.useForm();
  const [planForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [userForm] = Form.useForm();
  const [editPlanForm] = Form.useForm();

  // Selections
  const [selectedTenant, setSelectedTenant] = useState<any>(null);
  const [selectedPlan, setSelectedPlan] = useState<any>(null);
  const [editingUser, setEditingUser] = useState<any>(null);

  // Hooks
  const { data: allUsers, isLoading: isUsersLoading, refetch: refetchUsers } = useGetEmployeesQuery();
  const [createEmployee, { isLoading: isCreatingUser }] = useCreateEmployeeMutation();
  const [updateEmployee] = useUpdateEmployeeMutation();
  const [deleteEmployee] = useDeleteEmployeeMutation();

  useEffect(() => { fetchData(); }, []);

  // Data Fetching
  const fetchData = async () => {
    setLoading(true);
    try {
      const [statsRes, tenantsRes, plansRes, analyticsRes, productRes, risksRes, economicsRes] = await Promise.all([
        apiClient.get('/admin/stats'),
        apiClient.get('/admin/tenants/payments'),
        apiClient.get('/admin/plans'),
        apiClient.get('/admin/analytics'),
        apiClient.get('/admin/product-analytics'),
        apiClient.get('/admin/risks'), // Этап 4
        apiClient.get('/admin/economics') // Этап 5
      ]);
      
      setStats(statsRes.data);
      setTenants(tenantsRes.data || []);
      setPlans(plansRes.data || []);
      setAnalytics(analyticsRes.data || {});
      setProduct(productRes.data || {});
      setRisks(risksRes.data || []);
      setEconomics(economicsRes.data || {});
      
    } catch (e) { 
      message.error('Ошибка загрузки данных'); 
    } finally { 
      setLoading(false); 
    }
  };

  // --- Handlers (Shortened) ---
  const handleCreateTenant = async (values: any) => { try { await apiClient.post('/admin/tenants', values); message.success('Сеть создана'); setIsTenantModalVisible(false); fetchData(); } catch (e) { message.error('Ошибка'); } };
  const handleEditTenant = (record: any) => { setSelectedTenant(record); editForm.setFieldsValue({ name: record.name, plan_id: record.plan_id, paid_until: record.paid_until ? dayjs(record.paid_until) : null }); setIsEditModalVisible(true); };
  const handleUpdateTenant = async (values: any) => { try { const payload = { ...values, paid_until: values.paid_until ? values.paid_until.toISOString() : null }; await apiClient.put(`/admin/tenants/${selectedTenant.id}`, payload); message.success('Обновлено'); setIsEditModalVisible(false); fetchData(); } catch (e) { message.error('Ошибка'); } };
  const handleBlockTenant = async (id: string) => { try { await apiClient.post(`/admin/tenants/${id}/block`); message.success('Заблокировано'); fetchData(); } catch (e) { message.error('Ошибка'); } };
  const handleDeleteTenant = async (id: string) => { try { await apiClient.delete(`/admin/tenants/${id}`); message.success('Удалено'); fetchData(); } catch (e) { message.error('Ошибка'); } };
  const handleCreatePlan = async (values: any) => { try { const payload = { name: values.name, price: Number(values.price), max_users: Number(values.max_users) }; await apiClient.post('/admin/plans', payload); message.success('Тариф создан'); setIsPlanModalVisible(false); fetchData(); } catch (e) { message.error('Ошибка'); } };
  const handleEditPlan = (record: any) => { setSelectedPlan(record); editPlanForm.setFieldsValue(record); setIsEditPlanModalVisible(true); };
  const handleUpdatePlan = async (values: any) => { try { const payload = { name: values.name, price: Number(values.price), max_users: Number(values.max_users) }; await apiClient.put(`/admin/plans/${selectedPlan.id}`, payload); message.success('Тариф обновлен'); setIsEditPlanModalVisible(false); fetchData(); } catch (e) { message.error('Ошибка'); } };
  const handleDeletePlan = async (id: string) => { try { await apiClient.delete(`/admin/plans/${id}`); message.success('Тариф удален'); fetchData(); } catch (e) { message.error('Ошибка'); } };
  
  const showUserModal = (user?: any) => { 
    setEditingUser(user || null); 
    if (user) {
      // При редактировании конвертируем tenant_id в строку, если он есть (для Ant Design Select)
      userForm.setFieldsValue({
        ...user,
        tenant_id: user.tenant_id || undefined
      }); 
    } else {
      userForm.resetFields(); 
    }
    setIsUserModalVisible(true); 
  };

  const handleUserSubmit = async () => { 
    try { 
      const values = await userForm.validateFields();
      // Подготовка данных: если tenant_id пустой, не отправляем его вообще или отправляем null
      const payload = {
        ...values,
        tenant_id: values.tenant_id || null
      };

      if (editingUser) {
        await updateEmployee({ id: editingUser.id, ...payload }).unwrap(); 
        message.success('Обновлен'); 
      } else {
        await createEmployee(payload).unwrap(); 
        message.success('Создан'); 
      }
      setIsUserModalVisible(false); 
      userForm.resetFields(); 
      refetchUsers(); 
    } catch (e: any) { 
      message.error(e?.data?.error || 'Ошибка'); 
    } 
  };
  
  const handleUserDelete = async (id: string) => { try { await deleteEmployee(id).unwrap(); message.success('Удален'); refetchUsers(); } catch (e) { message.error('Ошибка'); } };

  // --- Columns ---
  const tenantsColumns = [
    { title: 'Название', dataIndex: 'name', key: 'name' },
    { title: 'Статус', dataIndex: 'status', key: 'status', render: (s: string) => <Tag color={s === 'active' ? 'green' : 'red'}>{s === 'active' ? 'Активен' : 'Заблокирован'}</Tag> },
    { title: 'Оплата до', dataIndex: 'paid_until', key: 'paid_until', render: (d: string) => d ? new Date(d).toLocaleDateString() : '-' },
    { title: 'Действия', key: 'actions', render: (_: any, record: any) => ( <Space> <Button size="small" icon={<EditOutlined />} onClick={() => handleEditTenant(record)} /> <Button size="small" danger icon={<StopOutlined />} onClick={() => handleBlockTenant(record.id)} /> <Popconfirm title="Удалить?" onConfirm={() => handleDeleteTenant(record.id)}><Button size="small" icon={<DeleteOutlined />} /></Popconfirm> </Space> )},
  ];

  const plansColumns = [
    { title: 'Название', dataIndex: 'name', key: 'name' },
    { title: 'Цена', dataIndex: 'price', key: 'price', render: (v: number) => `${v} руб` },
    { title: 'Лимит юзеров', dataIndex: 'max_users', key: 'max_users' },
    { title: 'Действия', key: 'actions', render: (_: any, record: any) => ( <Space> <Button size="small" icon={<EditOutlined />} onClick={() => handleEditPlan(record)} /> <Popconfirm title="Удалить тариф?" onConfirm={() => handleDeletePlan(record.id)}><Button size="small" danger icon={<DeleteOutlined />} /></Popconfirm> </Space> )},
  ];

  const usersColumns = [
    { title: 'Имя', render: (r: any) => `${r.first_name || ''} ${r.last_name || ''}` },
    { title: 'Email', dataIndex: 'email' },
    { title: 'Роль', dataIndex: 'role', render: (r: string) => <Tag color={r === 'super_admin' ? 'magenta' : 'blue'}>{r}</Tag> },
    { title: 'Действия', render: (_: any, record: any) => ( <Space> <Button size="small" icon={<EditOutlined />} onClick={() => showUserModal(record)} /> <Popconfirm title="Удалить?" onConfirm={() => handleUserDelete(record.id)}><Button size="small" danger icon={<DeleteOutlined />} /></Popconfirm> </Space> )},
  ];

  // Этап 4: Колонки рисков
  const risksColumns = [
    { title: 'Сеть', dataIndex: 'name', key: 'name' },
    { title: 'Health Score', dataIndex: 'health_score', key: 'health_score', width: 200, render: (score: number) => {
        const color = score > 70 ? '#52c41a' : score > 40 ? '#faad14' : '#ff4d4f';
        return <Progress percent={score} strokeColor={color} size="small" />;
    }},
    { title: 'Причина риска', dataIndex: 'risk_reason', key: 'risk_reason', render: (text: string) => <Tag color="error">{text}</Tag> },
    { title: 'Последняя активность', dataIndex: 'last_activity', key: 'last_activity' },
    { title: 'Действия', key: 'actions', render: (_: any, record: any) => ( <Button size="small" type="primary" danger>Вмешаться</Button> )},
  ];

  return (
    <div style={{ padding: 24 }}>
      <Head><title>Админ Панель</title></Head>
      <Title level={2}>Панель Супер Администратора</Title>

      {/* ЭТАП 1: Финансы */}
      <Title level={4} style={{ marginTop: 20 }}>Финансы и Сети</Title>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={4}><Card><Statistic title="MRR (Доход/мес)" value={stats.mrr || 0} prefix={<DollarOutlined />} suffix="руб" valueStyle={{ color: '#3f8600' }} /></Card></Col>
        <Col span={4}><Card><Statistic title="ARPU (Средний чек)" value={stats.arpu ? stats.arpu.toFixed(0) : 0} prefix={<RiseOutlined />} suffix="руб" /></Card></Col>
        <Col span={4}><Card><Statistic title="Активные Сети" value={stats.active_tenants || 0} prefix={<CheckCircleOutlined />} valueStyle={{ color: '#52c41a' }} /></Card></Col>
        <Col span={4}><Card><Statistic title="Ушедшие (Churn)" value={stats.churned_tenants || 0} prefix={<FallOutlined />} valueStyle={{ color: '#cf1322' }} /></Card></Col>
        <Col span={4}><Card><Statistic title="Всего Юзеров" value={stats.total_users || 0} prefix={<UserOutlined />} /></Card></Col>
        <Col span={4}><Card><Statistic title="Новых за месяц" value={stats.new_this_month || 0} prefix={<ShopOutlined />} /></Card></Col>
      </Row>

      {/* ЭТАП 2: Активность */}
      <Title level={4} style={{ marginTop: 20 }}>Активность и Воронка</Title>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={4}><Card><Statistic title="DAU (Сегодня)" value={analytics.dau || 0} prefix={<UserOutlined />} /></Card></Col>
        <Col span={4}><Card><Statistic title="MAU (Месяц)" value={analytics.mau || 0} /></Card></Col>
        <Col span={4}><Card><Statistic title="Sticky Factor" value={analytics.sticky_factor ? analytics.sticky_factor.toFixed(1) : 0} suffix="%" /></Card></Col>
        <Col span={4}><Card><Statistic title="Конверсия (Trial→Paid)" value={analytics.funnel?.conversion ? analytics.funnel.conversion.toFixed(1) : 0} suffix="%" /></Card></Col>
      </Row>

      {/* ЭТАП 3: KPI */}
      <Title level={4} style={{ marginTop: 20 }}>Продукт и KPI</Title>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={4}>
          <Card>
            <Statistic 
              title="KPI (Выполнение задач)" 
              value={product.kpi?.completion_rate ? product.kpi.completion_rate.toFixed(0) : 0} 
              suffix="%" 
              prefix={<AimOutlined />}
            />
          </Card>
        </Col>
        <Col span={4}>
            <Card>
                <Statistic 
                    title="Активные Trials" 
                    value={product.trials?.active_count || 0} 
                    prefix={<ThunderboltOutlined />}
                />
            </Card>
        </Col>
      </Row>

      {/* ЭТАП 4: Риски и Health Score (Восстановлено) */}
      <Title level={4} style={{ marginTop: 20 }}>Риски и Удержание</Title>
      <Row gutter={16} style={{ marginBottom: 24 }}>
         <Col span={6}>
            <Card>
                <Statistic 
                    title="Сети в зоне риска" 
                    value={risks.length} 
                    prefix={<WarningOutlined />}
                    valueStyle={{ color: risks.length > 0 ? '#cf1322' : '#3f8600' }}
                />
                <div style={{ fontSize: 12, color: '#888', marginTop: 5 }}>
                    Требуют внимания
                </div>
            </Card>
        </Col>
        <Col span={6}>
            <Card>
                <Statistic 
                    title="Retention (Удержание)" 
                    value={product.product?.retention_rate ? product.product.retention_rate.toFixed(0) : 0} 
                    suffix="%"
                    prefix={<HeartOutlined />}
                />
            </Card>
        </Col>
      </Row>

      {/* ЭТАП 5: Unit Economics (НОВОЕ - добавлено) */}
      <Title level={4} style={{ marginTop: 20 }}>Unit Экономика (LTV/CAC)</Title>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={4}>
            <Card>
                <Statistic 
                    title="LTV (Lifetime Value)" 
                    value={economics.ltv ? economics.ltv.toFixed(0) : 0} 
                    prefix={<LineChartOutlined />}
                    suffix="руб"
                />
            </Card>
        </Col>
        <Col span={4}>
            <Card>
                <Statistic 
                    title="CAC (Cost of Acq)" 
                    value={economics.cac ? economics.cac.toFixed(0) : 0} 
                    prefix={<DollarOutlined />}
                    suffix="руб"
                />
            </Card>
        </Col>
        <Col span={4}>
            <Card>
                <Statistic 
                    title="LTV / CAC Ratio" 
                    value={economics.ltv_cac_ratio ? economics.ltv_cac_ratio.toFixed(1) : 0} 
                    suffix="x"
                    valueStyle={{ color: economics.ltv_cac_ratio > 3 ? '#3f8600' : '#faad14' }}
                />
                <div style={{ fontSize: 12, color: '#888', marginTop: 5 }}>
                    Норма: &gt; 3.0
                </div>
            </Card>
        </Col>
      </Row>

      {/* Табы управления */}
      <Card>
        <Tabs defaultActiveKey="tenants">
          <Tabs.TabPane tab="Управление Сетями" key="tenants">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsTenantModalVisible(true)} style={{ marginBottom: 16 }}>Добавить Сеть</Button>
            <Table dataSource={tenants} columns={tenantsColumns} rowKey="id" loading={loading} />
          </Tabs.TabPane>

          <Tabs.TabPane tab={`Зона Риска (${risks.length})`} key="risks">
             <div style={{ marginBottom: 16 }}>
                <WarningOutlined style={{ color: 'red', marginRight: 8 }} />
                <span>Сети с низким Health Score или проблемами оплаты</span>
             </div>
             <Table dataSource={risks} columns={risksColumns} rowKey="id" loading={loading} />
          </Tabs.TabPane>

          <Tabs.TabPane tab="Тарифные Планы" key="plans">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsPlanModalVisible(true)} style={{ marginBottom: 16 }}>Создать Тариф</Button>
            <Table dataSource={plans} columns={plansColumns} rowKey="id" loading={loading} />
          </Tabs.TabPane>

          <Tabs.TabPane tab="Все Пользователи" key="users">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => showUserModal()} style={{ marginBottom: 16 }}>Добавить Юзера</Button>
            <Table dataSource={allUsers} columns={usersColumns} rowKey="id" loading={isUsersLoading} />
          </Tabs.TabPane>
        </Tabs>
      </Card>

      {/* Modals */}
      <Modal title="Создать новую сеть" open={isTenantModalVisible} onCancel={() => setIsTenantModalVisible(false)} onOk={() => tenantForm.submit()}>
        <Form form={tenantForm} layout="vertical" onFinish={handleCreateTenant}>
          <Form.Item name="name" label="Название" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="plan_id" label="Тариф">
            <Select allowClear placeholder="Выберите тариф">
              {plans.map((p: any) => <Option key={p.id} value={p.id}>{p.name}</Option>)}
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="Редактировать клиента" open={isEditModalVisible} onCancel={() => setIsEditModalVisible(false)} onOk={() => editForm.submit()}>
        <Form form={editForm} layout="vertical" onFinish={handleUpdateTenant}>
          <Form.Item name="name" label="Название" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="plan_id" label="Тариф">
            <Select allowClear placeholder="Выберите тариф">
              {plans.map((p: any) => <Option key={p.id} value={p.id}>{p.name}</Option>)}
            </Select>
          </Form.Item>
          <Form.Item name="paid_until" label="Оплачено до"><DatePicker style={{ width: '100%' }} showTime /></Form.Item>
        </Form>
      </Modal>

      <Modal title="Создать тариф" open={isPlanModalVisible} onCancel={() => setIsPlanModalVisible(false)} onOk={() => planForm.submit()}>
        <Form form={planForm} layout="vertical" onFinish={handleCreatePlan}>
          <Form.Item name="name" label="Название" rules={[{ required: true }]}><Input placeholder="Например: Start" /></Form.Item>
          <Form.Item name="price" label="Цена (RUB)" rules={[{ required: true }]}><Input type="number" placeholder="5000" /></Form.Item>
          <Form.Item name="max_users" label="Лимит пользователей"><Input type="number" placeholder="10" /></Form.Item>
        </Form>
      </Modal>

      <Modal title="Редактировать тариф" open={isEditPlanModalVisible} onCancel={() => setIsEditPlanModalVisible(false)} onOk={() => editPlanForm.submit()}>
        <Form form={editPlanForm} layout="vertical" onFinish={handleUpdatePlan}>
          <Form.Item name="name" label="Название" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="price" label="Цена (RUB)" rules={[{ required: true }]}><Input type="number" /></Form.Item>
          <Form.Item name="max_users" label="Лимит пользователей"><Input type="number" /></Form.Item>
        </Form>
      </Modal>

      {/* Modal: Создать/Редактировать Юзера (ИЗМЕНЕНО) */}
      <Modal title={editingUser ? "Редактировать юзера" : "Новый юзер"} open={isUserModalVisible} onCancel={() => setIsUserModalVisible(false)} onOk={() => userForm.submit()} confirmLoading={isCreatingUser}>
        <Form form={userForm} layout="vertical" onFinish={handleUserSubmit}>
          <Form.Item name="first_name" label="Имя" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="last_name" label="Фамилия"><Input /></Form.Item>
          <Form.Item name="email" label="Email" rules={[{ required: true, type: 'email' }]}><Input disabled={!!editingUser} /></Form.Item>
          {!editingUser && <Form.Item name="password" label="Пароль" rules={[{ required: true, min: 6 }]}><Input.Password /></Form.Item>}
          
          <Form.Item name="role" label="Роль" rules={[{ required: true }]}>
            <Select placeholder="Выберите роль">
              <Option value="super_admin">Супер Админ</Option>
              <Option value="franchiser">Франчайзер</Option>
              <Option value="franchiser_manager">Менеджер Франчайзера</Option>
              <Option value="dealer">Дилер</Option>
              <Option value="salon_manager">Менеджер Салона</Option>
            </Select>
          </Form.Item>

          {/* НОВОЕ: Выбор сети */}
          <Form.Item 
            name="tenant_id" 
            label="Сеть (Tenant)" 
            // Обязательно только для Франчайзера. Для SuperAdmin можно оставить пустым.
            rules={[{ required: true, message: 'Выберите сеть для Франчайзера' }]} 
          >
            <Select allowClear placeholder="Выберите сеть (для Франчайзера)">
              {tenants.map((t: any) => <Option key={t.id} value={t.id}>{t.name}</Option>)}
            </Select>
          </Form.Item>

          <Form.Item name="phone" label="Телефон"><Input /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminDashboard;
// src/components/Dashboard/FranchiserManagerDashboard.tsx
import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Typography,
  Tabs,
  Button,
  Space,
  Popconfirm,
  Modal,
  Form,
  Input,
  message,
  Select,
  Statistic,
  Row,
  Divider,
  Col,
} from 'antd';
import {
  ShopOutlined,
  TeamOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckSquareOutlined,
} from '@ant-design/icons';
import GoalList from '@/components/Dashboard/GoalList';
import GoalCard from '@/components/Dashboard/GoalCard';
import {
  useGetEmployeesQuery,
  useCreateEmployeeMutation,
  useUpdateEmployeeMutation,
  useDeleteEmployeeMutation,
} from '@/services/api';
import { useGetVisibleGoalsQuery } from '@/services/goalApi';
import { User, Employee, Goal } from '@/types';
import ChecklistBoard from './ChecklistBoard';
import { getAssignableRoles } from '@/utils/rolePermissions';

const { Title } = Typography;
const { TabPane } = Tabs;
const { Option } = Select;

interface FranchiserManagerDashboardProps {
  user: User;          // пользователь с ролью franchiser_manager
  title?: string;
}

/* ---------- Функция агрегирования целей (по роли) ---------- */
const aggregateGoals = (goals: Goal[], role: string) => {
  const filtered = goals.filter((g) => g.role === role);
  const totalPlan = filtered.reduce((sum, g) => sum + g.sales_plan, 0);
  const totalFact = filtered.reduce((sum, g) => sum + (g.sales_fact ?? 0), 0);
  const totalForecast = filtered.reduce(
    (sum, g) => sum + (g.forecast ?? 0),
    0,
  );
  const percent = totalPlan ? Math.round((totalFact / totalPlan) * 100) : 0;
  return { totalPlan, totalFact, totalForecast, percent };
};

const FranchiserManagerDashboard: React.FC<FranchiserManagerDashboardProps> = ({
  user,
  title,
}) => {
  /** ------------------- Модальное окно сотрудника ------------------- */
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [form] = Form.useForm();

  /** ------------------- Запросы ------------------- */
  // Сотрудники (Dealer + Salon Manager)
  const {
    data: allUsers,
    isLoading: isUsersLoading,
    error: usersError,
    refetch: refetchUsers,
  } = useGetEmployeesQuery();

  // Цели (для агрегатов и GoalList)
  const {
    data: allGoals = [],
    isLoading: isGoalsLoading,
    error: goalsError,
    refetch: refetchGoals,
  } = useGetVisibleGoalsQuery();

  // Мутации сотрудников
  const [createEmployee, { isLoading: isCreating }] =
    useCreateEmployeeMutation();
  const [updateEmployee] = useUpdateEmployeeMutation();
  const [deleteEmployee] = useDeleteEmployeeMutation();

  /** ------------------- Ошибки ------------------- */
  useEffect(() => {
    if (usersError) {
      message.error('Ошибка загрузки сотрудников');
    }
    if (goalsError) {
      message.error('Ошибка загрузки целей');
    }
  }, [usersError, goalsError]);

  /** ------------------- Подготовка данных ------------------- */
  const safeUsers = Array.isArray(allUsers) ? allUsers : [];

  // Фильтрация по ролям, которые видит менеджер франчайзи
  const dealers = safeUsers.filter(
    (u) => (u.role || '').toLowerCase() === 'dealer',
  );
  const salonManagers = safeUsers.filter(
    (u) => (u.role || '').toLowerCase() === 'salon_manager',
  );

  // Кураторы (может назначать только уже существующего франчайзи‑менеджера)
  const potentialManagers = safeUsers.filter(
    (u) => (u.role || '').toLowerCase() === 'franchiser_manager',
  );

  /** ------------------- Открыть/закрыть модалку ------------------- */
  const showEmpModal = (employee?: Employee) => {
    setEditingEmployee(employee || null);
    if (employee) {
      form.setFieldsValue(employee);
    } else {
      form.resetFields();
    }
    setIsModalVisible(true);
  };

  /** ------------------- CRUD сотрудников ------------------- */
  const handleEmpSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingEmployee) {
        await updateEmployee({ id: editingEmployee.id, ...values }).unwrap();
        message.success('Сотрудник обновлён');
      } else {
        await createEmployee(values).unwrap();
        message.success('Сотрудник создан');
      }
      setIsModalVisible(false);
      form.resetFields();
      refetchUsers();
    } catch (e: any) {
      console.error(e);
      message.error(e?.data?.error || 'Ошибка сохранения');
    }
  };

  const handleEmpDelete = async (id: string) => {
    try {
      await deleteEmployee(id).unwrap();
      message.success('Сотрудник удалён');
      refetchUsers();
    } catch (e: any) {
      console.error(e);
      message.error(e?.data?.error || 'Ошибка удаления');
    }
  };

  /** ------------------- Таблица сотрудников ------------------- */
  const columns = [
    {
      title: 'Имя',
      dataIndex: 'first_name',
      key: 'first_name',
      render: (t: string) => t || '—',
    },
    { title: 'Email', dataIndex: 'email', key: 'email' },
    { title: 'Роль', dataIndex: 'role', key: 'role' },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, record: Employee) => (
        <Space>
          <Button
            icon={<EditOutlined />}
            size="small"
            onClick={() => showEmpModal(record)}
          />
          <Popconfirm
            title="Удалить сотрудника?"
            onConfirm={() => handleEmpDelete(record.id)}
          >
            <Button danger icon={<DeleteOutlined />} size="small" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  /** ------------------- Агрегация целей ------------------- */
  const dealerAgg = aggregateGoals(allGoals, 'dealer');
  const salonMgrAgg = aggregateGoals(allGoals, 'salon_manager');

  /** ------------------- Роли, которые может назначать ------------------- */
  const assignableRoles = getAssignableRoles(user.role); // -> ['dealer','salon_manager']
  const canAssign = assignableRoles.length > 0;

  /** ------------------- Рендер ------------------- */
  return (
    <div style={{ padding: 24 }}>
      {/* Шапка */}
      <Title level={2}>{title || 'Панель менеджера франчайзи'}</Title>

      {/* Статистика по сотрудникам */}
      <div style={{ marginBottom: 24, display: 'flex', gap: 16 }}>
        <Card style={{ width: 200 }}>
          Дилеров: <strong>{dealers.length}</strong>
        </Card>
        <Card style={{ width: 200 }}>
          Менеджеров салонов: <strong>{salonManagers.length}</strong>
        </Card>
      </div>

      {/* Личный план */}
      <Card style={{ marginBottom: 24 }}>
        <GoalCard />
      </Card>

      {/* Агрегированные карточки целей */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={12}>
          <Card title="Дилеры – план продаж">
            <Statistic title="План (₽)" value={dealerAgg.totalPlan} />
            <Statistic title="Факт (₽)" value={dealerAgg.totalFact} />
            <Statistic title="Прогноз (₽)" value={dealerAgg.totalForecast} />
            <Statistic title="Выполнено, %" value={dealerAgg.percent} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="Менеджеры салонов – план продаж">
            <Statistic title="План (₽)" value={salonMgrAgg.totalPlan} />
            <Statistic title="Факт (₽)" value={salonMgrAgg.totalFact} />
            <Statistic title="Прогноз (₽)" value={salonMgrAgg.totalForecast} />
            <Statistic title="Выполнено, %" value={salonMgrAgg.percent} />
          </Card>
        </Col>
      </Row>

      {/* Список целей и кнопка «Назначить план» */}
      <Card title="Планы (цели) в организации" style={{ marginBottom: 24 }}>
        {canAssign && (
          <Space style={{ marginBottom: 12 }}>
            {/* GoalList сам отобразит кнопку, но можно добавить пояснение */}
          </Space>
        )}
        <GoalList />
      </Card>

      {/* Вкладки для остальных функций */}
      <Tabs defaultActiveKey="tasks">
        <TabPane
          tab={
            <span>
              <CheckSquareOutlined /> Задачи
            </span>
          }
          key="tasks"
        >
          <Card>
            <ChecklistBoard employees={safeUsers} />
          </Card>
        </TabPane>

        {/* Таблица дилеров */}
        <TabPane
          tab={
            <span>
              <ShopOutlined /> Дилеры ({dealers.length})
            </span>
          }
          key="dealers"
        >
          <Card
            extra={
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => showEmpModal()}
              >
                Добавить
              </Button>
            }
          >
            <Table
              columns={columns}
              dataSource={dealers}
              rowKey="id"
              loading={isUsersLoading}
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </TabPane>

        {/* Таблица менеджеров салонов */}
        <TabPane
          tab={
            <span>
              <TeamOutlined /> Менеджеры салонов ({salonManagers.length})
            </span>
          }
          key="salonManagers"
        >
          <Card
            extra={
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => showEmpModal()}
              >
                Добавить
              </Button>
            }
          >
            <Table
              columns={columns}
              dataSource={salonManagers}
              rowKey="id"
              loading={isUsersLoading}
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </TabPane>
      </Tabs>

      {/* Модальное окно создания/редактирования сотрудника */}
      <Modal
        title={editingEmployee ? 'Редактировать' : 'Новый пользователь'}
        open={isModalVisible}
        onCancel={() => {
          setIsModalVisible(false);
          form.resetFields();
        }}
        onOk={handleEmpSubmit}
        confirmLoading={isCreating}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="first_name"
            label="Имя"
            rules={[{ required: true }]}
          >
            <Input />
          </Form.Item>

          <Form.Item name="last_name" label="Фамилия">
            <Input />
          </Form.Item>

          <Form.Item
            name="email"
            label="Email"
            rules={[{ required: true, type: 'email' }]}
          >
            <Input disabled={!!editingEmployee} />
          </Form.Item>

          {/* Пароль только при добавлении */}
          {!editingEmployee && (
            <Form.Item
              name="password"
              label="Пароль"
              rules={[{ required: true, min: 6 }]}
            >
              <Input.Password />
            </Form.Item>
          )}

          <Form.Item name="role" label="Роль" rules={[{ required: true }]}>
            <Select placeholder="Выберите роль">
              {/* Менеджер франчайзи может создавать только дилеров и менеджеров салонов */}
              <Option value="dealer">Дилер</Option>
              <Option value="salon_manager">Менеджер Салона</Option>
            </Select>
          </Form.Item>

          {/* Куратор (может быть только менеджер франчайзи) */}
          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) => prev.role !== cur.role}
          >
            {({ getFieldValue }) => {
              const sel = getFieldValue('role');
              if (
                (sel === 'dealer' || sel === 'salon_manager') &&
                potentialManagers.length > 0
              ) {
                return (
                  <Form.Item name="managed_by" label="Куратор (Менеджер)">
                    <Select placeholder="Выберите менеджера‑куратора">
                      {potentialManagers.map((m) => (
                        <Option key={m.id} value={m.id}>
                          {m.first_name} {m.last_name} ({m.email})
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                );
              }
              return null;
            }}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default FranchiserManagerDashboard;

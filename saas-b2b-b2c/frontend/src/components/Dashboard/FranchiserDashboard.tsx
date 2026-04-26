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
import ChecklistBoard from '@/components/Dashboard/ChecklistBoard';
import { getAssignableRoles } from '@/utils/rolePermissions';

const { Title } = Typography;
const { TabPane } = Tabs;
const { Option } = Select;

interface FranchiserDashboardProps {
  user: User;
  title?: string;
}

/* ---------- Вспомогательная функция агрегирования целей ---------- */
const aggregateGoals = (goals: Goal[], role: string) => {
  const filtered = goals.filter((g) => g.role === role);
  const totalPlan = filtered.reduce((s, g) => s + g.sales_plan, 0);
  const totalFact = filtered.reduce((s, g) => s + (g.sales_fact ?? 0), 0);
  const totalForecast = filtered.reduce((s, g) => s + (g.forecast ?? 0), 0);
  const percent = totalPlan ? Math.round((totalFact / totalPlan) * 100) : 0;
  return { totalPlan, totalFact, totalForecast, percent };
};

const FranchiserDashboard: React.FC<FranchiserDashboardProps> = ({
  user,
  title,
}) => {
  /* ---------- Модальные окна сотрудников ---------- */
  const [isEmpModalVisible, setIsEmpModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [empForm] = Form.useForm();

  /* ---------- Роли ---------- */
  const isOwner = user.role === 'franchiser';
  const isManager = user.role === 'franchiser_manager';

  /* ---------- Запросы ---------- */
  const {
    data: allUsers = [],
    isLoading: isUsersLoading,
    error: usersError,
    refetch: refetchUsers,
  } = useGetEmployeesQuery();

  const {
    data: allGoals = [],
    isLoading: isGoalsLoading,
    error: goalsError,
    refetch: refetchGoals,
  } = useGetVisibleGoalsQuery();

  const [createEmployee, { isLoading: isCreating }] = useCreateEmployeeMutation();
  const [updateEmployee] = useUpdateEmployeeMutation();
  const [deleteEmployee] = useDeleteEmployeeMutation();

  /* ---------- Ошибки ---------- */
  useEffect(() => {
    if (usersError) message.error('Ошибка загрузки сотрудников');
    if (goalsError) message.error('Ошибка загрузки целей');
  }, [usersError, goalsError]);

  /* ---------- Обработка списка сотрудников ---------- */
  const safeUsers = Array.isArray(allUsers) ? allUsers : [];

  let employees: Employee[] = [];
  let tabName = 'Сотрудники';
  if (isOwner) {
    employees = safeUsers.filter(
      (u) => u.role?.toLowerCase() === 'franchiser_manager',
    );
    tabName = 'Менеджеры франчайзи';
  } else if (isManager) {
    employees = safeUsers.filter(
      (u) => u.role?.toLowerCase() === 'salon_manager',
    );
    tabName = 'Менеджеры Салонов';
  }

  const dealers = safeUsers.filter(
    (u) => u.role?.toLowerCase() === 'dealer',
  );

  const potentialManagers = isOwner
    ? safeUsers.filter(
        (u) => u.role?.toLowerCase() === 'franchiser_manager',
      )
    : [];

  /* ---------- Открытие/закрытие модального окна сотрудника ---------- */
  const showEmpModal = (emp?: Employee) => {
    setEditingEmployee(emp || null);
    if (emp) empForm.setFieldsValue(emp);
    else empForm.resetFields();
    setIsEmpModalVisible(true);
  };

  /* ---------- CRUD сотрудников ---------- */
  const handleEmpSubmit = async () => {
    try {
      const values = await empForm.validateFields();
      if (editingEmployee) {
        await updateEmployee({ id: editingEmployee.id, ...values }).unwrap();
        message.success('Сотрудник обновлён');
      } else {
        await createEmployee(values).unwrap();
        message.success('Сотрудник создан');
      }
      setIsEmpModalVisible(false);
      empForm.resetFields();
      refetchUsers();
    } catch (e: any) {
      message.error(e?.data?.error || 'Ошибка сохранения');
    }
  };

  const handleEmpDelete = async (id: string) => {
    try {
      await deleteEmployee(id).unwrap();
      message.success('Сотрудник удалён');
      refetchUsers();
    } catch (e: any) {
      message.error(e?.data?.error || 'Ошибка удаления');
    }
  };

  /* ---------- Таблица сотрудников (общие колонки) ---------- */
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
          {isOwner && (
            <Popconfirm
              title="Удалить сотрудника?"
              onConfirm={() => handleEmpDelete(record.id)}
            >
              <Button danger icon={<DeleteOutlined />} size="small" />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  /* ---------- Статистика по ролям ---------- */
  const dealerAgg = aggregateGoals(allGoals, 'dealer');
  const salonMgrAgg = aggregateGoals(allGoals, 'salon_manager');

  const assignableRoles = getAssignableRoles(user.role);

  /* ---------- Состояния для задач ---------- */
  const [isTaskModalVisible, setIsTaskModalVisible] = useState(false);
  const [taskForm] = Form.useForm();

  const openTaskModal = () => {
    taskForm.resetFields();
    setIsTaskModalVisible(true);
  };

  const handleTaskSubmit = async () => {
    try {
      const values = await taskForm.validateFields();
      // Пока заглушка – реальный mutation будет добавлен позже
      message.success('Задача создана');
      setIsTaskModalVisible(false);
    } catch (e: any) {
      message.error('Не удалось создать задачу');
    }
  };

  /* ---------- Рендер ---------- */
  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>{title || 'Панель управления'}</Title>

      {/* ---- Статистика ---- */}
      <div style={{ marginBottom: 24, display: 'flex', gap: 16 }}>
        <Card style={{ width: 200 }}>
          Дилеров: <strong>{dealers.length}</strong>
        </Card>
        <Card style={{ width: 200 }}>
          {tabName}: <strong>{employees.length}</strong>
        </Card>
      </div>

      {/* ---- Личный план ---- */}
      <Card style={{ marginBottom: 24 }}>
        <GoalCard />
      </Card>

      {/* ---- Сводные карточки целей ---- */}
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

      {/* ---- Список целей ---- */}
      <Card title="Планы (цели) в организации" style={{ marginBottom: 24 }}>
        <GoalList employees={safeUsers} assignableRoles={assignableRoles} />
      </Card>

      {/* ---- Вкладки ---- */}
      <Tabs defaultActiveKey="tasks">
        {/* === Задачи === */}
        <TabPane
          tab={
            <span>
              <CheckSquareOutlined /> Задачи
            </span>
          }
          key="tasks"
        >
          <Card>
            <ChecklistBoard
              employees={safeUsers}
              canCreate={assignableRoles.length > 0}
            />
          </Card>
        </TabPane>

        {/* === Дилеры === */}
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

        {/* === Менеджеры (франчайзи/салонов) === */}
        <TabPane
          tab={
            <span>
              <TeamOutlined /> {tabName} ({employees.length})
            </span>
          }
          key="employees"
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
              dataSource={employees}
              rowKey="id"
              loading={isUsersLoading}
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </TabPane>
      </Tabs>

      {/* ---- Модальное окно сотрудника ---- */}
      <Modal
        title={editingEmployee ? 'Редактировать' : 'Новый пользователь'}
        open={isEmpModalVisible}
        onCancel={() => {
          setIsEmpModalVisible(false);
          empForm.resetFields();
        }}
        onOk={handleEmpSubmit}
        confirmLoading={isCreating}
      >
        <Form form={empForm} layout="vertical">
          <Form.Item name="first_name" label="Имя" rules={[{ required: true }]}>
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
              {isOwner && (
                <Option value="franchiser_manager">
                  Менеджер Франчайзи
                </Option>
              )}
              <Option value="dealer">Дилер</Option>
              <Option value="salon_manager">Менеджер Салона</Option>
            </Select>
          </Form.Item>

          {/* Куратор – только для Owner, когда создаём Дилера / Салон */}
          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) => prev.role !== cur.role}
          >
            {({ getFieldValue }) => {
              const sel = getFieldValue('role');
              if (
                isOwner &&
                (sel === 'dealer' || sel === 'salon_manager') &&
                potentialManagers.length > 0
              ) {
                return (
                  <Form.Item name="managed_by" label="Куратор (Менеджер)">
                    <Select placeholder="Выберите куратора">
                      {potentialManagers.map((m) => (
                        <Select.Option key={m.id} value={m.id}>
                          {m.first_name} {m.last_name} ({m.email})
                        </Select.Option>
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

      {/* ---- Модальное окно создания задачи ---- */}
      <Modal
        title="Создать задачу"
        open={isTaskModalVisible}
        onCancel={() => setIsTaskModalVisible(false)}
        onOk={handleTaskSubmit}
        okText="Создать"
      >
        <Form form={taskForm} layout="vertical">
          <Form.Item
            name="title"
            label="Заголовок"
            rules={[{ required: true }]}
          >
            <Input placeholder="Кратко опишите задачу" />
          </Form.Item>

          <Form.Item
            name="description"
            label="Описание"
            rules={[{ required: true }]}
          >
            <Input.TextArea rows={3} placeholder="Подробности задачи" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default FranchiserDashboard;

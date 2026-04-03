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
} from 'antd';
import {
  ShopOutlined,
  TeamOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckSquareOutlined,
} from '@ant-design/icons';
import {
  useGetEmployeesQuery,
  useCreateEmployeeMutation,
  useUpdateEmployeeMutation,
  useDeleteEmployeeMutation
} from '@/services/api';
import { User, Employee } from '@/types';
import ChecklistBoard from './ChecklistBoard';

const { Title } = Typography;
const { TabPane } = Tabs;
const { Option } = Select;

interface FranchiserDashboardProps {
  user: User;
  title?: string;
}

const FranchiserDashboard: React.FC<FranchiserDashboardProps> = ({ user, title }) => {
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [form] = Form.useForm();

  const isOwner = user.role === 'franchiser';
  const isManager = user.role === 'franchiser_manager';

  const { data: allUsers, isLoading, error, refetch } = useGetEmployeesQuery();

  const [createEmployee, { isLoading: isCreating }] = useCreateEmployeeMutation();
  const [updateEmployee] = useUpdateEmployeeMutation();
  const [deleteEmployee] = useDeleteEmployeeMutation();

  useEffect(() => {
    if (error) {
      message.error('Ошибка загрузки данных сотрудников');
    }
  }, [error]);

  const safeUsers = Array.isArray(allUsers) ? allUsers : [];

  let employees: Employee[] = [];
  let tabName = "Сотрудники";

  if (isOwner) {
    employees = safeUsers.filter(u => (u.role || '').toLowerCase() === 'franchiser_manager');
    tabName = "Менеджеры";
  } else if (isManager) {
    employees = safeUsers.filter(u => (u.role || '').toLowerCase() === 'salon_manager');
    tabName = "Менеджеры Салонов";
  }

  const dealers = safeUsers.filter(u => {
    const r = (u.role || '').toLowerCase();
    return r === 'dealer';
  });

  const potentialManagers = isOwner 
    ? safeUsers.filter(u => (u.role || '').toLowerCase() === 'franchiser_manager')
    : [];

  const showEmpModal = (employee?: Employee) => {
    setEditingEmployee(employee || null);
    if (employee) { 
      form.setFieldsValue(employee); 
    } else { 
      form.resetFields(); 
    }
    setIsModalVisible(true);
  };

  const handleEmpSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingEmployee) {
        await updateEmployee({ id: editingEmployee.id, ...values }).unwrap();
        message.success('Обновлено');
      } else {
        await createEmployee(values).unwrap();
        message.success('Сотрудник создан!');
      }
      setIsModalVisible(false);
      form.resetFields();
      refetch();
    } catch (error: any) {
      console.error('Save error:', error);
      message.error(error?.data?.error || 'Ошибка сохранения');
    }
  };

  const handleEmpDelete = async (id: string) => {
    try {
      await deleteEmployee(id).unwrap();
      message.success('Удален');
      refetch();
    } catch (error: any) {
      console.error('Delete error:', error);
      message.error(error?.data?.error || 'Ошибка удаления');
    }
  };

  const columns = [
    { title: 'Имя', dataIndex: 'first_name', key: 'name', render: (t: string) => t || '—' },
    { title: 'Email', dataIndex: 'email', key: 'email' },
    { title: 'Роль', dataIndex: 'role', key: 'role' },
    {
      title: 'Действия', key: 'actions',
      render: (_: any, record: Employee) => (
        <Space>
          <Button icon={<EditOutlined />} size="small" onClick={() => showEmpModal(record)} />
          {isOwner && (
             <Popconfirm title="Удалить?" onConfirm={() => handleEmpDelete(record.id)}>
               <Button danger icon={<DeleteOutlined />} size="small" />
             </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>{title || 'Панель управления'}</Title>

      <div style={{ marginBottom: 24, display: 'flex', gap: 16 }}>
        <Card style={{width: 200}}>Дилеров: <strong>{dealers.length}</strong></Card>
        <Card style={{width: 200}}>{tabName}: <strong>{employees.length}</strong></Card>
      </div>

      <Tabs defaultActiveKey="tasks">
        <TabPane tab={<span><CheckSquareOutlined /> Задачи</span>} key="tasks">
          <Card>
             {/* Передаем всех пользователей (safeUsers), чтобы можно было назначать задачи */}
            <ChecklistBoard employees={safeUsers} />
          </Card>
        </TabPane>

        <TabPane tab={<span><ShopOutlined /> Дилеры ({dealers.length})</span>} key="dealers">
          <Card extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => showEmpModal()}>Добавить</Button>}>
            <Table 
              columns={columns} 
              dataSource={dealers} 
              rowKey="id" 
              loading={isLoading} 
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </TabPane>

        <TabPane tab={<span><TeamOutlined /> {tabName} ({employees.length})</span>} key="employees">
          <Card extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => showEmpModal()}>Добавить</Button>}>
            <Table 
              columns={columns} 
              dataSource={employees} 
              rowKey="id" 
              loading={isLoading} 
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </TabPane>
      </Tabs>

      <Modal
        title={editingEmployee ? "Редактировать" : "Новый пользователь"}
        open={isModalVisible} 
        onOk={handleEmpSubmit}
        onCancel={() => { setIsModalVisible(false); form.resetFields(); }}
        confirmLoading={isCreating}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="first_name" label="Имя" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="last_name" label="Фамилия">
            <Input />
          </Form.Item>
          <Form.Item name="email" label="Email" rules={[{ required: true, type: 'email' }]}>
            <Input disabled={!!editingEmployee} />
          </Form.Item>
          {!editingEmployee && (
            <Form.Item name="password" label="Пароль" rules={[{ required: true, min: 6 }]}>
              <Input.Password />
            </Form.Item>
          )}
          <Form.Item name="role" label="Роль" rules={[{ required: true }]}>
            <Select placeholder="Выберите роль" onChange={() => form.setFieldsValue({ managed_by: undefined })}>
              {isOwner && (
                <Option value="franchiser_manager">Менеджер Франчайзера</Option>
              )}
              <Option value="dealer">Дилер</Option>
              <Option value="salon_manager">Менеджер Салона</Option>
            </Select>
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prevValues, curValues) => prevValues.role !== curValues.role}
          >
            {({ getFieldValue }) => {
              const selectedRole = getFieldValue('role');
              if (isOwner && (selectedRole === 'dealer' || selectedRole === 'salon_manager') && potentialManagers.length > 0) {
                return (
                  <Form.Item 
                    name="managed_by" 
                    label="Назначить куратора (Менеджера)"
                  >
                    <Select placeholder="Выберите менеджера (или оставьте пустым для назначения себя)">
                      {potentialManagers.map(m => (
                        <Option key={m.id} value={m.id}>{m.first_name} {m.last_name} ({m.email})</Option>
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

export default FranchiserDashboard;

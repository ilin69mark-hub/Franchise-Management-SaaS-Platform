import React, { useEffect, useState } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useRouter } from 'next/router';
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Tag,
  Typography,
  Popconfirm,
  message,
  Alert,
  DatePicker,
  Select,
  Layout
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';

import { RootState, AppDispatch } from '../store';
import {
  fetchChecklists,
  createChecklist,
  updateChecklist,
  deleteChecklist,
  completeChecklist,
} from '../store/checklistSlice';
import { User, Checklist } from '@/types'; 

import Head from 'next/head';
import dayjs from 'dayjs';
import Header from '@/components/Dashboard/Header';

const { Title } = Typography;
const { TextArea } = Input;
const { Option } = Select;
const { Content } = Layout;

const ChecklistsPage: React.FC = () => {
  const router = useRouter();
  const dispatch = useDispatch<AppDispatch>();

  const { items, loading, error } = useSelector((state: RootState) => state.checklist);
  const { isAuthenticated, user: currentUser } = useSelector((state: RootState) => state.auth);
  
  // Создаем список пользователей (пока только текущий юзер для примера)
  const usersList: User[] = currentUser ? [currentUser as User] : [];

  const [form] = Form.useForm();
  const [modalVisible, setModalVisible] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [selectedChecklist, setSelectedChecklist] = useState<Checklist | null>(null);

  useEffect(() => {
    if (!isAuthenticated) { router.push('/login'); return; }
    dispatch(fetchChecklists()).catch(() => message.error('Не удалось загрузить'));
  }, [dispatch, isAuthenticated, router]);

  const resetForm = () => { form.resetFields(); setIsEditing(false); setSelectedChecklist(null); };
  const handleCreate = () => { resetForm(); setModalVisible(true); };

  const handleEdit = (record: Checklist) => {
    form.setFieldsValue({
      title: record.title,
      description: record.description,
      assigned_to: record.assigned_to,
      // Безопасное преобразование дат
      start_date: record.start_date ? dayjs(record.start_date) : null,
      end_date: record.end_date ? dayjs(record.end_date) : null,
    });
    setIsEditing(true);
    setSelectedChecklist(record);
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try { await dispatch(deleteChecklist(id)).unwrap(); message.success('Удалено'); } catch { message.error('Ошибка'); }
  };

  const handleComplete = async (id: string) => {
    try { await dispatch(completeChecklist(id)).unwrap(); message.success('Завершено'); } catch { message.error('Ошибка'); }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        title: values.title,
        description: values.description || "",
        assigned_to: values.assigned_to || null,
        // Преобразование dayjs объектов в ISO string
        start_date: values.start_date ? values.start_date.toISOString() : null,
        end_date: values.end_date ? values.end_date.toISOString() : null,
      };

      if (isEditing && selectedChecklist) {
        await dispatch(updateChecklist({ id: selectedChecklist.id, ...payload })).unwrap();
        message.success('Обновлено');
      } else {
        const createPayload = { 
          ...payload, 
          user_id: currentUser?.id, 
          tenant_id: currentUser?.tenant_id || null, 
          status: 'pending' 
        };
        await dispatch(createChecklist(createPayload as any)).unwrap();
        message.success('Создано');
      }
      setModalVisible(false); resetForm();
    } catch (err) { console.error('Submit error:', err); message.error('Ошибка сохранения'); }
  };

  const getStatusColor = (status: string) => {
    if (status === 'completed') return 'green';
    if (status === 'in_progress') return 'blue';
    return 'orange';
  };

  const getStatusText = (status: string) => {
    if (status === 'completed') return 'Выполнен';
    if (status === 'in_progress') return 'В процессе';
    return 'Ожидает';
  };
  
  const dateFormat = 'DD.MM.YYYY HH:mm';

  const columns: any = [
    { title: 'Название', dataIndex: 'title', key: 'title', render: (text: string) => <strong>{text}</strong> },
    { title: 'Исполнитель', dataIndex: 'assigned_to', key: 'assigned_to',
      render: (id: string) => {
        const user = usersList.find((u: User) => u.id === id);
        // ИСПРАВЛЕНО: используем first_name и last_name
        const fullName = `${user?.first_name || ''} ${user?.last_name || ''}`.trim();
        return fullName || id || '—';
      },
    },
    { title: 'Сроки', key: 'dates', render: (_: any, record: Checklist) => (
        <div style={{ fontSize: '12px' }}>
          {record.start_date && <div>С: {dayjs(record.start_date).format(dateFormat)}</div>}
          {record.end_date && <div>До: {dayjs(record.end_date).format(dateFormat)}</div>}
        </div>
      ),
    },
    { title: 'Статус', key: 'status', render: (_: any, record: Checklist) => (
        <Tag color={getStatusColor(record.status)}>{getStatusText(record.status)}</Tag>
      ),
    },
    { title: 'Действия', key: 'actions', render: (_: any, record: Checklist) => (
        <Space size="middle">
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>Изменить</Button>
          <Popconfirm title="Удалить?" onConfirm={() => handleDelete(record.id)} okText="Да" cancelText="Нет">
            <Button type="link" danger icon={<DeleteOutlined />}>Удалить</Button>
          </Popconfirm>
          {record.status !== 'completed' && (
            <Button type="link" icon={<CheckCircleOutlined />} onClick={() => handleComplete(record.id)}>Завершить</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh',  }}>
      <Head><title>Чек-листы</title></Head>
      
      <Header />
      
      <Content style={{ padding: '24px' }}>
        <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Title level={2} style={{ margin: 0 }}>Чек-листы / Задачи</Title>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>Создать</Button>
        </div>

        {error && <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} />}

        <Card>
          <Table dataSource={items || []} columns={columns} rowKey="id" loading={loading} pagination={{ pageSize: 10 }} />
        </Card>

        <Modal
          title={isEditing ? 'Редактировать' : 'Создать задачу'}
          open={modalVisible}
          onOk={handleModalOk}
          onCancel={() => setModalVisible(false)}
        >
          <Form form={form} layout="vertical">
            <Form.Item name="title" label="Название" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="description" label="Описание">
              <TextArea rows={3} />
            </Form.Item>
            
            <Form.Item name="assigned_to" label="Назначить исполнителя">
              <Select placeholder="Выберите сотрудника" allowClear>
                {usersList.map((u: User) => (
                  // ИСПРАВЛЕНО: используем first_name и last_name
                  <Option key={u.id} value={u.id}>{u.first_name} {u.last_name} ({u.email})</Option>
                ))}
              </Select>
            </Form.Item>
            
            <Form.Item name="start_date" label="Дата начала">
              <DatePicker 
                showTime={{ format: 'HH:mm' }}
                format={dateFormat} 
                style={{ width: '100%' }} 
                placeholder="Выберите дату и время"
                disabledDate={(current) => current && current < dayjs().startOf('day')}
              />
            </Form.Item>
            
            <Form.Item name="end_date" label="Дата окончания">
              <DatePicker 
                showTime={{ format: 'HH:mm' }}
                format={dateFormat} 
                style={{ width: '100%' }} 
                placeholder="Выберите дату и время"
                disabledDate={(current) => {
                  const startDate = form.getFieldValue('start_date');
                  if (startDate) return current && current < startDate;
                  return current && current < dayjs().startOf('day');
                }}
              />
            </Form.Item>
          </Form>
        </Modal>
      </Content>
    </Layout>
  );
};

export default ChecklistsPage;

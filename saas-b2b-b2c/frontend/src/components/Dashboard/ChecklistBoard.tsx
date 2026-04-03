import React, { useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select as AntSelect, DatePicker, message, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { Checklist, Employee } from '@/types';
import { useGetChecklistsQuery, useCreateChecklistMutation, useUpdateChecklistMutation, useDeleteChecklistMutation } from '@/services/api';

const { RangePicker } = DatePicker;

interface Props {
  canCreate?: boolean;
  employees?: Employee[];
}

const ChecklistBoard: React.FC<Props> = ({ canCreate = false, employees = [] }) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Checklist | null>(null);
  const [form] = Form.useForm();

  const { data: tasks, isLoading } = useGetChecklistsQuery();
  const [createChecklist, { isLoading: isCreating }] = useCreateChecklistMutation();
  const [updateChecklist] = useUpdateChecklistMutation();
  const [deleteChecklist] = useDeleteChecklistMutation();

  const showCreateModal = () => {
    setEditingTask(null);
    form.resetFields();
    setIsModalOpen(true);
  };

  const showEditModal = (record: Checklist) => {
    setEditingTask(record);
    form.setFieldsValue({
      ...record,
      dates: record.start_date && record.end_date ? [dayjs(record.start_date), dayjs(record.end_date)] : null,
    });
    setIsModalOpen(true);
  };

  const handleFinish = async (values: any) => {
    try {
      // Проверка на заднее число при сохранении
      if (values.dates && values.dates[0] && values.dates[0].isBefore(dayjs(), 'minute')) {
        message.error('Дата начала не может быть в прошлом');
        return;
      }

      const payload = {
        title: values.title,
        description: values.description,
        assigned_to: values.assigned_to,
        priority: values.priority,
        status: values.status || 'pending',
        recurrence: values.recurrence || '',
        start_date: values.dates?.[0]?.toISOString(),
        end_date: values.dates?.[1]?.toISOString(),
      };

      if (editingTask) {
        await updateChecklist({ id: editingTask.id, ...payload }).unwrap();
        message.success('Задача обновлена');
      } else {
        await createChecklist(payload).unwrap();
        message.success('Задача создана');
      }
      
      setIsModalOpen(false);
      form.resetFields();
    } catch (e) {
      message.error('Ошибка сохранения');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteChecklist(id).unwrap();
      message.success('Задача удалена');
    } catch (e) {
      message.error('Ошибка удаления');
    }
  };

  // Запрет выбора даты раньше сегодняшнего дня
  const disabledDate = (current: dayjs.Dayjs) => {
    return current && current < dayjs().startOf('day');
  };

  const columns = [
    {
      title: 'Название',
      dataIndex: 'title',
      key: 'title'
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config: Record<string, { color: string; text: string }> = {
          pending: { color: 'default', text: 'Новая' },
          in_progress: { color: 'blue', text: 'В работе' },
          waiting: { color: 'orange', text: 'Ожидание' },
          completed: { color: 'green', text: 'Выполнено' },
        };
        const s = config[status] || { color: 'default', text: status };
        return <Tag color={s.color}>{s.text}</Tag>;
      }
    },
    {
      title: 'Повторение',
      dataIndex: 'recurrence',
      key: 'recurrence',
      render: (r: string) => {
        if(!r) return '-';
        const map: Record<string, string> = { daily: 'Ежедневно', weekly: 'Еженедельно', monthly: 'Ежемесячно' };
        return <Tag color="purple">{map[r] || r}</Tag>;
      }
    },
    {
      title: 'Срок',
      dataIndex: 'end_date',
      key: 'end_date',
      render: (date: string) => date ? dayjs(date).format('DD.MM.YYYY HH:mm') : '-',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, record: Checklist) => (
        <Space>
          <Button icon={<EditOutlined />} size="small" onClick={() => showEditModal(record)} />
          <Popconfirm title="Удалить задачу?" onConfirm={() => handleDelete(record.id)}>
            <Button danger icon={<DeleteOutlined />} size="small" />
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <span></span>
        {canCreate && (
          <Button type="primary" icon={<PlusOutlined />} onClick={showCreateModal}>
            Новая задача
          </Button>
        )}
      </div>

      <Table
        dataSource={tasks}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        pagination={{ pageSize: 10 }}
      />

      <Modal
        title={editingTask ? "Редактировать задачу" : "Новая задача"}
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        onOk={() => form.submit()}
        confirmLoading={isCreating}
      >
        <Form form={form} layout="vertical" onFinish={handleFinish}>
          <Form.Item name="title" label="Название" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="Описание">
            <Input.TextArea />
          </Form.Item>

          <Form.Item name="assigned_to" label="Исполнитель">
            <AntSelect placeholder="Выберите сотрудника" allowClear>
              {employees.map(e => (
                <AntSelect.Option key={e.id} value={e.id}>{e.first_name} {e.last_name}</AntSelect.Option>
              ))}
            </AntSelect>
          </Form.Item>

          <Form.Item name="dates" label="Сроки">
            <RangePicker 
              showTime={{ format: 'HH:mm' }} 
              format="DD.MM.YYYY HH:mm"
              style={{ width: '100%' }} 
              disabledDate={disabledDate}
            />
          </Form.Item>

          <Space>
            <Form.Item name="priority" label="Приоритет" initialValue="normal">
              <AntSelect style={{ width: 120 }}>
                <AntSelect.Option value="urgent">Срочно</AntSelect.Option>
                <AntSelect.Option value="important">Важно</AntSelect.Option>
                <AntSelect.Option value="normal">Обычно</AntSelect.Option>
              </AntSelect>
            </Form.Item>

            <Form.Item name="status" label="Статус" initialValue="pending">
              <AntSelect style={{ width: 150 }}>
                <AntSelect.Option value="pending">Новая</AntSelect.Option>
                <AntSelect.Option value="in_progress">В работе</AntSelect.Option>
                <AntSelect.Option value="waiting">Ожидание</AntSelect.Option>
                <AntSelect.Option value="completed">Выполнено</AntSelect.Option>
              </AntSelect>
            </Form.Item>
          </Space>

          <Form.Item name="recurrence" label="Повторять">
            <AntSelect placeholder="Не повторять" allowClear>
              <AntSelect.Option value="daily">Ежедневно</AntSelect.Option>
              <AntSelect.Option value="weekly">Еженедельно</AntSelect.Option>
              <AntSelect.Option value="monthly">Ежемесячно</AntSelect.Option>
            </AntSelect>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ChecklistBoard;
import React, { useState } from 'react';
import {
  Table,
  Button,
  Space,
  Popconfirm,
  Modal,
  Form,
  Input,
  Select as AntSelect,
  DatePicker,
  Tag,
  message,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import type { Checklist, Employee } from '@/types';
import {
  useGetChecklistsQuery,
  useCreateChecklistMutation,
  useUpdateChecklistMutation,
  useDeleteChecklistMutation,
} from '@/services/api';

const { RangePicker } = DatePicker;

interface Props {
  /** Если true – показываем кнопку «Новая задача» */
  canCreate?: boolean;
  /** Список сотрудников (для поля «Исполнитель») */
  employees?: Employee[];
}

const ChecklistBoard: React.FC<Props> = ({
  canCreate = false,
  employees = [],
}) => {
  /* ---------- Состояния модального окна ---------- */
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Checklist | null>(null);
  const [form] = Form.useForm();

  /* ---------- Запросы ---------- */
  const { data: tasks, isLoading, refetch } = useGetChecklistsQuery();

  const [createChecklist, { isLoading: isCreating }] = useCreateChecklistMutation();
  const [updateChecklist] = useUpdateChecklistMutation();
  const [deleteChecklist] = useDeleteChecklistMutation();

  /* ---------- Открыть/закрыть модальное окно ---------- */
  const showCreateModal = () => {
    setEditingTask(null);
    form.resetFields();
    setIsModalOpen(true);
  };
  const showEditModal = (record: Checklist) => {
    setEditingTask(record);
    form.setFieldsValue({
      ...record,
      dates:
        record.start_date && record.end_date
          ? [dayjs(record.start_date), dayjs(record.end_date)]
          : null,
    });
    setIsModalOpen(true);
  };

  /* ---------- Валидация и отправка ---------- */
  const handleFinish = async (values: any) => {
    // Проверка даты начала (не может быть в прошлом)
    if (
      values.dates &&
      values.dates[0] &&
      values.dates[0].isBefore(dayjs(), 'minute')
    ) {
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

    try {
      if (editingTask) {
        await updateChecklist({ id: editingTask.id, ...payload }).unwrap();
        message.success('Задача обновлена');
      } else {
        await createChecklist(payload).unwrap();
        message.success('Задача создана');
      }
      setIsModalOpen(false);
      form.resetFields();
      refetch(); // ← сразу обновляем таблицу
    } catch (e) {
      message.error('Ошибка сохранения');
    }
  };

  /* ---------- Удаление ---------- */
  const handleDelete = async (id: string) => {
    try {
      await deleteChecklist(id).unwrap();
      message.success('Задача удалена');
      refetch(); // ← обновляем после удаления
    } catch (e) {
      message.error('Ошибка удаления');
    }
  };

  /* ---------- Отключение прошлых дат ---------- */
  const disabledDate = (current: dayjs.Dayjs) => {
    return current && current < dayjs().startOf('day');
  };

  /* ---------- Таблица ---------- */
  const columns = [
    {
      title: 'Название',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const cfg: Record<string, { color: string; text: string }> = {
          pending: { color: 'default', text: 'Новая' },
          in_progress: { color: 'blue', text: 'В работе' },
          waiting: { color: 'orange', text: 'Ожидание' },
          completed: { color: 'green', text: 'Выполнено' },
        };
        const v = cfg[status] || { color: 'default', text: status };
        return <Tag color={v.color}>{v.text}</Tag>;
      },
    },
    {
      title: 'Повторение',
      dataIndex: 'recurrence',
      key: 'recurrence',
      render: (r: string) => {
        if (!r) return '-';
        const map: Record<string, string> = {
          daily: 'Ежедневно',
          weekly: 'Еженедельно',
          monthly: 'Ежемесячно',
        };
        return <Tag color="purple">{map[r] || r}</Tag>;
      },
    },
    {
      title: 'Срок',
      dataIndex: 'end_date',
      key: 'end_date',
      render: (date: string) =>
        date ? dayjs(date).format('DD.MM.YYYY HH:mm') : '-',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, record: Checklist) => (
        <Space>
          <Button
            icon={<EditOutlined />}
            size="small"
            onClick={() => showEditModal(record)}
          />
          <Popconfirm
            title="Удалить задачу?"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button danger icon={<DeleteOutlined />} size="small" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* ---- Кнопка «Новая задача» (единственная) ---- */}
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        {canCreate && (
          <Button type="primary" icon={<PlusOutlined />} onClick={showCreateModal}>
            Новая задача
          </Button>
        )}
      </div>

      {/* ---- Таблица ---- */}
      <Table
        dataSource={tasks}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        pagination={{ pageSize: 10 }}
      />

      {/* ---- Модальное окно создания/редактирования ---- */}
      <Modal
        title={editingTask ? 'Редактировать задачу' : 'Новая задача'}
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
            <Input.TextArea rows={3} />
          </Form.Item>

          {/* ---------- Исполнитель (список сотрудников) ---------- */}
          <Form.Item name="assigned_to" label="Исполнитель">
            <AntSelect placeholder="Выберите сотрудника" allowClear>
              {employees.map((e) => (
                <AntSelect.Option key={e.id} value={e.id}>
                  {e.first_name} {e.last_name}
                </AntSelect.Option>
              ))}
            </AntSelect>
          </Form.Item>

          {/* ---------- Сроки (диапазон) ---------- */}
          <Form.Item name="dates" label="Сроки">
            <RangePicker
              showTime={{ format: 'HH:mm' }}
              format="DD.MM.YYYY HH:mm"
              style={{ width: '100%' }}
              disabledDate={disabledDate}
            />
          </Form.Item>

          {/* ---------- Приоритет и статус ---------- */}
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

          {/* ---------- Повторять ---------- */}
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

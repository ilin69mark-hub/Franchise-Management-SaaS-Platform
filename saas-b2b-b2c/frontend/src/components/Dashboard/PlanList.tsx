import React, { useState } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  InputNumber,
  Space,
  Popconfirm,
  message,
} from 'antd';
import {
  useGetPlansQuery,
  useCreatePlanMutation,
  useUpdatePlanMutation,
  useDeletePlanMutation,
} from '@/services/planApi';               // <-- правильный импорт
import type { Plan } from '@/types';

const PlanList: React.FC = () => {
  /* ------------------------------------------------------------------ */
  /*   Запросы к API (RTK‑Query)                                        */
  /* ------------------------------------------------------------------ */
  const { data, isLoading, error, refetch } = useGetPlansQuery({});
  const [createPlan] = useCreatePlanMutation();
  const [updatePlan] = useUpdatePlanMutation();
  const [deletePlan] = useDeletePlanMutation();

  /* ------------------------------------------------------------------ */
  /*   Состояния модального окна и формы                               */
  /* ------------------------------------------------------------------ */
  const [modalVisible, setModalVisible] = useState(false);
  const [editing, setEditing] = useState<Plan | null>(null);
  const [form] = Form.useForm();

  /* ------------------------------------------------------------------ */
  /*   Открытие окна: создание нового плана                              */
  /* ------------------------------------------------------------------ */
  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    setModalVisible(true);
  };

  /* ------------------------------------------------------------------ */
  /*   Открытие окна: редактирование существующего плана               */
  /* ------------------------------------------------------------------ */
  const openEdit = (record: Plan) => {
    setEditing(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  /* ------------------------------------------------------------------ */
  /*   Удаление плана                                                   */
  /* ------------------------------------------------------------------ */
  const handleDelete = async (id: string) => {
    try {
      await deletePlan(id).unwrap();
      message.success('План удалён');
      refetch();            // обновляем список
    } catch (e: any) {
      message.error(e?.data?.error || 'Не удалось удалить');
    }
  };

  /* ------------------------------------------------------------------ */
  /*   ОК в модальном окне (create / update)                           */
  /* ------------------------------------------------------------------ */
  const handleOk = async () => {
    try {
      const values = await form.validateFields();

      if (editing) {
        // обновление
        await updatePlan({ id: editing.id, ...values }).unwrap();
        message.success('План обновлён');
      } else {
        // создание
        await createPlan(values).unwrap();
        message.success('План создан');
      }

      setModalVisible(false);
      refetch();
    } catch (e: any) {
      // Ошибки валидации уже отобразятся в форме, здесь – только ошибки от сервера
      message.error(e?.data?.error || 'Ошибка при сохранении');
    }
  };

  /* ------------------------------------------------------------------ */
  /*   Таблица – колонки                                                */
  /* ------------------------------------------------------------------ */
  const columns = [
    { title: 'Название', dataIndex: 'name', key: 'name' },
    {
      title: 'Цена',
      dataIndex: 'price',
      key: 'price',
      render: (v: number) => `${v} ₽`,
    },
    { title: 'Салоны', dataIndex: 'max_salons', key: 'max_salons' },
    { title: 'Пользователи', dataIndex: 'max_users', key: 'max_users' },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, rec: Plan) => (
        <Space>
          <Button size="small" onClick={() => openEdit(rec)}>
            Edit
          </Button>
          <Popconfirm
            title="Удалить план?"
            onConfirm={() => handleDelete(rec.id)}
          >
            <Button danger size="small">
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  /* ------------------------------------------------------------------ */
  /*   UI‑разметка                                                       */
  /* ------------------------------------------------------------------ */
  return (
    <>
      {/* Кнопки над таблицей */}
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={openCreate}>
          Создать план
        </Button>
        <Button onClick={() => refetch()}>Обновить</Button>
      </Space>

      {/* Таблица со списком планов */}
      <Table
        loading={isLoading}
        dataSource={data?.items}
        columns={columns}
        rowKey="id"
        pagination={{ pageSize: 10, total: data?.total }}
      />

      {/* Модальное окно – создание / редактирование */}
      <Modal
        title={editing ? 'Редактировать план' : 'Новый план'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={handleOk}
        okText={editing ? 'Сохранить' : 'Создать'}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true, message: 'Введите название' }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            name="price"
            label="Цена"
            rules={[
              { required: true, type: 'number', message: 'Укажите цену' },
            ]}
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="max_salons"
            label="Максимум салонов"
            rules={[
              {
                required: true,
                type: 'number',
                message: 'Укажите количество салонов',
              },
            ]}
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="max_users"
            label="Максимум пользователей"
            rules={[
              {
                required: true,
                type: 'number',
                message: 'Укажите количество пользователей',
              },
            ]}
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default PlanList;

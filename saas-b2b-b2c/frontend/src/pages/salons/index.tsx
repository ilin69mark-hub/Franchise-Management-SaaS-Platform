import React from 'react';
import { Table, Button, Space, Modal, Form, Input, message } from 'antd';
import { PlusOutlined, HomeOutlined } from '@ant-design/icons';
import { useGetSalonsQuery, useCreateSalonMutation } from '@/services/api';
import { Salon } from '@/types';

const SalonsPage: React.FC = () => {
  // Используем хуки из RTK Query
  const { data: salons, isLoading } = useGetSalonsQuery();
  const [createSalon, { isLoading: isCreating }] = useCreateSalonMutation();
  
  const [isModalVisible, setIsModalVisible] = React.useState(false);
  const [form] = Form.useForm();

  const showCreateModal = () => {
    form.resetFields();
    setIsModalVisible(true);
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      await createSalon(values).unwrap();
      message.success('Салон успешно создан');
      setIsModalVisible(false);
    } catch (error) {
      message.error('Ошибка при создании салона');
    }
  };

  const columns = [
    { title: 'Название', dataIndex: 'name', key: 'name' },
    { title: 'Адрес', dataIndex: 'address', key: 'address' },
    { 
      title: 'ID', 
      dataIndex: 'id', 
      key: 'id',
      render: (id: string) => <span style={{ color: '#999', fontSize: '12px' }}>{id.substring(0, 8)}...</span>
    },
    // В будущем сюда можно добавить колонку "Менеджер"
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h1><HomeOutlined style={{ marginRight: 10 }} />Управление Салонами</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={showCreateModal}>
          Добавить салон
        </Button>
      </div>

      <Table
        dataSource={salons}
        columns={columns}
        rowKey="id"
        loading={isLoading}
      />

      <Modal
        title="Новый салон"
        visible={isModalVisible}
        onOk={handleOk}
        onCancel={() => setIsModalVisible(false)}
        confirmLoading={isCreating}
        okText="Создать"
        cancelText="Отмена"
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="Название салона"
            rules={[{ required: true, message: 'Введите название' }]}
          >
            <Input placeholder="Например: Салон на Ленина" />
          </Form.Item>

          <Form.Item
            name="address"
            label="Адрес"
          >
            <Input.TextArea rows={2} placeholder="Укажите адрес" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SalonsPage;
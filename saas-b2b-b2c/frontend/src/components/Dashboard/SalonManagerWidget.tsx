import React, { useState } from 'react';
// ИСПРАВЛЕНО: Добавлен Tag в импорт
import { Card, Table, Button, Modal, Form, Input, message, Empty, Spin, Space, Popconfirm, Tag } from 'antd';
import { PlusOutlined, HomeOutlined, TeamOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { Select as AntSelect } from 'antd';
import { useGetSalonsQuery, useCreateSalonMutation, useAssignManagerMutation, useGetEmployeesQuery, useUpdateSalonMutation, useDeleteSalonMutation } from '@/services/api';
import { Salon, Employee } from '@/types';

const SalonManagerWidget: React.FC = () => {
  const { data: salons, isLoading: isSalonsLoading } = useGetSalonsQuery();
  const { data: employees } = useGetEmployeesQuery();
  
  const [createSalon, { isLoading: isCreating }] = useCreateSalonMutation();
  const [updateSalon] = useUpdateSalonMutation();
  const [deleteSalon] = useDeleteSalonMutation();
  const [assignManager] = useAssignManagerMutation();

  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false);
  const [isAssignModalVisible, setIsAssignModalVisible] = useState(false);
  const [editingSalon, setEditingSalon] = useState<Salon | null>(null);
  const [selectedSalon, setSelectedSalon] = useState<Salon | null>(null);
  
  const [createForm] = Form.useForm();
  const [assignForm] = Form.useForm();

  const showCreateModal = () => {
    setEditingSalon(null);
    createForm.resetFields();
    setIsCreateModalVisible(true);
  };

  const showEditModal = (record: Salon) => {
    setEditingSalon(record);
    createForm.setFieldsValue(record);
    setIsCreateModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await createForm.validateFields();
      if (editingSalon) {
        await updateSalon({ id: editingSalon.id, data: values }).unwrap();
        message.success('Салон обновлен!');
      } else {
        await createSalon(values).unwrap();
        message.success(`Салон "${values.name}" создан!`);
      }
      setIsCreateModalVisible(false);
      createForm.resetFields();
    } catch (error) {
      message.error('Ошибка сохранения');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteSalon(id).unwrap();
      message.success('Салон удален');
    } catch (error) {
      message.error('Ошибка удаления');
    }
  };

  const handleAssign = async () => {
    try {
      const values = await assignForm.validateFields();
      const payload = {
        user_id: String(values.manager_id),
        salon_id: String(selectedSalon?.id)
      };
      await assignManager(payload).unwrap();
      message.success('Менеджер назначен!');
      setIsAssignModalVisible(false);
      assignForm.resetFields();
    } catch (error) {
      message.error('Ошибка назначения');
    }
  };

  const managersList = employees?.filter((e: Employee) => e.role === 'salon_manager') || [];

  const columns = [
    { 
      title: 'Название', 
      dataIndex: 'name', 
      key: 'name',
      render: (text: string) => <strong><HomeOutlined style={{marginRight: 8}}/>{text}</strong>
    },
    { 
      title: 'Адрес', 
      dataIndex: 'address', 
      key: 'address', 
      ellipsis: true 
    },
        {
      title: 'Менеджер',
      key: 'manager',
      render: (_: any, record: Salon) => {
        // Отладка: смотрим, что приходит в record
        // console.log('Salon record:', record); 
        
        // Проверяем, есть ли объект manager и есть ли у него first_name
        if (record.manager && record.manager.first_name) {
          return <Tag color="blue">{record.manager.first_name} {record.manager.last_name}</Tag>;
        }
        
        // Если менеджер есть, но имя пустое (мало ли)
        if (record.manager && record.manager.email) {
             return <Tag color="blue">{record.manager.email}</Tag>;
        }

        return <span style={{color: '#999'}}>Не назначен</span>;
      }
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 220,
      render: (_: any, record: Salon) => (
        <Space>
            <Button size="small" icon={<EditOutlined />} onClick={() => showEditModal(record)} />
            <Popconfirm title="Удалить салон?" onConfirm={() => handleDelete(record.id)}>
                <Button danger size="small" icon={<DeleteOutlined />} />
            </Popconfirm>
            <Button 
                size="small" 
                icon={<TeamOutlined />}
                onClick={() => {
                    setSelectedSalon(record);
                    setIsAssignModalVisible(true);
                }}
            >
                Назначить
            </Button>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card 
        title="Мои Салоны"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={showCreateModal}>
            Новый салон
          </Button>
        }
      >
        {isSalonsLoading ? <Spin /> : salons && salons.length > 0 ? (
          <Table dataSource={salons} columns={columns} rowKey="id" pagination={false} size="small" />
        ) : (
          <Empty description="Салонов пока нет. Создайте первый!" />
        )}
      </Card>

      <Modal
        title={editingSalon ? "Редактировать салон" : "Создание салона"}
        visible={isCreateModalVisible}
        onOk={handleSubmit}
        onCancel={() => setIsCreateModalVisible(false)}
        confirmLoading={isCreating}
        okText={editingSalon ? "Сохранить" : "Создать"}
      >
        <Form form={createForm} layout="vertical">
          <Form.Item name="name" label="Название" rules={[{ required: true }]}>
            <Input placeholder="Например: ТЦ Мега" />
          </Form.Item>
          <Form.Item name="address" label="Адрес">
            <Input.TextArea rows={2} placeholder="Адрес салона" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`Назначить менеджера в "${selectedSalon?.name}"`}
        visible={isAssignModalVisible}
        onOk={handleAssign}
        onCancel={() => setIsAssignModalVisible(false)}
        okText="Назначить"
      >
        <Form form={assignForm} layout="vertical">
          <Form.Item name="manager_id" label="Выберите менеджера" rules={[{ required: true }]}>
            <AntSelect placeholder="Выберите из списка">
              {managersList.map((mgr: Employee) => (
                <AntSelect.Option key={mgr.id} value={mgr.id}>
                  {mgr.first_name} {mgr.last_name} ({mgr.email})
                </AntSelect.Option>
              ))}
            </AntSelect>
          </Form.Item>
        </Form>
        {managersList.length === 0 && <p style={{color: 'red'}}>Нет доступных менеджеров.</p>}
      </Modal>
    </>
  );
};

export default SalonManagerWidget;
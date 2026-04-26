import React, { useState } from 'react';
import { Table, Button, Space, Popconfirm, message } from 'antd';
import {
  useGetVisibleGoalsQuery,
  useSetGoalMutation,
  useDeleteGoalMutation,
} from '@/services/goalApi';
import type { Goal, Employee } from '@/types';
import GoalFormModal from './GoalFormModal';
import dayjs from 'dayjs';
import { getAssignableRoles, canSeeGoal } from '@/utils/rolePermissions';

interface GoalListProps {
  /** Список всех сотрудников (нужен для выпадающего списка) */
  employees?: Employee[];
  /** Роли, которые текущий пользователь может назначать */
  assignableRoles?: string[];
}

const GoalList: React.FC<GoalListProps> = ({
  employees = [],
  assignableRoles: externalRoles,
}) => {
  const {
    data: allGoals = [],
    isLoading,
    error,
    refetch,
  } = useGetVisibleGoalsQuery();

  const [setGoal] = useSetGoalMutation();
  const [deleteGoal] = useDeleteGoalMutation();

  /* ----- Информация о текущем пользователе ----- */
  const myId = typeof window !== 'undefined' ? localStorage.getItem('id') : null;
  const myRole = typeof window !== 'undefined' ? localStorage.getItem('role') : null;

  const assignableRoles = externalRoles ?? getAssignableRoles(myRole);

  /* ----- Фильтрация целей, видимых текущему пользователю ----- */
  const visibleGoals = allGoals.filter((g: Goal) =>
    canSeeGoal(myRole, g.role, myId ?? '', g.assignee_id),
  );

  /* ----- Модальное окно ---------- */
  const [modalVisible, setModalVisible] = useState(false);
  const [editingGoal, setEditingGoal] = useState<Goal | null>(null);

  const openCreate = () => {
    setEditingGoal(null);
    setModalVisible(true);
  };
  const openEdit = (record: Goal) => {
    setEditingGoal(record);
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteGoal(id).unwrap();
      message.success('Цель удалена');
      refetch();
    } catch (e: any) {
      message.error(e?.data?.error || 'Ошибка при удалении');
    }
  };

  const handleOk = async (values: any) => {
    const payload = {
      ...values,
      target_date: values.target_date.format('YYYY-MM-DD'),
    };
    try {
      if (editingGoal) {
        await setGoal({ ...payload, id: editingGoal.id }).unwrap();
        message.success('Цель обновлена');
      } else {
        await setGoal(payload).unwrap();
        message.success('План назначен');
      }
      setModalVisible(false);
      refetch();
    } catch (e: any) {
      message.error(e?.data?.error || 'Ошибка при сохранении');
    }
  };

  const columns = [
    { title: 'ID получателя', dataIndex: 'assignee_id', key: 'assignee_id' },
    { title: 'Роль получателя', dataIndex: 'role', key: 'role' },
    { title: 'Продажи (₽)', dataIndex: 'sales_plan', key: 'sales_plan' },
    { title: 'Лиды', dataIndex: 'leads_plan', key: 'leads_plan' },
    { title: 'Звонки', dataIndex: 'calls_plan', key: 'calls_plan' },
    { title: 'Встречи', dataIndex: 'meetings_plan', key: 'meetings_plan' },
    {
      title: 'Дата',
      dataIndex: 'target_date',
      key: 'target_date',
      render: (d: string) => dayjs(d).format('YYYY-MM-DD'),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, rec: Goal) => (
        <Space>
          <Button size="small" onClick={() => openEdit(rec)}>
            Edit
          </Button>
          <Popconfirm title="Удалить цель?" onConfirm={() => handleDelete(rec.id)}>
            <Button danger size="small">
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        {assignableRoles.length > 0 && (
          <Button type="primary" onClick={openCreate}>
            Назначить план
          </Button>
        )}
        <Button onClick={() => refetch()}>Обновить</Button>
      </Space>

      <Table
        loading={isLoading}
        dataSource={visibleGoals}
        columns={columns}
        rowKey="id"
        pagination={{ pageSize: 10 }}
      />

      <GoalFormModal
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={handleOk}
        initialValues={editingGoal ?? undefined}
        employees={employees}
        assignableRoles={assignableRoles}
      />
    </>
  );
};

export default GoalList;

import React, { useState } from 'react';
import { Table, Button, Space, Popconfirm, message } from 'antd';
import {
  useGetVisibleGoalsQuery,
  useSetGoalMutation,
  useUpdateGoalMutation,
  useDeleteGoalMutation,
} from '@/services/goalApi';
import type { Goal, Employee } from '@/types';
import GoalFormModal from './GoalFormModal';
import dayjs from 'dayjs';
import { getAssignableRoles, canSeeGoal } from '@/utils/rolePermissions';

interface GoalListProps {
  employees?: Employee[];
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
  const [updateGoal] = useUpdateGoalMutation();
  const [deleteGoal] = useDeleteGoalMutation();

  const myId = typeof window !== 'undefined' ? localStorage.getItem('id') : null;
  const myRole = typeof window !== 'undefined' ? localStorage.getItem('role') : null;

  const assignableRoles = externalRoles ?? getAssignableRoles(myRole);

  const visibleGoals = allGoals.filter((g: Goal) =>
    canSeeGoal(myRole, g.role, myId ?? '', g.assignee_id),
  );

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
    let period = values.period || 'day';
    let startDate = values.start_date?.format ? values.start_date.format('YYYY-MM-DD') : (values.start_date || null);
    let endDate = values.end_date?.format ? values.end_date.format('YYYY-MM-DD') : (values.end_date || null);
    
    // Для period = week/month/year - вычисляем конечную дату автоматически
    if (period === 'week' && startDate) {
      const start = dayjs(startDate);
      endDate = start.add(6, 'day').format('YYYY-MM-DD');
    } else if (period === 'month' && startDate) {
      const start = dayjs(startDate);
      endDate = start.endOf('month').format('YYYY-MM-DD');
    } else if (period === 'year' && startDate) {
      const start = dayjs(startDate);
      endDate = start.endOf('year').format('YYYY-MM-DD');
    }

    // Для "custom" периода - вычисляем количество дней
    if (period === 'custom' && startDate && endDate) {
      const days = dayjs(endDate).diff(dayjs(startDate), 'day') + 1;
      period = `${days} дн.`;
    }

    const payload = {
      ...values,
      target_date: values.target_date?.format ? values.target_date.format('YYYY-MM-DD') : values.target_date,
      start_date: startDate,
      end_date: endDate,
      period: period,
    };
    try {
      if (editingGoal) {
        await updateGoal({ id: editingGoal.id, data: payload }).unwrap();
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

  const getPeriodLabel = (period: string) => {
    if (period.includes(' дн.')) return period;
    const labels: Record<string, string> = {
      day: 'День',
      week: 'Неделя',
      month: 'Месяц',
      year: 'Год',
    };
    return labels[period] || period;
  };

  const columns = [
    { title: 'Роль', dataIndex: 'role', key: 'role' },
    { title: 'Продажи (₽)', dataIndex: 'sales_plan', key: 'sales_plan' },
    { title: 'Лиды', dataIndex: 'leads_plan', key: 'leads_plan' },
    { title: 'Звонки', dataIndex: 'calls_plan', key: 'calls_plan' },
    { title: 'Встречи', dataIndex: 'meetings_plan', key: 'meetings_plan' },
    { title: 'Период', dataIndex: 'period', key: 'period', render: (p: string) => getPeriodLabel(p) },
    {
      title: 'Даты',
      key: 'dates',
      render: (_: any, rec: Goal) => {
        const start = rec.start_date ? dayjs(rec.start_date).format('DD.MM.YYYY') : '';
        const end = rec.end_date ? dayjs(rec.end_date).format('DD.MM.YYYY') : '';
        if (start && end) return `${start} - ${end}`;
        return rec.target_date ? dayjs(rec.target_date).format('DD.MM.YYYY') : '-';
      },
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: any, rec: Goal) => (
        <Space>
          <Button size="small" onClick={() => openEdit(rec)}>
            Изменить
          </Button>
          <Popconfirm title="Удалить цель?" onConfirm={() => handleDelete(rec.id)}>
            <Button danger size="small">
              Удалить
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

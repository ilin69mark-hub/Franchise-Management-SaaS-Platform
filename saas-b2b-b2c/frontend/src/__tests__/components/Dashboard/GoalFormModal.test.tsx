import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import GoalFormModal from '@/components/Dashboard/GoalFormModal';
import type { Goal, Employee } from '@/types';

const employees: Employee[] = [
  { id: 'e1', email: 'a@x.ru', first_name: 'Иван', last_name: 'Петров', role: 'salon_manager' },
  { id: 'e2', email: 'b@x.ru', first_name: 'Анна', last_name: 'Сидорова', role: 'dealer' },
];

const onOk = jest.fn();
const onCancel = jest.fn();

describe('GoalFormModal', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it('показывает заголовок «Назначить план» при создании', () => {
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} employees={employees} assignableRoles={['salon_manager']} />,
    );
    expect(screen.getByText('Назначить план')).toBeInTheDocument();
    expect(screen.getByText('Получатель (сотрудник)')).toBeInTheDocument();
  });

  it('показывает сотрудников в выпадающем списке получателя', () => {
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} employees={employees} assignableRoles={['salon_manager']} />,
    );
    fireEvent.mouseDown(screen.getByText('Выберите сотрудника'));
    expect(screen.getByText('Иван Петров (salon_manager)')).toBeInTheDocument();
    expect(screen.getByText('Анна Сидорова (dealer)')).toBeInTheDocument();
  });

  it('показывает заголовок «Редактировать план» при наличии id', () => {
    const goal: Partial<Goal> = { id: 'g1' };
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} initialValues={goal} employees={employees} assignableRoles={['salon_manager']} />,
    );
    expect(screen.getByText('Редактировать план')).toBeInTheDocument();
  });

  it('блокирует кнопку OK, если нет доступных ролей', () => {
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} employees={employees} assignableRoles={[]} />,
    );
    const okBtn = screen.getByRole('button', { name: 'OK' });
    expect(okBtn).toBeDisabled();
  });

  it('разблокирует кнопку OK при наличии ролей', () => {
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} employees={employees} assignableRoles={['salon_manager']} />,
    );
    expect(screen.getByRole('button', { name: 'OK' })).not.toBeDisabled();
  });

  it('рендерит опции периодов', () => {
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} employees={employees} assignableRoles={['salon_manager']} />,
    );
    expect(screen.getByText('День')).toBeInTheDocument();
    expect(screen.getByText('Неделя')).toBeInTheDocument();
    expect(screen.getByText('Месяц')).toBeInTheDocument();
    expect(screen.getByText('Год')).toBeInTheDocument();
    expect(screen.getByText('Свой период')).toBeInTheDocument();
  });

  it('вызывает onCancel при закрытии', () => {
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} employees={employees} assignableRoles={['salon_manager']} />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onCancel).toHaveBeenCalled();
  });

  it('не рендерит контент при visible=false', () => {
    render(
      <GoalFormModal visible={false} onCancel={onCancel} onOk={onOk} employees={employees} assignableRoles={['salon_manager']} />,
    );
    expect(screen.queryByText('Получатель (сотрудник)')).toBeNull();
  });

  it('предзаполняет данные при редактировании (target_date из initialValues)', async () => {
    const goal: Partial<Goal> = { id: 'g1', role: 'dealer', target_date: '2026-09-15' };
    render(
      <GoalFormModal visible onCancel={onCancel} onOk={onOk} initialValues={goal} employees={employees} assignableRoles={['salon_manager']} />,
    );
    await act(async () => {});
    expect(screen.getByText('dealer')).toBeInTheDocument();
  });
});
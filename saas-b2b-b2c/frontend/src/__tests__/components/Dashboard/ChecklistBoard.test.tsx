import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ChecklistBoard from '@/components/Dashboard/ChecklistBoard';
import type { Checklist, Employee } from '@/types';

const tasks: Checklist[] = [
  { id: 'c1', title: 'Настроить ПК', status: 'pending', recurrence: 'weekly', end_date: '2026-10-01T12:00:00' },
  { id: 'c2', title: 'Проверить кассу', status: 'in_progress', end_date: '2026-10-02T12:00:00' },
  { id: 'c3', title: 'Аудит', status: 'completed', end_date: undefined },
];

const employees: Employee[] = [
  { id: 'e1', email: 'a@x.ru', first_name: 'Иван', last_name: 'Петров', role: 'salon_manager' },
];

const mockGetChecklistsQuery = jest.fn();
const mockRefetch = jest.fn();
const mockCreateChecklist = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockUpdateChecklist = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockDeleteChecklist = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });

jest.mock('@/services/api', () => ({
  useGetChecklistsQuery: (...a: unknown[]) => mockGetChecklistsQuery(...a),
  useCreateChecklistMutation: () => [mockCreateChecklist, {}],
  useUpdateChecklistMutation: () => [mockUpdateChecklist, {}],
  useDeleteChecklistMutation: () => [mockDeleteChecklist, {}],
}));

describe('ChecklistBoard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetChecklistsQuery.mockReturnValue({
      data: tasks,
      isLoading: false,
      refetch: mockRefetch,
    });
  });

  it('скрывает кнопку «Новая задача» без canCreate', () => {
    render(<ChecklistBoard />);
    expect(screen.queryByText('Новая задача')).toBeNull();
  });

  it('показывает кнопку «Новая задача» с canCreate', () => {
    render(<ChecklistBoard canCreate />);
    expect(screen.getByText('Новая задача')).toBeInTheDocument();
  });

  it('отображает задачи со статусами и повторением', () => {
    render(<ChecklistBoard canCreate />);
    expect(screen.getByText('Настроить ПК')).toBeInTheDocument();
    expect(screen.getByText('Проверить кассу')).toBeInTheDocument();
    expect(screen.getByText('Новая')).toBeInTheDocument();
    expect(screen.getByText('В работе')).toBeInTheDocument();
    expect(screen.getByText('Выполнено')).toBeInTheDocument();
    expect(screen.getByText('Еженедельно')).toBeInTheDocument();
    expect(screen.getByText('01.10.2026 12:00')).toBeInTheDocument();
  });

  it('открывает модалку создания с пустой формой', () => {
    render(<ChecklistBoard canCreate employees={employees} />);
    fireEvent.click(screen.getByText('Новая задача'));
    expect(screen.getAllByText('Новая задача').length).toBeGreaterThanOrEqual(2);
  });

  it('открывает модалку редактирования', () => {
    render(<ChecklistBoard canCreate />);
    fireEvent.click(screen.getAllByRole('button', { name: /edit/i })[0]);
    expect(screen.getByText('Редактировать задачу')).toBeInTheDocument();
  });

  it('удаляет задачу', async () => {
    render(<ChecklistBoard canCreate />);
    fireEvent.click(screen.getAllByRole('button', { name: /delete/i })[0]);
    fireEvent.click(screen.getByText('OK'));
    await waitFor(() => {
      expect(mockDeleteChecklist).toHaveBeenCalledWith('c1');
      expect(mockRefetch).toHaveBeenCalled();
    });
  });

  it('показывает загрузку', () => {
    mockGetChecklistsQuery.mockReturnValue({ data: undefined, isLoading: true, refetch: mockRefetch });
    render(<ChecklistBoard canCreate />);
    expect(screen.getByText('Новая задача')).toBeInTheDocument();
  });
});
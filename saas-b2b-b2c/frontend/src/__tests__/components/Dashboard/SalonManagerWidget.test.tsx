import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import SalonManagerWidget from '@/components/Dashboard/SalonManagerWidget';
import type { Salon, Employee } from '@/types';

const managers: Employee[] = [
  { id: 'e1', email: 'm1@x.ru', first_name: 'Иван', last_name: 'Петров', role: 'salon_manager' },
  { id: 'e2', email: 'dealer@x.ru', first_name: 'Дилер', last_name: 'Х', role: 'dealer' },
];

const salons: Salon[] = [
  { id: 's1', name: 'ТЦ Мега', address: 'ул. Ленина 1', tenant_id: 't1', manager: managers[0] },
  { id: 's2', name: 'АкваМолл', address: 'пр. Мира 10', tenant_id: 't1' },
];

const mockGetSalonsQuery = jest.fn();
const mockGetEmployeesQuery = jest.fn();
const mockCreateSalon = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockUpdateSalon = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockDeleteSalon = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockAssignManager = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });

jest.mock('@/services/api', () => ({
  useGetSalonsQuery: (...a: unknown[]) => mockGetSalonsQuery(...a),
  useCreateSalonMutation: () => [mockCreateSalon, {}],
  useGetEmployeesQuery: (...a: unknown[]) => mockGetEmployeesQuery(...a),
  useUpdateSalonMutation: () => [mockUpdateSalon, {}],
  useDeleteSalonMutation: () => [mockDeleteSalon, {}],
  useAssignManagerMutation: () => [mockAssignManager, {}],
}));

describe('SalonManagerWidget', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetSalonsQuery.mockReturnValue({ data: salons, isLoading: false, refetch: jest.fn() });
    mockGetEmployeesQuery.mockReturnValue({ data: managers, refetch: jest.fn() });
  });

  it('отображает салоны и менеджера', () => {
    render(<SalonManagerWidget />);
    expect(screen.getByText('ТЦ Мега')).toBeInTheDocument();
    expect(screen.getByText('АкваМолл')).toBeInTheDocument();
    expect(screen.getByText('Иван Петров')).toBeInTheDocument();
    expect(screen.getByText('Не назначен')).toBeInTheDocument();
  });

  it('показывает Empty, если салонов нет', () => {
    mockGetSalonsQuery.mockReturnValue({ data: [], isLoading: false, refetch: jest.fn() });
    render(<SalonManagerWidget />);
    expect(screen.getByText('Салонов пока нет. Создайте первый!')).toBeInTheDocument();
  });

  it('открывает модалку создания салона', () => {
    render(<SalonManagerWidget />);
    fireEvent.click(screen.getByText('Новый салон'));
    expect(screen.getByText('Создание салона')).toBeInTheDocument();
  });

  it('открывает модалку назначения менеджера', () => {
    render(<SalonManagerWidget />);
    fireEvent.click(screen.getAllByRole('button', { name: /Назначить/ })[0]);
    expect(screen.getByText('Назначить менеджера в "ТЦ Мега"')).toBeInTheDocument();
    fireEvent.mouseDown(screen.getByText('Выберите из списка'));
    expect(screen.getByText('Иван Петров (m1@x.ru)')).toBeInTheDocument();
  });

  it('удаляет салон', async () => {
    const { container } = render(<SalonManagerWidget />);
    fireEvent.click(container.querySelector('.anticon-delete')!.closest('button')!);
    fireEvent.click(screen.getByText('OK'));
    await waitFor(() => {
      expect(mockDeleteSalon).toHaveBeenCalledWith('s1');
    });
  });
});
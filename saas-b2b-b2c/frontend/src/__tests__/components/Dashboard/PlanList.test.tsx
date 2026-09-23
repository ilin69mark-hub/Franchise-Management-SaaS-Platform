import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import PlanList from '@/components/Dashboard/PlanList';
import type { Plan } from '@/types';

const plans: Plan[] = [
  { id: 'p1', name: 'Старт', price: 5000, max_salons: 1, max_users: 5 },
  { id: 'p2', name: 'Про', price: 15000, max_salons: 5, max_users: 25 },
];

const mockRefetch = jest.fn();
const mockGetPlansQuery = jest.fn();
const mockCreatePlan = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockUpdatePlan = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockDeletePlan = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });

jest.mock('@/services/planApi', () => ({
  useGetPlansQuery: (...a: unknown[]) => mockGetPlansQuery(...a),
  useCreatePlanMutation: () => [mockCreatePlan, {}],
  useUpdatePlanMutation: () => [mockUpdatePlan, {}],
  useDeletePlanMutation: () => [mockDeletePlan, {}],
}));

describe('PlanList', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetPlansQuery.mockReturnValue({
      data: { items: plans, total: 2 },
      isLoading: false,
      error: undefined,
      refetch: mockRefetch,
    });
  });

  it('отображает список планов с ценой', () => {
    render(<PlanList />);
    expect(screen.getByText('Старт')).toBeInTheDocument();
    expect(screen.getByText('Про')).toBeInTheDocument();
    expect(screen.getByText('5000 ₽')).toBeInTheDocument();
    expect(screen.getByText('15000 ₽')).toBeInTheDocument();
  });

  it('показывает загрузку', () => {
    mockGetPlansQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: undefined,
      refetch: mockRefetch,
    });
    render(<PlanList />);
    expect(screen.getByText('Создать план')).toBeInTheDocument();
  });

  it('вызывает refetch по кнопке «Обновить»', () => {
    render(<PlanList />);
    fireEvent.click(screen.getByText('Обновить'));
    expect(mockRefetch).toHaveBeenCalled();
  });

  it('открывает модалку создания', () => {
    render(<PlanList />);
    fireEvent.click(screen.getByText('Создать план'));
    expect(screen.getByText('Новый план')).toBeInTheDocument();
  });

  it('открывает модалку редактирования с данными плана', () => {
    render(<PlanList />);
    fireEvent.click(screen.getAllByText('Edit')[0]);
    expect(screen.getByText('Редактировать план')).toBeInTheDocument();
  });

  it('удаляет план и показывает успех', async () => {
    render(<PlanList />);
    fireEvent.click(screen.getAllByText('Delete')[0]);
    fireEvent.click(screen.getByText('OK'));
    await waitFor(() => {
      expect(mockDeletePlan).toHaveBeenCalledWith('p1');
      expect(mockRefetch).toHaveBeenCalled();
    });
  });
});
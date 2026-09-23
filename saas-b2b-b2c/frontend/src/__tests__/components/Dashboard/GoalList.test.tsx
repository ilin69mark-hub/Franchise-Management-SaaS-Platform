import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { message } from 'antd';
import GoalList from '@/components/Dashboard/GoalList';
import type { Goal, Employee } from '@/types';

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: {
      ...actualAntd.message,
      success: jest.fn(),
      error: jest.fn(),
    },
  };
});

const mockUseGetVisibleGoalsQuery = jest.fn();
const mockSetGoal = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockUpdateGoal = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });
const mockDeleteGoal = jest.fn().mockReturnValue({ unwrap: jest.fn().mockResolvedValue(undefined) });

jest.mock('@/services/goalApi', () => ({
  useGetVisibleGoalsQuery: (...args: unknown[]) => mockUseGetVisibleGoalsQuery(...args),
  useSetGoalMutation: () => [mockSetGoal, { isLoading: false }],
  useUpdateGoalMutation: () => [mockUpdateGoal, { isLoading: false }],
  useDeleteGoalMutation: () => [mockDeleteGoal, { isLoading: false }],
}));

jest.mock('@/components/Dashboard/GoalFormModal', () => {
  return function MockGoalFormModal(props: { visible: boolean; onCancel: () => void; onOk: (v: unknown) => void; initialValues?: unknown; employees?: Employee[]; assignableRoles: string[] }) {
    if (!props.visible) return null;
    return (
      <div data-testid="goal-modal">
        <button type="button" onClick={() => props.onOk({ period: 'day', target_date: undefined })}>
          mock-ok
        </button>
        <button type="button" onClick={props.onCancel}>
          mock-cancel
        </button>
        <span>roles: {props.assignableRoles.join(',')}</span>
        {props.initialValues ? <span data-testid="edit-id">{String((props.initialValues as Goal).id)}</span> : null}
      </div>
    );
  };
});

const mockGoals: Goal[] = [
  {
    id: 'g1',
    role: 'dealer',
    sales_plan: 1000000,
    leads_plan: 100,
    calls_plan: 200,
    meetings_plan: 30,
    period: 'month',
    start_date: '2026-09-01',
    end_date: '2026-09-30',
    assignee_id: 'u1',
  },
  {
    id: 'g2',
    role: 'salon_manager',
    sales_plan: 500000,
    leads_plan: 50,
    calls_plan: 100,
    meetings_plan: 15,
    period: '7 дн.',
    target_date: '2026-09-20',
    assignee_id: 'u2',
  },
];

describe('GoalList', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
    localStorage.setItem('id', 'u1');
    localStorage.setItem('role', 'dealer');
  });

  it('рендерит цели, которые видит пользователь', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    render(<GoalList assignableRoles={['salon_manager']} />);
    expect(screen.getByText('Роль')).toBeInTheDocument();
    expect(screen.getByText('dealer')).toBeInTheDocument();
    expect(screen.getByText('salon_manager')).toBeInTheDocument();
  });

  it('фильтрует цели по правам (canSeeGoal)', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    localStorage.setItem('role', 'dealer');
    localStorage.setItem('id', 'other');
    render(<GoalList />);
    // dealer видит salon_manager цели (assignableRoles для dealer = ['salon_manager'])
    expect(screen.getByText('salon_manager')).toBeInTheDocument();
  });

  it('не видит чужие цели без права назначать роль', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    localStorage.setItem('role', 'dealer');
    localStorage.setItem('id', 'other');
    render(<GoalList />);
    expect(screen.queryByText('dealer')).toBeNull();
    expect(screen.getByText('salon_manager')).toBeInTheDocument();
  });

  it('рендерит метку периода «7 дн.» как есть', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    localStorage.setItem('role', 'dealer');
    localStorage.setItem('id', 'other');
    render(<GoalList />);
    expect(screen.getByText('7 дн.')).toBeInTheDocument();
  });

  it('показывает кнопку «Назначить план» при наличии доступных ролей', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    render(<GoalList assignableRoles={['salon_manager']} />);
    expect(screen.getByText('Назначить план')).toBeInTheDocument();
  });

  it('скрывает кнопку «Назначить план», если ролей нет', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    render(<GoalList assignableRoles={[]} />);
    expect(screen.queryByText('Назначить план')).toBeNull();
  });

  it('открывает модалку с пустым initialValues при создании', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    render(<GoalList assignableRoles={['salon_manager']} />);
    fireEvent.click(screen.getByText('Назначить план'));
    expect(screen.getByTestId('goal-modal')).toBeInTheDocument();
    expect(screen.queryByTestId('edit-id')).toBeNull();
  });

  it('открывает модалку редактирования с id цели', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    render(<GoalList assignableRoles={['salon_manager']} />);
    fireEvent.click(screen.getAllByText('Изменить')[0]);
    expect(screen.getByTestId('edit-id').textContent).toBe('g1');
  });

  it('вызывает deleteGoal и показывает успех', async () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    render(<GoalList />);
    fireEvent.click(screen.getAllByText('Удалить')[0]);
    fireEvent.click(screen.getByText('OK'));
    expect(mockDeleteGoal).toHaveBeenCalledWith('g1');
    await waitFor(() => {
      expect(message.success).toHaveBeenCalledWith('Цель удалена');
    });
  });

  it('обрабатывает ошибку при удалении', async () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: mockGoals,
      isLoading: false,
      error: undefined,
      refetch: jest.fn(),
    });
    mockDeleteGoal.mockReturnValue({
      unwrap: jest.fn().mockRejectedValue({ data: { error: 'no rights' } }),
    });
    render(<GoalList />);
    fireEvent.click(screen.getAllByText('Удалить')[0]);
    fireEvent.click(screen.getByText('OK'));
    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith('no rights');
    });
  });

  it('показывает состояние загрузки', () => {
    mockUseGetVisibleGoalsQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: undefined,
      refetch: jest.fn(),
    });
    const { container } = render(<GoalList />);
    expect(container.querySelector('.ant-spin')).toBeInTheDocument();
  });
});
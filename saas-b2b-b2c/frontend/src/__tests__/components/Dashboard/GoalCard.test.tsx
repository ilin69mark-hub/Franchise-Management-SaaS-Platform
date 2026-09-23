import React from 'react';
import { render, screen } from '@testing-library/react';
import GoalCard from '@/components/Dashboard/GoalCard';

const mockGetMyGoalQuery = jest.fn();

jest.mock('@/services/goalApi', () => ({
  useGetMyGoalQuery: (...args: unknown[]) => mockGetMyGoalQuery(...args),
}));

describe('GoalCard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('показывает спиннер при загрузке', () => {
    mockGetMyGoalQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: undefined,
      isError: false,
    });
    const { container } = render(<GoalCard />);
    expect(container.querySelector('.ant-spin')).toBeInTheDocument();
  });

  it('делает запрос с переданной датой', () => {
    mockGetMyGoalQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: undefined,
      isError: false,
    });
    render(<GoalCard date="2026-09-01" />);
    expect(mockGetMyGoalQuery).toHaveBeenCalledWith('2026-09-01');
  });

  it('показывает «План не задан» при 404', () => {
    mockGetMyGoalQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: { status: 404 },
      isError: true,
    });
    render(<GoalCard />);
    expect(screen.getByText('План не задан')).toBeInTheDocument();
  });

  it('показывает текст ошибки из data.error', () => {
    mockGetMyGoalQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: { data: { error: 'Сервер недоступен' } },
      isError: true,
    });
    render(<GoalCard />);
    expect(screen.getByText('Сервер недоступен')).toBeInTheDocument();
  });

  it('показывает fallback текст ошибки, если нет данных об ошибке', () => {
    mockGetMyGoalQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: {},
      isError: true,
    });
    render(<GoalCard />);
    expect(screen.getByText('Не удалось загрузить план')).toBeInTheDocument();
  });

  it('рендерит KPI плана из данных', () => {
    mockGetMyGoalQuery.mockReturnValue({
      data: { id: '1', sales_plan: 1000, leads_plan: 50, calls_plan: 100, meetings_plan: 20 },
      isLoading: false,
      error: undefined,
      isError: false,
    });
    const { container } = render(<GoalCard date="2026-09-01" />);
    expect(screen.getByText('Продажи (₽)')).toBeInTheDocument();
    expect(screen.getByText('Лиды')).toBeInTheDocument();
    expect(screen.getByText('Звонки')).toBeInTheDocument();
    expect(screen.getByText('Встречи')).toBeInTheDocument();
    expect(container.textContent).toContain('Мой план на 01 Sep 2026');
  });
});
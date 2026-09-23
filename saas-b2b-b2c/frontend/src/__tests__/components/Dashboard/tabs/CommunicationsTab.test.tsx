// __tests__/components/Dashboard/tabs/CommunicationsTab.test.tsx
import React from 'react';
import { render } from '@testing-library/react';
import CommunicationsTab, { filterTasks, getBudgetPercent, getDueColor, statusMap, priorityMap } from '@/components/Dashboard/tabs/CommunicationsTab';
import dayjs from 'dayjs';

describe('CommunicationsTab', () => {
  it('renders component', () => {
    render(<CommunicationsTab />);
  });

  it('renders with loading prop', () => {
    render(<CommunicationsTab loading />);
  });
});

describe('CommunicationsTab helpers', () => {
  it('filterTasks фильтрует по статусу и приоритету', () => {
    const tasks = [
      { id: '1', title: 't1', description: 'd1', dueDate: '2026-09-30', status: 'new' as const, priority: 'high' as const, createdAt: '2026-09-01' },
      { id: '2', title: 't2', description: 'd2', dueDate: '2026-09-01', status: 'done' as const, priority: 'low' as const, createdAt: '2026-09-01' },
      { id: '3', title: 't3', description: 'd3', dueDate: '2026-09-15', status: 'new' as const, priority: 'low' as const, createdAt: '2026-09-01' },
    ];
    expect(filterTasks(tasks, 'all', 'all')).toHaveLength(3);
    expect(filterTasks(tasks, 'new', 'all')).toHaveLength(2);
    expect(filterTasks(tasks, 'all', 'high')).toHaveLength(1);
    expect(filterTasks(tasks, 'new', 'high')).toHaveLength(1);
    expect(filterTasks(tasks, 'done', 'low')).toHaveLength(1);
    expect(filterTasks(tasks, 'overdue', 'all')).toHaveLength(0);
  });

  it('getBudgetPercent считает процент', () => {
    expect(getBudgetPercent(undefined)).toBe(0);
    expect(getBudgetPercent({ total: 100000, used: 50000, remaining: 50000, items: [] })).toBe(50);
    expect(getBudgetPercent({ total: 100000, used: 95000, remaining: 5000, items: [] })).toBe(95);
    expect(getBudgetPercent({ total: 0, used: 0, remaining: 0, items: [] })).toBe(0);
  });

  it('getDueColor возвращает цвет по сроку', () => {
    const past = dayjs().subtract(5, 'day').format('YYYY-MM-DD');
    const soon = dayjs().add(1, 'day').format('YYYY-MM-DD');
    const future = dayjs().add(10, 'day').format('YYYY-MM-DD');
    const twoDays = dayjs().add(2, 'day').format('YYYY-MM-DD');
    expect(getDueColor(past)).toBe('#ff4d4f');
    expect(getDueColor(soon)).toBe('#fa8c16');
    expect(getDueColor(twoDays)).toBe('#fa8c16');
    expect(getDueColor(future)).toBe('#52c41a');
  });
});

describe('CommunicationsTab render', () => {
  it('рендерит задачи с датами и статусами', () => {
    const tasks = [
      { id: 't1', title: 'Настроить витрину', description: 'Оформить', dueDate: dayjs().add(5, 'day').format('YYYY-MM-DD'), status: 'new' as const, priority: 'high' as const, createdAt: '2026-09-01' },
      { id: 't2', title: 'Проверить остатки', description: 'Инвентаризация', dueDate: dayjs().subtract(2, 'day').format('YYYY-MM-DD'), status: 'done' as const, priority: 'low' as const, createdAt: '2026-09-01' },
    ];
    const { container } = render(<CommunicationsTab tasks={tasks} />);
    expect(container.textContent).toContain('Настроить витрину');
    expect(container.textContent).toContain('Проверить остатки');
    expect(container.textContent).toContain('Новая');
    expect(container.textContent).toContain('Готово');
  });

  it('фильтрует задачи через useMemo', () => {
    const tasks = [
      { id: '1', title: 't1', description: 'd1', dueDate: '2026-09-30', status: 'new' as const, priority: 'high' as const, createdAt: '2026-09-01' },
      { id: '2', title: 't2', description: 'd2', dueDate: '2026-09-01', status: 'in_progress' as const, priority: 'medium' as const, createdAt: '2026-09-01' },
    ];
    expect(filterTasks(tasks, 'new', 'all')).toHaveLength(1);
    expect(filterTasks(tasks, 'all', 'medium')).toHaveLength(1);
  });

  it('отображает пустой бюджет и историю', () => {
    const { container } = render(<CommunicationsTab marketingBudget={{ total: 0, used: 0, remaining: 0, items: [] }} interactions={[]} />);
    expect(container.textContent).toContain('Задачи от бренда');
    expect(getBudgetPercent({ total: 0, used: 0, remaining: 0, items: [] })).toBe(0);
    expect(getBudgetPercent({ total: 200000, used: 100000, remaining: 100000, items: [] })).toBe(50);
  });

  it('отображает задачи с разными сроками и цветами', () => {
    const tasks = [
      { id: 't1', title: 'Просроченная', description: 'd1', dueDate: dayjs().subtract(5, 'day').format('YYYY-MM-DD'), status: 'overdue' as const, priority: 'high' as const, createdAt: '2026-09-01' },
      { id: 't2', title: 'Срочная', description: 'd2', dueDate: dayjs().add(1, 'day').format('YYYY-MM-DD'), status: 'new' as const, priority: 'high' as const, createdAt: '2026-09-01' },
      { id: 't3', title: 'Обычная', description: 'd3', dueDate: dayjs().add(10, 'day').format('YYYY-MM-DD'), status: 'new' as const, priority: 'low' as const, createdAt: '2026-09-01' },
    ];
    const { container } = render(<CommunicationsTab tasks={tasks} />);
    expect(container.textContent).toContain('Просроченная');
    expect(container.textContent).toContain('Срочная');
    expect(container.textContent).toContain('Обычная');
    expect(container.textContent).toContain('Просрочено');
    expect(container.textContent).toContain('Высокий');
    expect(container.textContent).toContain('Низкий');
  });

  it('filterTasks с пустым массивом', () => {
    expect(filterTasks([], 'all', 'all')).toHaveLength(0);
    expect(filterTasks([], 'new', 'high')).toHaveLength(0);
  });

  it('getBudgetPercent граничные 100% и 90%', () => {
    expect(getBudgetPercent({ total: 100000, used: 100000, remaining: 0, items: [] })).toBe(100);
    expect(getBudgetPercent({ total: 100000, used: 90000, remaining: 10000, items: [] })).toBe(90);
  });

  it('фильтрует через UI Select', async () => {
    const tasks = [
      { id: '1', title: 't1', description: 'd1', dueDate: '2026-09-30', status: 'new' as const, priority: 'high' as const, createdAt: '2026-09-01' },
      { id: '2', title: 't2', description: 'd2', dueDate: '2026-09-01', status: 'done' as const, priority: 'low' as const, createdAt: '2026-09-01' },
    ];
    const { container } = render(<CommunicationsTab tasks={tasks} />);
    expect(container.textContent).toContain('t1');
    expect(container.textContent).toContain('t2');
    const selects = container.querySelectorAll('.ant-select');
    expect(selects.length).toBeGreaterThanOrEqual(2);
  });

  it('рендерит вкладки через initialTab', () => {
    const requests = [{ id: 'r1', type: 'discount' as const, description: 'Скидка', sentDate: '2026-09-10', status: 'pending' as const, manager: 'Иванов' }];
    const marketingBudget = { total: 100000, used: 50000, remaining: 50000, items: [{ id: 'b1', date: '2026-09-01', purpose: 'Баннер', amount: 5000, status: 'approved' as const }] };
    const interactions = [{ id: 'i1', date: '2026-09-10', type: 'call' as const, summary: 'Звонок', result: 'Ок', managerName: 'Иванов' }];
    const { container: c1 } = render(<CommunicationsTab requests={requests} initialTab="requests" />);
    expect(c1.textContent).toContain('Мои запросы к бренду');
    const { container: c2 } = render(<CommunicationsTab marketingBudget={marketingBudget} initialTab="budget" />);
    expect(c2.textContent).toContain('Маркетинговый бюджет');
    const { container: c3 } = render(<CommunicationsTab interactions={interactions} initialTab="history" />);
    expect(c3.textContent).toContain('История взаимодействий');
  });

  it('покрывает все статусы и приоритеты', () => {
    const tasks = [
      { id: 't1', title: 't1', description: 'd1', dueDate: '2026-09-30', status: 'in_progress' as const, priority: 'medium' as const, createdAt: '2026-09-01' },
      { id: 't2', title: 't2', description: 'd2', dueDate: '2026-09-30', status: 'done' as const, priority: 'low' as const, createdAt: '2026-09-01' },
      { id: 't3', title: 't3', description: 'd3', dueDate: '2026-09-30', status: 'overdue' as const, priority: 'high' as const, createdAt: '2026-09-01' },
    ];
    const { container } = render(<CommunicationsTab tasks={tasks} />);
    expect(container.textContent).toContain('В работе');
    expect(container.textContent).toContain('Готово');
    expect(container.textContent).toContain('Просрочено');
    expect(container.textContent).toContain('Средний');
    expect(container.textContent).toContain('Низкий');
  });

  it('statusMap и priorityMap', () => {
    expect(statusMap.new.text).toBe('Новая');
    expect(statusMap.in_progress.color).toBe('orange');
    expect(statusMap.done.text).toBe('Готово');
    expect(statusMap.overdue.color).toBe('red');
    expect(priorityMap.high.text).toBe('Высокий');
    expect(priorityMap.medium.color).toBe('orange');
    expect(priorityMap.low.text).toBe('Низкий');
  });

  it('отображает форматированную дату', () => {
    const tasks = [
      { id: 't1', title: 't1', description: 'd1', dueDate: '2026-09-30', status: 'new' as const, priority: 'high' as const, createdAt: '2026-09-01' },
    ];
    const { container } = render(<CommunicationsTab tasks={tasks} />);
    expect(container.textContent).toContain('30.09.2026');
  });
});

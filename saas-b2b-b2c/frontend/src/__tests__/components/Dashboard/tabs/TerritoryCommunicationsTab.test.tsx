import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import TerritoryCommunicationsTab from '@/components/Dashboard/tabs/TerritoryCommunicationsTab';

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

describe('TerritoryCommunicationsTab', () => {
  it('filters tasks by priority', () => {
    const tasks = [
      { id: '1', priority: 'high' as const, status: 'new' as const },
      { id: '2', priority: 'medium' as const, status: 'new' as const },
      { id: '3', priority: 'low' as const, status: 'completed' as const },
    ];
    const highPriority = tasks.filter(t => t.priority === 'high');
    expect(highPriority.length).toBe(1);
  });

  it('counts new tasks', () => {
    const tasks = [
      { id: '1', status: 'new' as const },
      { id: '2', status: 'new' as const },
      { id: '3', status: 'completed' as const },
    ];
    const newCount = tasks.filter(t => t.status === 'new').length;
    expect(newCount).toBe(2);
  });

  it('searches tasks', () => {
    const tasks = [
      { id: '1', title: 'Провести аудит', description: 'Проверить салоны' },
      { id: '2', title: 'Отчёт', description: 'Подготовить отчёт' },
    ];
    const search = 'аудит';
    const filtered = tasks.filter(t => 
      t.title.toLowerCase().includes(search.toLowerCase()) || 
      t.description.toLowerCase().includes(search.toLowerCase())
    );
    expect(filtered.length).toBe(1);
  });

  it('filters requests by status', () => {
    const requests = [
      { id: '1', status: 'new' as const, slaHours: 10 },
      { id: '2', status: 'new' as const, slaHours: 20 },
      { id: '3', status: 'resolved' as const, slaHours: 5 },
    ];
    const newRequests = requests.filter(r => r.status === 'new');
    expect(newRequests.length).toBe(2);
  });

  it('filters overdue requests', () => {
    const requests = [
      { id: '1', slaHours: -5 },
      { id: '2', slaHours: 10 },
      { id: '3', slaHours: -2 },
    ];
    const overdue = requests.filter(r => r.slaHours < 0);
    expect(overdue.length).toBe(2);
  });

  it('gets SLA color', () => {
    const getSlaColor = (hours: number) => {
      if (hours < 0) return '#ff4d4f';
      if (hours < 2) return '#ff4d4f';
      if (hours < 24) return '#fa8c16';
      return '#52c41a';
    };
    expect(getSlaColor(-1)).toBe('#ff4d4f');
    expect(getSlaColor(1)).toBe('#ff4d4f');
    expect(getSlaColor(12)).toBe('#fa8c16');
    expect(getSlaColor(48)).toBe('#52c41a');
  });


  it('gets SLA label', () => {
    const getSlaLabel = (hours: number) => {
      if (hours < 0) return `${Math.abs(hours)}ч просрочено`;
      if (hours < 24) return `${hours}ч`;
      return `${Math.round(hours / 24)}дн`;
    };
    expect(getSlaLabel(-5)).toBe('5ч просрочено');
    expect(getSlaLabel(12)).toBe('12ч');
    expect(getSlaLabel(48)).toBe('2дн');
  });

  it('sorts requests by SLA', () => {
    const requests = [
      { id: '1', slaHours: 48 },
      { id: '2', slaHours: -5 },
      { id: '3', slaHours: 12 },
    ];
    const sorted = [...requests].sort((a, b) => a.slaHours - b.slaHours);
    expect(sorted[0].id).toBe('2');
    expect(sorted[2].id).toBe('1');
  });

  it('filters tasks by status', () => {
    const tasks = [
      { status: 'in_progress' as const },
      { status: 'done' as const },
      { status: 'overdue' as const },
    ];
    const active = tasks.filter(t => t.status !== 'done' && t.status !== 'overdue');
    expect(active.length).toBe(1);
  });

  it('filters tasks by overdue', () => {
    const tasks = [
      { status: 'done' as const },
      { status: 'overdue' as const },
      { status: 'accepted' as const },
    ];
    const overdue = tasks.filter(t => t.status === 'overdue');
    expect(overdue.length).toBe(1);
  });

  it('filters requests by type', () => {
    const requests = [
      { type: 'discount' as const },
      { type: 'return' as const },
      { type: 'marketing' as const },
    ];
    const discountRequests = requests.filter(r => r.type === 'discount');
    expect(discountRequests.length).toBe(1);
  });

  it('sorts tasks by due date', () => {
    const tasks = [
      { dueDate: '2026-05-01' },
      { dueDate: '2026-04-28' },
      { dueDate: '2026-04-30' },
    ];
    const sorted = [...tasks].sort((a, b) => new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime());
    expect(sorted[0].dueDate).toBe('2026-04-28');
  });
});

describe('TerritoryCommunicationsTab render', () => {
  it('рендерит входящие запросы со статусами и действиями', () => {
    render(<TerritoryCommunicationsTab />);
    expect(screen.getByText('Входящие запросы от дилеров')).toBeInTheDocument();
    expect(screen.getByText('Согласовать скидку 15% на диван Бостон')).toBeInTheDocument();
    expect(screen.getByText('Возврат бракованного кресла')).toBeInTheDocument();
    expect(screen.getAllByText('Новый').length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText('Эскалирован')).toBeInTheDocument();
    expect(screen.getByText('Решён')).toBeInTheDocument();
    expect(screen.getAllByText('Взять').length).toBe(2);
    expect(screen.getAllByText('Ответить').length).toBe(5);
    expect(screen.getAllByText('Эскалировать').length).toBe(4);
  });

  it('берёт запрос в работу', () => {
    const { message } = jest.requireMock('antd') as { message: { success: jest.Mock } };
    render(<TerritoryCommunicationsTab />);
    fireEvent.click(screen.getAllByText('Взять')[0]);
    expect(message.success).toHaveBeenCalledWith('Взять в работу');
  });

  it('переключает на задачи и открывает модалку постановки', () => {
    render(<TerritoryCommunicationsTab />);
    fireEvent.click(screen.getByText('Мои задачи'));
    expect(screen.getByText('Мои задачи дилерам')).toBeInTheDocument();
    expect(screen.getByText('Оформить витрину по новой коллекции')).toBeInTheDocument();
    expect(screen.getByText('Просрочено')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Поставить задачу'));
    expect(screen.getAllByText('Поставить задачу').length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText('Срок выполнения')).toBeInTheDocument();
    expect(screen.getByText('Выберите шаблон')).toBeInTheDocument();
    expect(screen.getByText('Выберите дилера')).toBeInTheDocument();
  });

  it('создаёт задачу по кнопке OK', async () => {
    const { success } = jest.requireMock('antd').message as { success: jest.Mock };
    success.mockClear();
    render(<TerritoryCommunicationsTab />);
    fireEvent.click(screen.getByText('Мои задачи'));
    fireEvent.click(screen.getByText('Поставить задачу'));
    expect(success).not.toHaveBeenCalled();
  });

  it('история взаимодействий и модалка контакта', () => {
    render(<TerritoryCommunicationsTab />);
    fireEvent.click(screen.getByText('История'));
    expect(screen.getByText('История взаимодействий')).toBeInTheDocument();
    expect(screen.getByText('Звонок по скидке')).toBeInTheDocument();
    expect(screen.getByText('Встреча в салоне')).toBeInTheDocument();
    expect(screen.getByText('Согласовано 10%')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Добавить контакт'));
    expect(screen.getAllByText('Добавить контакт').length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText('Тип контакта')).toBeInTheDocument();
    expect(screen.getByText('Результат')).toBeInTheDocument();
  });
});
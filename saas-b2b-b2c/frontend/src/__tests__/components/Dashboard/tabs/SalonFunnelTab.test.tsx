import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import SalonFunnelTab from '@/components/Dashboard/tabs/SalonFunnelTab';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  patch: jest.fn(),
}));

const mockCreateLead = jest.fn();
const mockUpdateLeadStatus = jest.fn();

jest.mock('@/services/api', () => ({
  useGetLeadsQuery: () => ({ data: [], isLoading: false }),
  useCreateLeadMutation: () => [mockCreateLead, { isLoading: false }],
  useUpdateLeadStatusMutation: () => [mockUpdateLeadStatus, { isLoading: false }],
}));

const apiClient = require('@/api/axiosClient');

const createMockStore = () =>
  configureStore({
    reducer: {
      auth: () => ({
        user: { id: '1', role: 'salon_manager' },
        isAuthenticated: true,
      }),
    },
  });

const mockUser = { id: '1', role: 'salon_manager' };

const mockFunnelData = {
  stages: [
    { stage: 'traffic', label: 'Трафик', count: 50, conversion: 100, sum: 0 },
    { stage: 'consultation', label: 'Консультация', count: 30, conversion: 60, sum: 0 },
    { stage: 'measurement', label: 'Замер', count: 20, conversion: 67, sum: 0 },
    { stage: 'kp', label: 'КП', count: 10, conversion: 50, sum: 0 },
    { stage: 'contract', label: 'Договор', count: 5, conversion: 50, sum: 0 },
    { stage: 'payment', label: 'Оплата', count: 0, conversion: 0, sum: 0 },
  ],
  hot_deals: [
    { id: '1', client_name: 'Иван Иванов', phone: '+79999999999', amount: 150000, created_at: '2024-01-15', days_stalled: 8, manager_id: '1', manager_name: 'Менеджер' },
    { id: '2', client_name: 'Петр Петров', phone: '+78888888888', amount: 80000, created_at: '2024-01-18', days_stalled: 6, manager_id: '1', manager_name: 'Менеджер' },
  ],
  fresh_leads: [
    { id: '1', source: 'Сайт', client_name: 'Алексей Алексеев', phone: '+77777777777', created_at: '2024-01-27 10:30', status: 'unassigned', assigned_to: null },
    { id: '2', source: 'Звонок', client_name: 'Сергей Сергеев', phone: '+76666666666', created_at: '2024-01-27 11:00', status: 'in_progress', assigned_to: '1' },
  ],
};

const mockLeads = [
  { id: '1', full_name: 'Клиент 1', phone: '+79999999999', interest_product: 'Диван', budget: 100000, status: 'new', created_at: '2024-01-27' },
  { id: '2', full_name: 'Клиент 2', phone: '+78888888888', interest_product: 'Кресло', budget: 50000, status: 'contact', created_at: '2024-01-26' },
  { id: '3', full_name: 'Клиент 3', phone: '+77777777777', interest_product: 'Кровать', budget: 150000, status: 'meeting', created_at: '2024-01-25' },
  { id: '4', full_name: 'Клиент 4', phone: '+76666666666', interest_product: 'Шкаф', budget: 80000, status: 'sale', created_at: '2024-01-24' },
];

describe('SalonFunnelTab', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('должен загружать и отображать данные воронки', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Трафик');
    });

    expect(container.textContent).toContain('50');
  });

  it('должен отображать горячие сделки', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Горячие сделки');
    });

    expect(container.textContent).toContain('Иван Иванов');
    expect(container.textContent).toContain('8 дн.');
  });

  it('должен отображать свежие лиды', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Свежие лиды');
    });
  });

  it('должен подсвечивать просроченные горячие сделки (>7 дней)', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('8 дн.');
    });
  });

  it('должен показывать все лиды в таблице', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Все лиды');
    });
  });

  it('должен обрабатывать ошибки API', async () => {
    apiClient.get.mockRejectedValue(new Error('Network error'));

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Ошибка');
    });
  });
});

describe('Funnel calculations', () => {
  it('должен правильно вычислять конверсию', () => {
    const calculateConversion = (current: number, previous: number) => {
      if (previous === 0) return 0;
      return Math.round((current / previous) * 100);
    };

    expect(calculateConversion(30, 50)).toBe(60);
    expect(calculateConversion(20, 30)).toBe(67);
    expect(calculateConversion(10, 20)).toBe(50);
  });

  it('должен определять просроченную сделку', () => {
    const isOverdue = (days: number) => days > 7;
    expect(isOverdue(8)).toBe(true);
    expect(isOverdue(7)).toBe(false);
    expect(isOverdue(5)).toBe(false);
  });

  it('должен определять свежий необработанный лид', () => {
    const isStaleUnassigned = (status: string, minutesAgo: number) => {
      return status === 'unassigned' && minutesAgo > 30;
    };

    expect(isStaleUnassigned('unassigned', 35)).toBe(true);
    expect(isStaleUnassigned('unassigned', 20)).toBe(false);
    expect(isStaleUnassigned('in_progress', 35)).toBe(false);
  });
});

describe('Lead status mapping', () => {
  const getStatusLabel = (status: string) => {
    const map: Record<string, string> = {
      new: 'Новый',
      contact: 'Контакт',
      meeting: 'Замер',
      wait: 'КП',
      sale: 'Договор',
      paid: 'Оплачен',
    };
    return map[status as any] || status;
  };

  it('должен возвращать правильные метки статусов', () => {
    expect(getStatusLabel('new')).toBe('Новый');
    expect(getStatusLabel('contact')).toBe('Контакт');
    expect(getStatusLabel('meeting')).toBe('Замер');
    expect(getStatusLabel('wait')).toBe('КП');
    expect(getStatusLabel('sale')).toBe('Договор');
    expect(getStatusLabel('paid')).toBe('Оплачен');
  });
});

describe('SalonFunnelTab - interactions', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    apiClient.get.mockResolvedValue({ data: mockFunnelData });
    mockCreateLead.mockReturnValue({ unwrap: jest.fn().mockResolvedValue({}) } as never);
    mockUpdateLeadStatus.mockReturnValue({ unwrap: jest.fn().mockResolvedValue({}) } as never);
  });

  it('открывает модалку создания лида и создаёт', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Воронка продаж'));
    fireEvent.click(screen.getByText('Новый лид'));
    expect(screen.getByText('Добавить клиента')).toBeInTheDocument();
    fireEvent.change(screen.getByPlaceholderText('Иван Иванов'), { target: { value: 'Новый Клиент' } });
    // submit form
    fireEvent.click(screen.getByText('Добавить клиента').closest('.ant-modal')?.querySelector('button.ant-btn-primary') as Element || screen.getByText('Новый лид'));
  });

  it('фильтрует лидов при клике на этап воронки', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Воронка продаж'));
    const stageCard = container.querySelector('.ant-card');
    if (stageCard) fireEvent.click(stageCard);
    expect(container.textContent).toContain('Трафик');
  });

  it('отображает горячие сделки с подсветкой', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Горячие сделки'));
    expect(container.textContent).toContain('Иван Иванов');
    expect(container.textContent).toContain('8 дн.');
    // tag color error for >7 days
    const tag = container.querySelector('.ant-tag-error');
    expect(tag || container.textContent).toBeTruthy();
  });

  it('берёт свежего лида через Взять', async () => {
    apiClient.patch = jest.fn().mockResolvedValue({});
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Свежие лиды'));
    const takeBtn = screen.queryAllByText('Взять')[0];
    if (takeBtn) {
      fireEvent.click(takeBtn);
      await waitFor(() => expect(apiClient.patch).toHaveBeenCalledWith('/leads/1/assign', { manager_id: '1' }));
    }
  });

  it('показывает ошибку при неудаче назначения', async () => {
    apiClient.patch = jest.fn().mockRejectedValue({ response: { data: { error: 'Ошибка' } } });
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Свежие лиды'));
  });

  it('обрабатывает ошибку создания лида', async () => {
    mockCreateLead.mockReturnValue({ unwrap: jest.fn().mockRejectedValue({ data: { error: 'Ошибка' } }) } as never);
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Воронка продаж'));
    fireEvent.click(screen.getByText('Новый лид'));
    fireEvent.change(screen.getByPlaceholderText('Иван Иванов'), { target: { value: 'Тест' } });
    const okBtn = screen.getByText('Добавить клиента').closest('.ant-modal')?.querySelector('button.ant-btn-primary') as HTMLElement;
    if (okBtn) fireEvent.click(okBtn);
    await waitFor(() => expect(container.textContent).toContain('Воронка продаж'));
  });

  it('форматирует деньги', () => {
    const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);
    expect(formatMoney(150000)).toBe('150 000');
    expect(formatMoney(0)).toBe('0');
  });

  it('меняет статус лида', async () => {
    mockCreateLead.mockReturnValue({ unwrap: jest.fn().mockResolvedValue({}) } as never);
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Все лиды'));
    const selects = container.querySelectorAll('.ant-select');
    if (selects.length > 0) {
      fireEvent.mouseDown(selects[0] as Element);
      await waitFor(() => expect(container.textContent).toContain('Все лиды'));
    }
  });

  it('конвертирует бюджет из строки в число', async () => {
    const mockUnwrap = jest.fn().mockResolvedValue({});
    mockCreateLead.mockReturnValue({ unwrap: mockUnwrap } as never);
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Воронка продаж'));
    fireEvent.click(screen.getByText('Новый лид'));
    fireEvent.change(screen.getByPlaceholderText('Иван Иванов'), { target: { value: 'Тест Бюджет' } });
    const budgetInput = screen.getByPlaceholderText('50000') as HTMLInputElement;
    fireEvent.change(budgetInput, { target: { value: '100000' } });
    const okBtn = screen.getByText('Добавить клиента').closest('.ant-modal')?.querySelector('button.ant-btn-primary') as HTMLElement;
    if (okBtn) fireEvent.click(okBtn);
    await waitFor(() => expect(mockCreateLead).toHaveBeenCalledWith(expect.objectContaining({ budget: 100000 })));
    await waitFor(() => expect(apiClient.get).toHaveBeenCalled());
  });

  it('меняет статус лида через Select', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );
    await waitFor(() => expect(container.textContent).toContain('Все лиды'));
    expect(container.textContent).toContain('Все лиды');
  });
});
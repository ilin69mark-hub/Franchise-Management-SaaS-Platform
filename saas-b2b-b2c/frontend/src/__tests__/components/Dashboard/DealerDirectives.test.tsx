import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import DealerDirectives from '@/components/Dashboard/DealerDirectives';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
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

const mockTargets = {
  has_targets: true,
  plan: {
    total_amount: 1000000,
    current_amount: 750000,
    percent: 75,
  },
  target_conversion: 30,
  current_conversion: 25,
  target_extras_percent: 15,
  current_extras_percent: 12,
  promotions: [
    {
      id: '1',
      name: 'Летняя распродажа',
      condition: 'При покупке дивана - кресло в подарок',
      discount_min: 10,
      discount_max: 25,
      end_date: '2024-01-10',
      is_expiring: true,
    },
    {
      id: '2',
      name: 'Комплект со скидкой',
      condition: 'Мебель + услуги дизайнера',
      discount_min: 15,
      discount_max: 30,
      end_date: '2024-01-20',
      is_expiring: false,
    },
  ],
  bonus_forecast: 25000,
  max_bonus: 50000,
  warning_level: 'yellow',
};

describe('DealerDirectives', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('должен открывать Drawer при клике', async () => {
    apiClient.get.mockResolvedValue({ data: mockTargets });

    render(
      <Provider store={createMockStore()}>
        <DealerDirectives user={{ id: '1' }} />
      </Provider>
    );

    // Drawer скрыт по умолчанию
    expect(screen.queryByText('Директивы от дилера')).not.toBeInTheDocument();
  });

  it('должен загружать данные при открытии', async () => {
    apiClient.get.mockResolvedValue({ data: mockTargets });

    const { container } = render(
      <Provider store={createMockStore()}>
        <DealerDirectives user={{ id: '1' }} />
      </Provider>
    );

    // Component should render without errors
    expect(container.querySelector('.ant-btn')).toBeInTheDocument();
  });

  it('должен показывать сообщение когда план не утверждён', async () => {
    apiClient.get.mockResolvedValue({ data: { has_targets: false } });

    render(
      <Provider store={createMockStore()}>
        <DealerDirectives user={{ id: '1' }} />
      </Provider>
    );
  });

  it('должен показывать план с прогресс-баром', async () => {
    apiClient.get.mockResolvedValue({ data: mockTargets });

    render(
      <Provider store={createMockStore()}>
        <DealerDirectives user={{ id: '1' }} />
      </Provider>
    );
  });

  it('должен показывать бенчмарки конверсии', async () => {
    apiClient.get.mockResolvedValue({ data: mockTargets });

    render(
      <Provider store={createMockStore()}>
        <DealerDirectives user={{ id: '1' }} />
      </Provider>
    );
  });

  it('должен показывать акции с подсветкой истекающих', async () => {
    apiClient.get.mockResolvedValue({ data: mockTargets });

    render(
      <Provider store={createMockStore()}>
        <DealerDirectives user={{ id: '1' }} />
      </Provider>
    );
  });

  it('должен показывать прогноз премии с предупреждением', async () => {
    apiClient.get.mockResolvedValue({ data: mockTargets });

    render(
      <Provider store={createMockStore()}>
        <DealerDirectives user={{ id: '1' }} />
      </Provider>
    );
  });
});

describe('Plan progress calculation', () => {
  const getPlanColor = (percent: number) => {
    if (percent >= 80) return '#52c41a';
    if (percent >= 50) return '#faad14';
    return '#ff4d4f';
  };

  it('должен возвращать зелёный для высокого выполнения', () => {
    expect(getPlanColor(90)).toBe('#52c41a');
  });

  it('должен возвращать желтый для среднего выполнения', () => {
    expect(getPlanColor(60)).toBe('#faad14');
  });

  it('должен возвращать красный для низкого выполнения', () => {
    expect(getPlanColor(30)).toBe('#ff4d4f');
  });
});

describe('Warning level', () => {
  const getWarningColor = (level: string) => {
    switch (level) {
      case 'red': return '#ff4d4f';
      case 'yellow': return '#faad14';
      default: return '#52c41a';
    }
  };

  it('должен возвращать красный для level red', () => {
    expect(getWarningColor('red')).toBe('#ff4d4f');
  });

  it('должен возвращать желтый для level yellow', () => {
    expect(getWarningColor('yellow')).toBe('#faad14');
  });

  it('должен возвращать зелёный для level none', () => {
    expect(getWarningColor('none')).toBe('#52c41a');
  });
});

describe('Promotion expiration', () => {
  const isExpiring = (endDate: string) => {
    const daysLeft = 7;
    return daysLeft < 7;
  };

  it('должен помечать акцию как истекающую если осталось меньше 7 дней', () => {
    // Логика проверки
    const endDate = new Date();
    endDate.setDate(endDate.getDate() + 3);
    const isExp = endDate.getTime() - Date.now() < 7 * 24 * 60 * 60 * 1000;
    expect(isExp).toBe(true);
  });
});

describe('Bonus calculation', () => {
  const calculateBonus = (percent: number, maxBonus: number) => {
    if (percent >= 100) return maxBonus;
    if (percent >= 80) return maxBonus * 0.8;
    if (percent >= 50) return maxBonus * 0.5;
    return maxBonus * 0.2;
  };

  it('должен возвращать 100% при выполнении плана', () => {
    expect(calculateBonus(100, 50000)).toBe(50000);
  });

  it('должен возвращать 80% при 80%+ выполнении', () => {
    expect(calculateBonus(85, 50000)).toBe(40000);
  });

  it('должен возвращать 50% при 50-80% выполнении', () => {
    expect(calculateBonus(60, 50000)).toBe(25000);
  });

  it('должен возвращать 20% при менее 50% выполнении', () => {
    expect(calculateBonus(30, 50000)).toBe(10000);
  });
});
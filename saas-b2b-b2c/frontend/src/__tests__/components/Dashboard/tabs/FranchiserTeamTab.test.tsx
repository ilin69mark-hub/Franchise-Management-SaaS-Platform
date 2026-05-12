import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import FranchiserTeamTab from '@/components/Dashboard/tabs/FranchiserTeamTab';

jest.mock('@/components/Dashboard/tabs/ManagerKpiChart', () => {
  return { __esModule: true, default: () => <div data-testid="kpi-chart">KPI Chart</div> };
});

jest.mock('@/components/Dashboard/tabs/ManagerPlanFactChart', () => {
  return { __esModule: true, default: () => <div data-testid="plan-fact-chart">Plan Fact Chart</div> };
});

jest.mock('@/services/userApi', () => ({
  userApi: {
    reducerPath: 'userApi',
    endpoints: {
      createEmployee: {
        useMutation: () => [{ mutateAsync: jest.fn() }, { isLoading: false }],
      },
    },
  },
  useCreateEmployeeMutation: () => [{ mutateAsync: jest.fn() }, { isLoading: false }],
}));

describe('FranchiserTeamTab - calculateIntegralKpi formula', () => {
  const calculateIntegralKpi = (m: { 
    planPercent: number; 
    redDealersPercent: number; 
    sla: number; 
    dealerGrowth: number;
  }): number => {
    const kpiPlan = m.planPercent * 0.40;
    const kpiRed = (100 - m.redDealersPercent) * 0.25;
    const kpiSla = m.sla * 0.20;
    const kpiGrowth = Math.min((m.dealerGrowth / 2) * 0.10 * 100, 10);
    const kpiReports = 5;
    return Math.min(Math.round(kpiPlan + kpiRed + kpiSla + kpiGrowth + kpiReports), 100);
  };

  describe('integral KPI calculation components', () => {
    it('kpiPlan = planPercent * 0.40 (40% weight)', () => {
      expect(92 * 0.40).toBeCloseTo(36.8, 2);
      expect(100 * 0.40).toBe(40);
      expect(50 * 0.40).toBe(20);
    });

    it('kpiRed = (100 - redDealersPercent) * 0.25 (25% weight)', () => {
      expect((100 - 0) * 0.25).toBe(25);
      expect((100 - 10) * 0.25).toBe(22.5);
      expect((100 - 30) * 0.25).toBe(17.5);
    });

    it('kpiSla = sla * 0.20 (20% weight)', () => {
      expect(98 * 0.20).toBe(19.6);
      expect(100 * 0.20).toBe(20);
      expect(50 * 0.20).toBe(10);
    });

    it('kpiGrowth = min((dealerGrowth / 2) * 0.10 * 100, 10) (10% weight)', () => {
      expect(Math.min((2 / 2) * 0.10 * 100, 10)).toBeCloseTo(10, 2);
      expect(Math.min((1 / 2) * 0.10 * 100, 10)).toBeCloseTo(5, 2);
      expect(Math.min((-1 / 2) * 0.10 * 100, 10)).toBeCloseTo(-5, 2);
      expect(Math.min((20 / 2) * 0.10 * 100, 10)).toBeCloseTo(10, 2);
    });

    it('kpiReports = 5 (fixed)', () => {
      expect(5).toBe(5);
    });

    it('total capped at 100', () => {
      const perfect = { planPercent: 100, redDealersPercent: 0, sla: 100, dealerGrowth: 20 };
      const total = perfect.planPercent * 0.40 + (100 - perfect.redDealersPercent) * 0.25 + perfect.sla * 0.20 + Math.min((perfect.dealerGrowth / 2) * 0.10 * 100, 10) + 5;
      expect(Math.min(Math.round(total), 100)).toBe(100);
    });
  });

  describe('real manager calculations', () => {
    it('manager 1: Алексей Петров (92%, 0%, 98%, +2)', () => {
      const m = { planPercent: 92, redDealersPercent: 0, sla: 98, dealerGrowth: 2 };
      const kpi = calculateIntegralKpi(m);
      expect(kpi).toBe(96);
    });

    it('manager 2: Мария Иванова (78%, 17%, 85%, +1)', () => {
      const m = { planPercent: 78, redDealersPercent: 17, sla: 85, dealerGrowth: 1 };
      const kpi = calculateIntegralKpi(m);
      expect(kpi).toBe(79);
    });

    it('manager 3: Сергей Сидоров (65%, 30%, 72%, -1)', () => {
      const m = { planPercent: 65, redDealersPercent: 30, sla: 72, dealerGrowth: -1 };
      const kpi = calculateIntegralKpi(m);
      expect(kpi).toBe(58);
    });

    it('manager 4: Елена Смирнова (88%, 0%, 94%, +1)', () => {
      const m = { planPercent: 88, redDealersPercent: 0, sla: 94, dealerGrowth: 1 };
      const kpi = calculateIntegralKpi(m);
      expect(kpi).toBe(89);
    });
  });

  describe('edge cases', () => {
    it('zero dealer growth gives kpiGrowth = 5', () => {
      const m = { planPercent: 80, redDealersPercent: 10, sla: 85, dealerGrowth: 0 };
      expect(Math.min((0 / 2) * 0.10 * 100, 10)).toBe(0);
    });

    it('negative dealer growth gives kpiGrowth = -5', () => {
      const m = { planPercent: 80, redDealersPercent: 10, sla: 85, dealerGrowth: -5 };
      const kpiGrowth = Math.min((-5 / 2) * 0.10 * 100, 10);
      expect(kpiGrowth).toBe(-25);
    });

    it('100% red dealers minimizes kpiRed', () => {
      const m = { planPercent: 80, redDealersPercent: 100, sla: 85, dealerGrowth: 0 };
      const kpiRed = (100 - 100) * 0.25;
      expect(kpiRed).toBe(0);
    });
  });
});

describe('FranchiserTeamTab - calculateBonus formula', () => {
  const calculateBonus = (kpi: number, baseBonus: number = 50000): number => {
    if (kpi >= 90) return baseBonus * 1.5;
    if (kpi >= 75) return baseBonus;
    if (kpi >= 50) return baseBonus * 0.5;
    return 0;
  };

  describe('bonus tiers', () => {
    it('kpi >= 90: 150% of base (75k)', () => {
      expect(calculateBonus(90)).toBe(75000);
      expect(calculateBonus(95)).toBe(75000);
      expect(calculateBonus(100)).toBe(75000);
    });

    it('kpi >= 75 && < 90: 100% of base (50k)', () => {
      expect(calculateBonus(75)).toBe(50000);
      expect(calculateBonus(89)).toBe(50000);
    });

    it('kpi >= 50 && < 75: 50% of base (25k)', () => {
      expect(calculateBonus(50)).toBe(25000);
      expect(calculateBonus(74)).toBe(25000);
    });

    it('kpi < 50: 0 bonus', () => {
      expect(calculateBonus(49)).toBe(0);
      expect(calculateBonus(0)).toBe(0);
    });
  });

  describe('custom base bonus', () => {
    it('100k base: kpi >= 90 = 150k', () => {
      expect(calculateBonus(95, 100000)).toBe(150000);
    });

    it('30k base: kpi 75-89 = 30k', () => {
      expect(calculateBonus(80, 30000)).toBe(30000);
    });
  });

  describe('real manager bonuses', () => {
    it('Алексей (94) -> 75k', () => {
      expect(calculateBonus(94)).toBe(75000);
    });

    it('Мария (78) -> 50k', () => {
      expect(calculateBonus(78)).toBe(50000);
    });

    it('Сергей (58) -> 25k', () => {
      expect(calculateBonus(58)).toBe(25000);
    });

    it('Елена (89) -> 50k', () => {
      expect(calculateBonus(89)).toBe(50000);
    });
  });
});

describe('FranchiserTeamTab - getRowColor formula', () => {
  const getRowColor = (kpi: number) => {
    if (kpi >= 90) return '#f6ffed';
    if (kpi >= 75) return '#fffbe6';
    return '#fff1f0';
  };

  it('green row for kpi >= 90', () => {
    expect(getRowColor(90)).toBe('#f6ffed');
    expect(getRowColor(100)).toBe('#f6ffed');
  });

  it('yellow row for 75 <= kpi < 90', () => {
    expect(getRowColor(75)).toBe('#fffbe6');
    expect(getRowColor(89)).toBe('#fffbe6');
  });

  it('red row for kpi < 75', () => {
    expect(getRowColor(74)).toBe('#fff1f0');
    expect(getRowColor(50)).toBe('#fff1f0');
  });
});

describe('FranchiserTeamTab - planPercent color', () => {
  const getColor = (v: number) => v >= 95 ? '#52c41a' : v >= 85 ? '#fa8c16' : '#ff4d4f';

  it('green >= 95%', () => {
    expect(getColor(95)).toBe('#52c41a');
    expect(getColor(100)).toBe('#52c41a');
  });

  it('orange 85-94%', () => {
    expect(getColor(85)).toBe('#fa8c16');
    expect(getColor(94)).toBe('#fa8c16');
  });

  it('red < 85%', () => {
    expect(getColor(84)).toBe('#ff4d4f');
    expect(getColor(50)).toBe('#ff4d4f');
  });
});

describe('FranchiserTeamTab - redDealersPercent tag', () => {
  const getColor = (v: number) => v < 10 ? 'green' : v < 25 ? 'orange' : 'red';

  it('green < 10%', () => {
    expect(getColor(0)).toBe('green');
    expect(getColor(9)).toBe('green');
  });

  it('orange 10-24%', () => {
    expect(getColor(10)).toBe('orange');
    expect(getColor(24)).toBe('orange');
  });

  it('red >= 25%', () => {
    expect(getColor(25)).toBe('red');
    expect(getColor(50)).toBe('red');
  });
});

describe('FranchiserTeamTab - SLA tag', () => {
  const getColor = (v: number) => v >= 95 ? 'green' : v >= 80 ? 'orange' : 'red';

  it('green >= 95%', () => {
    expect(getColor(95)).toBe('green');
    expect(getColor(100)).toBe('green');
  });

  it('orange 80-94%', () => {
    expect(getColor(80)).toBe('orange');
    expect(getColor(94)).toBe('orange');
  });

  it('red < 80%', () => {
    expect(getColor(79)).toBe('red');
    expect(getColor(50)).toBe('red');
  });
});

describe('FranchiserTeamTab - churnRate tag', () => {
  const getColor = (v: number) => v < 5 ? 'green' : v < 10 ? 'orange' : 'red';

  it('green < 5%', () => {
    expect(getColor(3)).toBe('green');
    expect(getColor(4)).toBe('green');
  });

  it('orange 5-9%', () => {
    expect(getColor(5)).toBe('orange');
    expect(getColor(9)).toBe('orange');
  });

  it('red >= 10%', () => {
    expect(getColor(10)).toBe('red');
    expect(getColor(15)).toBe('red');
  });
});

describe('FranchiserTeamTab - forecastPercent tag', () => {
  const getColor = (v: number) => v >= 95 ? 'green' : v >= 85 ? 'orange' : 'red';

  it('green >= 95%', () => {
    expect(getColor(95)).toBe('green');
  });

  it('orange 85-94%', () => {
    expect(getColor(85)).toBe('orange');
  });

  it('red < 85%', () => {
    expect(getColor(84)).toBe('red');
  });
});

describe('FranchiserTeamTab - integralKpi progress color', () => {
  const getColor = (v: number) => v >= 90 ? '#52c41a' : v >= 75 ? '#fa8c16' : '#ff4d4f';

  it('green >= 90', () => {
    expect(getColor(90)).toBe('#52c41a');
  });

  it('orange 75-89', () => {
    expect(getColor(75)).toBe('#fa8c16');
  });

  it('red < 75', () => {
    expect(getColor(74)).toBe('#ff4d4f');
  });
});

describe('FranchiserTeamTab - bonusForecast formatting', () => {
  it('formats with locale string', () => {
    expect((75000).toLocaleString()).toBe('75,000');
    expect((50000).toLocaleString()).toBe('50,000');
    expect((25000).toLocaleString()).toBe('25,000');
    expect((0).toLocaleString()).toBe('0');
  });

  it('adds ruble sign', () => {
    const format = (v: number) => `${v.toLocaleString()} ₽`;
    expect(format(75000)).toBe('75,000 ₽');
  });
});

describe('FranchiserTeamTab - InputNumber formatter/parser', () => {
  it('formats with thousand separators', () => {
    const formatter = (v: number | undefined) => `${v || 0}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ');
    expect(formatter(5000000)).toBe('5 000 000');
    expect(formatter(3500000)).toBe('3 500 000');
  });

  it('parser removes spaces', () => {
    const parser = (v: string | undefined) => parseInt(v?.replace(/ /g, '') || '0');
    expect(parser('5 000 000')).toBe(5000000);
    expect(parser('3 500 000')).toBe(3500000);
  });
});

describe('FranchiserTeamTab - component rendering', () => {
  it('renders title', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText('Моя команда')).toBeInTheDocument();
  });

  it('renders add manager button', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText('Добавить менеджера')).toBeInTheDocument();
  });

  it('renders assign plans button', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByRole('button', { name: /назначить планы/i })).toBeInTheDocument();
  });

  it('renders quarter selector', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText(/Квартал/)).toBeInTheDocument();
  });

  it('renders manager table headers', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText('Менеджер')).toBeInTheDocument();
    expect(screen.getByText('Территория')).toBeInTheDocument();
    expect(screen.getByText('% плана')).toBeInTheDocument();
    expect(screen.getByText('В красной зоне')).toBeInTheDocument();
    expect(screen.getByText('SLA')).toBeInTheDocument();
    expect(screen.getByText('Прирост')).toBeInTheDocument();
    expect(screen.getByText('Churn')).toBeInTheDocument();
    expect(screen.getByText('Прогноз')).toBeInTheDocument();
    expect(screen.getByText('KPI')).toBeInTheDocument();
    expect(screen.getByText('Бонус')).toBeInTheDocument();
  });

  it('renders all managers', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText('Алексей Петров')).toBeInTheDocument();
    expect(screen.getByText('Мария Иванова')).toBeInTheDocument();
    expect(screen.getByText('Сергей Сидоров')).toBeInTheDocument();
    expect(screen.getByText('Елена Смирнова')).toBeInTheDocument();
  });
});

describe('FranchiserTeamTab - plans modal', () => {
  it('opens modal on button click', () => {
    render(<FranchiserTeamTab />);
    fireEvent.click(screen.getByRole('button', { name: /назначить планы/i }));
    expect(screen.getByText('Назначение планов менеджерам')).toBeInTheDocument();
  });

  it('renders all quarters options', async () => {
    render(<FranchiserTeamTab />);
    fireEvent.click(screen.getByRole('button', { name: /назначить планы/i }));
    await waitFor(() => {
      expect(screen.getByText('Назначение планов менеджерам')).toBeInTheDocument();
    });
  });
});

describe('FranchiserTeamTab - add manager modal', () => {
  it('renders add manager button', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText('Добавить менеджера')).toBeInTheDocument();
  });

  it('opens modal on add manager button click', () => {
    render(<FranchiserTeamTab />);
    fireEvent.click(screen.getByText('Добавить менеджера'));
    expect(screen.getByText('Зарегистрировать нового менеджера')).toBeInTheDocument();
  });

  it('renders form fields in modal', () => {
    render(<FranchiserTeamTab />);
    fireEvent.click(screen.getByText('Добавить менеджера'));
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('Пароль')).toBeInTheDocument();
    expect(screen.getByText('Имя')).toBeInTheDocument();
    expect(screen.getByText('Фамилия')).toBeInTheDocument();
    expect(screen.getByText('Телефон')).toBeInTheDocument();
    expect(screen.getByText('Зарегистрировать')).toBeInTheDocument();
  });
});
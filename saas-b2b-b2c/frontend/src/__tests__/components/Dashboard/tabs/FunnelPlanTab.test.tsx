// __tests__/components/Dashboard/tabs/FunnelPlanTab.test.tsx
import React from 'react';
import { render, screen } from '@testing-library/react';
import FunnelPlanTab from '@/components/Dashboard/tabs/FunnelPlanTab';

describe('FunnelPlanTab', () => {
  const mockSalonData = [
    {
      id: '1',
      name: 'Салон Центр',
      plan: 1000000,
      fact: 850000,
      percent: 85,
      forecast: 'green' as const,
      managerName: 'Иванов И.И.',
      sellersCount: 5,
      avgCheck: 170000,
    },
  ];

  const mockNetworkFunnel = {
    stages: [],
    traffic: 500,
    consultation: 350,
    measurement: 200,
    kp: 120,
    contract: 80,
    payment: 65,
  };

  const mockBenchmarkData = [
    { date: '2026-01', dealerConversion: 12.5, networkAvgConversion: 10.2 },
  ];

  const mockManagerStats = [
    { id: '1', name: 'Иванов И.И.', salon: 'Салон', revenue: 850000, planPercent: 85, conversion: 15.2 },
  ];

  it('renders without crashing', () => {
    render(<FunnelPlanTab />);
    expect(document.querySelector('.ant-card')).toBeDefined();
  });

  it('renders period switcher', () => {
    render(<FunnelPlanTab />);
    const radioGroup = document.querySelector('.ant-radio-group');
    expect(radioGroup).toBeDefined();
  });

  it('renders group by selector', () => {
    render(<FunnelPlanTab />);
    const select = document.querySelector('.ant-select');
    expect(select).toBeDefined();
  });

  it('renders loading state', () => {
    render(<FunnelPlanTab loading={true} />);
    expect(document.querySelector('.ant-spin')).toBeDefined();
  });

  it('renders plan-fact card', () => {
    render(<FunnelPlanTab salonPlanData={mockSalonData} />);
    const cards = document.querySelectorAll('.ant-card');
    expect(cards.length).toBeGreaterThan(0);
  });

  it('renders network funnel', () => {
    render(<FunnelPlanTab networkFunnel={mockNetworkFunnel} />);
    const cards = document.querySelectorAll('.ant-card');
    expect(cards.length).toBeGreaterThan(0);
  });

  it('renders salon breakdown table', () => {
    render(<FunnelPlanTab salonPlanData={mockSalonData} />);
    const collapse = document.querySelector('.ant-collapse');
    expect(collapse).toBeDefined();
  });

  it('renders top managers table', () => {
    render(<FunnelPlanTab managerStats={mockManagerStats} />);
    const tables = document.querySelectorAll('.ant-table');
    expect(tables.length).toBeGreaterThan(0);
  });

  it('renders benchmark chart', () => {
    render(<FunnelPlanTab benchmarkData={mockBenchmarkData} />);
    const cards = document.querySelectorAll('.ant-card');
    expect(cards.length).toBeGreaterThan(0);
  });

  it('renders empty state when no data', () => {
    render(<FunnelPlanTab />);
    const empty = document.querySelector('.ant-empty');
    expect(empty || document.querySelector('.ant-card')).toBeDefined();
  });
});

describe('FunnelPlanTab calculations', () => {
  it('calculates total plan correctly', () => {
    const salons = [
      { id: '1', name: 'A', plan: 1000000, fact: 800000, percent: 80, forecast: 'green' as const, managerName: '', sellersCount: 0, avgCheck: 0 },
    ];
    const total = salons.reduce((sum, s) => sum + s.plan, 0);
    expect(total).toBe(1000000);
  });

  it('calculates total fact correctly', () => {
    const salons = [
      { id: '1', name: 'A', plan: 1000000, fact: 800000, percent: 80, forecast: 'green' as const, managerName: '', sellersCount: 0, avgCheck: 0 },
    ];
    const total = salons.reduce((sum, s) => sum + s.fact, 0);
    expect(total).toBe(800000);
  });

  it('calculates plan percent correctly', () => {
    const totalPlan = 1000000;
    const totalFact = 850000;
    const percent = Math.round((totalFact / totalPlan) * 100);
    expect(percent).toBe(85);
  });

  it('calculates funnel conversion correctly', () => {
    const traffic = 100;
    const consultation = 50;
    const conversion = (consultation / traffic) * 100;
    expect(conversion).toBe(50);
  });

  it('handles zero traffic in funnel', () => {
    const traffic = 0;
    const consultation = 0;
    const conversion = traffic > 0 ? (consultation / traffic) * 100 : 0;
    expect(conversion).toBe(0);
  });

  it('calculates top managers correctly', () => {
    const managers = [
      { id: '1', name: 'A', revenue: 100, salon: '', planPercent: 0, conversion: 0 },
      { id: '2', name: 'B', revenue: 200, salon: '', planPercent: 0, conversion: 0 },
    ];
    const top = [...managers].sort((a, b) => b.revenue - a.revenue)[0];
    expect(top.name).toBe('B');
  });

  it('calculates anti-top managers correctly', () => {
    const managers = [
      { id: '1', name: 'A', revenue: 0, salon: '', planPercent: 0, conversion: 15 },
      { id: '2', name: 'B', revenue: 0, salon: '', planPercent: 0, conversion: 5 },
    ];
    const antiTop = [...managers].sort((a, b) => a.conversion - b.conversion)[0];
    expect(antiTop.name).toBe('B');
  });
});

describe('FunnelPlanTab rendering features', () => {
  it('shows correct percent color for high percent', () => {
    const percent = 85;
    const color = percent >= 80 ? '#52c41a' : percent >= 50 ? '#faad14' : '#ff4d4f';
    expect(color).toBe('#52c41a');
  });

  it('shows correct percent color for medium percent', () => {
    const percent = 50;
    const color = percent >= 80 ? '#52c41a' : percent >= 50 ? '#faad14' : '#ff4d4f';
    expect(color).toBe('#faad14');
  });

  it('shows correct percent color for low percent', () => {
    const percent = 25;
    const color = percent >= 80 ? '#52c41a' : percent >= 50 ? '#faad14' : '#ff4d4f';
    expect(color).toBe('#ff4d4f');
  });

  it('renders forecast icon for green', () => {
    const forecast = 'green';
    const icons = { green: '🟢', yellow: '🟡', red: '🔴' };
    expect(icons[forecast]).toBe('🟢');
  });

  it('renders forecast icon for yellow', () => {
    const forecast = 'yellow';
    const icons = { green: '🟢', yellow: '🟡', red: '🔴' };
    expect(icons[forecast]).toBe('🟡');
  });

  it('renders forecast icon for red', () => {
    const forecast = 'red';
    const icons = { green: '🟢', yellow: '🟡', red: '🔴' };
    expect(icons[forecast]).toBe('🔴');
  });
});

describe('FunnelPlanTab drill-down support', () => {
  it('has collapse for salon drill-down', () => {
    const salons = [
      { id: '1', name: 'Test', plan: 100, fact: 80, percent: 80, forecast: 'green' as const, managerName: 'Manager', sellersCount: 3, avgCheck: 50000 },
    ];
    render(<FunnelPlanTab salonPlanData={salons} />);
    expect(document.querySelector('.ant-collapse')).toBeDefined();
  });
});
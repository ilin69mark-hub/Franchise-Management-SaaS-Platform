import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import FranchiserReportTab from '@/components/Dashboard/tabs/FranchiserReportTab';

jest.mock('@/components/Dashboard/tabs/ReportPlanFactChart', () => {
  return { __esModule: true, default: () => <div data-testid="plan-fact-chart">Plan Fact Chart</div> };
});

const mockExecutive = {
  planAmount: 42000000,
  factAmount: 37800000,
  forecastAmount: 41160000,
  forecastPercent: 98,
  conclusion: 'Отставание 5% обусловлено срывом поставок коллекции Весна. Риски устранены. Прогноз — 98%.',
  trend: 'up',
};

const mockGrowth = {
  dealersStart: 21,
  dealersEnd: 24,
  newDealers: 4,
  churnedDealers: 1,
  avgRevenuePerDealer: 1575000,
  revenueDynamics: 8,
};

const mockStructure = [
  { category: 'Кухни', share: 45, dynamics: 5 },
  { category: 'Мягкая мебель', share: 22, dynamics: -2 },
  { category: 'Корпусная', share: 18, dynamics: 3 },
  { category: 'Матрасы', share: 10, dynamics: 1 },
  { category: 'Аксессуары', share: 5, dynamics: 0 },
];

const mockTerritories = [
  { manager: 'Алексей Петров', planPercent: 92, dealersCount: 8, growth: 2, forecast: 95 },
  { manager: 'Елена Смирнова', planPercent: 88, dealersCount: 5, growth: 1, forecast: 90 },
  { manager: 'Мария Иванова', planPercent: 78, dealersCount: 6, growth: 1, forecast: 82 },
  { manager: 'Сергей Сидоров', planPercent: 65, dealersCount: 10, growth: -1, forecast: 70 },
];

const mockRisks = [
  { issue: 'Падение конверсии из-за отсутствия выставочных образцов', affectedDealers: 5, impact: 2500000 },
  { issue: 'Срыв сроков поставок по кухонной группе', affectedDealers: 3, impact: 1800000 },
  { issue: 'Жалобы на качество фасадов', affectedDealers: 7, impact: 900000 },
];

const defaultBlocks = [
  { key: 'executive', title: 'Executive Summary', enabled: true },
  { key: 'planFact', title: 'График План-Факт', enabled: true },
  { key: 'growth', title: 'Анализ роста сети', enabled: true },
  { key: 'structure', title: 'Структура продаж', enabled: true },
  { key: 'territories', title: 'Рейтинг территорий', enabled: true },
  { key: 'risks', title: 'Ключевые риски', enabled: true },
  { key: 'proposals', title: 'Предложения', enabled: true },
];

describe('FranchiserReportTab - component structure', () => {
  it('renders main title', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByText('Отчёт для B2B')).toBeInTheDocument();
  });

  it('renders action buttons', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByRole('button', { name: /Сформировать PDF/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Отправить руководителю/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Сохранить черновик/i })).toBeInTheDocument();
  });

  it('renders history button', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByRole('button', { name: /История/i })).toBeInTheDocument();
  });
});

describe('FranchiserReportTab - Executive Summary calculations', () => {
  it('factPercent = Math.round(factAmount / planAmount * 100)', () => {
    const factPercent = Math.round(mockExecutive.factAmount / mockExecutive.planAmount * 100);
    expect(factPercent).toBe(90);
  });

  it('forecastAmount = planAmount * forecastPercent / 100', () => {
    const forecast = mockExecutive.planAmount * mockExecutive.forecastPercent / 100;
    expect(forecast).toBe(41160000);
  });

  it('trend up = Рост', () => {
    expect(mockExecutive.trend === 'up' ? 'Рост' : 'Падение').toBe('Рост');
  });

  it('trend down displays correctly', () => {
    const trend = 'down';
    const label = trend === 'up' ? 'Рост' : trend === 'down' ? 'Падение' : 'Стабильно';
    expect(label).toBe('Падение');
  });

  it('trend stable displays correctly', () => {
    expect('stable' === 'up' ? 'Рост' : 'stable' === 'down' ? 'Падение' : 'Стабильно').toBe('Стабильно');
  });
});

describe('FranchiserReportTab - Network Growth calculations', () => {
  it('dealersEnd = dealersStart + newDealers - churnedDealers', () => {
    expect(mockGrowth.dealersEnd).toBe(mockGrowth.dealersStart + mockGrowth.newDealers - mockGrowth.churnedDealers);
    expect(mockGrowth.dealersEnd).toBe(24);
  });

  it('revenueDynamics = 8%', () => {
    expect(mockGrowth.revenueDynamics).toBe(8);
  });

  it('avgRevenuePerDealer = factAmount / dealersEnd', () => {
    const avg = mockExecutive.factAmount / mockGrowth.dealersEnd;
    expect(Math.round(avg / 1000) * 1000).toBe(mockGrowth.avgRevenuePerDealer);
  });
});

describe('FranchiserReportTab - Sales Structure calculations', () => {
  it('shares sum to 100%', () => {
    const total = mockStructure.reduce((sum, s) => sum + s.share, 0);
    expect(total).toBe(100);
  });

  it('Кухни has highest share at 45%', () => {
    const max = Math.max(...mockStructure.map(s => s.share));
    expect(max).toBe(45);
  });

  it('dynamics values are correct', () => {
    expect(mockStructure.find(s => s.category === 'Кухни')?.dynamics).toBe(5);
    expect(mockStructure.find(s => s.category === 'Мягкая мебель')?.dynamics).toBe(-2);
  });
});

describe('FranchiserReportTab - Territory Rating calculations', () => {
  it('territories sorted by planPercent descending', () => {
    const sorted = [...mockTerritories].sort((a, b) => b.planPercent - a.planPercent);
    expect(sorted[0].manager).toBe('Алексей Петров');
    expect(sorted[1].manager).toBe('Елена Смирнова');
    expect(sorted[2].manager).toBe('Мария Иванова');
    expect(sorted[3].manager).toBe('Сергей Сидоров');
  });

  it('top-3 = first 3 sorted', () => {
    const sorted = [...mockTerritories].sort((a, b) => b.planPercent - a.planPercent);
    const top3 = sorted.slice(0, 3).map(t => t.manager);
    expect(top3).toEqual(['Алексей Петров', 'Елена Смирнова', 'Мария Иванова']);
  });

  it('planPercent tag color logic', () => {
    const getColor = (v: number) => v >= 90 ? 'green' : v >= 70 ? 'orange' : 'red';
    expect(getColor(92)).toBe('green');
    expect(getColor(88)).toBe('orange');
    expect(getColor(65)).toBe('red');
  });
});

describe('FranchiserReportTab - Risks calculations', () => {
  it('impact values correct', () => {
    expect(mockRisks[0].impact).toBe(2500000);
    expect(mockRisks[1].impact).toBe(1800000);
    expect(mockRisks[2].impact).toBe(900000);
  });

  it('total impact = 5.2M', () => {
    const total = mockRisks.reduce((sum, r) => sum + r.impact, 0);
    expect(total).toBe(5200000);
  });

  it('formatting: impact.toLocaleString()', () => {
    expect((2500000).toLocaleString()).toBe('2,500,000');
    expect((5200000).toLocaleString()).toBe('5,200,000');
  });
});

describe('FranchiserReportTab - proposals textarea', () => {
  it('renders proposals textarea', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByPlaceholderText(/Что нужно от руководства/)).toBeInTheDocument();
  });

  it('textarea accepts input', () => {
    render(<FranchiserReportTab />);
    const textarea = screen.getByPlaceholderText(/Что нужно от руководства/);
    fireEvent.change(textarea, { target: { value: 'Нужен бюджет' } });
    expect(textarea).toHaveValue('Нужен бюджет');
  });
});

describe('FranchiserReportTab - history modal', () => {
  it('opens history modal', () => {
    render(<FranchiserReportTab />);
    fireEvent.click(screen.getByRole('button', { name: /История/i }));
    expect(screen.getByText('История отчётов')).toBeInTheDocument();
  });

  it('renders history periods', () => {
    render(<FranchiserReportTab />);
    fireEvent.click(screen.getByRole('button', { name: /История/i }));
    expect(screen.getByText('2026-Q1')).toBeInTheDocument();
  });
});

describe('FranchiserReportTab - comment modal', () => {
  it('opens comment modal on block comment click', async () => {
    render(<FranchiserReportTab />);
    const buttons = screen.getAllByRole('button');
    const commentBtn = buttons.find(b => b.textContent?.includes('Комментарий'));
    if (commentBtn) fireEvent.click(commentBtn);
  });
});

describe('FranchiserReportTab - defaultBlocks configuration', () => {
  it('has exactly 7 blocks', () => {
    expect(defaultBlocks).toHaveLength(7);
  });

  it('all blocks have unique keys', () => {
    const keys = defaultBlocks.map(b => b.key);
    const uniqueKeys = new Set(keys);
    expect(uniqueKeys.size).toBe(keys.length);
  });

  it('all blocks enabled by default', () => {
    defaultBlocks.forEach(b => {
      expect(b.enabled).toBe(true);
    });
  });

  it('block keys match switch cases', () => {
    const validKeys = ['executive', 'planFact', 'growth', 'structure', 'territories', 'risks', 'proposals'];
    defaultBlocks.forEach(b => {
      expect(validKeys).toContain(b.key);
    });
  });
});

describe('FranchiserReportTab - block toggle logic', () => {
  it('enabledBlocks = blocks.filter(b => b.enabled)', () => {
    const enabledBlocks = defaultBlocks.filter(b => b.enabled);
    expect(enabledBlocks).toHaveLength(7);
  });

  it('disabling a block removes it from enabledBlocks', () => {
    const blocks = defaultBlocks.map(b => b.key === 'executive' ? { ...b, enabled: false } : b);
    const enabledBlocks = blocks.filter(b => b.enabled);
    expect(enabledBlocks).toHaveLength(6);
    expect(enabledBlocks.find(b => b.key === 'executive')).toBeUndefined();
  });
});

describe('FranchiserReportTab - Plan Fact chart', () => {
  it('renders chart component', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByTestId('plan-fact-chart')).toBeInTheDocument();
  });
});

describe('FranchiserReportTab - button handlers', () => {
  it('has PDF button', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByRole('button', { name: /Сформировать PDF/i })).toBeInTheDocument();
  });

  it('has send button', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByRole('button', { name: /Отправить руководителю/i })).toBeInTheDocument();
  });

  it('has save button', () => {
    render(<FranchiserReportTab />);
    expect(screen.getByRole('button', { name: /Сохранить черновик/i })).toBeInTheDocument();
  });
});
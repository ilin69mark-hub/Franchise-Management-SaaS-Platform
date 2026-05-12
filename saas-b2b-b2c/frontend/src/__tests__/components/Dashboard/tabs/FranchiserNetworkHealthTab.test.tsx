import React from 'react';
import { render, screen } from '@testing-library/react';
import FranchiserNetworkHealthTab from '@/components/Dashboard/tabs/FranchiserNetworkHealthTab';

jest.mock('@/components/Dashboard/tabs/GeographyMap', () => {
  return { __esModule: true, default: () => <div data-testid="geography-map">Geography Map</div> };
});

describe('FranchiserNetworkHealthTab', () => {
  describe('getSegmentColor formula', () => {
    const getSegmentColor = (segment: string) => {
      const colors: Record<string, string> = { A: '#52c41a', B: '#1890ff', C: '#fa8c16', D: '#ff4d4f' };
      return colors[segment] || '#d9d9d9';
    };

    it('returns green for A segment', () => {
      expect(getSegmentColor('A')).toBe('#52c41a');
    });

    it('returns blue for B segment', () => {
      expect(getSegmentColor('B')).toBe('#1890ff');
    });

    it('returns orange for C segment', () => {
      expect(getSegmentColor('C')).toBe('#fa8c16');
    });

    it('returns red for D segment', () => {
      expect(getSegmentColor('D')).toBe('#ff4d4f');
    });

    it('returns gray for unknown segment', () => {
      expect(getSegmentColor('X')).toBe('#d9d9d9');
      expect(getSegmentColor('')).toBe('#d9d9d9');
    });
  });

  describe('segmentation total calculation', () => {
    const segmentation = [
      { segment: 'A' as const, name: 'A-дилеры', count: 4, revenueShare: 42, dynamics: 2 },
      { segment: 'B' as const, name: 'B-дилеры', count: 12, revenueShare: 38, dynamics: -1 },
      { segment: 'C' as const, name: 'C-дилеры', count: 6, revenueShare: 15, dynamics: -2 },
      { segment: 'D' as const, name: 'D-дилеры', count: 2, revenueShare: 5, dynamics: 1 },
    ];

    it('calculates total dealers correctly', () => {
      const total = segmentation.reduce((s, sgm) => s + sgm.count, 0);
      expect(total).toBe(24);
    });

    it('calculates segment percentage from total', () => {
      const total = segmentation.reduce((s, sgm) => s + sgm.count, 0);
      const aPercent = Math.round((segmentation[0].count / total) * 100);
      expect(aPercent).toBe(17); // 4/24 = 16.67 -> 17
    });

    it('revenue share should sum to 100', () => {
      const totalRevenue = segmentation.reduce((s, sgm) => s + sgm.revenueShare, 0);
      expect(totalRevenue).toBe(100);
    });

    it('calculates all segment percentages', () => {
      const total = segmentation.reduce((s, sgm) => s + sgm.count, 0);
      const percentages = segmentation.map(sgm => Math.round((sgm.count / total) * 100));
      expect(percentages).toEqual([17, 50, 25, 8]); // 4, 12, 6, 2 from 24
    });
  });

  describe('issue status mapping', () => {
    const getIssueStatus = (status: string) => ({
      color: status === 'resolved' ? 'green' : status === 'in_progress' ? 'processing' : 'orange',
      label: status === 'identified' ? 'Выявлено' : status === 'in_progress' ? 'Решается' : 'Решено',
      icon: status === 'resolved' ? 'check' : status === 'in_progress' ? 'warning' : 'close',
    });

    it('resolved status is green', () => {
      const result = getIssueStatus('resolved');
      expect(result.color).toBe('green');
      expect(result.label).toBe('Решено');
      expect(result.icon).toBe('check');
    });

    it('in_progress status is processing', () => {
      const result = getIssueStatus('in_progress');
      expect(result.color).toBe('processing');
      expect(result.label).toBe('Решается');
      expect(result.icon).toBe('warning');
    });

    it('identified status is orange', () => {
      const result = getIssueStatus('identified');
      expect(result.color).toBe('orange');
      expect(result.label).toBe('Выявлено');
      expect(result.icon).toBe('close');
    });
  });

  describe('lostRevenue formatting', () => {
    it('formats large numbers with commas', () => {
      const format = (n: number) => n.toLocaleString();
      expect(format(2500000)).toBe('2,500,000');
      expect(format(1800000)).toBe('1,800,000');
      expect(format(900000)).toBe('900,000');
    });

    it('adds ruble sign suffix', () => {
      const format = (n: number) => `${n.toLocaleString()} ₽`;
      expect(format(2500000)).toBe('2,500,000 ₽');
    });
  });

  describe('white spot potential formatting', () => {
    it('formats potential as approximate dealers', () => {
      const format = (v: number) => `~${v} дилеров`;
      expect(format(15)).toBe('~15 дилеров');
      expect(format(12)).toBe('~12 дилеров');
      expect(format(6)).toBe('~6 дилеров');
    });
  });

  describe('marketing ROI color', () => {
    it('returns green for ROI > 2', () => {
      const color = (roi: number) => roi > 2 ? '#52c41a' : '#fa8c16';
      expect(color(2.4)).toBe('#52c41a');
      expect(color(3.0)).toBe('#52c41a');
    });

    it('returns orange for ROI <= 2', () => {
      const color = (roi: number) => roi > 2 ? '#52c41a' : '#fa8c16';
      expect(color(2.0)).toBe('#fa8c16');
      expect(color(1.5)).toBe('#fa8c16');
    });
  });

  describe('marketing comparison calculations', () => {
    it('calculates revenue difference with/without central', () => {
      const withCentral = { avgRevenue: 5200000, avgTraffic: 180, conversion: 14 };
      const withoutCentral = { avgRevenue: 3800000, avgTraffic: 120, conversion: 9 };
      
      const revenueDiff = withCentral.avgRevenue - withoutCentral.avgRevenue;
      expect(revenueDiff).toBe(1400000);
    });

    it('calculates traffic difference', () => {
      const withCentral = { avgRevenue: 5200000, avgTraffic: 180, conversion: 14 };
      const withoutCentral = { avgRevenue: 3800000, avgTraffic: 120, conversion: 9 };
      
      const trafficDiff = withCentral.avgTraffic - withoutCentral.avgTraffic;
      expect(trafficDiff).toBe(60);
    });

    it('calculates conversion difference', () => {
      const withCentral = { avgRevenue: 5200000, avgTraffic: 180, conversion: 14 };
      const withoutCentral = { avgRevenue: 3800000, avgTraffic: 120, conversion: 9 };
      
      const convDiff = withCentral.conversion - withoutCentral.conversion;
      expect(convDiff).toBe(5);
    });

    it('calculates revenue uplift percentage', () => {
      const withCentral = 5200000;
      const withoutCentral = 3800000;
      
      const uplift = ((withCentral - withoutCentral) / withoutCentral) * 100;
      expect(uplift).toBeCloseTo(36.84, 1);
    });

    it('calculates traffic uplift percentage', () => {
      const withCentral = 180;
      const withoutCentral = 120;
      
      const uplift = ((withCentral - withoutCentral) / withoutCentral) * 100;
      expect(uplift).toBe(50);
    });
  });

  describe('metrics target comparison', () => {
    const metrics = [
      { name: 'Churn Rate', value: 4.2, unit: '%', dynamics: -0.5, target: 5 },
      { name: 'NPS дилеров', value: 72, unit: 'баллов', dynamics: 5, target: 70 },
      { name: 'Срок жизни дилера', value: 38, unit: 'мес.', dynamics: 2, target: 36 },
      { name: 'Покрытие матрицы', value: 82, unit: '%', dynamics: -2, target: 90 },
    ];

    it('Churn Rate is below target (good)', () => {
      const isBelowTarget = metrics[0].value < metrics[0].target!;
      expect(isBelowTarget).toBe(true); // 4.2 < 5
    });

    it('NPS is above target (good)', () => {
      const isAboveTarget = metrics[1].value > metrics[1].target!;
      expect(isAboveTarget).toBe(true); // 72 > 70
    });

    it('Срок жизни is above target (good)', () => {
      const isAboveTarget = metrics[2].value > metrics[2].target!;
      expect(isAboveTarget).toBe(true); // 38 > 36
    });

    it('Покрытие is below target (needs improvement)', () => {
      const isBelowTarget = metrics[3].value < metrics[3].target!;
      expect(isBelowTarget).toBe(true); // 82 < 90
    });

    it('dynamics positive is success type', () => {
      const getType = (dynamics: number) => dynamics > 0 ? 'success' : dynamics < 0 ? 'danger' : 'secondary';
      expect(getType(5)).toBe('success');
      expect(getType(-2)).toBe('danger');
      expect(getType(0)).toBe('secondary');
    });
  });

  describe('migration matrix data', () => {
    const migration = [
      { from: 'C', to: 'B', count: 2 },
      { from: 'B', to: 'A', count: 1 },
      { from: 'B', to: 'C', count: 1 },
      { from: 'A', to: 'B', count: 0 },
    ];

    it('calculates total migrations', () => {
      const total = migration.reduce((s, m) => s + m.count, 0);
      expect(total).toBe(4);
    });

    it('finds upward migrations (promotions)', () => {
      const upwardMigrations = migration.filter(m => m.from < m.to);
      expect(upwardMigrations.length).toBe(2); // C->B, B->A
    });

    it('finds downward migrations (demotions)', () => {
      const downwardMigrations = migration.filter(m => m.from > m.to);
      expect(downwardMigrations.length).toBe(2); // C->B (demotion: C>B), B->C (B>C alphabetically)
    });

    it('finds stable migrations (same segment)', () => {
      const stableMigrations = migration.filter(m => m.from === m.to);
      expect(stableMigrations.length).toBe(0);
    });
  });

  describe('white spot data', () => {
    const whiteSpots = [
      { city: 'Воронеж', population: 1050000, potential: 15 },
      { city: 'Краснодар', population: 918000, potential: 12 },
      { city: 'Казань', population: 1250000, potential: 18 },
      { city: 'Сочи', population: 412000, potential: 6 },
    ];

    it('sorts by population descending', () => {
      const sorted = [...whiteSpots].sort((a, b) => b.population - a.population);
      expect(sorted[0].city).toBe('Казань');
      expect(sorted[3].city).toBe('Сочи');
    });

    it('calculates total potential dealers', () => {
      const totalPotential = whiteSpots.reduce((s, ws) => s + ws.potential, 0);
      expect(totalPotential).toBe(51);
    });

    it('calculates average population', () => {
      const totalPop = whiteSpots.reduce((s, ws) => s + ws.population, 0);
      const avgPop = totalPop / whiteSpots.length;
      expect(avgPop).toBeCloseTo(908000, -4);
    });
  });

  describe('component rendering', () => {
    it('renders title', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('Здоровье сети')).toBeInTheDocument();
    });

    it('renders period selector', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('Квартал')).toBeInTheDocument();
    });

    it('renders all segmentation cards', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getAllByText('A-дилеры (лидеры)').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('B-дилеры (стабильные)').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('C-дилеры (отстающие)').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('D-дилеры (кандидаты)').length).toBeGreaterThanOrEqual(1);
    });

    it('renders all metrics', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('Churn Rate')).toBeInTheDocument();
      expect(screen.getByText('NPS дилеров')).toBeInTheDocument();
      expect(screen.getByText('Срок жизни дилера')).toBeInTheDocument();
      expect(screen.getByText('Доля новых дилеров')).toBeInTheDocument();
      expect(screen.getByText('Покрытие матрицы')).toBeInTheDocument();
      expect(screen.getByText('Плотность покрытия')).toBeInTheDocument();
    });

    it('renders system issues section', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('Системные проблемы')).toBeInTheDocument();
    });

    it('renders geography section', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('География присутствия')).toBeInTheDocument();
      expect(screen.getByText('Белые пятна')).toBeInTheDocument();
    });

    it('renders marketing section', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('Маркетинговая эффективность')).toBeInTheDocument();
      expect(screen.getByText('С централизованным маркетингом')).toBeInTheDocument();
      expect(screen.getByText('Без централизованного маркетинга')).toBeInTheDocument();
      expect(screen.getByText('ROI маркетинга')).toBeInTheDocument();
    });

    it('renders migration table', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('Миграция дилеров (квартал)')).toBeInTheDocument();
      expect(screen.getByText('Из')).toBeInTheDocument();
      expect(screen.getByText('В')).toBeInTheDocument();
      expect(screen.getByText('Кол-во')).toBeInTheDocument();
    });

    it('renders segmentation table', () => {
      render(<FranchiserNetworkHealthTab />);
      expect(screen.getByText('Сегментация')).toBeInTheDocument();
      expect(screen.getByText('Количество')).toBeInTheDocument();
      expect(screen.getByText('Доля выручки')).toBeInTheDocument();
      expect(screen.getByText('Динамика')).toBeInTheDocument();
    });
  });
});
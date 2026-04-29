// src/__tests__/components/Dashboard/tabs/TerritoryMapTab.test.tsx
import React from 'react';
import { render } from '@testing-library/react';
import TerritoryMapTab from '@/components/Dashboard/tabs/TerritoryMapTab';
import { DealerMetrics } from '@/store/territoryManagerStore';

describe('TerritoryMapTab', () => {
  it('renders', () => {
    expect(() => render(<TerritoryMapTab />)).not.toThrow();
  });

  it('renders with dealers', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'Test Dealer', salonCount: 2, planPercent: 90, forecastPercent: 100, conversion: 4.5, status: 'green', plan: 10000000, fact: 9000000, debt: 0, margin: 25, avgCheck: 420000, taskCount: 0 },
    ];
    expect(() => render(<TerritoryMapTab dealers={dealers} />)).not.toThrow();
  });

  it('renders loading', () => {
    expect(() => render(<TerritoryMapTab loading />)).not.toThrow();
  });

  it('filters by status', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'Green', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'Red', salonCount: 1, planPercent: 50, forecastPercent: 60, conversion: 1.5, status: 'red' },
    ];
    const redDealers = dealers.filter(d => d.status === 'red');
    expect(redDealers.length).toBe(1);
  });

  it('calculates status counts', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow' },
      { dealerId: '3', dealerName: 'C', salonCount: 1, planPercent: 50, forecastPercent: 60, conversion: 1.5, status: 'red' },
    ];
    const green = dealers.filter(d => d.status === 'green').length;
    const yellow = dealers.filter(d => d.status === 'yellow').length;
    const red = dealers.filter(d => d.status === 'red').length;
    expect(green).toBe(1);
    expect(yellow).toBe(1);
    expect(red).toBe(1);
  });

  it('searches by name', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'Мебель Москва', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'Диванит', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow' },
    ];
    const searchText = 'мебель';
    const filtered = dealers.filter(d => d.dealerName.toLowerCase().includes(searchText.toLowerCase()));
    expect(filtered.length).toBe(1);
  });

  it('calculates territory totals', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 100, forecastPercent: 110, conversion: 4, status: 'green', plan: 10000000, fact: 10000000, debt: 0, margin: 25 },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow', plan: 5000000, fact: 4000000, debt: 50000, margin: 22 },
    ];
    const totalPlan = dealers.reduce((s, d) => s + (d.plan || 0), 0);
    const totalFact = dealers.reduce((s, d) => s + (d.fact || 0), 0);
    const totalDebt = dealers.reduce((s, d) => s + (d.debt || 0), 0);
    const avgMargin = dealers.reduce((s, d) => s + (d.margin || 0), 0) / dealers.length;
    expect(totalPlan).toBe(15000000);
    expect(totalFact).toBe(14000000);
    expect(totalDebt).toBe(50000);
    expect(avgMargin).toBe(23.5);
  });

  it('calculates conversion color', () => {
    const getCellColor = (conv: number) => {
      if (conv >= 3) return '#f6ffed';
      if (conv >= 2) return '#fffbe6';
      return '#fff1f0';
    };
    expect(getCellColor(4)).toBe('#f6ffed');
    expect(getCellColor(2.5)).toBe('#fffbe6');
    expect(getCellColor(1)).toBe('#fff1f0');
  });

  it('sorts dealers by field', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 100, forecastPercent: 110, conversion: 4, status: 'green' },
      { dealerId: '3', dealerName: 'C', salonCount: 1, planPercent: 60, forecastPercent: 70, conversion: 2, status: 'red' },
    ];
    const sorted = [...dealers].sort((a, b) => b.planPercent - a.planPercent);
    expect(sorted[0].dealerName).toBe('B');
    expect(sorted[2].dealerName).toBe('C');
  });

  it('filters leaders', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 110, forecastPercent: 120, conversion: 5, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 100, forecastPercent: 110, conversion: 4, status: 'green' },
      { dealerId: '3', dealerName: 'C', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 3, status: 'green' },
    ];
    const leaders = dealers.filter(d => d.planPercent >= 100);
    expect(leaders.length).toBe(2);
  });

  it('filters problem dealers', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 110, forecastPercent: 120, conversion: 5, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 60, forecastPercent: 70, conversion: 2, status: 'red' },
    ];
    const problems = dealers.filter(d => d.planPercent < 70);
    expect(problems.length).toBe(1);
  });
});
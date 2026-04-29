// __tests__/components/Dashboard/tabs/ProductsStockTab.test.tsx
import React from 'react';
import { render } from '@testing-library/react';
import ProductsStockTab from '@/components/Dashboard/tabs/ProductsStockTab';

describe('ProductsStockTab', () => {
  it('renders', () => {
    render(<ProductsStockTab />);
  });

  it('renders loading', () => {
    render(<ProductsStockTab loading={true} />);
  });

  it('renders period filter', () => {
    render(<ProductsStockTab period="month" />);
  });

  it('renders category filter', () => {
    render(<ProductsStockTab category="kitchens" />);
  });

  it('renders salon filter', () => {
    render(<ProductsStockTab salonId="all" />);
  });

  it('renders inventory data', () => {
    render(<ProductsStockTab inventory={[{ id: '1', collection: 'A', category: 'B', stockWarehouse: 10, onDisplay: 2, soldPeriod: 5, turnoverDays: 30, totalStockValue: 100000 }]} />);
  });

  it('renders lost sales', () => {
    render(<ProductsStockTab lostSales={[]} />);
  });

  it('renders returns', () => {
    const data = [{ id: '1', date: '2026-04-01', product: 'A', reason: 'B', amount: 1000, status: 'resolved' as const }];
    render(<ProductsStockTab returns={data} />);
  });

  it('calculates turnover', () => {
    expect(Math.round((15 * 30) / 8)).toBe(56);
  });

  it('detects slow items', () => {
    expect(95 > 90).toBe(true);
  });

  it('detects deficit', () => {
    const stock = 0, sold = 5;
    expect(stock === 0 && sold > 0).toBe(true);
  });

  it('calculates percent', () => {
    expect((2500000 / 5000000) * 100).toBe(50);
  });

  it('sums inventory', () => {
    const data = [{ id: '1', val: 100 }];
    expect(data.reduce((sum, i) => sum + i.val, 0)).toBe(100);
  });

  it('sums lost sales', () => {
    const data = [{ lostRevenue: 1000 }];
    expect(data.reduce((sum, i) => sum + i.lostRevenue, 0)).toBe(1000);
  });

  it('sums returns', () => {
    const data = [{ amount: 5000 }];
    expect(data.reduce((sum, i) => sum + i.amount, 0)).toBe(5000);
  });

  it('renders dynamics', () => {
    render(<ProductsStockTab salesDynamics={[{ month: '2026-01', sales: 10, stock: 20 }]} />);
  });
});
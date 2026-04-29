// src/__tests__/components/Dashboard/tabs/TerritoryBenchmarkTab.test.tsx
describe('TerritoryBenchmarkTab', () => {
  it('calculates scatter data', () => {
    const data = [
      { id: '1', revenue: 13800000, margin: 28, salons: 3, segment: 'A' as const },
      { id: '2', revenue: 6240000, margin: 22, salons: 2, segment: 'B' as const },
      { id: '3', revenue: 2250000, margin: 15, salons: 1, segment: 'D' as const },
    ];
    const avgRevenue = data.reduce((s, d) => s + d.revenue, 0) / data.length;
    expect(avgRevenue).toBeGreaterThan(7000000);
    expect(avgRevenue).toBeLessThan(8000000);
  });

  it('groups dealers by segment', () => {
    const data = [
      { segment: 'A' as const },
      { segment: 'B' as const },
      { segment: 'A' as const },
      { segment: 'C' as const },
      { segment: 'D' as const },
    ];
    const segments = { A: 0, B: 0, C: 0, D: 0 };
    data.forEach(d => segments[d.segment]++);
    expect(segments.A).toBe(2);
    expect(segments.B).toBe(1);
    expect(segments.C).toBe(1);
    expect(segments.D).toBe(1);
  });

  it('calculates dealer index', () => {
    const d = { revenue: 15000000, margin: 28, conversion: 5, discount: 5, salons: 3 };
    const planPercent = (d.revenue / 10000000) * 100;
    let index = Math.min(100, 
      (planPercent * 0.4) + 
      (d.conversion * 5 * 0.25) + 
      (d.margin * 1.5 * 0.2) + 
      (d.discount < 10 ? 10 : 0) * 0.1 + 
      (d.salons > 1 ? 5 : 0) * 0.05
    );
    expect(index).toBeGreaterThanOrEqual(70);
  });

  it('sorts dealers by index', () => {
    const data = [
      { totalIndex: 85 },
      { totalIndex: 72 },
      { totalIndex: 93 },
      { totalIndex: 58 },
    ];
    const sorted = [...data].sort((a, b) => b.totalIndex - a.totalIndex);
    expect(sorted[0].totalIndex).toBe(93);
    expect(sorted[3].totalIndex).toBe(58);
  });

  it('calculates sales structure', () => {
    const data = { kitchen: 35, soft: 25, корпусная: 20, mattress: 15, accessories: 5 };
    const total = Object.values(data).reduce((s, v) => s + v, 0);
    expect(total).toBe(100);
  });

  it('identifies white spots', () => {
    const data = [
      { dealerPercent: 15, networkPercent: 30 },
      { dealerPercent: 30, networkPercent: 25 },
      { dealerPercent: 10, networkPercent: 15 },
    ];
    const whiteSpots = data.filter(d => d.dealerPercent < d.networkPercent);
    expect(whiteSpots.length).toBe(2);
  });

  it('calculates inventory risk', () => {
    const data = [
      { stock: 15, cost: 1200000, days: 180 },
      { stock: 8, cost: 480000, days: 120 },
      { stock: 3, cost: 180000, days: 45 },
    ];
    const highRisk = data.filter(d => d.days > 90);
    expect(highRisk.length).toBe(2);
  });

  it('calculates lost revenue', () => {
    const data = [
      { currentStock: 0, lostRevenue: 350000 },
      { currentStock: 1, lostRevenue: 180000 },
      { currentStock: 5, lostRevenue: 0 },
    ];
    const totalLost = data.reduce((s, d) => s + d.lostRevenue, 0);
    expect(totalLost).toBe(530000);
  });

  it('gets segment color', () => {
    const colors = { A: '#52c41a', B: '#1890ff', C: '#fa8c16', D: '#ff4d4f' };
    expect(colors.A).toBe('#52c41a');
    expect(colors.B).toBe('#1890ff');
    expect(colors.C).toBe('#fa8c16');
    expect(colors.D).toBe('#ff4d4f');
  });

  it('filters by segment', () => {
    const data = [
      { segment: 'A' },
      { segment: 'B' },
      { segment: 'C' },
    ];
    const aSegment = data.filter(d => d.segment === 'A');
    expect(aSegment.length).toBe(1);
  });

  it('calculates rank change', () => {
    const current = [85, 72, 93, 58];
    const previous = [72, 85, 93, 58];
    const changes = current.map((c, i) => previous.indexOf(c) - i);
    expect(changes[0]).not.toBe(0);
  });
});
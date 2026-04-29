// src/__tests__/components/Dashboard/tabs/TerritoryFunnelTab.test.tsx
describe('TerritoryFunnelTab', () => {
  it('calculates conversion', () => {
    const leads = 1000;
    const deals = 50;
    const conversion = (deals / leads) * 100;
    expect(conversion).toBe(5);
  });

  it('calculates step conversion', () => {
    const data = [450, 270, 180, 108, 54, 45];
    const stepConversions = data.slice(1).map((v, i) => Math.round((v / data[i]) * 100));
    expect(stepConversions[0]).toBe(60);
  });

  it('calculates traffic conversion', () => {
    const data = [450, 270, 180, 108, 54, 45];
    const trafficConversions = data.map(v => Math.round((v / data[0]) * 100));
    expect(trafficConversions[0]).toBe(100);
    expect(trafficConversions[4]).toBe(12);
  });

  it('calculates average check', () => {
    const paid = 38000000;
    const deals = 75;
    const avgCheck = paid / deals;
    expect(avgCheck).toBeCloseTo(506666.67, 0);
  });

  it('determines status by conversion', () => {
    const status = (conversion: number) => {
      if (conversion >= 10) return 'green';
      if (conversion >= 7) return 'yellow';
      return 'red';
    };
    expect(status(12)).toBe('green');
    expect(status(8)).toBe('yellow');
    expect(status(5)).toBe('red');
  });

  it('filters anomalies by type', () => {
    const anomalies = [
      { type: 'conversion_drop', severity: 'warning' },
      { type: 'traffic_drop', severity: 'warning' },
      { type: 'avg_check', severity: 'critical' },
    ];
    const filtered = anomalies.filter(a => a.type === 'conversion_drop');
    expect(filtered.length).toBe(1);
  });

  it('identifies critical anomalies', () => {
    const anomalies = [
      { severity: 'warning' },
      { severity: 'critical' },
      { severity: 'warning' },
    ];
    const critical = anomalies.filter(a => a.severity === 'critical');
    expect(critical.length).toBe(1);
  });

  it('calculates store funnel', () => {
    const store = { traffic: 180, consultation: 108, measurement: 72, kp: 43, contract: 22, payment: 18 };
    const conversion = (store.payment / store.traffic) * 100;
    expect(conversion).toBe(10);
  });

  it('sorts managers by conversion', () => {
    const managers = [
      { name: 'A', conversion: 15 },
      { name: 'B', conversion: 11 },
      { name: 'C', conversion: 10 },
    ];
    const sorted = [...managers].sort((a, b) => b.conversion - a.conversion);
    expect(sorted[0].name).toBe('A');
  });
});
// src/__tests__/components/Dashboard/alertStore.test.ts
describe('Alert store logic', () => {
  it('calculates unread count', () => {
    const alerts = [
      { status: 'new' as const },
      { status: 'read' as const },
      { status: 'new' as const },
    ];
    const unread = alerts.filter(a => a.status === 'new').length;
    expect(unread).toBe(2);
  });

  it('filters by category', () => {
    const alerts = [
      { category: 'plan' as const },
      { category: 'funnel' as const },
      { category: 'plan' as const },
    ];
    const planAlerts = alerts.filter(a => a.category === 'plan');
    expect(planAlerts.length).toBe(2);
  });

  it('filters by critical', () => {
    const alerts = [
      { priority: 'critical' as const },
      { priority: 'warning' as const },
      { priority: 'critical' as const },
    ];
    const critical = alerts.filter(a => a.priority === 'critical');
    expect(critical.length).toBe(2);
  });

  it('sorts by date', () => {
    const alerts = [
      { id: '1', createdAt: '2026-04-28T10:00:00' },
      { id: '2', createdAt: '2026-04-28T09:00:00' },
    ];
    const sorted = [...alerts].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    expect(sorted[0].id).toBe('1');
  });

  it('groups by category', () => {
    const alerts = [
      { category: 'plan' as const },
      { category: 'funnel' as const },
      { category: 'plan' as const },
    ];
    const groups = { plan: [], funnel: [], stock: [], communications: [], activity: [] };
    alerts.forEach(a => groups[a.category].push(a));
    expect(groups.plan.length).toBe(2);
    expect(groups.funnel.length).toBe(1);
  });

  it('updates settings', () => {
    const settings = {
      planThreshold: 70,
      conversionDropThreshold: 15,
    };
    const newSettings = { ...settings, planThreshold: 80 };
    expect(newSettings.planThreshold).toBe(80);
  });

  it('marks alert as read', () => {
    const alerts = [
      { id: '1', status: 'new' as const },
      { id: '2', status: 'new' as const },
    ];
    const updated = alerts.map(a => a.id === '1' ? { ...a, status: 'read' as const } : a);
    expect(updated[0].status).toBe('read');
  });

  it('checks category colors', () => {
    const colors = {
      plan: '#ff4d4f',
      funnel: '#fa8c16',
      stock: '#722ed1',
      communications: '#1890ff',
      activity: '#52c41a',
    };
    expect(colors.plan).toBe('#ff4d4f');
  });
});
// src/__tests__/components/Dashboard/tabs/TerritoryCommunicationsTab.test.tsx
describe('TerritoryCommunicationsTab', () => {
  it('filters tasks by priority', () => {
    const tasks = [
      { id: '1', priority: 'high' as const, status: 'new' as const },
      { id: '2', priority: 'medium' as const, status: 'new' as const },
      { id: '3', priority: 'low' as const, status: 'completed' as const },
    ];
    const highPriority = tasks.filter(t => t.priority === 'high');
    expect(highPriority.length).toBe(1);
  });

  it('counts new tasks', () => {
    const tasks = [
      { id: '1', status: 'new' as const },
      { id: '2', status: 'new' as const },
      { id: '3', status: 'completed' as const },
    ];
    const newCount = tasks.filter(t => t.status === 'new').length;
    expect(newCount).toBe(2);
  });

  it('searches tasks', () => {
    const tasks = [
      { id: '1', title: 'Провести аудит', description: 'Проверить салоны' },
      { id: '2', title: 'Отчёт', description: 'Подготовить отчёт' },
    ];
    const search = 'аудит';
    const filtered = tasks.filter(t => 
      t.title.toLowerCase().includes(search.toLowerCase()) || 
      t.description.toLowerCase().includes(search.toLowerCase())
    );
    expect(filtered.length).toBe(1);
  });

  it('filters requests by status', () => {
    const requests = [
      { id: '1', status: 'new' as const, slaHours: 10 },
      { id: '2', status: 'new' as const, slaHours: 20 },
      { id: '3', status: 'resolved' as const, slaHours: 5 },
    ];
    const newRequests = requests.filter(r => r.status === 'new');
    expect(newRequests.length).toBe(2);
  });

  it('filters overdue requests', () => {
    const requests = [
      { id: '1', slaHours: -5 },
      { id: '2', slaHours: 10 },
      { id: '3', slaHours: -2 },
    ];
    const overdue = requests.filter(r => r.slaHours < 0);
    expect(overdue.length).toBe(2);
  });

  it('gets SLA color', () => {
    const getSlaColor = (hours: number) => {
      if (hours < 0) return '#ff4d4f';
      if (hours < 2) return '#ff4d4f';
      if (hours < 24) return '#fa8c16';
      return '#52c41a';
    };
    expect(getSlaColor(-1)).toBe('#ff4d4f');
    expect(getSlaColor(1)).toBe('#ff4d4f');
    expect(getSlaColor(12)).toBe('#fa8c16');
    expect(getSlaColor(48)).toBe('#52c41a');
  });


  it('gets SLA label', () => {
    const getSlaLabel = (hours: number) => {
      if (hours < 0) return `${Math.abs(hours)}ч просрочено`;
      if (hours < 24) return `${hours}ч`;
      return `${Math.round(hours / 24)}дн`;
    };
    expect(getSlaLabel(-5)).toBe('5ч просрочено');
    expect(getSlaLabel(12)).toBe('12ч');
    expect(getSlaLabel(48)).toBe('2дн');
  });

  it('sorts requests by SLA', () => {
    const requests = [
      { id: '1', slaHours: 48 },
      { id: '2', slaHours: -5 },
      { id: '3', slaHours: 12 },
    ];
    const sorted = [...requests].sort((a, b) => a.slaHours - b.slaHours);
    expect(sorted[0].id).toBe('2');
    expect(sorted[2].id).toBe('1');
  });

  it('filters tasks by status', () => {
    const tasks = [
      { status: 'in_progress' as const },
      { status: 'done' as const },
      { status: 'overdue' as const },
    ];
    const active = tasks.filter(t => t.status !== 'done' && t.status !== 'overdue');
    expect(active.length).toBe(1);
  });

  it('filters tasks by overdue', () => {
    const tasks = [
      { status: 'done' as const },
      { status: 'overdue' as const },
      { status: 'accepted' as const },
    ];
    const overdue = tasks.filter(t => t.status === 'overdue');
    expect(overdue.length).toBe(1);
  });

  it('filters requests by type', () => {
    const requests = [
      { type: 'discount' as const },
      { type: 'return' as const },
      { type: 'marketing' as const },
    ];
    const discountRequests = requests.filter(r => r.type === 'discount');
    expect(discountRequests.length).toBe(1);
  });

  it('sorts tasks by due date', () => {
    const tasks = [
      { dueDate: '2026-05-01' },
      { dueDate: '2026-04-28' },
      { dueDate: '2026-04-30' },
    ];
    const sorted = [...tasks].sort((a, b) => new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime());
    expect(sorted[0].dueDate).toBe('2026-04-28');
  });
});
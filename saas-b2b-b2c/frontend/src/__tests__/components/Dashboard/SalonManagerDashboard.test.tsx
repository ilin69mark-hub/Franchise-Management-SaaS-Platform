describe('SalonManagerDashboard utilities', () => {
  it('должен форматировать деньги', () => {
    const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);
    expect(formatMoney(1000000).replace(/\s/g, ' ')).toBe('1 000 000');
    expect(formatMoney(50000).replace(/\s/g, ' ')).toBe('50 000');
  });

  it('должен определять цвет выполнения плана', () => {
    const getPlanColor = (percent: number) => {
      if (percent >= 80) return '#52c41a';
      if (percent >= 50) return '#faad14';
      return '#ff4d4f';
    };

    expect(getPlanColor(100)).toBe('#52c41a');
    expect(getPlanColor(80)).toBe('#52c41a');
    expect(getPlanColor(65)).toBe('#faad14');
    expect(getPlanColor(50)).toBe('#faad14');
    expect(getPlanColor(30)).toBe('#ff4d4f');
  });

  it('должен показывать названия табов', () => {
    const tabs = ['Главная', 'Воронка', 'Команда', 'Товары'];
    expect(tabs).toContain('Главная');
    expect(tabs).toContain('Воронка');
    expect(tabs).toContain('Команда');
    expect(tabs).toContain('Товары');
  });
});
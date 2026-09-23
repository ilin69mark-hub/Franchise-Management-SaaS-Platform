import React from 'react';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import TerritoryPlanFactTab from '@/components/Dashboard/tabs/TerritoryPlanFactTab';

jest.mock('@/api/axiosClient', () => ({
  __esModule: true,
  default: {
    get: jest.fn(() => Promise.reject(new Error('network'))),
    post: jest.fn(() => Promise.reject(new Error('network'))),
  },
}));

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: {
      ...actualAntd.message,
      success: jest.fn(),
      error: jest.fn(),
    },
  };
});

const pickOption = (text: string) => {
  const dropdowns = Array.from(document.querySelectorAll('.ant-select-dropdown:not(.ant-select-dropdown-hidden)'));
  const last = dropdowns[dropdowns.length - 1] as HTMLElement;
  fireEvent.click(within(last).getByText(text));
};

describe('TerritoryPlanFactTab', () => {
  it('calculates totals', () => {
    const data = [
      { plan: 10000000, fact: 8000000, forecast: 9000000 },
      { plan: 5000000, fact: 4000000, forecast: 4500000 },
    ];
    const totalPlan = data.reduce((s, d) => s + d.plan, 0);
    const totalFact = data.reduce((s, d) => s + d.fact, 0);
    const totalForecast = data.reduce((s, d) => s + d.forecast, 0);
    expect(totalPlan).toBe(15000000);
    expect(totalFact).toBe(12000000);
    expect(totalForecast).toBe(13500000);
  });

  it('calculates percent', () => {
    const totalFact = 12000000;
    const totalPlan = 15000000;
    const percent = (totalFact / totalPlan) * 100;
    expect(percent).toBe(80);
  });

  it('calculates optimistic scenario', () => {
    const data = [
      { plan: 10000000, fact: 12000000, forecast: 13000000 },
      { plan: 5000000, fact: 3000000, forecast: 3500000 },
    ];
    const totalPlan = data.reduce((s, d) => s + d.plan, 0);
    const leaders = data.filter(d => d.fact / d.plan >= 1.1);
    const problems = data.filter(d => d.fact / d.plan < 0.7);
    const optimistic = leaders.reduce((s, d) => s + d.forecast, 0) + problems.reduce((s, d) => s + d.fact * 1.1, 0);
    const percent = (optimistic / totalPlan) * 100;
    expect(optimistic).toBe(13000000 + 3300000);
    expect(percent).toBeCloseTo(108.67, 1);
  });

  it('calculates scenario amounts', () => {
    const data = [
      { plan: 10000000, fact: 12000000, forecast: 13000000 },
      { plan: 5000000, fact: 3000000, forecast: 3500000 },
    ];
    const totalPlan = data.reduce((s, d) => s + d.plan, 0);
    const totalFact = data.reduce((s, d) => s + d.fact, 0);
    const totalForecast = data.reduce((s, d) => s + d.forecast, 0);
    expect(totalPlan).toBe(15000000);
    expect(totalFact).toBe(15000000);
    expect(totalForecast).toBe(16500000);
  });

  it('identifies underperformers', () => {
    const data = [
      { dealerName: 'A', percent: 90 },
      { dealerName: 'B', percent: 60 },
      { dealerName: 'C', percent: 45 },
    ];
    const under = data.filter(d => d.percent < 70);
    expect(under.length).toBe(2);
  });

  it('identifies overperformers', () => {
    const data = [
      { dealerName: 'A', percent: 110 },
      { dealerName: 'B', percent: 90 },
      { dealerName: 'C', percent: 105 },
    ];
    const over = data.filter(d => d.percent >= 100);
    expect(over.length).toBe(2);
  });

  it('calculates deviation text', () => {
    const data = [
      { dealerName: 'A', plan: 10000000, fact: 8000000 },
      { dealerName: 'B', plan: 5000000, fact: 6000000 },
    ];
    const deviations = data.filter(d => d.fact < d.plan);
    const totalGap = deviations.reduce((s, d) => s + d.plan - d.fact, 0);
    expect(totalGap).toBe(2000000);
  });

  it('calculates top 3 contribution', () => {
    const data = [
      { dealerName: 'A', fact: 12000000 },
      { dealerName: 'B', fact: 8000000 },
      { dealerName: 'C', fact: 5000000 },
      { dealerName: 'D', fact: 3000000 },
      { dealerName: 'E', fact: 1000000 },
    ];
    const sorted = [...data].sort((a, b) => b.fact - a.fact);
    const top3 = sorted.slice(0, 3).reduce((s, d) => s + d.fact, 0);
    const total = data.reduce((s, d) => s + d.fact, 0);
    const percent = (top3 / total) * 100;
    expect(top3).toBe(25000000);
    expect(percent).toBeCloseTo(86.2, 1);
  });

  it('calculates deviation from plan', () => {
    const plan = 10000000;
    const fact = 8000000;
    const deviation = fact - plan;
    const deviationPercent = ((fact - plan) / plan) * 100;
    expect(deviation).toBe(-2000000);
    expect(deviationPercent).toBe(-20);
  });

  it('sorts dealers by plan contribution', () => {
    const data = [
      { name: 'A', plan: 5000000 },
      { name: 'B', plan: 15000000 },
      { name: 'C', plan: 10000000 },
    ];
    const sorted = [...data].sort((a, b) => b.plan - a.plan);
    expect(sorted[0].name).toBe('B');
    expect(sorted[2].name).toBe('A');
  });
});

describe('TerritoryPlanFactTab render', () => {
  it('рендерит сценарии, сводку и таблицу отклонений', () => {
    render(<TerritoryPlanFactTab />);
    expect(screen.getByText('Сформировать PDF-отчёт')).toBeInTheDocument();
    expect(screen.getByText('Квартал')).toBeInTheDocument();
    expect(screen.getByText('Оптимистичный')).toBeInTheDocument();
    expect(screen.getByText('Реалистичный')).toBeInTheDocument();
    expect(screen.getByText('Пессимистичный')).toBeInTheDocument();
    expect(screen.getByText(/Отставание от плана 8\.0 млн руб\. сформировано из-за/)).toBeInTheDocument();
    expect(screen.getByText('Вклад дилеров в план')).toBeInTheDocument();
    expect(screen.getByText('План-факт динамика')).toBeInTheDocument();
    expect(screen.getByText('Детализация отклонений')).toBeInTheDocument();
    expect(screen.getByText('Мебель Москва')).toBeInTheDocument();
    expect(screen.getByText('Салон мебели Казань')).toBeInTheDocument();
    expect(screen.getAllByText('Выберите').length).toBe(5);
    expect(screen.getAllByText(/\d+\.\d млн ₽/).length).toBeGreaterThan(0);
  });

  it('рендерит тэг топ-3 дилеров', () => {
    render(<TerritoryPlanFactTab />);
    expect(screen.getByText(/Топ-3 дилера дают \d+% результата/)).toBeInTheDocument();
  });

  it('переключает период на месяц', () => {
    render(<TerritoryPlanFactTab />);
    fireEvent.mouseDown(screen.getByText('Квартал'));
    pickOption('Месяц');
    fireEvent.click(screen.getByText('Сформировать PDF-отчёт'));
    expect(screen.getByText('Период: Месяц')).toBeInTheDocument();
    expect(screen.getByText('План: 47.0 млн ₽')).toBeInTheDocument();
    expect(screen.getByText('Факт: 40.2 млн ₽')).toBeInTheDocument();
  });

  it('открывает предпросмотр и показывает ошибку генерации PDF', async () => {
    render(<TerritoryPlanFactTab />);
    fireEvent.click(screen.getByText('Сформировать PDF-отчёт'));
    expect(screen.getByText('Предпросмотр отчёта')).toBeInTheDocument();
    expect(screen.getByText('Период: Квартал')).toBeInTheDocument();
    const { error } = jest.requireMock('antd').message as { error: jest.Mock };
    fireEvent.click(screen.getByText('Скачать PDF'));
    await waitFor(() => {
      expect(error).toHaveBeenCalledWith('Ошибка генерации отчёта');
    });
  });

  it('закрывает и снова открывает модалку предпросмотра', () => {
    render(<TerritoryPlanFactTab />);
    fireEvent.click(screen.getByText('Сформировать PDF-отчёт'));
    fireEvent.click(screen.getByText('Закрыть'));
    fireEvent.click(screen.getByText('Сформировать PDF-отчёт'));
    expect(screen.getByText('Предпросмотр отчёта')).toBeInTheDocument();
    expect(screen.getByText('Период: Квартал')).toBeInTheDocument();
  });
});
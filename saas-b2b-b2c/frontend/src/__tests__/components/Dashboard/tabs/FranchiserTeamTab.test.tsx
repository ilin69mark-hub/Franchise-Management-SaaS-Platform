import React from 'react';
import { render, screen } from '@testing-library/react';
import FranchiserTeamTab from '@/components/Dashboard/tabs/FranchiserTeamTab';

jest.mock('@/components/Dashboard/tabs/ManagerKpiChart', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="kpi-chart">KPI Chart</div>,
  };
});

jest.mock('@/components/Dashboard/tabs/ManagerPlanFactChart', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="plan-fact-chart">Plan Fact Chart</div>,
  };
});

describe('FranchiserTeamTab', () => {
  it('renders title', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText('Моя команда')).toBeInTheDocument();
  });

  it('renders managers table', () => {
    render(<FranchiserTeamTab />);
    const content = document.body.textContent || '';
    expect(content).toContain('Алексей Петров');
    expect(content).toContain('Мария Иванова');
    expect(content).toContain('Сергей Сидоров');
    expect(content).toContain('Елена Смирнова');
  });

  it('renders table columns', () => {
    render(<FranchiserTeamTab />);
    const content = document.body.textContent || '';
    expect(content).toContain('Территория');
    expect(content).toContain('% плана');
    expect(content).toContain('В красной зоне');
    expect(content).toContain('SLA');
    expect(content).toContain('Прирост');
    expect(content).toContain('Churn');
    expect(content).toContain('Прогноз');
    expect(content).toContain('KPI');
    expect(content).toContain('Бонус');
  });

  it('renders assign plans button', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText('Назначить планы')).toBeInTheDocument();
  });

  it('renders quarterly selector', () => {
    render(<FranchiserTeamTab />);
    expect(screen.getByText(/Q/)).toBeInTheDocument();
  });
});
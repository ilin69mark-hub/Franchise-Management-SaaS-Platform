import React from 'react';
import { render, screen } from '@testing-library/react';
import FranchiserNetworkHealthTab from '@/components/Dashboard/tabs/FranchiserNetworkHealthTab';

jest.mock('@/components/Dashboard/tabs/GeographyMap', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="geography-map">Geography Map</div>,
  };
});

describe('FranchiserNetworkHealthTab', () => {
  it('renders title', () => {
    render(<FranchiserNetworkHealthTab />);
    expect(screen.getByText('Здоровье сети')).toBeInTheDocument();
  });

  it('renders segmentation cards', () => {
    render(<FranchiserNetworkHealthTab />);
    const content = document.body.textContent || '';
    expect(content).toContain('A-дилеры');
    expect(content).toContain('B-дилеры');
    expect(content).toContain('C-дилеры');
    expect(content).toContain('D-дилеры');
  });

  it('renders metrics widgets', () => {
    render(<FranchiserNetworkHealthTab />);
    const content = document.body.textContent || '';
    expect(content).toContain('Churn Rate');
    expect(content).toContain('NPS дилеров');
    expect(content).toContain('Срок жизни дилера');
    expect(content).toContain('Доля новых дилеров');
    expect(content).toContain('Покрытие матрицы');
    expect(content).toContain('Плотность покрытия');
  });

  it('renders system issues', () => {
    render(<FranchiserNetworkHealthTab />);
    expect(screen.getByText('Системные проблемы')).toBeInTheDocument();
  });

  it('renders geography section', () => {
    render(<FranchiserNetworkHealthTab />);
    expect(screen.getByText('География присутствия')).toBeInTheDocument();
    expect(screen.getByText('Белые пятна')).toBeInTheDocument();
  });

  it('renders marketing ROI', () => {
    render(<FranchiserNetworkHealthTab />);
    expect(screen.getByText('Маркетинговая эффективность')).toBeInTheDocument();
    expect(screen.getByText('ROI маркетинга')).toBeInTheDocument();
  });

  it('renders period filter', () => {
    render(<FranchiserNetworkHealthTab />);
    expect(screen.getAllByText('Квартал').length).toBeGreaterThan(0);
  });
});
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import SalesDynamicsChart from '@/components/Dashboard/tabs/SalesDynamicsChart';

const salesData = [
  { month: 'Янв', plan: 1000000, fact: 900000, forecast: null },
  { month: 'Фев', plan: 1000000, fact: null, forecast: 1100000 },
];

describe('SalesDynamicsChart', () => {
  it('рендерит переключатель вида', () => {
    render(<SalesDynamicsChart data={salesData} />);
    expect(screen.getByText('Помесячно')).toBeInTheDocument();
    expect(screen.getByText('Поквартально')).toBeInTheDocument();
    expect(screen.getByText('Накопленным итогом')).toBeInTheDocument();
  });

  it('переключает на накопленный итог', () => {
    render(<SalesDynamicsChart data={salesData} />);
    fireEvent.click(screen.getByText('Накопленным итогом'));
    expect(screen.getByText('Накопленным итогом')).toBeInTheDocument();
  });

  it('работает с кумулятивным видом по умолчанию', () => {
    render(<SalesDynamicsChart data={salesData} view="cumulative" />);
    expect(screen.getByText('Накопленным итогом')).toBeInTheDocument();
  });
});
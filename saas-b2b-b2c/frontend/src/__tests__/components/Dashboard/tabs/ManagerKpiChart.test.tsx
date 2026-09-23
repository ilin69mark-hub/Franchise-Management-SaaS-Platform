import React from 'react';
import { render, screen } from '@testing-library/react';
import ManagerKpiChart from '@/components/Dashboard/tabs/ManagerKpiChart';

describe('ManagerKpiChart', () => {
  it('показывает «Нет данных» при пустом массиве', () => {
    render(<ManagerKpiChart data={[]} />);
    expect(screen.getByText('Нет данных')).toBeInTheDocument();
  });

  it('не выводит «Нет данных» при наличии данных', () => {
    render(
      <ManagerKpiChart data={[{ month: 'Янв', kpi: 85 }, { month: 'Фев', kpi: 90 }]} />,
    );
    expect(screen.queryByText('Нет данных')).toBeNull();
  });
});
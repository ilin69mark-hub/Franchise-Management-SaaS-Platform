import React from 'react';
import { render, screen } from '@testing-library/react';
import ManagerPlanFactChart from '@/components/Dashboard/tabs/ManagerPlanFactChart';

describe('ManagerPlanFactChart', () => {
  it('показывает «Нет данных» при пустом массиве', () => {
    render(<ManagerPlanFactChart data={[]} />);
    expect(screen.getByText('Нет данных')).toBeInTheDocument();
  });

  it('не выводит «Нет данных» при наличии данных', () => {
    render(
      <ManagerPlanFactChart data={[{ month: 'Янв', plan: 1000, fact: 900 }, { month: 'Фев', plan: 1200, fact: 1100 }]} />,
    );
    expect(screen.queryByText('Нет данных')).toBeNull();
  });
});
import React from 'react';
import { render } from '@testing-library/react';
import ReportPlanFactChart from '@/components/Dashboard/tabs/ReportPlanFactChart';

describe('ReportPlanFactChart', () => {
  it('рендерится без ошибок с моковыми данными', () => {
    const { container } = render(<ReportPlanFactChart />);
    expect(container.firstChild).not.toBeNull();
  });
});
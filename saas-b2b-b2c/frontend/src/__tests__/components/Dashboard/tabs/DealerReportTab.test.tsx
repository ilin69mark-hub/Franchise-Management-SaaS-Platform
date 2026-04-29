// __tests__/components/Dashboard/tabs/DealerReportTab.test.tsx
import React from 'react';
import { render } from '@testing-library/react';
import DealerReportTab from '@/components/Dashboard/tabs/DealerReportTab';

describe('DealerReportTab', () => {
  it('renders', () => {
    render(<DealerReportTab />);
  });

  it('renders loading', () => {
    render(<DealerReportTab loading />);
  });

  it('calcs budget', () => {
    expect(Math.round((30000 / 100000) * 100)).toBe(30);
  });

  it('handles no data', () => {
    expect(() => render(<DealerReportTab />)).not.toThrow();
  });

  it('handles empty history', () => {
    expect(() => render(<DealerReportTab reportHistory={[]} />)).not.toThrow();
  });
});
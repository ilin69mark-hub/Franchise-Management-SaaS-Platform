import React from 'react';
import { render } from '@testing-library/react';
import AlertDropdown from '@/components/Dashboard/Alerts/FranchiserAlerts';

describe('AlertDropdown', () => {
  it('renders without errors', () => {
    const { container } = render(<AlertDropdown />);
    expect(container).toBeInTheDocument();
  });
});
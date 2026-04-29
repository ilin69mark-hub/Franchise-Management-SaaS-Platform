// __tests__/components/Dashboard/tabs/CommunicationsTab.test.tsx
import React from 'react';
import { render } from '@testing-library/react';
import CommunicationsTab from '@/components/Dashboard/tabs/CommunicationsTab';

describe('CommunicationsTab', () => {
  it('renders component', () => {
    render(<CommunicationsTab />);
  });

  it('renders with loading prop', () => {
    render(<CommunicationsTab loading />);
  });
});
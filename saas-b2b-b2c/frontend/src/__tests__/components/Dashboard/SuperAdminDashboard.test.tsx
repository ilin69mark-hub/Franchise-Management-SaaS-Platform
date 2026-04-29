import React from 'react';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import SuperAdminDashboard from '@/components/Dashboard/SuperAdminDashboard';

const createTestStore = () => {
  return configureStore({
    reducer: {
      auth: (state = {}, action) => state,
    },
  });
};

jest.mock('@/store/superAdminStore', () => ({
  useSuperAdminStore: () => ({
    activeSection: 'metrics',
    setActiveSection: jest.fn(),
    stats: {
      mrr: 2500000,
      arr: 30000000,
      churnRate: 5.2,
      overduePayments: 3,
      arpu: 15000,
    },
    isLoading: false,
    alertCount: 2,
    setStats: jest.fn(),
    setLoading: jest.fn(),
    setLastUpdated: jest.fn(),
  }),
}));

jest.mock('@/store/authSlice', () => ({
  logout: jest.fn(),
}));

jest.mock('next/router', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}));

const mockUser = {
  id: '1',
  email: 'admin@example.com',
  first_name: 'Админ',
  last_name: 'Системы',
  role: 'super_admin',
};

describe('SuperAdminDashboard', () => {
  const renderWithProvider = (component: React.ReactElement) => {
    const store = createTestStore();
    return render(<Provider store={store}>{component}</Provider>);
  };

  it('renders header with platform title', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    expect(screen.getByText('SaaS Платформа')).toBeInTheDocument();
  });

  it('renders Super Admin role label', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    expect(screen.getByText('Супер Админ')).toBeInTheDocument();
  });

  it('renders KPI widgets in header', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    const content = document.body.textContent || '';
    expect(content).toContain('MRR');
    expect(content).toContain('ARR');
    expect(content).toContain('Churn Rate');
    expect(content).toContain('Просроч. платежи');
    expect(content).toContain('ARPU');
  });

  it('renders sidebar navigation items', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    const content = document.body.textContent || '';
    expect(content).toContain('SaaS-метрики');
    expect(content).toContain('Тенанты');
    expect(content).toContain('Активность');
    expect(content).toContain('Тех. здоровье');
    expect(content).toContain('Биллинг');
    expect(content).toContain('Аудит');
  });

  it('renders notification bell icon', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    const bellIcon = document.querySelector('.anticon-bell');
    expect(bellIcon).toBeInTheDocument();
  });

  it('renders avatar', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    const avatar = document.querySelector('.ant-avatar');
    expect(avatar).toBeInTheDocument();
  });

  it('displays current date and time', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    const content = document.body.textContent || '';
    expect(content).toMatch(/\d+:\d+/);
  });

  it('highlights active menu item', () => {
    renderWithProvider(<SuperAdminDashboard user={mockUser} />);
    const menuItems = document.querySelectorAll('.ant-menu-item');
    expect(menuItems.length).toBeGreaterThan(0);
  });
});
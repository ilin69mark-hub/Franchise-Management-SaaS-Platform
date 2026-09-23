import React from 'react';
import { render, screen } from '@testing-library/react';
import DashboardPage from '@/pages/index';

const mockPush = jest.fn();
jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock('next/head', () => ({ __esModule: true, default: () => null }));

jest.mock('@/components/Dashboard/FranchiserDashboard', () => {
  const Comp: React.FC<{ title: string }> = ({ title }) => <div>FranchiserDashboard {title}</div>;
  return { __esModule: true, default: Comp };
});
jest.mock('@/components/Dashboard/FranchiserManagerDashboard', () => {
  const Comp: React.FC<{ title: string }> = ({ title }) => <div>FranchiserManagerDashboard {title}</div>;
  return { __esModule: true, default: Comp };
});
jest.mock('@/components/Dashboard/DealerDashboard', () => {
  const Comp: React.FC<{ title: string }> = ({ title }) => <div>DealerDashboard {title}</div>;
  return { __esModule: true, default: Comp };
});
jest.mock('@/components/Dashboard/SalonManagerDashboard', () => {
  const Comp: React.FC<{ title: string }> = ({ title }) => <div>SalonManagerDashboard {title}</div>;
  return { __esModule: true, default: Comp };
});

let mockAuth: { isAuthenticated: boolean; user: { role: string; id: string } | null };
jest.mock('react-redux', () => ({
  useSelector: (selector: (s: { auth: typeof mockAuth }) => unknown) => selector({ auth: mockAuth }),
}));

describe('DashboardPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => 'token') }, writable: true });
  });

  it('показывает загрузку когда нет user', () => {
    mockAuth = { isAuthenticated: true, user: null };
    const { container } = render(<DashboardPage />);
    expect(container.querySelector('.ant-layout')).toBeInTheDocument();
  });

  it('редиректит super_admin', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'super_admin', id: '1' } };
    render(<DashboardPage />);
    expect(mockPush).toHaveBeenCalledWith('/admin');
  });

  it('рендерит franchiser', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'franchiser', id: '1' } };
    render(<DashboardPage />);
    expect(screen.getByText(/FranchiserDashboard/)).toBeInTheDocument();
  });

  it('рендерит franchiser_manager', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'franchiser_manager', id: '1' } };
    render(<DashboardPage />);
    expect(screen.getByText(/FranchiserManagerDashboard/)).toBeInTheDocument();
  });

  it('рендерит dealer', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'dealer', id: '1' } };
    render(<DashboardPage />);
    expect(screen.getByText(/DealerDashboard/)).toBeInTheDocument();
  });

  it('рендерит salon_manager', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'salon_manager', id: '1' } };
    render(<DashboardPage />);
    expect(screen.getByText(/SalonManagerDashboard/)).toBeInTheDocument();
  });

  it('рендерит дефолтный franchiser для неизвестной роли', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'unknown', id: '1' } };
    render(<DashboardPage />);
    expect(screen.getByText(/FranchiserDashboard/)).toBeInTheDocument();
  });

  it('редиректит неавторизованного', () => {
    mockAuth = { isAuthenticated: false, user: null };
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => null) }, writable: true });
    render(<DashboardPage />);
    expect(mockPush).toHaveBeenCalledWith('/login');
  });
});

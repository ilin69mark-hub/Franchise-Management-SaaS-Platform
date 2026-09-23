import React from 'react';
import { render, screen } from '@testing-library/react';
import SalonManagerPage from '@/pages/salon-manager';

const mockPush = jest.fn();
jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock('next/head', () => ({ __esModule: true, default: () => null }));

jest.mock('@/components/Dashboard/Header', () => {
  const Comp: React.FC = () => <div>Header</div>;
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

describe('SalonManagerPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => 'token') }, writable: true });
  });

  it('показывает загрузку когда нет user', () => {
    mockAuth = { isAuthenticated: true, user: null };
    render(<SalonManagerPage />);
    expect(screen.getByText('Header')).toBeInTheDocument();
  });

  it('рендерит salon_manager', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'salon_manager', id: '1' } };
    render(<SalonManagerPage />);
    expect(screen.getByText(/SalonManagerDashboard/)).toBeInTheDocument();
  });

  it('редиректит super_admin', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'super_admin', id: '1' } };
    render(<SalonManagerPage />);
    expect(mockPush).toHaveBeenCalledWith('/admin');
  });

  it('редиректит franchiser', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'franchiser', id: '1' } };
    render(<SalonManagerPage />);
    expect(mockPush).toHaveBeenCalledWith('/franchiser-manager');
  });

  it('редиректит dealer', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'dealer', id: '1' } };
    render(<SalonManagerPage />);
    expect(mockPush).toHaveBeenCalledWith('/dealer');
  });

  it('редиректит неизвестную роль', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'unknown', id: '1' } };
    render(<SalonManagerPage />);
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('редиректит неавторизованного', () => {
    mockAuth = { isAuthenticated: false, user: null };
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => null) }, writable: true });
    render(<SalonManagerPage />);
    expect(mockPush).toHaveBeenCalledWith('/login');
  });
});

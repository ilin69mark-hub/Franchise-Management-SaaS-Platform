import React from 'react';
import { render, screen } from '@testing-library/react';
import AdminPage from '@/pages/admin';

const mockPush = jest.fn();
jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock('next/head', () => ({ __esModule: true, default: () => null }));

jest.mock('@/components/Dashboard/Header', () => {
  const Comp: React.FC = () => <div>Header</div>;
  return { __esModule: true, default: Comp };
});

jest.mock('@/components/Dashboard/SuperAdminDashboard', () => {
  const Comp: React.FC<{ user: unknown }> = ({ user }) => <div>SuperAdminDashboard {(user as { role: string }).role}</div>;
  return { __esModule: true, default: Comp };
});

let mockAuth: { isAuthenticated: boolean; user: { role: string; id: string } | null };
jest.mock('react-redux', () => ({
  useSelector: (selector: (s: { auth: typeof mockAuth }) => unknown) => selector({ auth: mockAuth }),
}));

describe('AdminPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => 'token') }, writable: true });
  });

  it('показывает загрузку когда нет user', () => {
    mockAuth = { isAuthenticated: true, user: null };
    render(<AdminPage />);
    expect(screen.getByText('Header')).toBeInTheDocument();
  });

  it('рендерит супер-админа', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'super_admin', id: '1' } };
    render(<AdminPage />);
    expect(screen.getByText(/SuperAdminDashboard/)).toBeInTheDocument();
  });

  it('редиректит дилера', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'dealer', id: '1' } };
    render(<AdminPage />);
    expect(mockPush).toHaveBeenCalledWith('/dealer');
  });

  it('редиректит франчайзера', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'franchiser', id: '1' } };
    render(<AdminPage />);
    expect(mockPush).toHaveBeenCalledWith('/franchiser-manager');
  });

  it('редиректит franchiser_manager', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'franchiser_manager', id: '1' } };
    render(<AdminPage />);
    expect(mockPush).toHaveBeenCalledWith('/franchiser-manager');
  });

  it('редиректит salon_manager', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'salon_manager', id: '1' } };
    render(<AdminPage />);
    expect(mockPush).toHaveBeenCalledWith('/salon-manager');
  });

  it('редиректит неизвестную роль на /', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'unknown', id: '1' } };
    render(<AdminPage />);
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('редиректит неавторизованного на login', () => {
    mockAuth = { isAuthenticated: false, user: null };
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => null) }, writable: true });
    render(<AdminPage />);
    expect(mockPush).toHaveBeenCalledWith('/login');
  });
});

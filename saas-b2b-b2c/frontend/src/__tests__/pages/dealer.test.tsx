import React from 'react';
import { render, screen } from '@testing-library/react';
import DealerPage from '@/pages/dealer';

const mockPush = jest.fn();
jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock('next/head', () => ({ __esModule: true, default: () => null }));

jest.mock('@/components/Dashboard/Header', () => {
  const Comp: React.FC = () => <div>Header</div>;
  return { __esModule: true, default: Comp };
});

jest.mock('@/components/Dashboard/DealerDashboard', () => {
  const Comp: React.FC<{ user: unknown; title: string }> = ({ title }) => <div>DealerDashboard {title}</div>;
  return { __esModule: true, default: Comp };
});

let mockAuth: { isAuthenticated: boolean; user: { role: string; id: string } | null };
jest.mock('react-redux', () => ({
  useSelector: (selector: (s: { auth: typeof mockAuth }) => unknown) => selector({ auth: mockAuth }),
}));

describe('DealerPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => 'token') }, writable: true });
  });

  it('показывает загрузку когда нет user', () => {
    mockAuth = { isAuthenticated: true, user: null };
    render(<DealerPage />);
    expect(screen.getByText('Header')).toBeInTheDocument();
  });

  it('рендерит дилера', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'dealer', id: '1' } };
    render(<DealerPage />);
    expect(screen.getByText(/DealerDashboard/)).toBeInTheDocument();
  });

  it('редиректит super_admin', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'super_admin', id: '1' } };
    render(<DealerPage />);
    expect(mockPush).toHaveBeenCalledWith('/admin');
  });

  it('редиректит franchiser', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'franchiser', id: '1' } };
    render(<DealerPage />);
    expect(mockPush).toHaveBeenCalledWith('/franchiser-manager');
  });

  it('редиректит salon_manager', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'salon_manager', id: '1' } };
    render(<DealerPage />);
    expect(mockPush).toHaveBeenCalledWith('/salon-manager');
  });

  it('редиректит неизвестную роль', () => {
    mockAuth = { isAuthenticated: true, user: { role: 'unknown', id: '1' } };
    render(<DealerPage />);
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('редиректит неавторизованного', () => {
    mockAuth = { isAuthenticated: false, user: null };
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => null) }, writable: true });
    render(<DealerPage />);
    expect(mockPush).toHaveBeenCalledWith('/login');
  });
});

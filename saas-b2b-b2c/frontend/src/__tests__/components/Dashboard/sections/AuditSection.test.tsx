import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import AuditSection from '@/components/Dashboard/sections/AuditSection';

jest.mock('@/store/auditStore', () => ({
  useAuditStore: () => ({
    adminActions: [],
    impersonations: [],
    activeSessions: [],
    userLogins: [],
    filters: { admin: '', action: '', dateRange: null, tenant: '' },
    isLoading: false,
    setAdminActions: jest.fn(),
    setImpersonations: jest.fn(),
    setActiveSessions: jest.fn(),
    setUserLogins: jest.fn(),
    setFilters: jest.fn(),
    setLoading: jest.fn(),
  }),
}));

describe('AuditSection', () => {
  it('renders section title', async () => {
    render(<AuditSection />);
    await waitFor(() => {
      expect(screen.getByText('Аудит')).toBeInTheDocument();
    });
  });

  it('renders admin actions table', async () => {
    render(<AuditSection />);
    await waitFor(() => {
      expect(screen.getByText('Журнал действий суперадминов')).toBeInTheDocument();
    });
  });

  it('renders impersonations table', async () => {
    render(<AuditSection />);
    await waitFor(() => {
      expect(screen.getByText('Журнал имперсонаций')).toBeInTheDocument();
    });
  });

  it('renders active sessions table', async () => {
    render(<AuditSection />);
    await waitFor(() => {
      expect(screen.getByText('Активные сессии')).toBeInTheDocument();
    });
  });

  it('renders user logins table', async () => {
    render(<AuditSection />);
    await waitFor(() => {
      expect(screen.getByText('Входы пользователей')).toBeInTheDocument();
    });
  });
});
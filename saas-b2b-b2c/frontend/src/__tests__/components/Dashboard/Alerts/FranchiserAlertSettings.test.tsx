import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import FranchiserAlertSettings from '@/components/Dashboard/Alerts/FranchiserAlertSettings';
import type { FranchiserAlertSettings as Settings } from '@/store/franchiserAlertStore';

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: {
      ...actualAntd.message,
      success: jest.fn(),
      error: jest.fn(),
    },
  };
});

const mockSetSettings = jest.fn();

jest.mock('@/store/franchiserAlertStore', () => ({
  useFranchiserAlertStore: () => ({ settings: mockSettings, setSettings: mockSetSettings }),
}));

let mockSettings: Settings = {
  enabled: true,
  channel: { inApp: true, email: false, push: false },
  thresholds: { criticalForecastPercent: 60, dealerChurnPercent: 15, managerKpiPercent: 50 },
};

describe('FranchiserAlertSettings', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockSettings = {
      enabled: true,
      channel: { inApp: true, email: false, push: false },
      thresholds: { criticalForecastPercent: 60, dealerChurnPercent: 15, managerKpiPercent: 50 },
    };
  });

  it('рендерит каналы и пороговые значения', () => {
    render(<FranchiserAlertSettings />);
    expect(screen.getByText('Настройки уведомлений')).toBeInTheDocument();
    expect(screen.getByText('Каналы уведомлений')).toBeInTheDocument();
    expect(screen.getByText('In-app уведомления')).toBeInTheDocument();
    expect(screen.getByText('Email дайджест (раз в неделю)')).toBeInTheDocument();
    expect(screen.getByText('Push-уведомления')).toBeInTheDocument();
    expect(screen.getByText('Пороговые значения')).toBeInTheDocument();
    expect(screen.getByText('Критический прогноз сети (%)')).toBeInTheDocument();
    expect(screen.getByText('Порог оттока дилеров (%)')).toBeInTheDocument();
    expect(screen.getByText('Порог KPI менеджера (%)')).toBeInTheDocument();
    expect(screen.getByText('Сохранить настройки')).toBeInTheDocument();
  });

  it('сохраняет настройки', async () => {
    render(<FranchiserAlertSettings />);
    fireEvent.click(screen.getByText('Сохранить настройки'));
    await waitFor(() => {
      expect(mockSetSettings).toHaveBeenCalled();
      expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Настройки сохранены');
    });
  });
});
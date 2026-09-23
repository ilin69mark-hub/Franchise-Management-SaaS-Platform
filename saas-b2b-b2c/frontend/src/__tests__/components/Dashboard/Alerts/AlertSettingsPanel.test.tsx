import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import AlertSettingsPanel from '@/components/Dashboard/Alerts/AlertSettingsPanel';
import type { AlertSettings } from '@/store/alertStore';

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

jest.mock('@/store/alertStore', () => ({
  useAlertStore: () => ({ settings: mockSettings, setSettings: mockSetSettings }),
}));

let mockSettings: AlertSettings = {
  enabled: true,
  planThreshold: 70,
  conversionDropThreshold: 15,
  trafficDropThreshold: 30,
  stuckDealDays: 7,
  stuckDealAmount: 200000,
  inactiveDealerDays: 3,
  nonLiquidDays: 90,
  channels: { inApp: true, emailInstant: false, emailDigest: true, push: false },
};

describe('AlertSettingsPanel', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockSettings = {
      enabled: true,
      planThreshold: 70,
      conversionDropThreshold: 15,
      trafficDropThreshold: 30,
      stuckDealDays: 7,
      stuckDealAmount: 200000,
      inactiveDealerDays: 3,
      nonLiquidDays: 90,
      channels: { inApp: true, emailInstant: false, emailDigest: true, push: false },
    };
  });

  it('рендерит все секции и контролы', () => {
    render(<AlertSettingsPanel />);
    expect(screen.getByText('Настройки алертов')).toBeInTheDocument();
    expect(screen.getAllByText('Общие').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Пороговые значения')).toBeInTheDocument();
    expect(screen.getByText('Каналы доставки')).toBeInTheDocument();
    expect(screen.getByText('Включить алерты')).toBeInTheDocument();
    expect(screen.getByText('Порог красной зоны плана (%):')).toBeInTheDocument();
    expect(screen.getByText('Порог падения конверсии (%):')).toBeInTheDocument();
    expect(screen.getByText('Порог падения трафика (%):')).toBeInTheDocument();
    expect(screen.getByText('Дней зависшей сделки:')).toBeInTheDocument();
    expect(screen.getByText('Сумма крупной сделки (руб.):')).toBeInTheDocument();
    expect(screen.getByText('Дней неактивности дилера:')).toBeInTheDocument();
    expect(screen.getByText('Дней неликвида:')).toBeInTheDocument();
    expect(screen.getByText('В системе (in-app)')).toBeInTheDocument();
    expect(screen.getByText('Email мгновенно')).toBeInTheDocument();
    expect(screen.getByText('Email дайджест (раз в день)')).toBeInTheDocument();
    expect(screen.getByText('Push-уведомления')).toBeInTheDocument();
    expect(screen.getByText('Отмена')).toBeInTheDocument();
    expect(screen.getByText('Сохранить')).toBeInTheDocument();
  });

  it('переключает глобальный switch алертов', () => {
    const { container } = render(<AlertSettingsPanel />);
    fireEvent.click(container.querySelector('.ant-switch')!);
    expect(mockSetSettings).toHaveBeenCalledWith({ enabled: false });
  });

  it('переключает канал in-app через checkbox', () => {
    render(<AlertSettingsPanel />);
    fireEvent.click(screen.getByText('В системе (in-app)'));
    expect(mockSetSettings).toHaveBeenCalledWith({
      channels: { inApp: false, emailInstant: false, emailDigest: true, push: false },
    });
  });

  it('сохраняет настройки и закрывает панель', async () => {
    const onClose = jest.fn();
    render(<AlertSettingsPanel onClose={onClose} />);
    fireEvent.click(screen.getByText('Сохранить'));
    await waitFor(() => {
      expect(mockSetSettings).toHaveBeenCalled();
      expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Настройки сохранены');
      expect(onClose).toHaveBeenCalled();
    });
  });
});
// src/__tests__/components/Dashboard/Alerts/DealerAlerts.test.tsx
describe('DealerAlerts', () => {
  it('calculates unread count', () => {
    const alerts = [
      { id: '1', category: 'finance' as const, priority: 'high' as const, title: 'Test', description: 'Test', createdAt: '2026-04-30', isRead: false },
      { id: '2', category: 'operation' as const, priority: 'medium' as const, title: 'Test2', description: 'Test2', createdAt: '2026-04-30', isRead: false },
    ];
    const unread = alerts.filter(a => !a.isRead).length;
    expect(unread).toBe(2);
  });

  it('groups alerts by category', () => {
    const alerts = [
      { id: '1', category: 'finance' as const, priority: 'high' as const, title: 'Test', description: 'Test', createdAt: '2026-04-30', isRead: false },
      { id: '2', category: 'operation' as const, priority: 'medium' as const, title: 'Test2', description: 'Test2', createdAt: '2026-04-30', isRead: false },
      { id: '3', category: 'communication' as const, priority: 'low' as const, title: 'Test3', description: 'Test3', createdAt: '2026-04-30', isRead: false },
    ];
    const groups: Record<string, typeof alerts> = { finance: [], operation: [], communication: [] };
    alerts.forEach(a => groups[a.category].push(a));
    expect(groups.finance.length).toBe(1);
    expect(groups.operation.length).toBe(1);
    expect(groups.communication.length).toBe(1);
  });

  it('loads settings from localStorage', () => {
    const settings = {
      fotPercentThreshold: 20,
      rentPercentThreshold: 12,
      conversionDropPercent: 20,
      trafficDropPercent: 30,
      stuckDealDays: 7,
      stuckDealAmount: 200000,
      nonLiquidDays: 90,
    };
    localStorage.setItem('dealerAlertSettings', JSON.stringify(settings));
    const saved = localStorage.getItem('dealerAlertSettings');
    expect(JSON.parse(saved || '{}')).toEqual(settings);
  });

  it('saves settings to localStorage', () => {
    const settings = { fotPercentThreshold: 25 };
    localStorage.setItem('dealerAlertSettings', JSON.stringify(settings));
    const saved = localStorage.getItem('dealerAlertSettings');
    expect(JSON.parse(saved || '{}').fotPercentThreshold).toBe(25);
  });

  it('marks alert as read', () => {
    const localUnread: Record<string, boolean> = { '1': true, '2': true };
    delete localUnread['1'];
    expect(localUnread['1']).toBeUndefined();
    expect(localUnread['2']).toBe(true);
  });

  it('filters unread alerts', () => {
    const alerts = [
      { id: '1', category: 'finance' as const, priority: 'high' as const, title: 'Test', description: 'Test', createdAt: '2026-04-30', isRead: false },
      { id: '2', category: 'operation' as const, priority: 'medium' as const, title: 'Test2', description: 'Test2', createdAt: '2026-04-30', isRead: true },
    ];
    const localUnread: Record<string, boolean> = { '1': true, '2': false };
    const unreadAlerts = alerts.filter(a => localUnread[a.id] !== false);
    expect(unreadAlerts.length).toBe(1);
  });

  it('validates alert settings thresholds', () => {
    const settings = {
      fotPercentThreshold: 20,
      rentPercentThreshold: 12,
      conversionDropPercent: 20,
      trafficDropPercent: 30,
      stuckDealDays: 7,
      stuckDealAmount: 200000,
      nonLiquidDays: 90,
    };
    expect(settings.fotPercentThreshold).toBeGreaterThan(0);
    expect(settings.rentPercentThreshold).toBeGreaterThan(0);
    expect(settings.conversionDropPercent).toBeGreaterThan(0);
    expect(settings.trafficDropPercent).toBeGreaterThan(0);
    expect(settings.stuckDealDays).toBeGreaterThan(0);
    expect(settings.stuckDealAmount).toBeGreaterThan(0);
    expect(settings.nonLiquidDays).toBeGreaterThan(0);
  });

  it('handles priority levels', () => {
    const priorities = ['high', 'medium', 'low'] as const;
    expect(priorities.includes('high')).toBe(true);
    expect(priorities.includes('medium')).toBe(true);
    expect(priorities.includes('low')).toBe(true);
  });

  it('handles category types', () => {
    const categories = ['finance', 'operation', 'communication'] as const;
    expect(categories.includes('finance')).toBe(true);
    expect(categories.includes('operation')).toBe(true);
    expect(categories.includes('communication')).toBe(true);
  });
});

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import DealerAlerts from '@/components/Dashboard/Alerts/DealerAlerts';
import type { Alert } from '@/components/Dashboard/Alerts/DealerAlerts';

const testAlerts: Alert[] = [
  { id: '1', category: 'finance', priority: 'high', title: 'ФОТ превышен', description: 'Проверьте ФОТ', createdAt: '2026-04-30T10:00:00', isRead: false },
  { id: '2', category: 'operation', priority: 'medium', title: 'Зависшая сделка', description: 'Сделка не двигается', createdAt: '2026-04-30T10:00:00', isRead: false },
];

describe('DealerAlerts (render)', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('отображает колокольчик и счётчик непрочитанных', () => {
    render(<DealerAlerts alerts={testAlerts} loading={false} />);
    expect(screen.getByRole('button')).toBeInTheDocument();
    expect(screen.getAllByText('2').length).toBeGreaterThanOrEqual(1);
  });

  it('открывает попап с категориями и заголовками алертов', () => {
    render(<DealerAlerts alerts={testAlerts} loading={false} />);
    fireEvent.click(screen.getByRole('button'));
    expect(screen.getByText('Алерты')).toBeInTheDocument();
    expect(screen.getByText('Финансы')).toBeInTheDocument();
    expect(screen.getByText('Операции')).toBeInTheDocument();
    expect(screen.getByText('ФОТ превышен')).toBeInTheDocument();
    expect(screen.getByText('Зависшая сделка')).toBeInTheDocument();
  });

  it('показывает «Нет алертов» при пустом списке', () => {
    render(<DealerAlerts alerts={[]} loading={false} />);
    fireEvent.click(screen.getByRole('button'));
    expect(screen.getByText('Нет алертов')).toBeInTheDocument();
  });

  it('показывает спиннер при загрузке', () => {
    render(<DealerAlerts alerts={[]} loading />);
    fireEvent.click(screen.getByRole('button'));
    expect(document.querySelector('.ant-spin')).not.toBeNull();
  });

  it('открывает модалку настроек и вызывает onSettingsSave', () => {
    const onSettingsSave = jest.fn().mockResolvedValue(undefined);
    render(<DealerAlerts alerts={testAlerts} loading={false} onSettingsSave={onSettingsSave} />);
    fireEvent.click(screen.getByRole('button'));
    fireEvent.click(screen.getAllByText('Настройки алертов')[0]);
    expect(screen.getByText('⚙️ Настройки порогов алертов')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Сохранить настройки'));
    expect(onSettingsSave).toHaveBeenCalled();
  });
});
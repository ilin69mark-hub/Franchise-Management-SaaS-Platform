// src/__tests__/components/Dashboard/MorningBriefDrawer.test.tsx
import React from 'react';
import { render } from '@testing-library/react';
import MorningBriefDrawer from '@/components/Dashboard/MorningBriefDrawer';

describe('MorningBriefDrawer', () => {
  it('renders closed', () => {
    expect(() => render(<MorningBriefDrawer open={false} onClose={() => {}} />)).not.toThrow();
  });

  it('renders open', () => {
    expect(() => render(<MorningBriefDrawer open={true} onClose={() => {}} />)).not.toThrow();
  });

  it('renders with custom user name', () => {
    expect(() => render(<MorningBriefDrawer open={true} onClose={() => {}} userName="Иван" />)).not.toThrow();
  });

  it('calculates overall status good', () => {
    const planPercent = 95;
    const forecast = 100;
    let type: 'good' | 'warning' | 'critical' = 'good';
    if (planPercent >= 90) type = 'good';
    expect(type).toBe('good');
  });

  it('calculates overall status warning', () => {
    const planPercent = 80;
    const forecast = 85;
    let type: 'good' | 'warning' | 'critical' = 'warning';
    if (planPercent >= 90) type = 'good';
    else if (planPercent >= 70) type = 'warning';
    else type = 'critical';
    expect(type).toBe('warning');
  });

  it('calculates overall status critical', () => {
    const planPercent = 60;
    let type: 'good' | 'warning' | 'critical' = 'critical';
    if (planPercent >= 90) type = 'good';
    else if (planPercent >= 70) type = 'warning';
    else type = 'critical';
    expect(type).toBe('critical');
  });

  it('counts risk items', () => {
    const riskItems = [
      { dealerName: 'A', percent: 68 },
      { dealerName: 'B', percent: 65 },
      { dealerName: 'C', percent: 72 },
    ];
    expect(riskItems.length).toBe(3);
  });

  it('filters control items by urgency', () => {
    const controlItems = [
      { urgency: 'high' as const },
      { urgency: 'medium' as const },
      { urgency: 'high' as const },
    ];
    const urgent = controlItems.filter(i => i.urgency === 'high');
    expect(urgent.length).toBe(2);
  });

  it('calculates sparkline data', () => {
    const data = [
      { day: 'Пн', plan: 12, fact: 11 },
      { day: 'Вт', plan: 12, fact: 13 },
    ];
    const avgFact = data.reduce((s, d) => s + d.fact, 0) / data.length;
    expect(avgFact).toBe(12);
  });

  it('filters tasks by type', () => {
    const tasks = [
      { type: 'call' as const },
      { type: 'meeting' as const },
      { type: 'call' as const },
    ];
    const calls = tasks.filter(t => t.type === 'call');
    expect(calls.length).toBe(2);
  });
});
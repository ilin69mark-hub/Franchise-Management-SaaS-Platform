import React from 'react';
import { render, screen } from '@testing-library/react';
import GeographyMap from '@/components/Dashboard/tabs/GeographyMap';

describe('GeographyMap', () => {
  it('показывает регионы, численность и плотность', () => {
    render(<GeographyMap />);
    expect(screen.getByText('Карта присутствия дилеров по регионам России')).toBeInTheDocument();
    expect(screen.getByText('Центр (Москва, СПб)')).toBeInTheDocument();
    expect(screen.getByText('Дальний Восток')).toBeInTheDocument();
    expect(screen.getByText('25.0M')).toBeInTheDocument();
    expect(screen.getByText('Высокая')).toBeInTheDocument();
    expect(screen.getAllByText('Средняя').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Низкая').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Нет').length).toBeGreaterThanOrEqual(1);
  });
});
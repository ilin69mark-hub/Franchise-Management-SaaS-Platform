import React from 'react';
import { render, screen } from '@testing-library/react';
import Header from '@/components/Dashboard/Header';

describe('Header', () => {
  it('показывает название бренда ivan.ru', () => {
    render(<Header />);
    expect(screen.getByText('ivan.ru')).toBeInTheDocument();
  });
});
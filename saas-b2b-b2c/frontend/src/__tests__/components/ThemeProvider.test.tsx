import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import ThemeProvider, { useThemeMode } from '@/components/ThemeProvider';

const Consumer: React.FC = () => {
  const { theme, toggleTheme, setTheme } = useThemeMode();
  return (
    <div>
      <span data-testid="theme">{theme}</span>
      <button onClick={toggleTheme}>toggle</button>
      <button onClick={() => setTheme('light')}>force light</button>
    </div>
  );
};

const renderWithProvider = () => {
  const utils = render(
    <ThemeProvider>
      <Consumer />
    </ThemeProvider>,
  );
  act(() => {});
  return utils;
};

describe('ThemeProvider', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('по умолчанию светлая тема', () => {
    renderWithProvider();
    expect(screen.getByTestId('theme')).toHaveTextContent('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('читает тему из localStorage', () => {
    localStorage.setItem('franchiseTheme', 'dark');
    renderWithProvider();
    expect(screen.getByTestId('theme')).toHaveTextContent('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('переключает тему', () => {
    renderWithProvider();
    fireEvent.click(screen.getByText('toggle'));
    expect(screen.getByTestId('theme')).toHaveTextContent('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('устанавливает тему через setTheme', () => {
    renderWithProvider();
    fireEvent.click(screen.getByText('force light'));
    expect(screen.getByTestId('theme')).toHaveTextContent('light');
  });
});
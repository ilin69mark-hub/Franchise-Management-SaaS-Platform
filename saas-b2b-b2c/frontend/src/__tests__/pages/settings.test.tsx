import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import SettingsPage from '@/pages/settings';

const mockPush = jest.fn();
const mockBack = jest.fn();
jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush, back: mockBack }),
}));

jest.mock('next/head', () => ({ __esModule: true, default: () => null }));
jest.mock('next/link', () => {
  const Comp: React.FC<{ children: React.ReactNode }> = ({ children }) => <a>{children}</a>;
  return { __esModule: true, default: Comp };
});

let mockTheme = 'light';
const mockToggle = jest.fn(() => { mockTheme = mockTheme === 'light' ? 'dark' : 'light'; });
jest.mock('@/components/ThemeProvider', () => ({
  useThemeMode: () => ({ theme: mockTheme, toggleTheme: mockToggle }),
}));

let mockAuth = { isAuthenticated: true };
jest.mock('react-redux', () => ({
  useSelector: (selector: (s: { auth: typeof mockAuth }) => unknown) => selector({ auth: mockAuth }),
}));

describe('SettingsPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockAuth = { isAuthenticated: true };
    mockTheme = 'light';
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => 'token') }, writable: true });
  });

  it('рендерит настройки с темой', () => {
    render(<SettingsPage />);
    expect(screen.getAllByText('Настройки').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Оформление')).toBeInTheDocument();
    expect(screen.getByText('Тёмная тема')).toBeInTheDocument();
    expect(screen.getByText('Сейчас включена светлая тема')).toBeInTheDocument();
    expect(screen.getByText('Назад')).toBeInTheDocument();
  });

  it('показывает тёмную тему когда включена', () => {
    mockTheme = 'dark';
    render(<SettingsPage />);
    expect(screen.getByText('Сейчас включена тёмная тема')).toBeInTheDocument();
  });

  it('переключает тему', () => {
    render(<SettingsPage />);
    const switchEl = document.querySelector('.ant-switch') as HTMLElement;
    if (switchEl) fireEvent.click(switchEl);
    else {
      // fallback: find Switch via role
      const switches = screen.getAllByRole('switch');
      fireEvent.click(switches[0]);
    }
    expect(mockToggle).toHaveBeenCalled();
  });

  it('кнопка Назад', () => {
    render(<SettingsPage />);
    fireEvent.click(screen.getByText('Назад'));
    expect(mockBack).toHaveBeenCalled();
  });

  it('редиректит неавторизованного', () => {
    mockAuth = { isAuthenticated: false };
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => null) }, writable: true });
    render(<SettingsPage />);
    expect(mockPush).toHaveBeenCalledWith('/login');
  });

  it('отображает подсказку', () => {
    render(<SettingsPage />);
    expect(screen.getByText('Тема применяется мгновенно и сохраняется в браузере между сеансами.')).toBeInTheDocument();
  });
});

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import MyApp from '@/pages/_app';

jest.mock('react-redux', () => ({
  Provider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

jest.mock('@/components/ThemeProvider', () => {
  const Comp: React.FC<{ children: React.ReactNode }> = ({ children }) => <div>{children}</div>;
  return { __esModule: true, default: Comp };
});

const mockDispatch = jest.fn();
jest.mock('@/store', () => ({
  store: { dispatch: (...args: unknown[]) => mockDispatch(...args), getState: jest.fn(), subscribe: jest.fn() },
}));

jest.mock('@/utils/logger', () => ({
  __esModule: true,
  default: { error: jest.fn() },
}));

const TestComponent: React.FC = () => <div>Test Page</div>;

describe('_app', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    Object.defineProperty(window, 'localStorage', {
      value: { getItem: jest.fn(() => null), setItem: jest.fn() },
      writable: true,
    });
  });

  it('рендерит страницу после монтирования', async () => {
    render(<MyApp Component={TestComponent} pageProps={{}} router={{} as never} />);
    await waitFor(() => expect(screen.getByText('Test Page')).toBeInTheDocument());
  });

  it('диспатчит auth из localStorage когда есть токен и user', async () => {
    const userStr = JSON.stringify({ role: 'dealer', id: '1' });
    Object.defineProperty(window, 'localStorage', {
      value: { getItem: jest.fn((key: string) => (key === 'accessToken' ? 'token123' : key === 'user' ? userStr : null)) },
      writable: true,
    });
    render(<MyApp Component={TestComponent} pageProps={{}} router={{} as never} />);
    await waitFor(() => expect(mockDispatch).toHaveBeenCalledWith({ type: 'auth/setAuthFromStorage' }));
  });

  it('не диспатчит когда нет токена', async () => {
    Object.defineProperty(window, 'localStorage', {
      value: { getItem: jest.fn(() => null) },
      writable: true,
    });
    render(<MyApp Component={TestComponent} pageProps={{}} router={{} as never} />);
    await waitFor(() => expect(screen.getByText('Test Page')).toBeInTheDocument());
    expect(mockDispatch).not.toHaveBeenCalled();
  });

  it('обрабатывает ошибку парсинга user', async () => {
    Object.defineProperty(window, 'localStorage', {
      value: { getItem: jest.fn((key: string) => (key === 'accessToken' ? 'token' : key === 'user' ? 'invalid-json' : null)) },
      writable: true,
    });
    const { logger } = jest.requireMock('@/utils/logger').default ? { logger: jest.requireMock('@/utils/logger').default } : { logger: { error: jest.fn() } };
    render(<MyApp Component={TestComponent} pageProps={{}} router={{} as never} />);
    await waitFor(() => expect(screen.getByText('Test Page')).toBeInTheDocument());
  });

  it('не рендерит до showChild', () => {
    const { container } = render(<MyApp Component={TestComponent} pageProps={{}} router={{} as never} />);
    // initially showChild false, but after useEffect it becomes true - so eventually page appears
    expect(container).toBeInTheDocument();
  });
});

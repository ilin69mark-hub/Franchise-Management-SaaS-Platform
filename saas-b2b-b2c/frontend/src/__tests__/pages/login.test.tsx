import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import LoginPage from '@/pages/login';

const mockReplace = jest.fn();
const mockPush = jest.fn();

jest.mock('next/router', () => ({
  useRouter: () => ({ replace: mockReplace, push: mockPush }),
}));

jest.mock('next/head', () => ({ __esModule: true, default: () => null }));

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: { ...actualAntd.message, success: jest.fn(), error: jest.fn() },
  };
});

jest.mock('@/components/Dashboard/BackButton', () => {
  const Comp: React.FC = () => <div>BackButton</div>;
  return { __esModule: true, default: Comp };
});

const mockLogin = jest.fn();
jest.mock('@/store/authSlice', () => ({
  login: (payload: unknown) => ({ type: 'auth/login/pending', payload, unwrap: mockLogin }),
}));

let mockState: { auth: { loading: boolean; error: string | null; isAuthenticated: boolean; user: { role: string } | null } };

jest.mock('react-redux', () => ({
  useSelector: (selector: (s: typeof mockState) => unknown) => selector(mockState),
  useDispatch: () => jest.fn(() => ({ unwrap: mockLogin })),
}));

describe('LoginPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockState = { auth: { loading: false, error: null, isAuthenticated: false, user: null } };
    mockLogin.mockResolvedValue({ user: { role: 'dealer' } });
  });

  it('рендерит форму входа', () => {
    render(<LoginPage />);
    expect(screen.getByText('Вход в систему')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Пароль')).toBeInTheDocument();
    expect(screen.getByText('Войти')).toBeInTheDocument();
    expect(screen.getByText('BackButton')).toBeInTheDocument();
    expect(screen.getByText('Зарегистрироваться')).toBeInTheDocument();
  });

  it('показывает ошибку', () => {
    mockState = { auth: { loading: false, error: 'Неверный пароль', isAuthenticated: false, user: null } };
    render(<LoginPage />);
    expect(screen.getByText('Неверный пароль')).toBeInTheDocument();
  });

  it('показывает loading на кнопке', () => {
    mockState = { auth: { loading: true, error: null, isAuthenticated: false, user: null } };
    render(<LoginPage />);
    expect(screen.getByText('Войти').closest('button')).toHaveClass('ant-btn-loading');
  });

  it('редиректит залогиненного super_admin', async () => {
    mockState = { auth: { loading: false, error: null, isAuthenticated: true, user: { role: 'super_admin' } } };
    render(<LoginPage />);
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/admin'));
  });

  it('редиректит франчайзера', async () => {
    mockState = { auth: { loading: false, error: null, isAuthenticated: true, user: { role: 'franchiser' } } };
    render(<LoginPage />);
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/franchiser-manager'));
  });

  it('редиректит дилера', async () => {
    mockState = { auth: { loading: false, error: null, isAuthenticated: true, user: { role: 'dealer' } } };
    render(<LoginPage />);
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/dealer'));
  });

  it('редиректит salon_manager', async () => {
    mockState = { auth: { loading: false, error: null, isAuthenticated: true, user: { role: 'salon_manager' } } };
    render(<LoginPage />);
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/salon-manager'));
  });

  it('переходит на регистрацию', () => {
    render(<LoginPage />);
    fireEvent.click(screen.getByText('Зарегистрироваться'));
    expect(mockPush).toHaveBeenCalledWith('/register');
  });
});

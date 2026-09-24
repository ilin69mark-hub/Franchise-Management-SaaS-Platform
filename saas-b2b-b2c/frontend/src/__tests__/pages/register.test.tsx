import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import RegisterPage from '@/pages/register';

const mockPush = jest.fn();
jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush }),
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

const mockDispatch = jest.fn(() => Promise.resolve());
let mockState: { auth: { loading: boolean; error: string | null; isAuthenticated: boolean } };

jest.mock('react-redux', () => ({
  useSelector: (selector: (s: typeof mockState) => unknown) => selector(mockState),
  useDispatch: () => mockDispatch,
}));

jest.mock('@/store/authSlice', () => ({
  register: (payload: unknown) => ({ type: 'auth/register/pending', payload }),
}));

describe('RegisterPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockState = { auth: { loading: false, error: null, isAuthenticated: false } };
  });

  it('рендерит форму регистрации', () => {
    render(<RegisterPage />);
    expect(screen.getByText('Создать аккаунт')).toBeInTheDocument();
    expect(screen.getByText('Имя')).toBeInTheDocument();
    expect(screen.getByText('Фамилия')).toBeInTheDocument();
    expect(screen.getByText('Название компании')).toBeInTheDocument();
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('Пароль')).toBeInTheDocument();
    expect(screen.getByText('Зарегистрироваться')).toBeInTheDocument();
    expect(screen.getByText('BackButton')).toBeInTheDocument();
  });

  it('показывает ошибку регистрации', () => {
    mockState = { auth: { loading: false, error: 'Email занят', isAuthenticated: false } };
    render(<RegisterPage />);
    expect(screen.getByText('Ошибка регистрации')).toBeInTheDocument();
  });

  it('переходит на логин', () => {
    render(<RegisterPage />);
    fireEvent.click(screen.getByText('Войти'));
    expect(mockPush).toHaveBeenCalledWith('/login');
  });

  it('редиректит если уже залогинен', () => {
    mockState = { auth: { loading: false, error: null, isAuthenticated: true } };
    render(<RegisterPage />);
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('показывает loading', () => {
    mockState = { auth: { loading: true, error: null, isAuthenticated: false } };
    render(<RegisterPage />);
    expect(screen.getByText('Зарегистрироваться').closest('button')).toHaveClass('ant-btn-loading');
  });
});

describe('RegisterPage submit', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockState = { auth: { loading: false, error: null, isAuthenticated: false } };
  });

  const fillValidForm = () => {
    fireEvent.change(screen.getByLabelText('Имя'), { target: { value: 'Ivan' } });
    fireEvent.change(screen.getByLabelText('Фамилия'), { target: { value: 'Petrov' } });
    fireEvent.change(screen.getByLabelText('Название компании'), { target: { value: 'Acme' } });
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'ivan@acme.test' } });
    fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: 'long-password-12' } });
  };

  it('редиректит на /login только при успехе (RE-AUDIT)', async () => {
    mockDispatch.mockResolvedValue({ meta: { requestStatus: 'fulfilled' } });
    const { container } = render(<RegisterPage />);
    fillValidForm();
    fireEvent.submit(container.querySelector('form') as HTMLFormElement);
    await screen.findByText('Зарегистрироваться');
    expect(mockPush).toHaveBeenCalledWith('/login');
  });

  it('не редиректит при провале регистрации (RE-AUDIT)', async () => {
    mockDispatch.mockResolvedValue({ meta: { requestStatus: 'rejected' } });
    const { container } = render(<RegisterPage />);
    fillValidForm();
    fireEvent.submit(container.querySelector('form') as HTMLFormElement);
    await new Promise((r) => setTimeout(r, 50));
    expect(mockPush).not.toHaveBeenCalledWith('/login');
  });
});

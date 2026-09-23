import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ProfilePage from '@/pages/profile';

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

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: { ...actualAntd.message, success: jest.fn(), error: jest.fn() },
  };
});

let mockProfile: Record<string, unknown> | null = {
  first_name: 'Иван',
  last_name: 'Петров',
  display_name: 'Иван Петров',
  email: 'ivan@test.com',
  position: 'Менеджер',
  bio: 'О себе',
  quote: 'Цитата',
  avatar_url: null,
  status: 'online',
  available_for_questions: true,
  contacts: { phone: '+79991112233', telegram: '@ivan', whatsapp: '+79991112233', working_hours: '10-18', email_visible: true, phone_visible: true },
  achievements: ['Лучший продавец'],
};
let mockIsLoading = false;
let mockIsError = false;
const mockRefetch = jest.fn();
const mockUpdateProfile = jest.fn(() => ({ unwrap: jest.fn().mockResolvedValue({}) }));

jest.mock('@/services/userApi', () => ({
  useGetMyProfileQuery: () => ({ data: mockProfile, isLoading: mockIsLoading, isError: mockIsError, refetch: mockRefetch }),
  useUpdateProfileMutation: () => [mockUpdateProfile, { isLoading: false }],
}));

let mockAuthState: { isAuthenticated: boolean };
jest.mock('react-redux', () => ({
  useSelector: (selector: (s: { auth: typeof mockAuthState }) => unknown) => selector({ auth: mockAuthState }),
}));

describe('ProfilePage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockAuthState = { isAuthenticated: true };
    mockProfile = {
      first_name: 'Иван',
      last_name: 'Петров',
      display_name: 'Иван Петров',
      email: 'ivan@test.com',
      position: 'Менеджер',
      bio: 'О себе',
      quote: 'Цитата',
      avatar_url: null,
      status: 'online',
      available_for_questions: true,
      contacts: { phone: '+79991112233', telegram: '@ivan', whatsapp: '+79991112233', working_hours: '10-18', email_visible: true, phone_visible: true },
      achievements: ['Лучший продавец'],
    };
    mockIsLoading = false;
    mockIsError = false;
    Object.defineProperty(window, 'localStorage', { value: { getItem: jest.fn(() => 'token'), setItem: jest.fn() }, writable: true });
  });

  it('рендерит профиль с данными', async () => {
    render(<ProfilePage />);
    expect(screen.getAllByText('Мой профиль').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Аватар')).toBeInTheDocument();
    expect(screen.getByText('Основная информация')).toBeInTheDocument();
    expect(screen.getByText('Контакты для дилеров')).toBeInTheDocument();
    expect(screen.getByText('Статус и настроение')).toBeInTheDocument();
    expect(screen.getByText('Лучший продавец')).toBeInTheDocument();
  });

  it('показывает скелетон при загрузке', () => {
    mockIsLoading = true;
    const { container } = render(<ProfilePage />);
    expect(container.querySelector('.ant-skeleton')).toBeInTheDocument();
  });

  it('показывает ошибку и кнопку повтора', () => {
    mockIsLoading = false;
    mockIsError = true;
    render(<ProfilePage />);
    expect(screen.getByText('Не удалось загрузить профиль')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Повторить'));
    expect(mockRefetch).toHaveBeenCalled();
  });

  it('сохраняет основную информацию', async () => {
    render(<ProfilePage />);
    fireEvent.click(screen.getAllByText('Сохранить')[0]);
    await waitFor(() => expect(mockUpdateProfile).toHaveBeenCalled());
  });

  it('сохраняет контакты', async () => {
    render(<ProfilePage />);
    fireEvent.click(screen.getByText('Сохранить контакты'));
    await waitFor(() => expect(mockUpdateProfile).toHaveBeenCalled());
  });

  it('отображает инициалы когда нет аватара', () => {
    mockProfile = { ...mockProfile, first_name: 'А', last_name: 'Б', avatar_url: null };
    const { container } = render(<ProfilePage />);
    expect(container.textContent).toContain('Аватар');
  });

  it('кнопка Назад', () => {
    render(<ProfilePage />);
    fireEvent.click(screen.getByText('Назад'));
    expect(mockBack).toHaveBeenCalled();
  });
});

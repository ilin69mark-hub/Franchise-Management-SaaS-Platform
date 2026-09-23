import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import SalonsPage from '@/pages/salons';

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: { ...actualAntd.message, success: jest.fn(), error: jest.fn() },
  };
});

const mockSalons = [{ id: '1234567890', name: 'Салон на Ленина', address: 'ул. Ленина 1' }];
let mockIsLoading = false;
const mockCreate = jest.fn(() => ({ unwrap: jest.fn().mockResolvedValue({}) }));

jest.mock('@/services/api', () => ({
  useGetSalonsQuery: () => ({ data: mockSalons, isLoading: mockIsLoading }),
  useCreateSalonMutation: () => [mockCreate, { isLoading: false }],
}));

describe('SalonsPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockIsLoading = false;
  });

  it('рендерит заголовок и таблицу', () => {
    render(<SalonsPage />);
    expect(screen.getByText('Управление Салонами')).toBeInTheDocument();
    expect(screen.getByText('Добавить салон')).toBeInTheDocument();
    expect(screen.getByText('Салон на Ленина')).toBeInTheDocument();
    expect(screen.getByText('ул. Ленина 1')).toBeInTheDocument();
  });

  it('показывает loading', () => {
    mockIsLoading = true;
    const { container } = render(<SalonsPage />);
    expect(container.querySelector('.ant-spin')).toBeInTheDocument();
  });

  it('открывает модалку создания', () => {
    render(<SalonsPage />);
    fireEvent.click(screen.getByText('Добавить салон'));
    expect(screen.getByText('Новый салон')).toBeInTheDocument();
    expect(screen.getByLabelText('Название салона')).toBeInTheDocument();
    expect(screen.getByLabelText('Адрес')).toBeInTheDocument();
  });

  it('создаёт салон успешно', async () => {
    mockCreate.mockReturnValue({ unwrap: jest.fn().mockResolvedValue({}) });
    render(<SalonsPage />);
    fireEvent.click(screen.getByText('Добавить салон'));
    fireEvent.change(screen.getByPlaceholderText('Например: Салон на Ленина'), { target: { value: 'Новый салон' } });
    fireEvent.click(screen.getByText('Создать'));
    await waitFor(() => expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Салон успешно создан'));
  });

  it('обрабатывает ошибку создания', async () => {
    mockCreate.mockReturnValue({ unwrap: jest.fn().mockRejectedValue(new Error('fail')) });
    render(<SalonsPage />);
    fireEvent.click(screen.getByText('Добавить салон'));
    fireEvent.change(screen.getByPlaceholderText('Например: Салон на Ленина'), { target: { value: 'Тест' } });
    fireEvent.click(screen.getByText('Создать'));
    await waitFor(() => expect(jest.requireMock('antd').message.error).toHaveBeenCalledWith('Ошибка при создании салона'));
  });

  it('отображает ID с обрезкой', () => {
    render(<SalonsPage />);
    expect(screen.getByText('12345678...')).toBeInTheDocument();
  });
});

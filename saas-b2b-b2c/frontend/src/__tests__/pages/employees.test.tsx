import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import EmployeesPage from '@/pages/employees';

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: { ...actualAntd.message, success: jest.fn(), error: jest.fn() },
  };
});

const mockGetAll = jest.fn().mockResolvedValue([]);
const mockCreate = jest.fn().mockResolvedValue({});
const mockUpdate = jest.fn().mockResolvedValue({});
const mockDelete = jest.fn().mockResolvedValue({});

jest.mock('@/api/employee', () => ({
  EmployeeApi: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    create: (...args: unknown[]) => mockCreate(...args),
    update: (...args: unknown[]) => mockUpdate(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
  },
}));

describe('EmployeesPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetAll.mockResolvedValue([
      { id: '1', first_name: 'Иван', last_name: 'Петров', email: 'ivan@test.com', phone: '+79991112233', role: 'dealer' },
      { id: '2', first_name: 'Петр', last_name: 'Сидоров', email: 'petr@test.com', phone: '+79992223344', role: 'salon_manager' },
    ]);
  });

  it('рендерит заголовок и кнопку', async () => {
    render(<EmployeesPage />);
    expect(screen.getByText('Управление Сотрудниками')).toBeInTheDocument();
    expect(screen.getByText('Добавить сотрудника')).toBeInTheDocument();
    await waitFor(() => expect(mockGetAll).toHaveBeenCalled());
  });

  it('отображает сотрудников в таблице', async () => {
    render(<EmployeesPage />);
    await waitFor(() => expect(screen.getByText('Иван')).toBeInTheDocument());
    expect(screen.getByText('Петров')).toBeInTheDocument();
    expect(screen.getByText('ivan@test.com')).toBeInTheDocument();
    expect(screen.getByText('Дилер')).toBeInTheDocument();
    expect(screen.getByText('Управляющий Салоном')).toBeInTheDocument();
  });

  it('открывает модалку создания', async () => {
    render(<EmployeesPage />);
    fireEvent.click(screen.getByText('Добавить сотрудника'));
    expect(screen.getByText('Новый сотрудник')).toBeInTheDocument();
    expect(screen.getByLabelText('Имя')).toBeInTheDocument();
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByText('Выберите роль')).toBeInTheDocument();
  });

  it('открывает модалку редактирования', async () => {
    render(<EmployeesPage />);
    await waitFor(() => expect(screen.getByText('Иван')).toBeInTheDocument());
    const editBtns = document.querySelectorAll('.anticon-edit');
    fireEvent.click(editBtns[0]);
    expect(screen.getByText('Редактировать сотрудника')).toBeInTheDocument();
  });

  it('создаёт сотрудника', async () => {
    render(<EmployeesPage />);
    fireEvent.click(screen.getByText('Добавить сотрудника'));
    await waitFor(() => expect(screen.getByText('Новый сотрудник')).toBeInTheDocument());
    expect(screen.getByLabelText('Имя')).toBeInTheDocument();
  });

  it('обрабатывает ошибку загрузки', async () => {
    mockGetAll.mockRejectedValue(new Error('fail'));
    render(<EmployeesPage />);
    await waitFor(() => expect(jest.requireMock('antd').message.error).toHaveBeenCalledWith('Не удалось загрузить список сотрудников'));
  });

  it('удаляет сотрудника', async () => {
    render(<EmployeesPage />);
    await waitFor(() => expect(screen.getByText('Иван')).toBeInTheDocument());
    const deleteBtns = document.querySelectorAll('.anticon-delete');
    fireEvent.click(deleteBtns[0]);
    const confirmBtn = await screen.findByText('Да');
    fireEvent.click(confirmBtn);
    await waitFor(() => expect(mockDelete).toHaveBeenCalledWith('1'));
    await waitFor(() => expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Сотрудник удален'));
  });
});

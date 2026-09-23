import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ChecklistsPage from '@/pages/checklists';

jest.mock('next/router', () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock('next/head', () => ({ __esModule: true, default: () => null }));

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: {
      ...actualAntd.message,
      success: jest.fn(),
      error: jest.fn(),
    },
  };
});

jest.mock('@/components/Dashboard/Header', () => {
  const Header: React.FC = () => <div>Header-page</div>;
  return { __esModule: true, default: Header };
});

jest.mock('@/store/checklistSlice', () => ({
  fetchChecklists: () => ({ type: 'checklist/fetchChecklists/pending' }),
  createChecklist: () => ({ type: 'checklist/createChecklist/pending' }),
  updateChecklist: () => ({ type: 'checklist/updateChecklist/pending' }),
  deleteChecklist: (id: string) => ({ type: `checklist/deleteChecklist/pending:${id}` }),
  completeChecklist: (id: string) => ({ type: `checklist/completeChecklist/pending:${id}` }),
}));

const mockPush = jest.fn();
const mockDispatch = jest.fn();

let mockState: {
  checklist: { items: Array<Record<string, unknown>>; loading: boolean; error?: string };
  auth: { isAuthenticated: boolean; user: Record<string, unknown> | null };
};

jest.mock('react-redux', () => ({
  useSelector: (selector: (s: typeof mockState) => unknown) => selector(mockState),
  useDispatch: () => mockDispatch,
}));

const result = (value: unknown) => ({
  unwrap: jest.fn().mockResolvedValue(value),
  catch: jest.fn(() => Promise.resolve()),
});

const items = [
  {
    id: 'c1',
    title: 'Настроить баннер',
    description: 'Новый баннер',
    assigned_to: 'u1',
    start_date: '2026-09-01T10:00:00',
    end_date: '2026-09-10T18:00:00',
    status: 'in_progress',
  },
  { id: 'c2', title: 'Аудит магазина', assigned_to: 'u1', status: 'completed' },
  { id: 'c3', title: 'Обновить прайс', assigned_to: null, status: 'pending' },
];

describe('ChecklistsPage', () => {
  jest.setTimeout(30000);

  beforeEach(() => {
    jest.clearAllMocks();
    mockPush.mockImplementation(() => Promise.resolve());
    mockDispatch.mockImplementation(() => result(undefined));
    mockState = {
      checklist: { items, loading: false },
      auth: {
        isAuthenticated: true,
        user: { id: 'u1', tenant_id: 't1', first_name: 'Иван', last_name: 'Петров', email: 'a@x.ru' },
      },
    };
  });

  it('отображает чек-листы со статусами и исполнителями', () => {
    render(<ChecklistsPage />);
    expect(screen.getByText('Чек-листы / Задачи')).toBeInTheDocument();
    expect(screen.getByText('Header-page')).toBeInTheDocument();
    expect(screen.getByText('Создать')).toBeInTheDocument();
    expect(screen.getByText('Настроить баннер')).toBeInTheDocument();
    expect(screen.getAllByText('Иван Петров').length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText('С: 01.09.2026 10:00')).toBeInTheDocument();
    expect(screen.getByText('До: 10.09.2026 18:00')).toBeInTheDocument();
    expect(screen.getByText('В процессе')).toBeInTheDocument();
    expect(screen.getByText('Выполнен')).toBeInTheDocument();
    expect(screen.getByText('Ожидает')).toBeInTheDocument();
    expect(screen.getAllByText('Изменить').length).toBe(3);
    expect(screen.getAllByText('Завершить').length).toBe(2);
  });

  it('загружает чек-листы при авторизации', () => {
    render(<ChecklistsPage />);
    expect(mockDispatch).toHaveBeenCalled();
  });

  it('редиректит на login без авторизации', () => {
    mockState = {
      checklist: { items: [], loading: false },
      auth: { isAuthenticated: false, user: null },
    };
    render(<ChecklistsPage />);
    expect(mockPush).toHaveBeenCalledWith('/login');
  });

  it('показывает ошибку загрузки', () => {
    mockState = {
      checklist: { items: [], loading: false, error: 'Сеть недоступна' },
      auth: { isAuthenticated: true, user: { id: 'u1' } },
    };
    render(<ChecklistsPage />);
    expect(screen.getByText('Сеть недоступна')).toBeInTheDocument();
  });

  it('создаёт задачу через модалку', async () => {
    render(<ChecklistsPage />);
    fireEvent.click(screen.getByText('Создать'));
    const titleInput = await screen.findByLabelText('Название');
    fireEvent.change(titleInput, { target: { value: 'Новая задача' } });
    const okButton = (await screen.findAllByRole('button', { name: /^OK$/ })).pop()!;
    fireEvent.click(okButton);
    await waitFor(
      () => expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Создано'),
      { timeout: 5000 },
    );
  });

  it('открывает редактирование с предзаполненной формой', async () => {
    render(<ChecklistsPage />);
    fireEvent.click(screen.getAllByText('Изменить')[0]);
    await screen.findByText('Редактировать');
    expect(await screen.findByDisplayValue('Настроить баннер')).toBeInTheDocument();
  });

  it('завершает задачу', async () => {
    render(<ChecklistsPage />);
    fireEvent.click(screen.getAllByText('Завершить')[0]);
    await waitFor(
      () => expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Завершено'),
      { timeout: 5000 },
    );
  });

  it('удаляет задачу через подтверждение', async () => {
    render(<ChecklistsPage />);
    fireEvent.click(screen.getAllByText('Удалить')[0]);
    fireEvent.click(await screen.findByText('Да'));
    await waitFor(
      () => expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Удалено'),
      { timeout: 5000 },
    );
  });
});
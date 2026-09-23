import { EmployeeApi } from '@/api/employee';
import apiClient from '@/api/axiosClient';

jest.mock('@/api/axiosClient', () => ({
  __esModule: true,
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

const mockClient = apiClient as jest.Mocked<typeof apiClient>;

describe('EmployeeApi', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('getAll получает список сотрудников', async () => {
    mockClient.get.mockResolvedValue({ data: [{ id: 'e1' }] });
    await expect(EmployeeApi.getAll()).resolves.toEqual([{ id: 'e1' }]);
    expect(mockClient.get).toHaveBeenCalledWith('/users');
  });

  it('create отправляет данные с password', async () => {
    mockClient.post.mockResolvedValue({ data: { id: 'e2' } });
    await expect(
      EmployeeApi.create({ email: 'a@x.ru', password: 'secret1' }),
    ).resolves.toEqual({ id: 'e2' });
    expect(mockClient.post).toHaveBeenCalledWith('/users', { email: 'a@x.ru', password: 'secret1' });
  });

  it('update обращается к /users/:id', async () => {
    mockClient.put.mockResolvedValue({ data: { id: 'e1' } });
    await expect(EmployeeApi.update('e1', { first_name: 'Иван' })).resolves.toEqual({ id: 'e1' });
    expect(mockClient.put).toHaveBeenCalledWith('/users/e1', { first_name: 'Иван' });
  });

  it('delete вызывает DELETE /users/:id', async () => {
    mockClient.delete.mockResolvedValue(undefined);
    await EmployeeApi.delete('e1');
    expect(mockClient.delete).toHaveBeenCalledWith('/users/e1');
  });
});
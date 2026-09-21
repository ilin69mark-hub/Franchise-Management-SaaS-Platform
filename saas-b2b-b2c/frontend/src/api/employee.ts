import apiClient from './axiosClient'; // Используем твой настроенный инстанс
import { Employee } from '@/types';

export const EmployeeApi = {
  // Получить список сотрудников моей сети
  getAll: async (): Promise<Employee[]> => {
    const response = await apiClient.get('/users');
    return response.data;
  },

  // Создать сотрудника
  create: async (data: Partial<Employee> & { password: string }) => {
    const response = await apiClient.post('/users', data);
    return response.data;
  },

  // Обновить данные сотрудника
  update: async (id: string, data: Partial<Employee>) => {
    const response = await apiClient.put(`/users/${id}`, data);
    return response.data;
  },

  // Удалить сотрудника
  delete: async (id: string) => {
    await apiClient.delete(`/users/${id}`);
  },
};

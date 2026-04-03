import axiosClient from './axiosClient';

export const kpiApi = {
  // Получить статистику для менеджера
  getMyStats: () => axiosClient.get('/stats/my'),
  
  // Получить статистику для салона (Дилер)
  getSalonStats: (salonId: string) => axiosClient.get(`/stats/salon?salon_id=${salonId}`),

  // Задать план (Дилер)
  setGoal: (data: any) => axiosClient.post('/goals', data),

  // Расписание
  getSchedule: (date: string) => axiosClient.get(`/schedule?date=${date}`),
  createEvent: (data: any) => axiosClient.post('/schedule', data),
  updateEventStatus: (id: string, status: string) => axiosClient.put(`/schedule/${id}/status`, { status }),
};
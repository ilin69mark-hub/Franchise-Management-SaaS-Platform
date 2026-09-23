import { kpiApi } from '@/api/kpi';
import apiClient from '@/api/axiosClient';

jest.mock('@/api/axiosClient', () => ({
  __esModule: true,
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn() },
}));

const mockClient = apiClient as jest.Mocked<typeof apiClient>;

describe('kpiApi', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('getMyStats получает статистику менеджера', async () => {
    mockClient.get.mockResolvedValue({ data: { plan: 50 } });
    await expect(kpiApi.getMyStats()).resolves.toEqual({ data: { plan: 50 } });
    expect(mockClient.get).toHaveBeenCalledWith('/stats/my');
  });

  it('getSalonStats формирует URL с salon_id', async () => {
    mockClient.get.mockResolvedValue({ data: {} });
    await kpiApi.getSalonStats('s1');
    expect(mockClient.get).toHaveBeenCalledWith('/stats/salon?salon_id=s1');
  });

  it('setGoal отправляет POST /goals', async () => {
    mockClient.post.mockResolvedValue({ data: { ok: true } });
    await kpiApi.setGoal({ plan: 1 });
    expect(mockClient.post).toHaveBeenCalledWith('/goals', { plan: 1 });
  });

  it('getSchedule запрашивает расписание по дате', async () => {
    mockClient.get.mockResolvedValue({ data: [] });
    await kpiApi.getSchedule('2026-09-01');
    expect(mockClient.get).toHaveBeenCalledWith('/schedule?date=2026-09-01');
  });

  it('createEvent создаёт событие', async () => {
    mockClient.post.mockResolvedValue({ data: { id: 'ev1' } });
    await kpiApi.createEvent({ title: 'Встреча' });
    expect(mockClient.post).toHaveBeenCalledWith('/schedule', { title: 'Встреча' });
  });

  it('updateEventStatus обновляет статус события', async () => {
    mockClient.put.mockResolvedValue({ data: {} });
    await kpiApi.updateEventStatus('ev1', 'done');
    expect(mockClient.put).toHaveBeenCalledWith('/schedule/ev1/status', { status: 'done' });
  });
});
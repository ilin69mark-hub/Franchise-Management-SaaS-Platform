import { getLeads, createLead, updateLeadStatus, addLeadActivity, getLeadDetails } from '@/api/leads';
import apiClient from '@/api/axiosClient';

jest.mock('@/api/axiosClient', () => ({
  __esModule: true,
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn() },
}));

const mockClient = apiClient as jest.Mocked<typeof apiClient>;

describe('leads API', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('getLeads возвращает список лидов', async () => {
    mockClient.get.mockResolvedValue({ data: [{ id: 'l1' }] });
    await expect(getLeads()).resolves.toEqual([{ id: 'l1' }]);
    expect(mockClient.get).toHaveBeenCalledWith('/leads');
  });

  it('createLead создаёт лида через POST', async () => {
    mockClient.post.mockResolvedValue({ data: { id: 'l2' } });
    await expect(createLead({ name: 'Анна' })).resolves.toEqual({ id: 'l2' });
    expect(mockClient.post).toHaveBeenCalledWith('/leads', { name: 'Анна' });
  });

  it('updateLeadStatus обновляет статус лида', async () => {
    mockClient.put.mockResolvedValue({ data: {} });
    await updateLeadStatus('l1', 'converted');
    expect(mockClient.put).toHaveBeenCalledWith('/leads/l1/status', { status: 'converted' });
  });

  it('addLeadActivity добавляет активность', async () => {
    mockClient.post.mockResolvedValue({ data: {} });
    await addLeadActivity('l1', 'call', 'Позвонили');
    expect(mockClient.post).toHaveBeenCalledWith('/leads/l1/activities', { type: 'call', description: 'Позвонили' });
  });

  it('getLeadDetails возвращает лида с активностями', async () => {
    mockClient.get.mockResolvedValue({ data: { lead: { id: 'l1' }, activities: [{ id: 'a1' }] } });
    await expect(getLeadDetails('l1')).resolves.toEqual({ lead: { id: 'l1' }, activities: [{ id: 'a1' }] });
    expect(mockClient.get).toHaveBeenCalledWith('/leads/l1');
  });
});
import { configureStore } from '@reduxjs/toolkit';
import { AxiosError } from 'axios';
import checklistReducer, {
  fetchChecklists,
  fetchChecklistById,
  createChecklist,
  updateChecklist,
  deleteChecklist,
  completeChecklist,
  setCurrentChecklist,
  resetCurrentChecklist,
  clearError,
  Checklist,
} from '@/store/checklistSlice';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('checklistSlice', () => {
  let store: ReturnType<typeof configureStore>;

  const createMockChecklist = (overrides: Partial<Checklist> = {}): Checklist => ({
    id: 'checklist-1',
    title: 'Test Checklist',
    description: 'Test description',
    user_id: 'user-1',
    tenant_id: 'tenant-1',
    status: 'pending',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    tasks: [],
    kpi_score: 0,
    ...overrides,
  });

  beforeEach(() => {
    store = configureStore({
      reducer: {
        checklist: checklistReducer,
      },
    });
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('has items as empty array', () => {
      expect(store.getState().checklist.items).toEqual([]);
    });

    it('has currentChecklist as null', () => {
      expect(store.getState().checklist.currentChecklist).toBeNull();
    });

    it('has loading as false', () => {
      expect(store.getState().checklist.loading).toBe(false);
    });

    it('has error as null', () => {
      expect(store.getState().checklist.error).toBeNull();
    });
  });

  describe('setCurrentChecklist action', () => {
    it('sets currentChecklist', () => {
      const checklist = createMockChecklist({ id: 'checklist-2', title: 'Current' });
      store.dispatch(setCurrentChecklist(checklist));
      expect(store.getState().checklist.currentChecklist).toEqual(checklist);
    });
  });

  describe('resetCurrentChecklist action', () => {
    it('resets currentChecklist to null', () => {
      store.dispatch(setCurrentChecklist(createMockChecklist()));
      store.dispatch(resetCurrentChecklist());
      expect(store.getState().checklist.currentChecklist).toBeNull();
    });
  });

  describe('clearError action', () => {
    it('clears error to null', () => {
      store.dispatch(clearError());
      expect(store.getState().checklist.error).toBeNull();
    });

    it('clears existing error', () => {
      store.dispatch(clearError());
      store.dispatch(clearError());
      expect(store.getState().checklist.error).toBeNull();
    });
  });

  describe('fetchChecklists thunk', () => {
    const mockChecklists: Checklist[] = [
      createMockChecklist({ id: 'cl-1', title: 'Checklist 1' }),
      createMockChecklist({ id: 'cl-2', title: 'Checklist 2' }),
    ];

    it('sets loading = true on pending', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));

      store.dispatch(fetchChecklists());

      expect(store.getState().checklist.loading).toBe(true);
      expect(store.getState().checklist.error).toBeNull();
    });

    it('sets items array on fulfilled', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockChecklists });

      await store.dispatch(fetchChecklists());

      expect(store.getState().checklist.loading).toBe(false);
      expect(store.getState().checklist.items).toHaveLength(2);
      expect(store.getState().checklist.items[0].title).toBe('Checklist 1');
    });

    it('handles empty response', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] });

      await store.dispatch(fetchChecklists());

      expect(store.getState().checklist.items).toEqual([]);
    });

it('sets error on rejected', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Request failed'));

      await store.dispatch(fetchChecklists());

      expect(store.getState().checklist.error).toBe('Ошибка получения чек-листов');
    });

    it('handles error with message in response', async () => {
      const error = {
        isAxiosError: true,
        response: { data: { message: 'Custom error message' } },
        message: 'Request failed',
      };
      mockApiClient.get.mockRejectedValue(error);

      await store.dispatch(fetchChecklists());

      expect(store.getState().checklist.error).toBe('Custom error message');
    });

    it('sets custom error message from response', async () => {
      const error = new AxiosError('Request failed', '400', {}, {}, { data: { message: 'Custom error message' } });
      mockApiClient.get.mockRejectedValue(error);

      await store.dispatch(fetchChecklists());

      expect(store.getState().checklist.error).toBe('Custom error message');
    });
  });

  describe('fetchChecklistById thunk', () => {
    const mockChecklist = createMockChecklist({ id: 'cl-1', title: 'Single Checklist' });

    it('sets loading = true on pending', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));

      store.dispatch(fetchChecklistById('cl-1'));

      expect(store.getState().checklist.loading).toBe(true);
    });

    it('sets currentChecklist on fulfilled', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockChecklist });

      await store.dispatch(fetchChecklistById('cl-1'));

      expect(store.getState().checklist.loading).toBe(false);
      expect(store.getState().checklist.currentChecklist).toEqual(mockChecklist);
    });

    it('updates item in items array if exists', async () => {
      const storeWithItems = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: [createMockChecklist({ id: 'cl-existing' })],
            currentChecklist: null,
            loading: false,
            error: null,
          },
        } as any,
      });

      const updatedChecklist = createMockChecklist({ id: 'cl-existing', title: 'Updated Title' });
      mockApiClient.get.mockResolvedValue({ data: updatedChecklist });

      await storeWithItems.dispatch(fetchChecklistById('cl-existing'));

      const item = storeWithItems.getState().checklist.items.find(i => i.id === 'cl-existing');
      expect(item?.title).toBe('Updated Title');
    });

    it('sets error on rejected (404)', async () => {
      mockApiClient.get.mockRejectedValue({
        isAxiosError: true,
        response: { status: 404, data: { message: 'Checklist not found' } },
      });

      await store.dispatch(fetchChecklistById('nonexistent'));

      expect(store.getState().checklist.error).toBe('Checklist not found');
    });
  });

  describe('createChecklist thunk', () => {
    const newChecklistData = {
      title: 'New Checklist',
      description: 'New Description',
      tasks: [
        { id: 'task-1', title: 'Task 1', status: 'pending' as const, order: 1, created_at: '', updated_at: '' },
      ],
    };

    it('sets loading = true on pending', async () => {
      mockApiClient.post.mockImplementation(() => new Promise(() => {}));

      store.dispatch(createChecklist(newChecklistData));

      expect(store.getState().checklist.loading).toBe(true);
    });

    it('adds new checklist to beginning of items on fulfilled', async () => {
      const createdChecklist = createMockChecklist({ id: 'new-cl-1', title: 'New Checklist' });
      mockApiClient.post.mockResolvedValue({ data: createdChecklist });

      await store.dispatch(createChecklist(newChecklistData));

      expect(store.getState().checklist.loading).toBe(false);
      expect(store.getState().checklist.items).toHaveLength(1);
      expect(store.getState().checklist.items[0].title).toBe('New Checklist');
    });

    it('adds multiple checklists preserving order', async () => {
      const first = createMockChecklist({ id: 'cl-1', title: 'First' });
      const second = createMockChecklist({ id: 'cl-2', title: 'Second' });

      mockApiClient.post.mockResolvedValue({ data: first });
      await store.dispatch(createChecklist(newChecklistData));
      expect(store.getState().checklist.items).toHaveLength(1);
      expect(store.getState().checklist.items[0].title).toBe('First');

      mockApiClient.post.mockResolvedValue({ data: second });
      await store.dispatch(createChecklist(newChecklistData));
      expect(store.getState().checklist.items).toHaveLength(2);
      expect(store.getState().checklist.items[0].title).toBe('Second');
      expect(store.getState().checklist.items[1].title).toBe('First');
    });

    it('sets error on rejected (validation error)', async () => {
      mockApiClient.post.mockRejectedValue({
        isAxiosError: true,
        response: { data: { message: 'Title is required' } },
      });

      await store.dispatch(createChecklist({ title: '' }));

      expect(store.getState().checklist.loading).toBe(false);
      expect(store.getState().checklist.error).toBe('Title is required');
    });

    it('sets default error when no custom message', async () => {
      mockApiClient.post.mockRejectedValue(new Error('Server error'));

      await store.dispatch(createChecklist(newChecklistData));

      expect(store.getState().checklist.error).toBe('Ошибка создания чек-листа');
    });
  });

  describe('updateChecklist thunk', () => {
    const existingChecklist = createMockChecklist({ id: 'cl-1', title: 'Original Title' });

    it('sets loading = true on pending', async () => {
      mockApiClient.put.mockImplementation(() => new Promise(() => {}));

      store.dispatch(updateChecklist({ id: 'cl-1', title: 'Updated' }));

      expect(store.getState().checklist.loading).toBe(true);
    });

    it('updates item in items array on fulfilled', async () => {
      const storeWithItems = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: [existingChecklist],
            currentChecklist: existingChecklist,
            loading: false,
            error: null,
          },
        } as any,
      });

      const updatedChecklist = { ...existingChecklist, title: 'Updated Title' };
      mockApiClient.put.mockResolvedValue({ data: updatedChecklist });

      await storeWithItems.dispatch(updateChecklist({ id: 'cl-1', title: 'Updated Title' }));

      const item = storeWithItems.getState().checklist.items.find(i => i.id === 'cl-1');
      expect(item?.title).toBe('Updated Title');
    });

    it('updates currentChecklist if it is the same item', async () => {
      store.dispatch(setCurrentChecklist(existingChecklist));
      const updatedChecklist = { ...existingChecklist, title: 'Updated Title' };
      mockApiClient.put.mockResolvedValue({ data: updatedChecklist });

      await store.dispatch(updateChecklist({ id: 'cl-1', title: 'Updated Title' }));

      expect(store.getState().checklist.currentChecklist?.title).toBe('Updated Title');
    });

    it('does not update currentChecklist if different item', async () => {
      store.dispatch(setCurrentChecklist(createMockChecklist({ id: 'other-cl' })));
      const updatedChecklist = { ...existingChecklist, title: 'Updated' };
      mockApiClient.put.mockResolvedValue({ data: updatedChecklist });

      await store.dispatch(updateChecklist({ id: 'cl-1', title: 'Updated' }));

      expect(store.getState().checklist.currentChecklist?.id).toBe('other-cl');
    });

    it('sets error on rejected', async () => {
      mockApiClient.put.mockRejectedValue({
        isAxiosError: true,
        response: { data: { message: 'Permission denied' } },
      });

      await store.dispatch(updateChecklist({ id: 'cl-1', title: 'Updated' }));

      expect(store.getState().checklist.error).toBe('Permission denied');
    });
  });

  describe('deleteChecklist thunk', () => {
    const existingChecklists = [
      createMockChecklist({ id: 'cl-1', title: 'To Delete' }),
      createMockChecklist({ id: 'cl-2', title: 'To Keep' }),
    ];

    it('sets loading = true on pending', async () => {
      mockApiClient.delete.mockImplementation(() => new Promise(() => {}));

      store.dispatch(deleteChecklist('cl-1'));

      expect(store.getState().checklist.loading).toBe(true);
    });

    it('removes item from items array on fulfilled', async () => {
      const storeWithItems = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: existingChecklists,
            currentChecklist: existingChecklists[0],
            loading: false,
            error: null,
          },
        } as any,
      });

      mockApiClient.delete.mockResolvedValue({});

      await storeWithItems.dispatch(deleteChecklist('cl-1'));

      expect(storeWithItems.getState().checklist.items).toHaveLength(1);
      expect(storeWithItems.getState().checklist.items[0].id).toBe('cl-2');
    });

    it('clears currentChecklist if deleted item was current', async () => {
      const storeWithItems = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: existingChecklists,
            currentChecklist: existingChecklists[0],
            loading: false,
            error: null,
          },
        } as any,
      });

      mockApiClient.delete.mockResolvedValue({});

      await storeWithItems.dispatch(deleteChecklist('cl-1'));

      expect(storeWithItems.getState().checklist.currentChecklist).toBeNull();
    });

    it('does not clear currentChecklist if different item deleted', async () => {
      const storeWithCurrent = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: existingChecklists,
            currentChecklist: createMockChecklist({ id: 'cl-3' }),
            loading: false,
            error: null,
          },
        } as any,
      });

      mockApiClient.delete.mockResolvedValue({});

      await storeWithCurrent.dispatch(deleteChecklist('cl-1'));

      expect(storeWithCurrent.getState().checklist.currentChecklist?.id).toBe('cl-3');
    });

    it('sets error on rejected', async () => {
      mockApiClient.delete.mockRejectedValue({
        isAxiosError: true,
        response: { data: { message: 'Cannot delete' } },
      });

      await store.dispatch(deleteChecklist('cl-1'));

      expect(store.getState().checklist.error).toBe('Cannot delete');
    });
  });

  describe('completeChecklist thunk', () => {
    const pendingChecklist = createMockChecklist({ id: 'cl-1', status: 'pending' });

    it('sets loading = true on pending', async () => {
      mockApiClient.post.mockImplementation(() => new Promise(() => {}));

      store.dispatch(completeChecklist('cl-1'));

      expect(store.getState().checklist.loading).toBe(true);
    });

    it('updates item status to completed on fulfilled', async () => {
      const storeWithItems = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: [pendingChecklist],
            currentChecklist: null,
            loading: false,
            error: null,
          },
        } as any,
      });

      const completedChecklist = { ...pendingChecklist, status: 'completed' as const };
      mockApiClient.post.mockResolvedValue({ data: completedChecklist });

      await storeWithItems.dispatch(completeChecklist('cl-1'));

      expect(storeWithItems.getState().checklist.items[0].status).toBe('completed');
    });

    it('updates currentChecklist if it is the same item', async () => {
      const storeWithCurrent = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: [],
            currentChecklist: pendingChecklist,
            loading: false,
            error: null,
          },
        } as any,
      });

      const completedChecklist = { ...pendingChecklist, status: 'completed' as const };
      mockApiClient.post.mockResolvedValue({ data: completedChecklist });

      await storeWithCurrent.dispatch(completeChecklist('cl-1'));

      expect(storeWithCurrent.getState().checklist.currentChecklist?.status).toBe('completed');
    });

    it('sets error on rejected (already completed)', async () => {
      mockApiClient.post.mockRejectedValue({
        isAxiosError: true,
        response: { data: { message: 'Checklist already completed' } },
      });

      await store.dispatch(completeChecklist('cl-1'));

      expect(store.getState().checklist.error).toBe('Checklist already completed');
    });

    it('sets default error when no custom message', async () => {
      mockApiClient.post.mockRejectedValue(new Error('Server error'));

      await store.dispatch(completeChecklist('cl-1'));

      expect(store.getState().checklist.error).toBe('Ошибка завершения чек-листа');
    });
  });

  describe('edge cases', () => {
    it('handles non-array response in fetchChecklists', async () => {
      mockApiClient.get.mockResolvedValue({ data: 'not-an-array' });

      await store.dispatch(fetchChecklists());

      expect(store.getState().checklist.items).toEqual([]);
    });

    it('handles error without response data', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Unknown error'));

      await store.dispatch(fetchChecklists());

      expect(store.getState().checklist.error).toBe('Ошибка получения чек-листов');
    });

    it('preserves items when update fails', async () => {
      const storeWithItems = configureStore({
        reducer: { checklist: checklistReducer },
        preloadedState: {
          checklist: {
            items: [createMockChecklist({ id: 'cl-1', title: 'Test Checklist' })],
            currentChecklist: null,
            loading: false,
            error: null,
          },
        } as any,
      });

      mockApiClient.put.mockRejectedValue(new Error('Error'));

      await storeWithItems.dispatch(updateChecklist({ id: 'cl-1', title: 'New' }));

      expect(storeWithItems.getState().checklist.items[0].title).toBe('Test Checklist');
    });
  });
});
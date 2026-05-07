import { useGoalStore, Goal, GoalFormData } from '@/store/goalStore';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('goalStore', () => {
  beforeEach(() => {
    useGoalStore.setState({
      goals: [],
      activeGoal: null,
      isLoading: false,
      filter: 'all',
      sortBy: 'deadline',
    });
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('has empty goals array', () => {
      expect(useGoalStore.getState().goals).toEqual([]);
    });

    it('has activeGoal as null', () => {
      expect(useGoalStore.getState().activeGoal).toBeNull();
    });

    it('has isLoading as false', () => {
      expect(useGoalStore.getState().isLoading).toBe(false);
    });

    it('has filter as all', () => {
      expect(useGoalStore.getState().filter).toBe('all');
    });

    it('has sortBy as deadline', () => {
      expect(useGoalStore.getState().sortBy).toBe('deadline');
    });
  });

  describe('setGoals', () => {
    it('sets goals list', () => {
      const goals: Goal[] = [
        { id: 'g1', title: 'Goal 1', description: 'Desc 1', targetValue: 100, currentValue: 50, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' },
        { id: 'g2', title: 'Goal 2', description: 'Desc 2', targetValue: 200, currentValue: 100, deadline: '2026-06-30', status: 'completed' as const, category: 'marketing' as const, createdAt: '2026-01-15' },
      ];
      
      useGoalStore.getState().setGoals(goals);
      expect(useGoalStore.getState().goals).toHaveLength(2);
    });
  });

  describe('setActiveGoal', () => {
    it('sets active goal', () => {
      const goal: Goal = { id: 'g1', title: 'Active Goal', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' };
      
      useGoalStore.getState().setActiveGoal(goal);
      expect(useGoalStore.getState().activeGoal).toEqual(goal);
    });

    it('clears active goal', () => {
      useGoalStore.setState({ activeGoal: { id: 'g1', title: 'Goal', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' } });
      useGoalStore.getState().setActiveGoal(null);
      expect(useGoalStore.getState().activeGoal).toBeNull();
    });
  });

  describe('setLoading', () => {
    it('sets loading to true', () => {
      useGoalStore.getState().setLoading(true);
      expect(useGoalStore.getState().isLoading).toBe(true);
    });

    it('sets loading to false', () => {
      useGoalStore.setState({ isLoading: true });
      useGoalStore.getState().setLoading(false);
      expect(useGoalStore.getState().isLoading).toBe(false);
    });
  });

  describe('setFilter', () => {
    it('sets filter to active', () => {
      useGoalStore.getState().setFilter('active');
      expect(useGoalStore.getState().filter).toBe('active');
    });

    it('sets filter to completed', () => {
      useGoalStore.getState().setFilter('completed');
      expect(useGoalStore.getState().filter).toBe('completed');
    });

    it('sets filter to overdue', () => {
      useGoalStore.getState().setFilter('overdue');
      expect(useGoalStore.getState().filter).toBe('overdue');
    });
  });

  describe('setSortBy', () => {
    it('sets sortBy to priority', () => {
      useGoalStore.getState().setSortBy('priority');
      expect(useGoalStore.getState().sortBy).toBe('priority');
    });

    it('sets sortBy to progress', () => {
      useGoalStore.getState().setSortBy('progress');
      expect(useGoalStore.getState().sortBy).toBe('progress');
    });

    it('sets sortBy to created', () => {
      useGoalStore.getState().setSortBy('created');
      expect(useGoalStore.getState().sortBy).toBe('created');
    });
  });

  describe('fetchGoals async', () => {
    it('fetches and sets goals', async () => {
      const mockGoals: Goal[] = [
        { id: 'g1', title: 'Goal 1', description: 'Desc 1', targetValue: 100, currentValue: 50, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' },
      ];

      mockApiClient.get.mockResolvedValue({ data: mockGoals });

      await useGoalStore.getState().fetchGoals();

      expect(useGoalStore.getState().goals).toEqual(mockGoals);
      expect(useGoalStore.getState().isLoading).toBe(false);
    });

    it('handles empty goals response', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] });

      await useGoalStore.getState().fetchGoals();

      expect(useGoalStore.getState().goals).toEqual([]);
      expect(useGoalStore.getState().isLoading).toBe(false);
    });

    it('sets loading during fetch', async () => {
      let resolvePromise: (value: unknown) => void;
      mockApiClient.get.mockImplementation(() => new Promise((resolve) => {
        resolvePromise = resolve;
      }));

      const fetchPromise = useGoalStore.getState().fetchGoals();
      expect(useGoalStore.getState().isLoading).toBe(true);

      resolvePromise!({ data: [] });
      await fetchPromise;
      expect(useGoalStore.getState().isLoading).toBe(false);
    });

    it('handles API error gracefully', async () => {
      mockApiClient.get.mockRejectedValue(new Error('API Error'));

      await useGoalStore.getState().fetchGoals();

      expect(useGoalStore.getState().goals).toEqual([]);
      expect(useGoalStore.getState().isLoading).toBe(false);
    });
  });

  describe('createGoal async', () => {
    it('creates goal and adds to list', async () => {
      const formData: GoalFormData = {
        title: 'New Goal',
        description: 'New Description',
        targetValue: 150,
        deadline: '2026-12-31',
        category: 'operations',
      };

      const newGoal: Goal = { ...formData, id: 'new-id', currentValue: 0, status: 'pending' as const, createdAt: '2026-05-01' };
      mockApiClient.post.mockResolvedValue({ data: newGoal });

      const result = await useGoalStore.getState().createGoal(formData);

      expect(mockApiClient.post).toHaveBeenCalledWith('/goals', formData);
      expect(useGoalStore.getState().goals).toHaveLength(1);
      expect(useGoalStore.getState().goals[0].title).toBe('New Goal');
      expect(result).toEqual(newGoal);
    });
  });

  describe('updateGoal async', () => {
    it('updates goal in list', async () => {
      useGoalStore.setState({
        goals: [
          { id: 'g1', title: 'Old Title', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' },
        ],
      });

      const updatedGoal: Goal = { id: 'g1', title: 'New Title', description: 'Desc', targetValue: 100, currentValue: 75, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' };
      mockApiClient.put.mockResolvedValue({ data: updatedGoal });

      const result = await useGoalStore.getState().updateGoal('g1', { currentValue: 75 });

      expect(mockApiClient.put).toHaveBeenCalledWith('/goals/g1', { currentValue: 75 });
      expect(useGoalStore.getState().goals[0].title).toBe('New Title');
      expect(useGoalStore.getState().goals[0].currentValue).toBe(75);
      expect(result).toEqual(updatedGoal);
    });
  });

  describe('deleteGoal async', () => {
    it('removes goal from list', async () => {
      useGoalStore.setState({
        goals: [
          { id: 'g1', title: 'Goal 1', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' },
          { id: 'g2', title: 'Goal 2', description: 'Desc 2', targetValue: 200, currentValue: 100, deadline: '2026-06-30', status: 'completed' as const, category: 'marketing' as const, createdAt: '2026-01-15' },
        ],
      });

      mockApiClient.delete.mockResolvedValue({});

      await useGoalStore.getState().deleteGoal('g1');

      expect(mockApiClient.delete).toHaveBeenCalledWith('/goals/g1');
      expect(useGoalStore.getState().goals).toHaveLength(1);
      expect(useGoalStore.getState().goals[0].id).toBe('g2');
    });
  });

  describe('optimistic progress update', () => {
    it('updates goal progress optimistically in state', async () => {
      useGoalStore.setState({
        goals: [
          { id: 'g1', title: 'Goal 1', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' },
        ],
      });

      const updatedGoal: Goal = { id: 'g1', title: 'Goal 1', description: 'Desc', targetValue: 100, currentValue: 70, deadline: '2026-12-31', status: 'in_progress' as const, category: 'sales' as const, createdAt: '2026-01-01' };
      mockApiClient.put.mockResolvedValue({ data: updatedGoal });

      await useGoalStore.getState().updateGoal('g1', { currentValue: 70 });

      const goal = useGoalStore.getState().goals.find(g => g.id === 'g1');
      expect(goal?.currentValue).toBe(70);
    });
  });
});
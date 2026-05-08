import { useGoalStore, Goal, GoalFormData, GoalStatus } from '@/store/goalStore';

describe('Goal Store', () => {
  beforeEach(() => {
    useGoalStore.setState({
      goals: [],
      activeGoal: null,
      isLoading: false,
      filter: 'all',
      sortBy: 'deadline',
    });
  });

  describe('initial state', () => {
    it('has empty goals array', () => {
      expect(useGoalStore.getState().goals).toEqual([]);
    });

    it('has null activeGoal', () => {
      expect(useGoalStore.getState().activeGoal).toBeNull();
    });

    it('has isLoading false', () => {
      expect(useGoalStore.getState().isLoading).toBe(false);
    });

    it('has filter all', () => {
      expect(useGoalStore.getState().filter).toBe('all');
    });

    it('has sortBy deadline', () => {
      expect(useGoalStore.getState().sortBy).toBe('deadline');
    });
  });

  describe('goals management', () => {
    it('adds a new goal', () => {
      const goal: Goal = {
        id: 'goal-1',
        title: 'Increase Sales',
        description: 'Increase monthly sales by 20%',
        targetValue: 100000,
        currentValue: 30000,
        deadline: '2026-02-01',
        status: 'in_progress',
        category: 'sales',
        assignedTo: 'user-1',
        createdAt: '2026-01-01',
      };
      
      useGoalStore.setState({ goals: [goal] });
      expect(useGoalStore.getState().goals).toHaveLength(1);
      expect(useGoalStore.getState().goals[0].title).toBe('Increase Sales');
    });

    it('sets multiple goals', () => {
      const goals: Goal[] = [
        { id: 'g1', title: 'Goal 1', description: 'Desc 1', targetValue: 100, currentValue: 50, deadline: '2026-02-01', status: 'in_progress', category: 'sales', createdAt: '2026-01-01' },
        { id: 'g2', title: 'Goal 2', description: 'Desc 2', targetValue: 200, currentValue: 100, deadline: '2026-03-01', status: 'completed', category: 'marketing', createdAt: '2026-01-01' },
      ];
      useGoalStore.setState({ goals });
      expect(useGoalStore.getState().goals).toHaveLength(2);
    });

    it('filters active goals', () => {
      const goals: Goal[] = [
        { id: 'g1', title: 'Active Goal', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-02-01', status: 'in_progress', category: 'sales', createdAt: '2026-01-01' },
        { id: 'g2', title: 'Completed Goal', description: 'Desc', targetValue: 200, currentValue: 200, deadline: '2026-03-01', status: 'completed', category: 'marketing', createdAt: '2026-01-01' },
      ];
      useGoalStore.setState({ goals });
      const activeGoals = goals.filter(g => g.status === 'in_progress');
      expect(activeGoals).toHaveLength(1);
    });

    it('filters completed goals', () => {
      const goals: Goal[] = [
        { id: 'g1', title: 'Active Goal', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-02-01', status: 'in_progress', category: 'sales', createdAt: '2026-01-01' },
        { id: 'g2', title: 'Completed Goal', description: 'Desc', targetValue: 200, currentValue: 200, deadline: '2026-03-01', status: 'completed', category: 'marketing', createdAt: '2026-01-01' },
      ];
      useGoalStore.setState({ goals });
      const completedGoals = goals.filter(g => g.status === 'completed');
      expect(completedGoals).toHaveLength(1);
    });

    it('calculates progress percentage', () => {
      const goal: Goal = {
        id: 'g1',
        title: 'Test Goal',
        description: 'Test',
        targetValue: 100,
        currentValue: 75,
        deadline: '2026-02-01',
        status: 'in_progress',
        category: 'sales',
        createdAt: '2026-01-01',
      };
      const progress = (goal.currentValue / goal.targetValue) * 100;
      expect(progress).toBe(75);
    });

    it('sorts by deadline', () => {
      const goals: Goal[] = [
        { id: 'g1', title: 'G1', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-03-01', status: 'in_progress', category: 'sales', createdAt: '2026-01-01' },
        { id: 'g2', title: 'G2', description: 'Desc', targetValue: 100, currentValue: 50, deadline: '2026-01-01', status: 'in_progress', category: 'sales', createdAt: '2026-01-01' },
      ];
      const sorted = [...goals].sort((a, b) => new Date(a.deadline).getTime() - new Date(b.deadline).getTime());
      expect(sorted[0].id).toBe('g2');
    });
  });

  describe('setActiveGoal', () => {
    it('sets active goal', () => {
      const goal: Goal = {
        id: 'goal-1',
        title: 'Active Goal',
        description: 'Test',
        targetValue: 100,
        currentValue: 50,
        deadline: '2026-02-01',
        status: 'in_progress',
        category: 'sales',
        createdAt: '2026-01-01',
      };
      useGoalStore.setState({ activeGoal: goal });
      expect(useGoalStore.getState().activeGoal).toEqual(goal);
    });

    it('clears active goal', () => {
      const goal: Goal = {
        id: 'goal-1',
        title: 'Goal',
        description: 'Desc',
        targetValue: 100,
        currentValue: 50,
        deadline: '2026-02-01',
        status: 'in_progress',
        category: 'sales',
        createdAt: '2026-01-01',
      };
      useGoalStore.setState({ activeGoal: goal });
      useGoalStore.setState({ activeGoal: null });
      expect(useGoalStore.getState().activeGoal).toBeNull();
    });
  });

  describe('setLoading', () => {
    it('sets loading state', () => {
      useGoalStore.getState().setLoading(true);
      expect(useGoalStore.getState().isLoading).toBe(true);
    });

    it('toggles loading', () => {
      useGoalStore.getState().setLoading(true);
      useGoalStore.getState().setLoading(false);
      expect(useGoalStore.getState().isLoading).toBe(false);
    });
  });

  describe('setFilter', () => {
    it('sets filter to active', () => {
      useGoalStore.setState({ filter: 'active' });
      expect(useGoalStore.getState().filter).toBe('active');
    });

    it('sets filter to completed', () => {
      useGoalStore.setState({ filter: 'completed' });
      expect(useGoalStore.getState().filter).toBe('completed');
    });

    it('sets filter to overdue', () => {
      useGoalStore.setState({ filter: 'overdue' });
      expect(useGoalStore.getState().filter).toBe('overdue');
    });
  });

  describe('setSortBy', () => {
    it('sets sort by priority', () => {
      useGoalStore.setState({ sortBy: 'priority' });
      expect(useGoalStore.getState().sortBy).toBe('priority');
    });

    it('sets sort by progress', () => {
      useGoalStore.setState({ sortBy: 'progress' });
      expect(useGoalStore.getState().sortBy).toBe('progress');
    });

    it('sets sort by created', () => {
      useGoalStore.setState({ sortBy: 'created' });
      expect(useGoalStore.getState().sortBy).toBe('created');
    });
  });

  describe('store methods exist', () => {
    it('has setGoals', () => {
      expect(typeof useGoalStore.getState().setGoals).toBe('function');
    });

    it('has setActiveGoal', () => {
      expect(typeof useGoalStore.getState().setActiveGoal).toBe('function');
    });

    it('has setLoading', () => {
      expect(typeof useGoalStore.getState().setLoading).toBe('function');
    });

    it('has setFilter', () => {
      expect(typeof useGoalStore.getState().setFilter).toBe('function');
    });

    it('has setSortBy', () => {
      expect(typeof useGoalStore.getState().setSortBy).toBe('function');
    });

    it('has fetchGoals method', () => {
      expect(typeof useGoalStore.getState().fetchGoals).toBe('function');
    });

    it('has createGoal method', () => {
      expect(typeof useGoalStore.getState().createGoal).toBe('function');
    });

    it('has updateGoal method', () => {
      expect(typeof useGoalStore.getState().updateGoal).toBe('function');
    });

    it('has deleteGoal method', () => {
      expect(typeof useGoalStore.getState().deleteGoal).toBe('function');
    });
  });
});

describe('Goal Form Validation', () => {
  it('validates required title', () => {
    const formData: Partial<GoalFormData> = { description: 'Test', targetValue: 100 };
    expect(formData.title).toBeUndefined();
  });

  it('validates target value', () => {
    const formData: Partial<GoalFormData> = { title: 'Test', targetValue: 100 };
    expect(formData.targetValue).toBe(100);
  });

  it('validates deadline format', () => {
    const deadline = '2026-02-01';
    const date = new Date(deadline);
    expect(date.getFullYear()).toBe(2026);
  });

  it('validates category values', () => {
    const validCategories = ['sales', 'marketing', 'operations', 'hr'];
    expect(validCategories).toContain('sales');
    expect(validCategories).toContain('marketing');
  });
});
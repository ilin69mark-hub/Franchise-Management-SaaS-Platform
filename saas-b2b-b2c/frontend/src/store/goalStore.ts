import { create } from 'zustand';
import apiClient from '@/api/axiosClient';

export type GoalStatus = 'pending' | 'in_progress' | 'completed' | 'overdue';
export type GoalCategory = 'sales' | 'marketing' | 'operations' | 'hr';

export interface Goal {
  id: string;
  title: string;
  description: string;
  targetValue: number;
  currentValue: number;
  deadline: string;
  status: GoalStatus;
  category: GoalCategory;
  assignedTo?: string;
  createdAt: string;
}

export interface GoalFormData {
  title: string;
  description: string;
  targetValue: number;
  deadline: string;
  category: GoalCategory;
  assignedTo?: string;
}

interface GoalState {
  goals: Goal[];
  activeGoal: Goal | null;
  isLoading: boolean;
  filter: 'all' | 'active' | 'completed' | 'overdue';
  sortBy: 'deadline' | 'priority' | 'progress' | 'created';
  setGoals: (goals: Goal[]) => void;
  setActiveGoal: (goal: Goal | null) => void;
  setLoading: (loading: boolean) => void;
  setFilter: (filter: GoalState['filter']) => void;
  setSortBy: (sortBy: GoalState['sortBy']) => void;
  fetchGoals: () => Promise<void>;
  createGoal: (data: GoalFormData) => Promise<Goal>;
  updateGoal: (id: string, data: Partial<Goal>) => Promise<Goal>;
  deleteGoal: (id: string) => Promise<void>;
}

export const useGoalStore = create<GoalState>((set, get) => ({
  goals: [],
  activeGoal: null,
  isLoading: false,
  filter: 'all',
  sortBy: 'deadline',

  setGoals: (goals) => set({ goals }),
  setActiveGoal: (goal) => set({ activeGoal: goal }),
  setLoading: (loading) => set({ isLoading: loading }),
  setFilter: (filter) => set({ filter }),
  setSortBy: (sortBy) => set({ sortBy }),

  fetchGoals: async () => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get('/goals');
      set({ goals: res.data });
    } catch (error) {
      console.error('Failed to fetch goals:', error);
    } finally {
      set({ isLoading: false });
    }
  },

  createGoal: async (data) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.post('/goals', data);
      const newGoal = res.data;
      set({ goals: [...get().goals, newGoal] });
      return newGoal;
    } finally {
      set({ isLoading: false });
    }
  },

  updateGoal: async (id, data) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.put(`/goals/${id}`, data);
      const updated = res.data;
      set({
        goals: get().goals.map(g => g.id === id ? updated : g),
      });
      return updated;
    } finally {
      set({ isLoading: false });
    }
  },

  deleteGoal: async (id) => {
    set({ isLoading: true });
    try {
      await apiClient.delete(`/goals/${id}`);
      set({ goals: get().goals.filter(g => g.id !== id) });
    } finally {
      set({ isLoading: false });
    }
  },
}));
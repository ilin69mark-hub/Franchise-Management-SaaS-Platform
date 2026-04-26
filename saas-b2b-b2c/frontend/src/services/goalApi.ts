import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Goal } from '@/types';

export const goalApi = createApi({
  reducerPath: 'goalApi',
  baseQuery: fetchBaseQuery({
    // Берём базовый URL из переменной окружения (NEXT_PUBLIC_API_URL),
    // fallback – обычный localhost:8080 (для локального запуска без Docker)
    baseUrl: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1`,
    prepareHeaders: (headers) => {
      const token = typeof window !== 'undefined' ? localStorage.getItem('accessToken') : null;
      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  tagTypes: ['Goal'],
  endpoints: (builder) => ({
    // Личный план текущего пользователя (по дате)
    getMyGoal: builder.query<Goal, string>({
      query: (date) => `goals/by-date/${date}`,
      providesTags: (result) => (result ? [{ type: 'Goal', id: result.id }] : []),
    }),

    // Список целей, которые видит пользователь
    getVisibleGoals: builder.query<Goal[], void>({
      query: () => 'goals/visible',
      providesTags: (result) =>
        result ? result.map((g) => ({ type: 'Goal' as const, id: g.id })) : [],
    }),

    // Создать/обновить цель
    setGoal: builder.mutation<Goal, Omit<Goal, 'id' | 'created_at' | 'updated_at'>>({
      query: (body) => ({
        url: 'goals',
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Goal', id: 'LIST' }],
    }),

    // Удалить цель
    deleteGoal: builder.mutation<void, string>({
      query: (id) => ({
        url: `goals/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: [{ type: 'Goal', id: 'LIST' }],
    }),
  }),
});

export const {
  useGetMyGoalQuery,
  useGetVisibleGoalsQuery,
  useSetGoalMutation,
  useDeleteGoalMutation,
} = goalApi;

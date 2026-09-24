import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Goal } from '@/types';
import { getCsrfToken, CSRF_HEADER } from '@/utils/csrf';

export const goalApi = createApi({
  reducerPath: 'goalApi',
  baseQuery: fetchBaseQuery({
    baseUrl: `${(process.env.NEXT_PUBLIC_API_URL || (process.env.NODE_ENV === 'production' ? (()=>{throw new Error('NEXT_PUBLIC_API_URL must be set in production')})() : 'http://localhost:8080'))}/api/v1`,
    prepareHeaders: (headers) => {
      // F7: сессия в httpOnly cookie; CSRF double-submit для мутаций.
      const csrf = getCsrfToken();
      if (csrf) {
        headers.set(CSRF_HEADER, csrf);
      }
      return headers;
    },
  }),
  tagTypes: ['Goal'],
  endpoints: (builder) => ({
    getMyGoal: builder.query<Goal, string>({
      query: (date) => `goals/by-date/${date}`,
      providesTags: (result) => (result ? [{ type: 'Goal', id: result.id }] : []),
    }),

    getVisibleGoals: builder.query<Goal[], void>({
      query: () => 'goals/visible',
      providesTags: (result) =>
        result ? result.map((g) => ({ type: 'Goal' as const, id: g.id })) : [],
    }),

    setGoal: builder.mutation<Goal, Omit<Goal, 'id' | 'created_at' | 'updated_at'>>({
      query: (body) => ({
        url: 'goals',
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Goal', id: 'LIST' }],
    }),

    updateGoal: builder.mutation<Goal, { id: string; data: Partial<Goal> }>({
      query: ({ id, data }) => ({
        url: `goals/${id}`,
        method: 'PUT',
        body: data,
      }),
      invalidatesTags: [{ type: 'Goal', id: 'LIST' }],
    }),

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
  useUpdateGoalMutation,
  useDeleteGoalMutation,
} = goalApi;

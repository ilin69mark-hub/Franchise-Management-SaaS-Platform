// src/services/planApi.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Plan } from '@/types';

/* -----------------------------------------------------------------
   1️⃣ Тип аргумента для PATCH /plans/:id
   id – обязателен,
   остальные поля – опциональны (кроме id, created_at, updated_at)
----------------------------------------------------------------- */
type UpdatePlanArg = { id: string } & Partial<
  Omit<Plan, 'id' | 'created_at' | 'updated_at'>
>;

/* -----------------------------------------------------------------
   2️⃣ RTK‑Query‑сервис для тарифных планов
----------------------------------------------------------------- */
export const planApi = createApi({
  reducerPath: 'planApi',
  baseQuery: fetchBaseQuery({
    baseUrl: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1`,
    prepareHeaders: (headers) => {
      const token =
        typeof window !== 'undefined' ? localStorage.getItem('accessToken') : null;
      if (token && token !== 'null') {
        headers.set('Authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  tagTypes: ['Plan'],
  endpoints: (builder) => ({
    /* --------------------- GET /plans --------------------- */
    getPlans: builder.query<
      { items: Plan[]; total: number },
      {
        search?: string;
        page?: number;
        size?: number;
        sort?: string;
        desc?: boolean;
      }
    >({
      query: (params) => ({
        url: 'plans',
        method: 'GET',
        params,
      }),
      providesTags: (result) =>
        result
          ? [
              { type: 'Plan', id: 'LIST' },
              ...result.items.map((p) => ({
                type: 'Plan' as const,
                id: p.id,
              })),
            ]
          : [{ type: 'Plan', id: 'LIST' }],
    }),

    /* --------------------- POST /plans --------------------- */
    createPlan: builder.mutation<
      Plan,
      Omit<Plan, 'id' | 'created_at' | 'updated_at'>
    >({
      query: (body) => ({
        url: 'plans',
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Plan', id: 'LIST' }],
    }),

    /* --------------------- PATCH /plans/:id --------------------- */
    updatePlan: builder.mutation<Plan, UpdatePlanArg>({
      query: ({ id, ...patch }) => ({
        url: `plans/${id}`,
        method: 'PATCH',
        body: patch,
      }),
      // Инвалидируем только тот план, который изменили
      invalidatesTags: (result, _error, arg) => [{ type: 'Plan', id: arg.id }],
    }),

    /* --------------------- DELETE /plans/:id --------------------- */
    deletePlan: builder.mutation<void, string>({
      query: (id) => ({
        url: `plans/${id}`,
        method: 'DELETE',
      }),
      // После удаления нужно обновить список
      invalidatesTags: [{ type: 'Plan', id: 'LIST' }],
      // Сервер не возвращает тело – говорим RTK‑Query, что ответ undefined
      transformResponse: () => undefined,
    }),
  }),
});

/* -----------------------------------------------------------------
   3️⃣ Экспорт готовых хуков
----------------------------------------------------------------- */
export const {
  useGetPlansQuery,
  useCreatePlanMutation,
  useUpdatePlanMutation,
  useDeletePlanMutation,
} = planApi;

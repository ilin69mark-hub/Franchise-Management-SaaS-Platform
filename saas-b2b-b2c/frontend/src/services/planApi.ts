// src/services/planApi.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Plan } from '@/types';
import { getCsrfToken, CSRF_HEADER } from '@/utils/csrf';

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
    credentials: 'include', // RE-AUDIT: без этого cookie не уходят cross-origin → 401 вне nginx
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

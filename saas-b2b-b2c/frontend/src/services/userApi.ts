import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { UserProfile, UpdateProfileRequest, CreateEmployeeRequest, EmployeeResponse } from '@/types';
import { getCsrfToken, CSRF_HEADER } from '@/utils/csrf';

export const userApi = createApi({
  reducerPath: 'userApi',
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
  tagTypes: ['UserProfile', 'Employees'],
  endpoints: (builder) => ({
    getMyProfile: builder.query<UserProfile, void>({
      query: () => 'users/me',
      providesTags: (result) => (result ? [{ type: 'UserProfile', id: result.id }] : []),
    }),

    updateProfile: builder.mutation<UserProfile, UpdateProfileRequest>({
      query: (data) => ({
        url: 'users/me',
        method: 'PUT',
        body: data,
      }),
      invalidatesTags: (result) => (result ? [{ type: 'UserProfile', id: result.id }] : []),
    }),

    createEmployee: builder.mutation<EmployeeResponse, CreateEmployeeRequest>({
      query: (data) => ({
        url: 'users',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: [{ type: 'Employees' }],
    }),

    getEmployees: builder.query<EmployeeResponse[], void>({
      query: () => 'users',
      providesTags: (result) =>
        result
          ? [...result.map(({ id }) => ({ type: 'Employees' as const, id })), { type: 'Employees' as const }]
          : [{ type: 'Employees' as const }],
    }),
  }),
});

export const {
  useGetMyProfileQuery,
  useUpdateProfileMutation,
  useCreateEmployeeMutation,
  useGetEmployeesQuery,
} = userApi;

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { UserProfile, UpdateProfileRequest, CreateEmployeeRequest, EmployeeResponse } from '@/types';

export const userApi = createApi({
  reducerPath: 'userApi',
  baseQuery: fetchBaseQuery({
    baseUrl: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1`,
    prepareHeaders: (headers) => {
      const token = typeof window !== 'undefined' ? localStorage.getItem('accessToken') : null;
      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
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

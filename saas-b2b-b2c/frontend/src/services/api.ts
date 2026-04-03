import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import {
  User,
  Checklist,
  Lead,
  Dealer,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  Employee,
  Salon,
} from '@/types';

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: 'http://localhost:8080/api/v1',
    prepareHeaders: (headers, { getState }) => {
      let token = (getState() as any).auth?.accessToken;

      if (!token && typeof window !== 'undefined') {
        const storedState = localStorage.getItem('reduxState');
        if (storedState) {
            try {
                const parsed = JSON.parse(storedState);
                token = parsed?.auth?.accessToken;
            } catch (e) {
                console.error('Error parsing token', e);
            }
        }
      }

      if (token) {
        headers.set('authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  tagTypes: ['User', 'Checklist', 'Lead', 'Dealer', 'Auth', 'Employee', 'Notification', 'Salon'],
  endpoints: (builder) => ({
    // === АВТОРИЗАЦИЯ ===
    login: builder.mutation<AuthResponse, LoginRequest>({
      query: (credentials) => ({ url: '/auth/login', method: 'POST', body: credentials }),
      invalidatesTags: ['Auth'],
    }),
    register: builder.mutation<AuthResponse, RegisterRequest>({
      query: (userData) => ({ url: '/auth/register', method: 'POST', body: userData }),
      invalidatesTags: ['Auth'],
    }),
    logout: builder.mutation<void, void>({
      query: () => ({ url: '/auth/logout', method: 'POST' }),
    }),
    getProfile: builder.query<User, void>({
      query: () => '/auth/me',
      providesTags: ['User'],
    }),

    // === УВЕДОМЛЕНИЯ ===
    getNotifications: builder.query<any[], void>({
      query: () => '/notifications',
      providesTags: ['Notification'],
    }),
    readNotification: builder.mutation<void, string>({
      query: (id) => ({ url: `/notifications/${id}/read`, method: 'POST' }),
      invalidatesTags: ['Notification'],
    }),

    // === ЧЕК-ЛИСТЫ ===
    getChecklists: builder.query<Checklist[], void>({
      query: () => '/checklists',
      providesTags: (result) => result ? [...result.map(({ id }) => ({ type: 'Checklist' as const, id })), 'Checklist'] : ['Checklist'],
    }),
    createChecklist: builder.mutation<Checklist, Omit<Checklist, 'id'>>({
      query: (newChecklist) => ({ url: '/checklists', method: 'POST', body: newChecklist }),
      invalidatesTags: ['Checklist'],
    }),
    updateChecklist: builder.mutation<Checklist, Partial<Checklist> & Pick<Checklist, 'id'>>({
      query: ({ id, ...patch }) => ({ url: `/checklists/${id}`, method: 'PUT', body: patch }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Checklist', id }],
    }),
    deleteChecklist: builder.mutation<void, string>({
      query: (id) => ({ url: `/checklists/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Checklist'],
    }),

    // === ЛИДЫ ===
    getLeads: builder.query<Lead[], void>({
      query: () => '/leads',
      providesTags: (result) => result ? [...result.map(({ id }) => ({ type: 'Lead' as const, id })), 'Lead'] : ['Lead'],
    }),
    createLead: builder.mutation<Lead, Omit<Lead, 'id'>>({
      query: (newLead) => ({ url: '/leads', method: 'POST', body: newLead }),
      invalidatesTags: ['Lead'],
    }),
    updateLeadStatus: builder.mutation<void, { id: string; status: string }>({
      query: ({ id, status }) => ({
        url: `/leads/${id}/status`,
        method: 'PUT',
        body: { status },
      }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Lead', id }],
    }),
    addLeadActivity: builder.mutation<void, { leadId: string; type: string; description: string }>({
      query: ({ leadId, ...body }) => ({
        url: `/leads/${leadId}/activities`,
        method: 'POST',
        body,
      }),
      invalidatesTags: (result, error, { leadId }) => [{ type: 'Lead', id: leadId }],
    }),

    // === ДИЛЕРЫ ===
    getDealers: builder.query<Dealer[], void>({
      query: () => '/dealers',
      providesTags: (result) => result ? [...result.map(({ id }) => ({ type: 'Dealer' as const, id })), 'Dealer'] : ['Dealer'],
    }),

    // === СОТРУДНИКИ (HR) ===
    getEmployees: builder.query<Employee[], void>({
      query: () => '/users',
      providesTags: (result) => result ? [...result.map(({ id }) => ({ type: 'Employee' as const, id })), 'Employee'] : ['Employee'],
    }),
    createEmployee: builder.mutation<Employee, Partial<Employee> & { password: string }>({
      query: (newEmployee) => ({ url: '/users', method: 'POST', body: newEmployee }),
      invalidatesTags: ['Employee'],
    }),
    updateEmployee: builder.mutation<Employee, Partial<Employee> & Pick<Employee, 'id'>>({
      query: ({ id, ...patch }) => ({ url: `/users/${id}`, method: 'PUT', body: patch }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Employee', id }],
    }),
    deleteEmployee: builder.mutation<void, string>({
      query: (id) => ({ url: `/users/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Employee'],
    }),

    // === САЛОНЫ ===
    getSalons: builder.query<Salon[], void>({
      query: () => '/salons',
      providesTags: (result) => result ? [...result.map(({ id }) => ({ type: 'Salon' as const, id })), 'Salon'] : ['Salon'],
    }),
    createSalon: builder.mutation<Salon, { name: string; address: string }>({
      query: (newSalon) => ({ url: '/salons', method: 'POST', body: newSalon }),
      invalidatesTags: ['Salon'],
    }),
    assignManager: builder.mutation<void, { user_id: string; salon_id: string }>({
      query: (body) => ({ url: '/salons/assign', method: 'POST', body }),
      invalidatesTags: ['Employee', 'Salon'], // ВАЖНО: Сбрасываем кэш салонов и сотрудников
    }),
    updateSalon: builder.mutation<Salon, { id: string; data: { name: string; address: string } }>({
      query: ({ id, data }) => ({ url: `/salons/${id}`, method: 'PUT', body: data }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Salon', id }],
    }),
    deleteSalon: builder.mutation<void, string>({
      query: (id) => ({ url: `/salons/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Salon'],
    }),
  }),
});


export const {
  useLoginMutation,
  useRegisterMutation,
  useLogoutMutation,
  useGetProfileQuery,
  useGetNotificationsQuery,
  useReadNotificationMutation,
  useGetChecklistsQuery,
  useCreateChecklistMutation,
  useUpdateChecklistMutation,
  useDeleteChecklistMutation,
  useGetLeadsQuery,
  useCreateLeadMutation,
  useGetDealersQuery,
  useGetEmployeesQuery,
  useCreateEmployeeMutation,
  useUpdateEmployeeMutation,
  useDeleteEmployeeMutation,
  useUpdateLeadStatusMutation,
  useAddLeadActivityMutation,
  useGetSalonsQuery,
  useCreateSalonMutation,
  useAssignManagerMutation,
  useUpdateSalonMutation,
  useDeleteSalonMutation,
} = apiSlice;
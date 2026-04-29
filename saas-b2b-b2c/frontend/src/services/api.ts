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
  UnitTemplate,
} from '@/types';
import dayjs from 'dayjs';

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    // Берём URL из env‑переменной (для Docker – имя сервиса `backend`);
    // fallback – localhost:8080 (для локального старта без Docker)
    baseUrl: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1`,
    prepareHeaders: (headers, { getState }) => {
      // Пытаемся взять токен из Redux‑стора
      let token = (getState() as any).auth?.accessToken;

      // Если в сторе ничего, ищем в localStorage (на случай полной перезагрузки)
      if (!token && typeof window !== 'undefined') {
        const storedState = localStorage.getItem('reduxState');
        if (storedState) {
          try {
            const parsed = JSON.parse(storedState);
            token = parsed?.auth?.accessToken;
          } catch (e) {
            console.error('Error parsing reduxState from localStorage', e);
          }
        }
      }

      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  tagTypes: [
    'User',
    'Checklist',
    'Lead',
    'Dealer',
    'Auth',
    'Employee',
    'Notification',
    'Salon',
    'UnitTemplate',
    'DealerPlanFact',
    'DealerFunnel',
    'DealerBenchmark',
    'TopManagers',
    'Inventory',
    'LostSales',
    'Returns',
    'Tasks',
    'Requests',
    'MarketingBudget',
    'Interactions',
    'ReportData',
    'Reports',
    'Alerts',
    'AlertSettings',
  ],
  endpoints: (builder) => ({
    // === АВТОРИЗАЦИЯ ===
    login: builder.mutation<AuthResponse, LoginRequest>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
      invalidatesTags: ['Auth'],
    }),
    register: builder.mutation<AuthResponse, RegisterRequest>({
      query: (userData) => ({
        url: '/auth/register',
        method: 'POST',
        body: userData,
      }),
      invalidatesTags: ['Auth'],
    }),
    logout: builder.mutation<void, void>({
      query: () => ({
        url: '/auth/logout',
        method: 'POST',
      }),
    }),
    getProfile: builder.query<User, void>({
      query: () => '/auth/me',
      providesTags: ['User'],
    }),

    // === УВЕДОМЛЕНИЯ ===
    getNotifications: builder.query<any[], void>({
      query: () => '/notifications',
      providesTags: (result) =>
        result ? [...result.map(({ id }) => ({ type: 'Notification' as const, id })), 'Notification'] : ['Notification'],
    }),
    readNotification: builder.mutation<void, string>({
      query: (id) => ({
        url: `/notifications/${id}/read`,
        method: 'POST',
      }),
      invalidatesTags: ['Notification'],
    }),

    // === ЧЕК‑ЛИСТЫ ===
    getChecklists: builder.query<Checklist[], void>({
      query: () => '/checklists',
      providesTags: (result) =>
        result ? [...result.map(({ id }) => ({ type: 'Checklist' as const, id })), 'Checklist'] : ['Checklist'],
    }),
    createChecklist: builder.mutation<Checklist, Omit<Checklist, 'id'>>({
      query: (newChecklist) => ({
        url: '/checklists',
        method: 'POST',
        body: newChecklist,
      }),
      invalidatesTags: ['Checklist'],
    }),
    updateChecklist: builder.mutation<Checklist, Partial<Checklist> & Pick<Checklist, 'id'>>({
      query: ({ id, ...patch }) => ({
        url: `/checklists/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Checklist', id }],
    }),
    deleteChecklist: builder.mutation<void, string>({
      query: (id) => ({
        url: `/checklists/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Checklist'],
    }),

    // === ЛИДЫ ===
    getLeads: builder.query<Lead[], void>({
      query: () => '/leads',
      providesTags: (result) =>
        result ? [...result.map(({ id }) => ({ type: 'Lead' as const, id })), 'Lead'] : ['Lead'],
    }),
    createLead: builder.mutation<Lead, Omit<Lead, 'id'>>({
      query: (newLead) => ({
        url: '/leads',
        method: 'POST',
        body: newLead,
      }),
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
      providesTags: (result) =>
        result ? [...result.map(({ id }) => ({ type: 'Dealer' as const, id })), 'Dealer'] : ['Dealer'],
    }),

    // === СОТРУДНИКИ (HR) ===
    getEmployees: builder.query<Employee[], void>({
      query: () => '/users',
      providesTags: (result) =>
        result ? [...result.map(({ id }) => ({ type: 'Employee' as const, id })), 'Employee'] : ['Employee'],
    }),
    createEmployee: builder.mutation<Employee, Partial<Employee> & { password: string }>({
      query: (newEmployee) => ({
        url: '/users',
        method: 'POST',
        body: newEmployee,
      }),
      invalidatesTags: ['Employee'],
    }),
    updateEmployee: builder.mutation<Employee, Partial<Employee> & Pick<Employee, 'id'>>({
      query: ({ id, ...patch }) => ({
        url: `/users/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Employee', id }],
    }),
    deleteEmployee: builder.mutation<void, string>({
      query: (id) => ({
        url: `/users/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Employee'],
    }),

    // === САЛОНЫ ===
    getSalons: builder.query<Salon[], void>({
      query: () => '/salons',
      providesTags: (result) =>
        result ? [...result.map(({ id }) => ({ type: 'Salon' as const, id })), 'Salon'] : ['Salon'],
    }),
    createSalon: builder.mutation<Salon, { name: string; address: string }>({
      query: (newSalon) => ({
        url: '/salons',
        method: 'POST',
        body: newSalon,
      }),
      invalidatesTags: ['Salon'],
    }),
    assignManager: builder.mutation<void, { user_id: string; salon_id: string }>({
      query: (body) => ({
        url: '/salons/assign',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Employee', 'Salon'],
    }),
    updateSalon: builder.mutation<Salon, { id: string; data: { name: string; address: string } }>({
      query: ({ id, data }) => ({
        url: `/salons/${id}`,
        method: 'PUT',
        body: data,
      }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Salon', id }],
    }),
    deleteSalon: builder.mutation<void, string>({
      query: (id) => ({
        url: `/salons/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Salon'],
    }),

    // === ШАБЛОНЫ UNIT-ЭКОНОМИКИ ===
    getUnitTemplates: builder.query<UnitTemplate[], void>({
      query: () => '/dealer/unit-templates',
      providesTags: (result) =>
        result ? [...result.map(({ id }) => ({ type: 'UnitTemplate' as const, id })), 'UnitTemplate'] : ['UnitTemplate'],
    }),
    createUnitTemplate: builder.mutation<UnitTemplate, Omit<UnitTemplate, 'id'>>({
      query: (newTemplate) => ({
        url: '/dealer/unit-templates',
        method: 'POST',
        body: newTemplate,
      }),
      invalidatesTags: ['UnitTemplate'],
    }),
    deleteUnitTemplate: builder.mutation<void, string>({
      query: (id) => ({
        url: `/dealer/unit-templates/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['UnitTemplate'],
    }),

    // === ПЛАН-ФАКТ ДИЛЕРА ===
    getDealerPlanFact: builder.query<any, { period: string; date: string }>({
      query: ({ period, date }) => `/dealer/plan-fact?period=${period}&date=${date}`,
      providesTags: ['DealerPlanFact'],
    }),

    // === ВОРОНКА СЕТИ ===
    getDealerFunnel: builder.query<any, string>({
      query: (date) => `/dealer/funnel?date=${date}`,
      providesTags: ['DealerFunnel'],
    }),

    // === БЕНЧМАРКИ ОТ ФРАНЧАЙЗЕРА ===
    getDealerBenchmark: builder.query<any, string>({
      query: (metric) => `/dealer/benchmark?metric=${metric}`,
      providesTags: ['DealerBenchmark'],
    }),

    // === ТОП-МЕНЕДЖЕРЫ ===
    getTopManagers: builder.query<any, { period: string; limit: number }>({
      query: ({ period, limit }) => `/dealer/top-managers?period=${period}&limit=${limit}`,
      providesTags: ['TopManagers'],
    }),

    // === СКЛАД И ИНВЕНТАРИ ===
    getInventory: builder.query<any, { period: string; store: string }>({
      query: ({ period, store }) => `/dealer/inventory?period=${period}&store=${store}`,
      providesTags: ['Inventory'],
    }),

    // === УПУЩЕННЫЕ ПРОДАЖИ ===
    getLostSales: builder.query<any, string>({
      query: (period) => `/dealer/lost-sales?period=${period}`,
      providesTags: ['LostSales'],
    }),

    // === ВОЗВРАТЫ ===
    getReturns: builder.query<any, string>({
      query: (period) => `/dealer/returns?period=${period}`,
      providesTags: ['Returns'],
    }),

    // === ЗАДАЧИ ОТ ФРАНЧАЙЗЕРА ===
    getTasks: builder.query<any, void>({
      query: () => '/dealer/tasks',
      providesTags: ['Tasks'],
    }),
    updateTaskStatus: builder.mutation<any, { id: string; status: string }>({
      query: ({ id, status }) => ({
        url: `/dealer/tasks/${id}`,
        method: 'PATCH',
        body: { status },
      }),
      invalidatesTags: ['Tasks'],
    }),
    addTaskComment: builder.mutation<any, { id: string; comment: string }>({
      query: ({ id, comment }) => ({
        url: `/dealer/tasks/${id}/comments`,
        method: 'POST',
        body: { comment },
      }),
      invalidatesTags: ['Tasks'],
    }),

    // === ЗАПРОСЫ К БРЕНДУ ===
    getRequests: builder.query<any, void>({
      query: () => '/dealer/requests',
      providesTags: ['Requests'],
    }),
    createRequest: builder.mutation<any, any>({
      query: (body) => ({
        url: '/dealer/requests',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Requests'],
    }),

    // === МАРКЕТИНГОВЫЙ БЮДЖЕТ ===
    getMarketingBudget: builder.query<any, string>({
      query: (quarter) => `/dealer/marketing-budget?quarter=${quarter}`,
      providesTags: ['MarketingBudget'],
    }),

    // === ИСТОРИЯ ВЗАИМОДЕЙСТВИЙ ===
    getInteractions: builder.query<any, string>({
      query: (managerId) => `/dealer/interactions?manager_id=${managerId}`,
      providesTags: ['Interactions'],
    }),

    // === ОТЧЁТЫ ДЛЯ БРЕНДА ===
    getReportData: builder.query<any, { period: string; date: string }>({
      query: ({ period, date }) => `/dealer/report-data?period=${period}&date=${date}`,
      providesTags: ['ReportData'],
    }),
    createReport: builder.mutation<any, any>({
      query: (body) => ({
        url: '/dealer/reports',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Reports'],
    }),
    getReportHistory: builder.query<any, number>({
      query: (limit) => `/dealer/reports?limit=${limit}`,
      providesTags: ['Reports'],
    }),

    // === АЛЕРТЫ ===
    getAlerts: builder.query<any, void>({
      query: () => '/dealer/alerts',
      providesTags: ['Alerts'],
    }),
    getUnreadAlerts: builder.query<any, void>({
      query: () => '/dealer/alerts/unread',
      providesTags: ['Alerts'],
    }),
    markAlertRead: builder.mutation<any, string>({
      query: (id) => ({
        url: `/dealer/alerts/${id}/read`,
        method: 'PATCH',
      }),
      invalidatesTags: ['Alerts'],
    }),
    getAlertSettings: builder.query<any, void>({
      query: () => '/dealer/alert-settings',
      providesTags: ['AlertSettings'],
    }),
    updateAlertSettings: builder.mutation<any, any>({
      query: (body) => ({
        url: '/dealer/alert-settings',
        method: 'PUT',
        body,
      }),
      invalidatesTags: ['AlertSettings'],
    }),
  }),
});

export const {
  // auth
  useLoginMutation,
  useRegisterMutation,
  useLogoutMutation,
  useGetProfileQuery,

  // notifications
  useGetNotificationsQuery,
  useReadNotificationMutation,

  // checklists
  useGetChecklistsQuery,
  useCreateChecklistMutation,
  useUpdateChecklistMutation,
  useDeleteChecklistMutation,

  // leads
  useGetLeadsQuery,
  useCreateLeadMutation,
  useUpdateLeadStatusMutation,
  useAddLeadActivityMutation,

  // dealers
  useGetDealersQuery,

  // employees
  useGetEmployeesQuery,
  useCreateEmployeeMutation,
  useUpdateEmployeeMutation,
  useDeleteEmployeeMutation,

  // salons
  useGetSalonsQuery,
  useCreateSalonMutation,
  useAssignManagerMutation,
  useUpdateSalonMutation,
  useDeleteSalonMutation,

  // unit templates
  useGetUnitTemplatesQuery,
  useCreateUnitTemplateMutation,
  useDeleteUnitTemplateMutation,

  // dealer plan-fact
  useGetDealerPlanFactQuery,

  // dealer funnel
  useGetDealerFunnelQuery,

  // dealer benchmark
  useGetDealerBenchmarkQuery,

  // top managers
  useGetTopManagersQuery,

  // inventory
  useGetInventoryQuery,

  // lost sales
  useGetLostSalesQuery,

  // returns
  useGetReturnsQuery,

  // tasks
  useGetTasksQuery,
  useUpdateTaskStatusMutation,
  useAddTaskCommentMutation,

  // requests
  useGetRequestsQuery,
  useCreateRequestMutation,

  // marketing budget
  useGetMarketingBudgetQuery,

  // interactions
  useGetInteractionsQuery,

  // dealer reports
  useGetReportDataQuery,
  useCreateReportMutation,
  useGetReportHistoryQuery,

  // alerts
  useGetAlertsQuery,
  useGetUnreadAlertsQuery,
  useMarkAlertReadMutation,
  useGetAlertSettingsQuery,
  useUpdateAlertSettingsMutation,
} = apiSlice;

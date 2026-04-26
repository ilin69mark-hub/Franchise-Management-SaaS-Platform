/* ==============================
   Общие типы, используемые в фронтенде
   ============================== */

/* ---------- Пользователь и аутентификация ---------- */
export interface User {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: string;
  phone?: string;
  managed_by?: string;
  tenant_id?: string;
}

/* ---------- Запросы аутентификации ---------- */
export interface LoginRequest {
  email: string;
  password: string;
}
export interface RegisterRequest {
  email: string;
  password: string;
  role?: string;
  first_name?: string;
  last_name?: string;
  phone?: string;
  managed_by?: string;
}

/* ---------- Ответ после логина/регистрации ---------- */
export interface AuthResponse {
  user: User;
  token: string;
  refresh_token?: string;
}

/* ---------- Сотрудники (Employee) ---------- */
// В проекте уже объявлен тип Employee, поэтому держим его один раз.
export interface Employee {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  phone?: string;
  created_at?: string;
  managed_by?: string;
}

/* ---------- Чек‑листы ---------- */
export interface Checklist {
  id: string;
  title: string;
  description?: string;
  status: 'pending' | 'in_progress' | 'waiting' | 'completed' | string;
  priority?: 'urgent' | 'important' | 'normal' | string;
  created_at?: string;
  updated_at?: string;
  assigned_to?: string; // id сотрудника
  start_date?: string;
  end_date?: string;
  user_id?: string;
  tenant_id?: string;
  is_overdue?: boolean;
}

/* ---------- Лиды и активности ---------- */
export interface Lead {
  id: string;
  salon_id: string;
  manager_id: string;
  full_name: string;
  phone?: string;
  email?: string;
  interest_product?: string;
  budget?: number;
  status: 'new' | 'contact' | 'meeting' | 'sale' | 'archive' | string;
  created_at?: string;
  updated_at?: string;
}
export interface LeadActivity {
  id: string;
  lead_id: string;
  user_id: string;
  type: 'call' | 'meeting' | 'note' | string;
  description: string;
  created_at: string;
}
export interface CreateLeadRequest {
  full_name: string;
  phone?: string;
  email?: string;
  interest_product?: string;
  budget?: number;
}

/* ---------- Дилеры ---------- */
export interface Dealer {
  id: string;
  businessName: string;
  address?: string;
  status: 'active' | 'inactive' | 'suspended';
}

/* ---------- Тарифные планы (Plan) ---------- */
export interface Plan {
  id: string;
  name: string;
  price: number;
  max_salons: number;
  max_users: number;
  created_at?: string;
  updated_at?: string;
}

/* ---------- Салоны ---------- */
export interface Salon {
  id: string;
  name: string;
  address: string;
  tenant_id: string;
  created_at?: string;
  manager?: Employee;
}

/* ---------- Goal (цель / план) ---------- */
export interface Goal {
  /** UUID цели */
  id: string;

  /** Кто поставил цель (UUID пользователя) */
  assigner_id: string;

  /** Кому назначена цель (UUID получателя) */
  assignee_id: string;

  /** Роль получателя */
  role:
    | 'franchise_manager'
    | 'dealer'
    | 'salon_manager'
    | string;

  /** План продаж (₽) */
  sales_plan: number;
  /** План лидов (кол‑во) */
  leads_plan: number;
  /** План звонков (кол‑во) */
  calls_plan: number;
  /** План встреч (кол‑во) */
  meetings_plan: number;

  /** Дата‑план (YYYY‑MM‑DD) */
  target_date: string;

  /** Фактические результаты (опциональны) */
  sales_fact?: number;
  leads_fact?: number;
  calls_fact?: number;
  meetings_fact?: number;

  /** Прогноз (может вводить пользователь) */
  forecast?: number;

  /** Даты создания/изменения */
  created_at?: string;
  updated_at?: string;
}

/* -------------------------------------------------
   Всё готово к импорту:
   import type {
     User, AuthResponse, Employee, Checklist, Lead, LeadActivity,
     CreateLeadRequest, Dealer, Plan, Salon, Goal
   } from '@/types';
   ------------------------------------------------- */

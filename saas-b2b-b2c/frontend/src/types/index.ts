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
  id: string;
  assigner_id: string;
  assignee_id: string;
  role: string;
  sales_plan: number;
  leads_plan: number;
  calls_plan: number;
  meetings_plan: number;
  period: string;
  start_date: string;
  end_date: string;
  target_date: string;
  sales_fact?: number;
  leads_fact?: number;
  calls_fact?: number;
  meetings_fact?: number;
  forecast?: number;
  created_at: string;
  updated_at: string;
}

/* ---------- Unit-экономика (калькулятор) ---------- */
export interface UnitEconomyInput {
  salePrice: number;
  purchasePrice: number;
  managerPercent: number;
  deliveryCost: number;
  acquiringPercent: number;
  rentAmortization: number;
  additionalServices: number;
}

export interface UnitEconomyResult {
  marginalProfit: number;
  netProfit: number;
  profitability: number;
  priceAfterDiscount: number;
}

export interface UnitTemplate {
  id?: string;
  name: string;
  category: 'sofa' | 'kitchen' | 'wardrobe' | 'other';
  input: UnitEconomyInput;
  created_at?: string;
}

/* ---------- Профиль пользователя ---------- */
export type UserStatus = 'online' | 'office' | 'meeting' | 'vacation' | 'dnd';

export interface UserContacts {
  email_visible: boolean;
  phone_visible: boolean;
  phone?: string;
  telegram?: string;
  whatsapp?: string;
  working_hours?: string;
}

export interface UserProfile {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  display_name?: string;
  position?: string;
  bio?: string;
  quote?: string;
  avatar_url?: string;
  status?: UserStatus;
  available_for_questions?: boolean;
  achievements?: string[];
  contacts?: UserContacts;
}

export interface UpdateProfileRequest {
  first_name?: string;
  last_name?: string;
  display_name?: string;
  position?: string;
  bio?: string;
  quote?: string;
  status?: UserStatus;
  available_for_questions?: boolean;
  contacts?: Partial<UserContacts>;
}

/* ---------- Создание сотрудника (для HR-модуля) ---------- */
export interface CreateEmployeeRequest {
  email: string;
  password: string;
  first_name: string;
  last_name?: string;
  phone?: string;
  role: string;
  managed_by?: string;
}

export interface EmployeeResponse {
  id: string;
  email: string;
  first_name: string;
  last_name?: string;
  phone?: string;
  role: string;
  managed_by?: string;
  created_at?: string;
}

/* -------------------------------------------------
   Всё готово к импорту:
   import type {
     User, AuthResponse, Employee, Checklist, Lead, LeadActivity,
     CreateLeadRequest, Dealer, Plan, Salon, Goal,
     UnitEconomyInput, UnitEconomyResult, UnitTemplate,
     UserProfile, UserStatus, UserContacts, UpdateProfileRequest
   } from '@/types';
   ------------------------------------------------- */

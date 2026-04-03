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

export interface AuthResponse {
  user: User;
  token: string;
  refresh_token?: string;
}

export interface Employee {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  phone: string;
  created_at: string;
  managed_by?: string;
}

export interface Checklist {
  id: string;
  title: string;
  description?: string;
  status: string; // ИСПРАВЛЕНО: используем string вместо строгого перечисления
  priority?: 'urgent' | 'important' | 'normal';
  created_at?: string;
  updated_at?: string;
  assigned_to?: string;
  start_date?: string;
  end_date?: string;
  user_id?: string;
  tenant_id?: string;
  is_overdue?: boolean;
}

export interface Lead {
  id: string;
  salon_id: string;
  manager_id: string;
  full_name: string;
  phone?: string;
  email?: string;
  interest_product?: string;
  budget?: number;
  status: string; // new, contact, meeting, sale, archive
  created_at?: string;
  updated_at?: string;
}

export interface LeadActivity {
  id: string;
  lead_id: string;
  user_id: string;
  type: string; // call, meeting, note
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

export interface Dealer {
  id: string;
  businessName: string;
  address?: string;
  status: 'active' | 'inactive' | 'suspended';
}

export interface Plan {
  id: string;
  name: string;
  price: number;
  max_users: number;
  max_salons?: number;
}

export interface Salon {
  id: string;
  name: string;
  address: string;
  tenant_id: string;
  created_at?: string;
  manager?: Employee;
}

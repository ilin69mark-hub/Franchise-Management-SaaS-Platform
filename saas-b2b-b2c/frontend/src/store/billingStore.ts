import { create } from 'zustand';

interface Invoice {
  id: string;
  tenant: string;
  invoiceNumber: string;
  period: string;
  amount: number;
  issuedAt: string;
  paidAt?: string;
  status: 'paid' | 'pending' | 'overdue';
  paymentMethod: 'auto' | 'manual' | 'other';
}

interface BillingSettings {
  notifyAfterDays: number;
  secondNotifyAfterDays: number;
  suspendAfterDays: number;
  autoInvoiceDays: number;
  emailTemplate?: string;
}

interface PaymentHistoryItem {
  id: string;
  date: string;
  tenant: string;
  amount: number;
  type: 'income' | 'refund';
}

interface BillingState {
  invoices: Invoice[];
  settings: BillingSettings | null;
  history: PaymentHistoryItem[];
  period: string;
  isLoading: boolean;
  showInvoiceModal: boolean;
  showPaymentModal: boolean;
  setInvoices: (invoices: Invoice[]) => void;
  setSettings: (settings: BillingSettings) => void;
  setHistory: (history: PaymentHistoryItem[]) => void;
  setPeriod: (period: string) => void;
  setLoading: (loading: boolean) => void;
  setShowInvoiceModal: (show: boolean) => void;
  setShowPaymentModal: (show: boolean) => void;
}

export const useBillingStore = create<BillingState>((set) => ({
  invoices: [],
  settings: null,
  history: [],
  period: '30',
  isLoading: false,
  showInvoiceModal: false,
  showPaymentModal: false,
  setInvoices: (invoices) => set({ invoices }),
  setSettings: (settings) => set({ settings }),
  setHistory: (history) => set({ history }),
  setPeriod: (period) => set({ period }),
  setLoading: (loading) => set({ isLoading: loading }),
  setShowInvoiceModal: (show) => set({ showInvoiceModal: show }),
  setShowPaymentModal: (show) => set({ showPaymentModal: show }),
}));
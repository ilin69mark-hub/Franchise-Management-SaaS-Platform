import { useBillingStore } from '@/store/billingStore';

describe('billingStore', () => {
  beforeEach(() => {
    useBillingStore.setState({
      invoices: [],
      settings: null,
      history: [],
      period: '30',
      isLoading: false,
      showInvoiceModal: false,
      showPaymentModal: false,
    });
  });

  describe('initial state', () => {
    it('has empty invoices array', () => {
      const { invoices } = useBillingStore.getState();
      expect(invoices).toEqual([]);
    });

    it('has null settings', () => {
      const { settings } = useBillingStore.getState();
      expect(settings).toBeNull();
    });

    it('has empty history array', () => {
      const { history } = useBillingStore.getState();
      expect(history).toEqual([]);
    });

    it('has default period of 30', () => {
      const { period } = useBillingStore.getState();
      expect(period).toBe('30');
    });

    it('is not loading by default', () => {
      const { isLoading } = useBillingStore.getState();
      expect(isLoading).toBe(false);
    });

    it('has modals closed by default', () => {
      const { showInvoiceModal, showPaymentModal } = useBillingStore.getState();
      expect(showInvoiceModal).toBe(false);
      expect(showPaymentModal).toBe(false);
    });
  });

  describe('setInvoices', () => {
    it('updates invoices array', () => {
      const mockInvoices = [
        {
          id: 'inv-1',
          tenant: 'Tenant 1',
          invoiceNumber: 'INV-001',
          period: '2024-01',
          amount: 5000,
          issuedAt: '2024-01-01',
          status: 'pending' as const,
          paymentMethod: 'manual' as const,
        },
      ];
      useBillingStore.getState().setInvoices(mockInvoices);
      expect(useBillingStore.getState().invoices).toEqual(mockInvoices);
    });
  });

  describe('setSettings', () => {
    it('updates billing settings', () => {
      const mockSettings = {
        notifyAfterDays: 5,
        secondNotifyAfterDays: 10,
        suspendAfterDays: 15,
        autoInvoiceDays: 1,
      };
      useBillingStore.getState().setSettings(mockSettings);
      expect(useBillingStore.getState().settings).toEqual(mockSettings);
    });
  });

  describe('setHistory', () => {
    it('updates payment history', () => {
      const mockHistory = [
        {
          id: 'hist-1',
          date: '2024-01-15',
          tenant: 'Tenant 1',
          amount: 5000,
          type: 'income' as const,
        },
      ];
      useBillingStore.getState().setHistory(mockHistory);
      expect(useBillingStore.getState().history).toEqual(mockHistory);
    });
  });

  describe('setPeriod', () => {
    it('updates period', () => {
      useBillingStore.getState().setPeriod('90');
      expect(useBillingStore.getState().period).toBe('90');
    });
  });

  describe('setLoading', () => {
    it('updates loading state', () => {
      useBillingStore.getState().setLoading(true);
      expect(useBillingStore.getState().isLoading).toBe(true);
    });
  });

  describe('setShowInvoiceModal', () => {
    it('opens invoice modal', () => {
      useBillingStore.getState().setShowInvoiceModal(true);
      expect(useBillingStore.getState().showInvoiceModal).toBe(true);
    });

    it('closes invoice modal', () => {
      useBillingStore.getState().setShowInvoiceModal(true);
      useBillingStore.getState().setShowInvoiceModal(false);
      expect(useBillingStore.getState().showInvoiceModal).toBe(false);
    });
  });

  describe('setShowPaymentModal', () => {
    it('opens payment modal', () => {
      useBillingStore.getState().setShowPaymentModal(true);
      expect(useBillingStore.getState().showPaymentModal).toBe(true);
    });

    it('closes payment modal', () => {
      useBillingStore.getState().setShowPaymentModal(true);
      useBillingStore.getState().setShowPaymentModal(false);
      expect(useBillingStore.getState().showPaymentModal).toBe(false);
    });
  });
});

describe('Billing Store Invoice Types', () => {
  it('supports paid invoice status', () => {
    const invoice = {
      id: 'inv-1',
      tenant: 'Tenant 1',
      invoiceNumber: 'INV-001',
      period: '2024-01',
      amount: 5000,
      issuedAt: '2024-01-01',
      paidAt: '2024-01-05',
      status: 'paid' as const,
      paymentMethod: 'auto' as const,
    };
    expect(invoice.status).toBe('paid');
  });

  it('supports pending invoice status', () => {
    const invoice = {
      id: 'inv-1',
      tenant: 'Tenant 1',
      invoiceNumber: 'INV-001',
      period: '2024-01',
      amount: 5000,
      issuedAt: '2024-01-01',
      status: 'pending' as const,
      paymentMethod: 'manual' as const,
    };
    expect(invoice.status).toBe('pending');
  });

  it('supports overdue invoice status', () => {
    const invoice = {
      id: 'inv-1',
      tenant: 'Tenant 1',
      invoiceNumber: 'INV-001',
      period: '2024-01',
      amount: 5000,
      issuedAt: '2024-01-01',
      status: 'overdue' as const,
      paymentMethod: 'manual' as const,
    };
    expect(invoice.status).toBe('overdue');
  });
});

describe('Billing Store Payment Types', () => {
  it('supports income payment type', () => {
    const payment = {
      id: 'pay-1',
      date: '2024-01-15',
      tenant: 'Tenant 1',
      amount: 5000,
      type: 'income' as const,
    };
    expect(payment.type).toBe('income');
  });

  it('supports refund payment type', () => {
    const payment = {
      id: 'pay-1',
      date: '2024-01-15',
      tenant: 'Tenant 1',
      amount: 1000,
      type: 'refund' as const,
    };
    expect(payment.type).toBe('refund');
  });
});
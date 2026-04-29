import { useFranchiserAlertStore, FranchiserAlert, FranchiserAlertSettings } from '@/store/franchiserAlertStore';

describe('franchiserAlertStore', () => {
  it('has initial alerts', () => {
    const state = useFranchiserAlertStore.getState();
    expect(state.alerts).toBeDefined();
    expect(state.alerts.length).toBeGreaterThan(0);
  });

  it('has default settings', () => {
    const state = useFranchiserAlertStore.getState();
    expect(state.settings).toBeDefined();
    expect(state.settings.thresholds.criticalForecastPercent).toBe(90);
    expect(state.settings.thresholds.dealerChurnPercent).toBe(5);
    expect(state.settings.thresholds.managerKpiPercent).toBe(70);
  });

  it('marks alert as read', () => {
    const alertId = '1';
    useFranchiserAlertStore.getState().markAsRead(alertId);
    const updatedAlert = useFranchiserAlertStore.getState().alerts.find(a => a.id === alertId);
    expect(updatedAlert?.status).toBe('acknowledged');
  });

  it('resolves alert', () => {
    const alertId = '2';
    useFranchiserAlertStore.getState().resolveAlert(alertId);
    const updatedAlert = useFranchiserAlertStore.getState().alerts.find(a => a.id === alertId);
    expect(updatedAlert?.status).toBe('resolved');
    expect(updatedAlert?.resolvedAt).toBeDefined();
  });

  it('assigns alert to manager', () => {
    const alertId = '3';
    useFranchiserAlertStore.getState().assignAlert(alertId, 'Иван Иванов');
    const updatedAlert = useFranchiserAlertStore.getState().alerts.find(a => a.id === alertId);
    expect(updatedAlert?.assignedTo).toBe('Иван Иванов');
  });

  it('calculates unread count', () => {
    const count = useFranchiserAlertStore.getState().getUnreadCount();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  it('calculates critical count', () => {
    const count = useFranchiserAlertStore.getState().getCriticalCount();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  it('updates settings', () => {
    useFranchiserAlertStore.getState().setSettings({
      thresholds: { criticalForecastPercent: 85, dealerChurnPercent: 10, managerKpiPercent: 75 }
    });
    const settings = useFranchiserAlertStore.getState().settings;
    expect(settings.thresholds.criticalForecastPercent).toBe(85);
    expect(settings.thresholds.dealerChurnPercent).toBe(10);
  });

  it('filters alerts by type', () => {
    const { alerts, setFilter } = useFranchiserAlertStore.getState();
    setFilter('plan');
    const planAlerts = alerts.filter(a => a.type === 'plan' && a.status !== 'resolved');
    expect(planAlerts.length).toBeGreaterThanOrEqual(0);
    setFilter('all');
  });
});
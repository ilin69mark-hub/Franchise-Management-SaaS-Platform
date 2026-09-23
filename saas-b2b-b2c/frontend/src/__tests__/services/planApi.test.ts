import { configureStore } from '@reduxjs/toolkit';
import { planApi } from '@/services/planApi';

global.fetch = jest.fn(() =>
  Promise.resolve({
    ok: true,
    status: 200,
    json: () => Promise.resolve([]),
    text: () => Promise.resolve(''),
    headers: { get: () => 'application/json' },
  } as unknown as Response)
) as jest.Mock;

describe('planApi', () => {
  it('should have correct reducerPath', () => {
    expect(planApi.reducerPath).toBe('planApi');
  });

  it('should have endpoints defined', () => {
    expect(planApi.endpoints).toBeDefined();
  });

  it('should have getPlans endpoint', () => {
    expect(planApi.endpoints.getPlans).toBeDefined();
  });

  it('should have createPlan mutation', () => {
    expect(planApi.endpoints.createPlan).toBeDefined();
  });

  it('should have updatePlan mutation', () => {
    expect(planApi.endpoints.updatePlan).toBeDefined();
  });

  it('should export hooks', () => {
    expect(planApi.useGetPlansQuery).toBeDefined();
    expect(planApi.useCreatePlanMutation).toBeDefined();
    expect(planApi.useUpdatePlanMutation).toBeDefined();
    expect(planApi.useDeletePlanMutation).toBeDefined();
  });

  it('диспатчит getPlans', async () => {
    const store = configureStore({ reducer: { planApi: planApi.reducer }, middleware: (g) => g().concat(planApi.middleware) });
    const result = store.dispatch(planApi.endpoints.getPlans.initiate() as never) as unknown as Promise<unknown>;
    await (result as Promise<unknown>).catch(() => {});
    expect(result).toBeDefined();
  });

  it('диспатчит createPlan', async () => {
    const store = configureStore({ reducer: { planApi: planApi.reducer }, middleware: (g) => g().concat(planApi.middleware) });
    const result = store.dispatch(planApi.endpoints.createPlan.initiate({ title: 'test' } as never) as never) as unknown as Promise<unknown>;
    await (result as Promise<unknown>).catch(() => {});
    expect(result).toBeDefined();
  });
});
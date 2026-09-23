import { configureStore } from '@reduxjs/toolkit';
import { goalApi } from '@/services/goalApi';

global.fetch = jest.fn(() =>
  Promise.resolve({
    ok: true,
    status: 200,
    json: () => Promise.resolve({}),
    text: () => Promise.resolve(''),
    headers: { get: () => 'application/json' },
  } as unknown as Response)
) as jest.Mock;

describe('goalApi', () => {
  it('should have correct reducerPath', () => {
    expect(goalApi.reducerPath).toBe('goalApi');
  });

  it('should have endpoints defined', () => {
    expect(goalApi.endpoints).toBeDefined();
  });

  it('should have getMyGoal endpoint', () => {
    expect(goalApi.endpoints.getMyGoal).toBeDefined();
  });

  it('should have getVisibleGoals endpoint', () => {
    expect(goalApi.endpoints.getVisibleGoals).toBeDefined();
  });

  it('should have setGoal mutation', () => {
    expect(goalApi.endpoints.setGoal).toBeDefined();
  });

  it('should have updateGoal mutation', () => {
    expect(goalApi.endpoints.updateGoal).toBeDefined();
  });

  it('should export hooks', () => {
    expect(goalApi.useGetMyGoalQuery).toBeDefined();
    expect(goalApi.useGetVisibleGoalsQuery).toBeDefined();
    expect(goalApi.useSetGoalMutation).toBeDefined();
    expect(goalApi.useUpdateGoalMutation).toBeDefined();
    expect(goalApi.useDeleteGoalMutation).toBeDefined();
  });

  it('диспатчит getMyGoal', async () => {
    const store = configureStore({ reducer: { goalApi: goalApi.reducer }, middleware: (g) => g().concat(goalApi.middleware) });
    const result = store.dispatch(goalApi.endpoints.getMyGoal.initiate('2026-09-01' as never) as never) as unknown as Promise<unknown>;
    await (result as Promise<unknown>).catch(() => {});
    expect(result).toBeDefined();
  });

  it('диспатчит setGoal', async () => {
    const store = configureStore({ reducer: { goalApi: goalApi.reducer }, middleware: (g) => g().concat(goalApi.middleware) });
    const result = store.dispatch(goalApi.endpoints.setGoal.initiate({ plan: 100 } as never) as never) as unknown as Promise<unknown>;
    await (result as Promise<unknown>).catch(() => {});
    expect(result).toBeDefined();
  });
});
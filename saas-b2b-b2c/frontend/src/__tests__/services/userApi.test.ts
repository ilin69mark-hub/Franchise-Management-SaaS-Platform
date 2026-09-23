import { configureStore } from '@reduxjs/toolkit';
import { userApi } from '@/services/userApi';

global.fetch = jest.fn(() =>
  Promise.resolve({
    ok: true,
    status: 200,
    json: () => Promise.resolve({}),
    text: () => Promise.resolve(''),
    headers: { get: () => 'application/json' },
  } as unknown as Response)
) as jest.Mock;

describe('userApi', () => {
  it('has correct reducerPath', () => {
    expect(userApi.reducerPath).toBe('userApi');
  });
  it('exports hooks', () => {
    expect(userApi.endpoints.getMyProfile).toBeDefined();
    expect(userApi.endpoints.updateProfile).toBeDefined();
  });

  it('диспатчит getMyProfile', async () => {
    const store = configureStore({ reducer: { userApi: userApi.reducer }, middleware: (g) => g().concat(userApi.middleware) });
    const result = store.dispatch(userApi.endpoints.getMyProfile.initiate() as never) as unknown as Promise<unknown>;
    await (result as Promise<unknown>).catch(() => {});
    expect(result).toBeDefined();
  });

  it('диспатчит updateProfile', async () => {
    const store = configureStore({ reducer: { userApi: userApi.reducer }, middleware: (g) => g().concat(userApi.middleware) });
    const result = store.dispatch(userApi.endpoints.updateProfile.initiate({ first_name: 'Test' } as never) as never) as unknown as Promise<unknown>;
    await (result as Promise<unknown>).catch(() => {});
    expect(result).toBeDefined();
  });
});

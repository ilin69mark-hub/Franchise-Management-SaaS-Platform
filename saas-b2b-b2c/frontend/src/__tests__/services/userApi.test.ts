import { userApi } from '@/services/userApi';

describe('userApi', () => {
  it('has correct reducerPath', () => {
    expect(userApi.reducerPath).toBe('userApi');
  });
  it('exports hooks', () => {
    expect(userApi.endpoints.getMyProfile).toBeDefined();
    expect(userApi.endpoints.updateProfile).toBeDefined();
  });
});

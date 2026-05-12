import { planApi } from '@/services/planApi';

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
});
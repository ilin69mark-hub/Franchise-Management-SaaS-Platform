import { goalApi } from '@/services/goalApi';

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
});
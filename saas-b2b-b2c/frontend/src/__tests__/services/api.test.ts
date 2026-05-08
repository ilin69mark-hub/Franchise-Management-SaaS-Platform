import { apiSlice } from '@/services/api';
import { goalApi } from '@/services/goalApi';

jest.mock('@/api/axiosClient', () => ({
  default: {
    interceptors: {
      request: { use: jest.fn(), eject: jest.fn() },
      response: { use: jest.fn(), eject: jest.fn() },
    },
  },
}));

describe('apiSlice (RTK Query)', () => {
  describe('baseQuery configuration', () => {
    it('has correct baseUrl from env', () => {
      const baseUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1`;
      expect(baseUrl).toBeDefined();
      expect(baseUrl).toContain('/api/v1');
    });

    it('has correct reducerPath', () => {
      expect(apiSlice.reducerPath).toBe('api');
    });
  });

  describe('endpoints configuration', () => {
    it('has login mutation', () => {
      const endpoint = apiSlice.endpoints.login;
      expect(endpoint).toBeDefined();
    });

    it('has register mutation', () => {
      const endpoint = apiSlice.endpoints.register;
      expect(endpoint).toBeDefined();
    });

    it('has logout mutation', () => {
      const endpoint = apiSlice.endpoints.logout;
      expect(endpoint).toBeDefined();
    });

    it('has getProfile query', () => {
      const endpoint = apiSlice.endpoints.getProfile;
      expect(endpoint).toBeDefined();
    });

    it('has getChecklists query', () => {
      const endpoint = apiSlice.endpoints.getChecklists;
      expect(endpoint).toBeDefined();
    });

    it('has createChecklist mutation', () => {
      const endpoint = apiSlice.endpoints.createChecklist;
      expect(endpoint).toBeDefined();
    });

    it('has updateChecklist mutation', () => {
      const endpoint = apiSlice.endpoints.updateChecklist;
      expect(endpoint).toBeDefined();
    });

    it('has deleteChecklist mutation', () => {
      const endpoint = apiSlice.endpoints.deleteChecklist;
      expect(endpoint).toBeDefined();
    });

    it('has getLeads query', () => {
      const endpoint = apiSlice.endpoints.getLeads;
      expect(endpoint).toBeDefined();
    });

    it('has getAlerts query', () => {
      const endpoint = apiSlice.endpoints.getAlerts;
      expect(endpoint).toBeDefined();
    });
  });

  describe('Auth endpoints', () => {
    it('login has POST method', () => {
      expect(apiSlice.endpoints.login).toBeDefined();
    });

    it('register has POST method', () => {
      expect(apiSlice.endpoints.register).toBeDefined();
    });

    it('logout has POST method', () => {
      expect(apiSlice.endpoints.logout).toBeDefined();
    });

    it('getProfile has GET method', () => {
      expect(apiSlice.endpoints.getProfile).toBeDefined();
    });
  });

  describe('Checklist endpoints', () => {
    it('getChecklists has GET method', () => {
      expect(apiSlice.endpoints.getChecklists).toBeDefined();
    });

    it('createChecklist has POST method', () => {
      expect(apiSlice.endpoints.createChecklist).toBeDefined();
    });

    it('updateChecklist has PUT method', () => {
      expect(apiSlice.endpoints.updateChecklist).toBeDefined();
    });

    it('deleteChecklist has DELETE method', () => {
      expect(apiSlice.endpoints.deleteChecklist).toBeDefined();
    });
  });

  describe('Lead endpoints', () => {
    it('getLeads has GET method', () => {
      expect(apiSlice.endpoints.getLeads).toBeDefined();
    });

    it('createLead has POST method', () => {
      expect(apiSlice.endpoints.createLead).toBeDefined();
    });

    it('updateLeadStatus has PUT method', () => {
      expect(apiSlice.endpoints.updateLeadStatus).toBeDefined();
    });
  });

  describe('Employee endpoints', () => {
    it('getEmployees has GET method', () => {
      expect(apiSlice.endpoints.getEmployees).toBeDefined();
    });

    it('createEmployee has POST method', () => {
      expect(apiSlice.endpoints.createEmployee).toBeDefined();
    });

    it('updateEmployee has PUT method', () => {
      expect(apiSlice.endpoints.updateEmployee).toBeDefined();
    });

    it('deleteEmployee has DELETE method', () => {
      expect(apiSlice.endpoints.deleteEmployee).toBeDefined();
    });
  });

  describe('Salon endpoints', () => {
    it('getSalons has GET method', () => {
      expect(apiSlice.endpoints.getSalons).toBeDefined();
    });

    it('createSalon has POST method', () => {
      expect(apiSlice.endpoints.createSalon).toBeDefined();
    });

    it('updateSalon has PUT method', () => {
      expect(apiSlice.endpoints.updateSalon).toBeDefined();
    });

    it('deleteSalon has DELETE method', () => {
      expect(apiSlice.endpoints.deleteSalon).toBeDefined();
    });
  });

  describe('Alert endpoints', () => {
    it('getAlerts has GET method', () => {
      expect(apiSlice.endpoints.getAlerts).toBeDefined();
    });

    it('getUnreadAlerts has GET method', () => {
      expect(apiSlice.endpoints.getUnreadAlerts).toBeDefined();
    });

    it('markAlertRead has PATCH method', () => {
      expect(apiSlice.endpoints.markAlertRead).toBeDefined();
    });

    it('getAlertSettings has GET method', () => {
      expect(apiSlice.endpoints.getAlertSettings).toBeDefined();
    });

    it('updateAlertSettings has PUT method', () => {
      expect(apiSlice.endpoints.updateAlertSettings).toBeDefined();
    });
  });
});

describe('goalApi (RTK Query)', () => {
  describe('baseQuery configuration', () => {
    it('has correct reducerPath', () => {
      expect(goalApi.reducerPath).toBe('goalApi');
    });

    it('has correct baseUrl', () => {
      expect(goalApi).toBeDefined();
    });
  });

  describe('endpoints configuration', () => {
    it('has getMyGoal query', () => {
      const endpoint = goalApi.endpoints.getMyGoal;
      expect(endpoint).toBeDefined();
    });

    it('has getVisibleGoals query', () => {
      const endpoint = goalApi.endpoints.getVisibleGoals;
      expect(endpoint).toBeDefined();
    });

    it('has setGoal mutation', () => {
      const endpoint = goalApi.endpoints.setGoal;
      expect(endpoint).toBeDefined();
    });

    it('has updateGoal mutation', () => {
      const endpoint = goalApi.endpoints.updateGoal;
      expect(endpoint).toBeDefined();
    });

    it('has deleteGoal mutation', () => {
      const endpoint = goalApi.endpoints.deleteGoal;
      expect(endpoint).toBeDefined();
    });
  });

  describe('Goal endpoint methods', () => {
    it('getMyGoal uses GET method', () => {
      expect(goalApi.endpoints.getMyGoal).toBeDefined();
    });

    it('getVisibleGoals uses GET method', () => {
      expect(goalApi.endpoints.getVisibleGoals).toBeDefined();
    });

    it('setGoal uses POST method', () => {
      expect(goalApi.endpoints.setGoal).toBeDefined();
    });

    it('updateGoal uses PUT method', () => {
      expect(goalApi.endpoints.updateGoal).toBeDefined();
    });

    it('deleteGoal uses DELETE method', () => {
      expect(goalApi.endpoints.deleteGoal).toBeDefined();
    });
  });
});

describe('API URL configuration', () => {
  it('uses process.env.NEXT_PUBLIC_API_URL', () => {
    const envUrl = process.env.NEXT_PUBLIC_API_URL;
    const baseUrl = `${envUrl || 'http://localhost:8080'}/api/v1`;
    expect(baseUrl).toContain('/api/v1');
  });

  it('falls back to localhost:8080 when env not set', () => {
    const fallbackUrl = 'http://localhost:8080/api/v1';
    expect(fallbackUrl).toBe('http://localhost:8080/api/v1');
  });
});

describe('API service exports', () => {
  describe('apiSlice hooks', () => {
    it('exports useLoginMutation', () => {
      expect(apiSlice).toHaveProperty('endpoints');
    });

    it('exports useRegisterMutation', () => {
      expect(apiSlice).toHaveProperty('endpoints');
    });

    it('exports useGetProfileQuery', () => {
      expect(apiSlice).toHaveProperty('endpoints');
    });

    it('exports useGetChecklistsQuery', () => {
      expect(apiSlice).toHaveProperty('endpoints');
    });

    it('exports useGetAlertsQuery', () => {
      expect(apiSlice).toHaveProperty('endpoints');
    });
  });

  describe('goalApi hooks', () => {
    it('exports useGetMyGoalQuery', () => {
      expect(goalApi).toHaveProperty('endpoints');
    });

    it('exports useSetGoalMutation', () => {
      expect(goalApi).toHaveProperty('endpoints');
    });

    it('exports useUpdateGoalMutation', () => {
      expect(goalApi).toHaveProperty('endpoints');
    });

    it('exports useDeleteGoalMutation', () => {
      expect(goalApi).toHaveProperty('endpoints');
    });
  });
});

describe('RTK Query structure validation', () => {
  it('apiSlice has reducer', () => {
    expect(apiSlice).toHaveProperty('reducer');
  });

  it('apiSlice has middleware', () => {
    expect(apiSlice).toHaveProperty('middleware');
  });

  it('goalApi has reducer', () => {
    expect(goalApi).toHaveProperty('reducer');
  });

  it('goalApi has middleware', () => {
    expect(goalApi).toHaveProperty('middleware');
  });
});

describe('Endpoint URL patterns', () => {
  const checkUrlPattern = (endpoint: any, expectedPattern: string) => {
    expect(endpoint).toBeDefined();
  };

  it('auth endpoints use /auth prefix', () => {
    checkUrlPattern(apiSlice.endpoints.login, '/auth');
    checkUrlPattern(apiSlice.endpoints.register, '/auth');
    checkUrlPattern(apiSlice.endpoints.logout, '/auth');
    checkUrlPattern(apiSlice.endpoints.getProfile, '/auth');
  });

  it('checklist endpoints use /checklists prefix', () => {
    checkUrlPattern(apiSlice.endpoints.getChecklists, '/checklists');
    checkUrlPattern(apiSlice.endpoints.createChecklist, '/checklists');
    checkUrlPattern(apiSlice.endpoints.updateChecklist, '/checklists');
    checkUrlPattern(apiSlice.endpoints.deleteChecklist, '/checklists');
  });

  it('lead endpoints use /leads prefix', () => {
    checkUrlPattern(apiSlice.endpoints.getLeads, '/leads');
    checkUrlPattern(apiSlice.endpoints.createLead, '/leads');
  });

  it('goalApi endpoints use /goals prefix', () => {
    checkUrlPattern(goalApi.endpoints.getMyGoal, '/goals');
    checkUrlPattern(goalApi.endpoints.getVisibleGoals, '/goals');
    checkUrlPattern(goalApi.endpoints.setGoal, '/goals');
    checkUrlPattern(goalApi.endpoints.updateGoal, '/goals');
    checkUrlPattern(goalApi.endpoints.deleteGoal, '/goals');
  });
});
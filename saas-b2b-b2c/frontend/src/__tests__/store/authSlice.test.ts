import { configureStore } from '@reduxjs/toolkit';
import authReducer, { login, register, logout, setAuthFromStorage } from '@/store/authSlice';
import apiClient from '@/api/axiosClient';

jest.mock('@/api/axiosClient', () => ({
  post: jest.fn(),
  get: jest.fn(),
}));

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('authSlice', () => {
  let store: ReturnType<typeof configureStore>;

  beforeEach(() => {
    store = configureStore({
      reducer: {
        auth: authReducer,
      },
    });
    jest.clearAllMocks();
    localStorage.clear();
  });

  describe('initial state', () => {
    it('has user as null', () => {
      expect(store.getState().auth.user).toBeNull();
    });

    it('has accessToken as null', () => {
      expect(store.getState().auth.accessToken).toBeNull();
    });

    it('has refreshToken as null', () => {
      expect(store.getState().auth.refreshToken).toBeNull();
    });

    it('has isAuthenticated as false', () => {
      expect(store.getState().auth.isAuthenticated).toBe(false);
    });

    it('has loading as false', () => {
      expect(store.getState().auth.loading).toBe(false);
    });

    it('has error as null', () => {
      expect(store.getState().auth.error).toBeNull();
    });
  });

  describe('logout action', () => {
    it('clears user', () => {
      store.dispatch(logout());
      expect(store.getState().auth.user).toBeNull();
    });

    it('clears accessToken', () => {
      store.dispatch(logout());
      expect(store.getState().auth.accessToken).toBeNull();
    });

    it('clears refreshToken', () => {
      store.dispatch(logout());
      expect(store.getState().auth.refreshToken).toBeNull();
    });

    it('sets isAuthenticated to false', () => {
      store.dispatch(logout());
      expect(store.getState().auth.isAuthenticated).toBe(false);
    });

    it('clears error', () => {
      store.dispatch(logout());
      expect(store.getState().auth.error).toBeNull();
    });

    it('clears localStorage items', () => {
      localStorage.setItem('accessToken', 'test-token');
      localStorage.setItem('user', '{"id": "1"}');
      localStorage.setItem('id', '1');
      localStorage.setItem('role', 'admin');

      store.dispatch(logout());

      expect(localStorage.getItem('accessToken')).toBeNull();
      expect(localStorage.getItem('user')).toBeNull();
      expect(localStorage.getItem('id')).toBeNull();
      expect(localStorage.getItem('role')).toBeNull();
    });
  });

  describe('setAuthFromStorage action', () => {
    it('restores auth from localStorage when token exists', () => {
      const mockUser = { id: 'user-1', email: 'test@example.com', role: 'admin' };
      localStorage.setItem('accessToken', 'stored-token');
      localStorage.setItem('user', JSON.stringify(mockUser));
      localStorage.setItem('id', 'user-1');
      localStorage.setItem('role', 'admin');

      store.dispatch(setAuthFromStorage());

      expect(store.getState().auth.accessToken).toBe('stored-token');
      expect(store.getState().auth.isAuthenticated).toBe(true);
    });

    it('does not restore when no token in localStorage', () => {
      store.dispatch(setAuthFromStorage());

      expect(store.getState().auth.accessToken).toBeNull();
      expect(store.getState().auth.isAuthenticated).toBe(false);
    });

    it('restores user with id and role from localStorage', () => {
      const mockUser = { email: 'test@example.com' };
      localStorage.setItem('accessToken', 'stored-token');
      localStorage.setItem('user', JSON.stringify(mockUser));
      localStorage.setItem('id', 'user-123');
      localStorage.setItem('role', 'dealer');

      store.dispatch(setAuthFromStorage());

      const user = store.getState().auth.user as any;
      expect(user.id).toBe('user-123');
      expect(user.role).toBe('dealer');
    });
  });

  describe('login thunk', () => {
    const mockUser = {
      id: 'user-1',
      email: 'test@example.com',
      firstName: 'John',
      lastName: 'Doe',
      role: 'dealer',
    };

    const mockResponse = {
      Token: 'access-token-123',
      RefreshToken: 'refresh-token-456',
      user: mockUser,
    };

    it('sets loading = true on pending', async () => {
      mockApiClient.post.mockImplementation(() => new Promise(() => {}));

      store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.loading).toBe(true);
      expect(store.getState().auth.error).toBeNull();
    });

    it('sets user, token, isAuthenticated = true on fulfilled', async () => {
      mockApiClient.post.mockResolvedValue({ data: mockResponse });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.loading).toBe(false);
      expect(store.getState().auth.isAuthenticated).toBe(true);
      expect(store.getState().auth.accessToken).toBe('access-token-123');
      expect(store.getState().auth.refreshToken).toBe('refresh-token-456');
      expect(store.getState().auth.user).toEqual(mockUser);
    });

    it('saves to localStorage on successful login', async () => {
      mockApiClient.post.mockResolvedValue({ data: mockResponse });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(localStorage.getItem('accessToken')).toBe('access-token-123');
      expect(localStorage.getItem('user')).toContain('test@example.com');
    });

    it('saves user id and role to localStorage', async () => {
      mockApiClient.post.mockResolvedValue({ data: mockResponse });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(localStorage.getItem('id')).toBe('user-1');
      expect(localStorage.getItem('role')).toBe('dealer');
    });

    it('handles different token formats (accessToken)', async () => {
      const responseWithAccessToken = {
        accessToken: 'access-token-via-accessToken',
        refreshToken: 'refresh-token-via-refreshToken',
        user: mockUser,
      };
      mockApiClient.post.mockResolvedValue({ data: responseWithAccessToken });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.accessToken).toBe('access-token-via-accessToken');
    });

    it('handles different token formats (token)', async () => {
      const responseWithToken = {
        token: 'access-token-via-token',
        refresh_token: 'refresh-token-via-underscore',
        user: mockUser,
      };
      mockApiClient.post.mockResolvedValue({ data: responseWithToken });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.accessToken).toBe('access-token-via-token');
    });

    it('sets error message on rejected', async () => {
      mockApiClient.post.mockRejectedValue({
        response: { data: { error: 'Invalid credentials' } },
      });

      await store.dispatch(login({ email: 'test@example.com', password: 'wrong' }));

      expect(store.getState().auth.loading).toBe(false);
      expect(store.getState().auth.error).toBe('Invalid credentials');
      expect(store.getState().auth.isAuthenticated).toBe(false);
    });

    it('sets default error message when no specific error', async () => {
      mockApiClient.post.mockRejectedValue(new Error('Network error'));

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.error).toBe('Ошибка входа');
    });

    it('sets isAuthenticated = false when token not in response', async () => {
      mockApiClient.post.mockResolvedValue({ data: { user: mockUser } });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.isAuthenticated).toBe(false);
      expect(store.getState().auth.error).toBe('Ошибка авторизации: токен не получен');
    });
  });

  describe('register thunk', () => {
    const mockUser = {
      id: 'user-new',
      email: 'new@example.com',
      firstName: 'New',
      lastName: 'User',
      role: 'dealer',
    };

    const mockRegisterResponse = {
      Token: 'new-access-token',
      RefreshToken: 'new-refresh-token',
      user: mockUser,
    };

    it('sets loading = true on pending', async () => {
      mockApiClient.post.mockImplementation(() => new Promise(() => {}));

      store.dispatch(register({ email: 'new@example.com', password: 'password123' }));

      expect(store.getState().auth.loading).toBe(true);
    });

    it('sets user, token, isAuthenticated = true on fulfilled', async () => {
      mockApiClient.post.mockResolvedValue({ data: mockRegisterResponse });

      await store.dispatch(register({ email: 'new@example.com', password: 'password123' }));

      expect(store.getState().auth.loading).toBe(false);
      expect(store.getState().auth.isAuthenticated).toBe(true);
      expect(store.getState().auth.accessToken).toBe('new-access-token');
      expect(store.getState().auth.user).toEqual(mockUser);
    });

    it('saves to localStorage on successful registration', async () => {
      mockApiClient.post.mockResolvedValue({ data: mockRegisterResponse });

      await store.dispatch(register({ email: 'new@example.com', password: 'password123' }));

      expect(localStorage.getItem('accessToken')).toBe('new-access-token');
    });

    it('sets error message on rejected', async () => {
      mockApiClient.post.mockRejectedValue({
        response: { data: { error: 'Email already exists' } },
      });

      await store.dispatch(register({ email: 'exists@example.com', password: 'password' }));

      expect(store.getState().auth.loading).toBe(false);
      expect(store.getState().auth.error).toBe('Email already exists');
    });

    it('sets isAuthenticated = false when token not in response', async () => {
      mockApiClient.post.mockResolvedValue({ data: { user: mockUser } });

      await store.dispatch(register({ email: 'new@example.com', password: 'password' }));

      expect(store.getState().auth.isAuthenticated).toBe(false);
      expect(store.getState().auth.error).toBe('Ошибка регистрации: токен не получен');
    });
  });
});

describe('authSlice selectors', () => {
  it('selectCurrentUser returns user from state', () => {
    const preloadedState = {
      auth: {
        user: { id: 'user-1', email: 'test@example.com', role: 'admin' },
        accessToken: 'token-123',
        refreshToken: null,
        isAuthenticated: true,
        loading: false,
        error: null,
      },
    };

    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: preloadedState as any,
    });

    const selectCurrentUser = (state: any) => state.auth.user;
    expect(selectCurrentUser(store.getState())).toEqual({ id: 'user-1', email: 'test@example.com', role: 'admin' });
  });

  it('selectIsAuthenticated returns isAuthenticated from state', () => {
    const preloadedState = {
      auth: {
        user: { id: 'user-1', email: 'test@example.com' },
        accessToken: 'token-123',
        refreshToken: 'refresh-123',
        isAuthenticated: true,
        loading: false,
        error: null,
      },
    };

    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: preloadedState as any,
    });

    const selectIsAuthenticated = (state: any) => state.auth.isAuthenticated;
    expect(selectIsAuthenticated(store.getState())).toBe(true);
  });

  it('selectToken returns accessToken from state', () => {
    const preloadedState = {
      auth: {
        user: { id: 'user-1', email: 'test@example.com' },
        accessToken: 'token-123',
        refreshToken: 'refresh-456',
        isAuthenticated: true,
        loading: false,
        error: null,
      },
    };

    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: preloadedState as any,
    });

    const selectToken = (state: any) => state.auth.accessToken;
    expect(selectToken(store.getState())).toBe('token-123');
  });

  it('selectToken returns null when not authenticated', () => {
    const store = configureStore({
      reducer: { auth: authReducer },
    });

    const selectToken = (state: any) => state.auth.accessToken;
    expect(selectToken(store.getState())).toBeNull();
  });
});
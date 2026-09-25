import { configureStore } from '@reduxjs/toolkit';
import authReducer, { login, register, logout, setAuthFromStorage } from '@/store/authSlice';
import apiClient from '@/api/axiosClient';

jest.mock('@/api/axiosClient', () => ({
  post: jest.fn(),
  get: jest.fn(),
}));

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

// F7: сессия в httpOnly cookie, токенов в сторе/localStorage/теле больше нет.
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

    it('has no token fields (F7 regression)', () => {
      const state = store.getState().auth as Record<string, unknown>;
      expect('accessToken' in state).toBe(false);
      expect('refreshToken' in state).toBe(false);
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

    it('sets isAuthenticated to false', () => {
      store.dispatch(logout());
      expect(store.getState().auth.isAuthenticated).toBe(false);
    });

    it('clears error', () => {
      store.dispatch(logout());
      expect(store.getState().auth.error).toBeNull();
    });

    it('clears cached user and legacy token keys', () => {
      localStorage.setItem('accessToken', 'test-token');
      localStorage.setItem('refreshToken', 'test-refresh');
      localStorage.setItem('user', '{"id": "1"}');
      localStorage.setItem('id', '1');
      localStorage.setItem('role', 'admin');

      store.dispatch(logout());

      expect(localStorage.getItem('accessToken')).toBeNull();
      expect(localStorage.getItem('refreshToken')).toBeNull();
      expect(localStorage.getItem('user')).toBeNull();
      expect(localStorage.getItem('id')).toBeNull();
      expect(localStorage.getItem('role')).toBeNull();
    });
  });

  describe('setAuthFromStorage action', () => {
    it('restores auth from cached user without any token', () => {
      const mockUser = { id: 'user-1', email: 'test@example.com', role: 'admin' };
      localStorage.setItem('user', JSON.stringify(mockUser));

      store.dispatch(setAuthFromStorage());

      expect(store.getState().auth.user).toEqual(mockUser);
      expect(store.getState().auth.isAuthenticated).toBe(true);
    });

    it('does not restore when no cached user', () => {
      store.dispatch(setAuthFromStorage());

      expect(store.getState().auth.user).toBeNull();
      expect(store.getState().auth.isAuthenticated).toBe(false);
    });

    it('does not restore user without id/role', () => {
      localStorage.setItem('user', JSON.stringify({ email: 'test@example.com' }));

      store.dispatch(setAuthFromStorage());

      expect(store.getState().auth.isAuthenticated).toBe(false);
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

    it('sets loading = true on pending', async () => {
      mockApiClient.post.mockImplementation(() => new Promise(() => {}));

      store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.loading).toBe(true);
      expect(store.getState().auth.error).toBeNull();
    });

    it('sets user and isAuthenticated = true on fulfilled with {user} only', async () => {
      mockApiClient.post.mockResolvedValue({ data: { user: mockUser } });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/login', { email: 'test@example.com', password: 'password' });
      expect(store.getState().auth.loading).toBe(false);
      expect(store.getState().auth.isAuthenticated).toBe(true);
      expect(store.getState().auth.user).toEqual(mockUser);
    });

    it('caches user but never tokens on successful login', async () => {
      mockApiClient.post.mockResolvedValue({ data: { user: mockUser } });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(localStorage.getItem('accessToken')).toBeNull();
      expect(localStorage.getItem('user')).toContain('test@example.com');
      expect(localStorage.getItem('id')).toBe('user-1');
      expect(localStorage.getItem('role')).toBe('dealer');
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

    it('sets isAuthenticated = false when user missing in response', async () => {
      mockApiClient.post.mockResolvedValue({ data: {} });

      await store.dispatch(login({ email: 'test@example.com', password: 'password' }));

      expect(store.getState().auth.isAuthenticated).toBe(false);
      expect(store.getState().auth.error).toBe('Ошибка авторизации: профиль не получен');
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

    it('sets loading = true on pending', async () => {
      mockApiClient.post.mockImplementation(() => new Promise(() => {}));

      store.dispatch(register({ email: 'new@example.com', password: 'password123' }));

      expect(store.getState().auth.loading).toBe(true);
    });

    it('sets user and isAuthenticated = true on fulfilled with {user} only', async () => {
      mockApiClient.post.mockResolvedValue({ data: { user: mockUser } });

      await store.dispatch(register({ email: 'new@example.com', password: 'password123' }));

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/register', { email: 'new@example.com', password: 'password123' });
      expect(store.getState().auth.loading).toBe(false);
      expect(store.getState().auth.isAuthenticated).toBe(true);
      expect(store.getState().auth.user).toEqual(mockUser);
    });

    it('never stores tokens on successful registration', async () => {
      mockApiClient.post.mockResolvedValue({ data: { user: mockUser } });

      await store.dispatch(register({ email: 'new@example.com', password: 'password123' }));

      expect(localStorage.getItem('accessToken')).toBeNull();
      const state = store.getState().auth as Record<string, unknown>;
      expect('accessToken' in state).toBe(false);
    });

    it('sets error message on rejected', async () => {
      mockApiClient.post.mockRejectedValue({
        response: { data: { error: 'Email already exists' } },
      });

      await store.dispatch(register({ email: 'exists@example.com', password: 'password' }));

      expect(store.getState().auth.loading).toBe(false);
      expect(store.getState().auth.error).toBe('Email already exists');
    });

    it('не считает отсутствие user ошибкой (анти-enumeration ответ 202)', async () => {
      // REAUDIT-3: бэкенд отвечает одинаковым 202 {"message"} и при успехе,
      // и при занятом email, чтобы нельзя было перечислить пользователей.
      mockApiClient.post.mockResolvedValue({ data: { message: 'If this email is available, the account has been created' } });

      await store.dispatch(register({ email: 'new@example.com', password: 'password' }));

      expect(store.getState().auth.isAuthenticated).toBe(false);
      expect(store.getState().auth.error).toBeNull();
    });
  });
});

describe('authSlice selectors', () => {
  it('selectCurrentUser returns user from state', () => {
    const preloadedState = {
      auth: {
        user: { id: 'user-1', email: 'test@example.com', role: 'admin' },
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

  it('has no token selector — tokens do not exist in state (F7)', () => {
    const store = configureStore({
      reducer: { auth: authReducer },
    });

    expect((store.getState().auth as any).accessToken).toBeUndefined();
    expect((store.getState().auth as any).refreshToken).toBeUndefined();
  });
});

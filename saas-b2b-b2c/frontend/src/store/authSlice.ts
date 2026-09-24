import logger from '@/utils/logger';
import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { User, AuthResponse } from '@/types';
import apiClient from '@/api/axiosClient';

// F7: сессия живёт в httpOnly cookie (бэкенд), в сторе — только профиль.
// Токены здесь НЕ хранятся (ни в памяти стора, ни в localStorage): раньше
// их крал любой XSS через localStorage.accessToken / тело ответа / WS-URL.
interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  loading: false,
  error: null,
};

function cacheUser(user: User) {
  if (typeof window === 'undefined') return;
  try {
    // Кэшируем ТОЛЬКО профиль (не секрет) — для мгновенного гейтинга страниц
    // до первой серверной проверки. Источник правды — /auth/me + API 401.
    localStorage.setItem('user', JSON.stringify(user));
    if (user?.id) localStorage.setItem('id', user.id);
    if (user?.role) localStorage.setItem('role', user.role);
  } catch (e) {
    logger.error('Failed to cache user', e);
  }
}

function clearUserCache() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem('user');
  localStorage.removeItem('id');
  localStorage.removeItem('role');
  localStorage.removeItem('userId');
  localStorage.removeItem('user_id');
  localStorage.removeItem('reduxState');
  // Legacy-ключи эпохи токенов — зачищаем при выходе/старте.
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
}

/* ---------- Асинхронные Thunk‑ы ---------- */
export const login = createAsyncThunk(
  'auth/login',
  async (
    credentials: { email: string; password: string; captcha_token?: string },
    { rejectWithValue },
  ) => {
    try {
      const response = await apiClient.post<AuthResponse>('/auth/login', credentials);
      return response.data;
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string } } };
      return rejectWithValue(e.response?.data?.error || 'Ошибка входа');
    }
  },
);

export const register = createAsyncThunk(
  'auth/register',
  async (
    userData: { email: string; password: string; role?: string },
    { rejectWithValue },
  ) => {
    try {
      const response = await apiClient.post<AuthResponse>('/auth/register', userData);
      return response.data;
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string } } };
      return rejectWithValue(e.response?.data?.error || 'Ошибка регистрации');
    }
  },
);

/* ---------- Слайс ---------- */
const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    logout: (state) => {
      state.user = null;
      state.isAuthenticated = false;
      state.error = null;
      clearUserCache();
    },
    setAuthFromStorage: (state) => {
      if (typeof window !== 'undefined') {
        const userStr = localStorage.getItem('user');
        if (!userStr) return;
        try {
          const user = JSON.parse(userStr) as User;
          if (user && user.id && user.role) {
            state.user = user;
            state.isAuthenticated = true;
          }
        } catch (e) {
          logger.error('Failed to restore user', e);
          clearUserCache();
        }
      }
    },
  },

  extraReducers: (builder) => {
    /* ---------- Login ---------- */
    builder.addCase(login.pending, (state) => {
      state.loading = true;
      state.error = null;
    });
    builder.addCase(login.fulfilled, (state, { payload }) => {
      state.loading = false;
      // F7: бэкенд возвращает {user}, сессия — в cookie. Токен в теле больше
      // не ждём и не требуем (раньше отсутствие токена считалось ошибкой).
      const user = (payload as AuthResponse).user;
      if (user && user.id) {
        state.user = user;
        state.isAuthenticated = true;
        cacheUser(user);
      } else {
        logger.error('User not found in login response', payload);
        state.error = 'Ошибка авторизации: профиль не получен';
        state.isAuthenticated = false;
      }
    });
    builder.addCase(login.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload as string;
    });

    /* ---------- Register ---------- */
    builder.addCase(register.pending, (state) => {
      state.loading = true;
      state.error = null;
    });
    builder.addCase(register.fulfilled, (state, { payload }) => {
      state.loading = false;
      const user = (payload as AuthResponse).user;
      if (user && user.id) {
        state.user = user;
        state.isAuthenticated = true;
        cacheUser(user);
      } else {
        logger.error('User not found in register response', payload);
        state.error = 'Ошибка регистрации: профиль не получен';
        state.isAuthenticated = false;
      }
    });
    builder.addCase(register.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload as string;
    });
  },
});

export const { logout, setAuthFromStorage } = authSlice.actions;
export default authSlice.reducer;

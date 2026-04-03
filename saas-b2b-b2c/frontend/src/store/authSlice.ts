import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { User, AuthResponse } from '@/types';
import apiClient from '@/api/axiosClient';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  user: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,
  loading: false,
  error: null,
};

// === АСИНХРОННЫЕ THUNK ===

export const login = createAsyncThunk(
  'auth/login',
  async (credentials: { email: string; password: string }, { rejectWithValue }) => {
    try {
      const response = await apiClient.post<AuthResponse>('/auth/login', credentials);
      return response.data;
    } catch (err: any) {
      return rejectWithValue(err.response?.data?.error || 'Ошибка входа');
    }
  }
);

export const register = createAsyncThunk(
  'auth/register',
  async (userData: { email: string; password: string; role?: string }, { rejectWithValue }) => {
    try {
      const response = await apiClient.post<AuthResponse>('/auth/register', userData);
      return response.data;
    } catch (err: any) {
      return rejectWithValue(err.response?.data?.error || 'Ошибка регистрации');
    }
  }
);

// === SLICE ===

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    logout: (state) => {
      state.user = null;
      state.accessToken = null;
      state.refreshToken = null;
      state.isAuthenticated = false;
      state.error = null;
      if (typeof window !== 'undefined') {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('user');
        localStorage.removeItem('reduxState');
      }
    },
    setAuthFromStorage: (state) => {
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('accessToken');
            const user = localStorage.getItem('user');
            if (token && user) {
                state.accessToken = token;
                state.user = JSON.parse(user);
                state.isAuthenticated = true;
            }
        }
    }
  },
  extraReducers: (builder) => {
    // Login
    builder.addCase(login.pending, (state) => {
      state.loading = true;
      state.error = null;
    });
    builder.addCase(login.fulfilled, (state, action) => {
      state.loading = false;
      state.isAuthenticated = true;
      
      // ОТЛАДКА: Смотрим весь ответ от сервера
      console.log('Login response payload:', action.payload);

      // Приводим к any, чтобы обойти строгую типизацию
      const payload: any = action.payload;
      
      // Ищем токен в разных форматах
      const token = payload.Token || payload.accessToken || payload.token;
      const refresh = payload.RefreshToken || payload.refreshToken || payload.refresh_token;

      if (token) {
        state.accessToken = token;
        state.refreshToken = refresh || null;
        state.user = payload.user;

        if (typeof window !== 'undefined') {
          localStorage.setItem('accessToken', token);
          localStorage.setItem('user', JSON.stringify(payload.user));
        }
      } else {
        console.error('Token not found in response!', payload);
        state.error = 'Ошибка авторизации: токен не получен';
        state.isAuthenticated = false;
      }
    });
    builder.addCase(login.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload as string;
    });

    // Register
    builder.addCase(register.pending, (state) => {
      state.loading = true;
      state.error = null;
    });
    builder.addCase(register.fulfilled, (state, action) => {
      state.loading = false;
      state.isAuthenticated = true;
      
      const payload: any = action.payload;
      const token = payload.Token || payload.accessToken || payload.token;
      const refresh = payload.RefreshToken || payload.refreshToken;

      if (token) {
        state.accessToken = token;
        state.refreshToken = refresh || null;
        state.user = payload.user;
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

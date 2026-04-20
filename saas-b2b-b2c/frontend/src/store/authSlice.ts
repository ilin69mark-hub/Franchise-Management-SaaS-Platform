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

/* ---------- Асинхронные Thunk‑ы ---------- */
export const login = createAsyncThunk(
  'auth/login',
  async (
    credentials: { email: string; password: string },
    { rejectWithValue },
  ) => {
    try {
      const response = await apiClient.post<AuthResponse>('/auth/login', credentials);
      return response.data;
    } catch (err: any) {
      return rejectWithValue(err.response?.data?.error || 'Ошибка входа');
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
    } catch (err: any) {
      return rejectWithValue(err.response?.data?.error || 'Ошибка регистрации');
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
      state.accessToken = null;
      state.refreshToken = null;
      state.isAuthenticated = false;
      state.error = null;
      if (typeof window !== 'undefined') {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('user');
        localStorage.removeItem('id');
        localStorage.removeItem('role');
        localStorage.removeItem('reduxState');
      }
    },
    setAuthFromStorage: (state) => {
      if (typeof window !== 'undefined') {
        const token = localStorage.getItem('accessToken');
        const userStr = localStorage.getItem('user');
        const id = localStorage.getItem('id');
        const role = localStorage.getItem('role');
        if (token && userStr) {
          state.accessToken = token;
          state.user = JSON.parse(userStr);
          state.isAuthenticated = true;
          // Если id/role есть в хранилище – восстанавливаем их в state
          if (id) (state.user as any).id = id;
          if (role) (state.user as any).role = role;
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
      state.isAuthenticated = true;

      // Payload может быть в разных формах – приводим к any
      const data: any = payload;

      // Токены могут быть `token` / `accessToken` / `Token`
      const token = data.Token || data.accessToken || data.token;
      const refresh = data.RefreshToken || data.refreshToken || data.refresh_token;

      if (token) {
        state.accessToken = token;
        state.refreshToken = refresh || null;
        state.user = data.user;

        if (typeof window !== 'undefined') {
          localStorage.setItem('accessToken', token);
          localStorage.setItem('user', JSON.stringify(data.user));

          // **Сохраняем id и role в localStorage**
          if (data.user?.id) localStorage.setItem('id', data.user.id);
          if (data.user?.role) localStorage.setItem('role', data.user.role);
        }
      } else {
        console.error('Token not found in login response', payload);
        state.error = 'Ошибка авторизации: токен не получен';
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
      state.isAuthenticated = true;

      const data: any = payload;
      const token = data.Token || data.accessToken || data.token;
      const refresh = data.RefreshToken || data.refreshToken || data.refresh_token;

      if (token) {
        state.accessToken = token;
        state.refreshToken = refresh || null;
        state.user = data.user;

        if (typeof window !== 'undefined') {
          localStorage.setItem('accessToken', token);
          localStorage.setItem('user', JSON.stringify(data.user));

          // **Сохраняем id и role после регистрации**
          if (data.user?.id) localStorage.setItem('id', data.user.id);
          if (data.user?.role) localStorage.setItem('role', data.user.role);
        }
      } else {
        console.error('Token not found in register response', payload);
        state.error = 'Ошибка регистрации: токен не получен';
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

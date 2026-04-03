// src/store/index.ts
import { configureStore } from '@reduxjs/toolkit';
import authReducer from './authSlice';
import checklistReducer from './checklistSlice';
// ИМПОРТИРУЕМ API SLICE
import { apiSlice } from '@/services/api';

// Безопасная обертка для localStorage (для SSR в Next.js)
const safeLocalStorage = {
  getItem: (key: string) => {
    if (typeof window === 'undefined') return null;
    try {
      return localStorage.getItem(key);
    } catch (error) {
      console.error('Error reading from localStorage', error);
      return null;
    }
  },
  setItem: (key: string, value: string) => {
    if (typeof window === 'undefined') return;
    try {
      localStorage.setItem(key, value);
    } catch (error) {
      console.error('Error writing to localStorage', error);
    }
  },
  removeItem: (key: string) => {
    if (typeof window === 'undefined') return;
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.error('Error removing from localStorage', error);
    }
  }
};

const loadState = () => {
  try {
    const serializedState = safeLocalStorage.getItem('reduxState');
    if (!serializedState) return undefined;
    const parsed = JSON.parse(serializedState);
    if (!parsed || typeof parsed !== 'object') {
        safeLocalStorage.removeItem('reduxState');
        return undefined;
    }
    return parsed;
  } catch (err) {
    console.error('Error loading state:', err);
    safeLocalStorage.removeItem('reduxState');
    return undefined;
  }
};

const preloadedState = loadState();

export const store = configureStore({
  reducer: {
    auth: authReducer,
    checklist: checklistReducer,
    // ДОБАВЛЯЕМ РЕДЮСЕР API
    [apiSlice.reducerPath]: apiSlice.reducer,
  },
  // ДОБАВЛЯЕМ MIDDLEWARE ДЛЯ RTK QUERY (ОБЯЗАТЕЛЬНО!)
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(apiSlice.middleware),
    
  preloadedState,
});

// Подписка на изменения
if (typeof window !== 'undefined') {
  store.subscribe(() => {
    try {
      const state = store.getState();
      if (state.auth) {
        const serializedState = JSON.stringify({
          auth: {
            user: state.auth.user,
            accessToken: state.auth.accessToken,
            refreshToken: state.auth.refreshToken,
            isAuthenticated: state.auth.isAuthenticated,
          }
        });
        safeLocalStorage.setItem('reduxState', serializedState);
      }
    } catch (err) {
      console.error('Could not save state', err);
    }
  });
}

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

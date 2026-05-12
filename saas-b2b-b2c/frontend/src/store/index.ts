import { configureStore } from '@reduxjs/toolkit';
import authReducer from './authSlice';
import checklistReducer from './checklistSlice';

// ------------------------------------------------------------------
//  RTK‑Query‑сервисы
// ------------------------------------------------------------------
import { apiSlice } from '@/services/api';          // общий API‑slice
import { planApi } from '@/services/planApi';      // сервис тарифных планов
import { goalApi } from '@/services/goalApi';      // <-- НОВОЕ: сервис целей (Goal)
import { userApi } from '@/services/userApi';      // <-- НОВОЕ: сервис профиля пользователя

/* -----------------------------------------------------------------
   Безопасная обертка localStorage (для SSR в Next.js)
----------------------------------------------------------------- */
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
  },
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

/* -----------------------------------------------------------------
   Store
----------------------------------------------------------------- */
export const store = configureStore({
  reducer: {
    auth: authReducer,
    checklist: checklistReducer,
    // RTK‑Query‑сервисы
    [apiSlice.reducerPath]: apiSlice.reducer,
    [planApi.reducerPath]: planApi.reducer,
    [goalApi.reducerPath]: goalApi.reducer,   // <-- НОВОЕ
    [userApi.reducerPath]: userApi.reducer,    // <-- НОВОЕ
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(apiSlice.middleware)
      .concat(planApi.middleware)
      .concat(goalApi.middleware)
      .concat(userApi.middleware),           // <-- НОВОЕ
  preloadedState,
});

/* -----------------------------------------------------------------
   Сохранение части state в localStorage (auth‑данные)
----------------------------------------------------------------- */
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
          },
        });
        safeLocalStorage.setItem('reduxState', serializedState);
      }
    } catch (err) {
      console.error('Could not save state', err);
    }
  });
}

/* -----------------------------------------------------------------
   Types
----------------------------------------------------------------- */
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

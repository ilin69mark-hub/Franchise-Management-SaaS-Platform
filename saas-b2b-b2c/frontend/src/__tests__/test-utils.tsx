import React, { ReactElement, ReactNode } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore, EnhancedStore } from '@reduxjs/toolkit';
import { BrowserRouter } from 'react-router-dom';
import authReducer from '@/store/authSlice';
import checklistReducer from '@/store/checklistSlice';

export type RootState = {
  auth: ReturnType<typeof authReducer>;
  checklist: ReturnType<typeof checklistReducer>;
};

export type AppStore = EnhancedStore<RootState>;

export function createTestStore(preloadedState?: Partial<RootState>): AppStore {
  return configureStore({
    reducer: {
      auth: authReducer,
      checklist: checklistReducer,
    },
    preloadedState: {
      auth: {
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        loading: false,
        error: null,
      },
      checklist: {
        items: [],
        loading: false,
        error: null,
        currentChecklist: null,
      },
      ...preloadedState,
    },
  });
}

interface WrapperProps {
  children: ReactNode;
  store?: AppStore;
}

function TestWrapper({ children, store }: WrapperProps) {
  const testStore = store || createTestStore();
  
  return (
    <Provider store={testStore}>
      <BrowserRouter>
        {children}
      </BrowserRouter>
    </Provider>
  );
}

export function renderWithProviders(
  ui: ReactElement,
  { store = createTestStore(), ...renderOptions } = {}
) {
  return {
    ...render(ui, {
      wrapper: ({ children }) => <TestWrapper store={store}>{children}</TestWrapper>,
      ...renderOptions,
    }),
    store,
  };
}

export * from '@testing-library/react';
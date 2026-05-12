import React from 'react';
import '../styles/globals.css';
import type { AppProps } from 'next/app';
import { Provider } from 'react-redux';
import { store } from '../store';
import { useState, useEffect } from 'react';
import ThemeProvider from '../components/ThemeProvider';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import weekday from 'dayjs/plugin/weekday';
import localeData from 'dayjs/plugin/localeData';

dayjs.locale('ru');
dayjs.extend(weekday);
dayjs.extend(localeData);

function MyApp({ Component, pageProps }: AppProps) {
  const [showChild, setShowChild] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    const userStr = localStorage.getItem('user');
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr);
        if (user && user.role) {
          store.dispatch({ type: 'auth/setAuthFromStorage' });
        }
      } catch (e) {
        console.error('Failed to parse user', e);
      }
    }
    setShowChild(true);
  }, []);

  if (!showChild) {
    return null;
  }

  return (
    <Provider store={store}>
      <ThemeProvider>
        <Component {...pageProps} />
      </ThemeProvider>
    </Provider>
  );
}

export default MyApp;

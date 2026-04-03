import '../styles/globals.css';
import type { AppProps } from 'next/app';
import { Provider } from 'react-redux';
import { store } from '../store';
import { useState, useEffect } from 'react';
import { Spin } from 'antd';

function MyApp({ Component, pageProps }: AppProps) {
  const [showChild, setShowChild] = useState(false);

  useEffect(() => {
    setShowChild(true);
  }, []);

  if (!showChild) {
    return null; // Возвращаем null, чтобы не было мигания спиннера, если загрузка быстрая
  }

  return (
    <Provider store={store}>
      <Component {...pageProps} />
    </Provider>
  );
}

export default MyApp;

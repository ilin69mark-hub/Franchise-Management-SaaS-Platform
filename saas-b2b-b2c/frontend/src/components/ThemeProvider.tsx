'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { ConfigProvider, theme as antdTheme } from 'antd';
import { useThemeStore, ThemeMode } from '@/store/themeStore';

const STORAGE_KEY = 'franchiseTheme';

export const ThemeContext = createContext<{
  theme: ThemeMode;
  setTheme: (t: ThemeMode) => void;
  toggleTheme: () => void;
}>({
  theme: 'light',
  setTheme: () => {},
  toggleTheme: () => {},
});

export const useThemeMode = () => useContext(ThemeContext);

const ThemeProviderInner: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [mounted, setMounted] = useState(false);
  const [theme, setThemeState] = useState<ThemeMode>('light');
  const { setTheme: storeSetTheme } = useThemeStore();

  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      const initial: ThemeMode =
        stored === 'dark' || stored === 'light'
          ? stored
          : window.matchMedia?.('(prefers-color-scheme: dark)').matches
          ? 'dark'
          : 'light';
      setThemeState(initial);
      document.documentElement.setAttribute('data-theme', initial);
    } catch {}
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!mounted) return;
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme, mounted]);

  const handleSetTheme = (t: ThemeMode) => {
    setThemeState(t);
    storeSetTheme(t);
  };

  const handleToggleTheme = () => {
    const next = theme === 'light' ? 'dark' : 'light';
    handleSetTheme(next);
  };

  if (!mounted) {
    return null;
  }

  return (
    <ConfigProvider
      theme={{
        algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: { colorPrimary: '#1677ff' },
      }}
    >
      <ThemeContext.Provider value={{ theme, setTheme: handleSetTheme, toggleTheme: handleToggleTheme }}>
        {children}
      </ThemeContext.Provider>
    </ConfigProvider>
  );
};

const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <ThemeProviderInner>{children}</ThemeProviderInner>
);

export default ThemeProvider;


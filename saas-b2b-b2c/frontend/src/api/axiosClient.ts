import axios from 'axios';

const apiClient = axios.create({
  // берём базовый URL из переменной окружения.
  // Если переменной нет – fallback к http://localhost:8080
  baseURL: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

/* ----- Добавляем токен к каждому запросу ----- */
apiClient.interceptors.request.use(
  (config) => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('accessToken');
      // Если токен пустой либо строка "null" – не отправляем заголовок
      if (token && token !== 'null') {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => Promise.reject(error),
);

export default apiClient;

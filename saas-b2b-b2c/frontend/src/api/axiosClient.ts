import axios from 'axios';

function getApiBase(): string {
  const env = process.env.NEXT_PUBLIC_API_URL;
  if (env) return env;
  if (process.env.NODE_ENV === 'production') {
    throw new Error('NEXT_PUBLIC_API_URL must be set in production (fail-closed)');
  }
  return 'http://localhost:8080';
}

const apiClient = axios.create({
  baseURL: `${getApiBase()}/api/v1`,
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

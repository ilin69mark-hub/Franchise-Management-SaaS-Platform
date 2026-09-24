import axios from 'axios';
import { getCsrfToken, CSRF_HEADER } from '@/utils/csrf';

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
  withCredentials: true,
});

/* ----- F7: CSRF-заголовок к каждому запросу (сессия — в httpOnly cookie,
   токены в localStorage больше не храним и не шлём) ----- */
apiClient.interceptors.request.use(
  (config) => {
    const token = getCsrfToken();
    if (token) {
      config.headers[CSRF_HEADER] = token;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

export default apiClient;

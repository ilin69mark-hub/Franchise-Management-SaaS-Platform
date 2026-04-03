import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Перехватчик для добавления токена
apiClient.interceptors.request.use(
  (config) => {
    if (typeof window !== 'undefined') {
      // Пробуем взять токен из localStorage (где его хранит authSlice)
      let token = localStorage.getItem('accessToken');
      
      // Если там строка "null" или пусто, чистим
      if (token === 'null' || !token) {
          token = null;
      }

      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    // ВАЖНО: Обязательно возвращаем config!
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default apiClient;

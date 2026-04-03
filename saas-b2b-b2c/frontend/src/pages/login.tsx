import React, { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useRouter } from 'next/router';
import { Card, Form, Input, Button, Alert } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { RootState } from '../store';
import { login } from '../store/authSlice';
import Head from 'next/head';
import BackButton from '@/components/Dashboard/BackButton';

const LoginPage: React.FC = () => {
  const router = useRouter();
  const dispatch = useDispatch();
  const { loading, error, isAuthenticated } = useSelector((state: RootState) => state.auth);

  useEffect(() => {
    // Если уже залогинены, пытаемся перенаправить
    if (isAuthenticated) {
      // Здесь можно добавить проверку роли из state, если нужно,
      // но основная логика редиректа сработает после успешного логина ниже
      router.push('/');
    }
  }, [isAuthenticated, router]);

  const onFinish = async (values: any) => {
    try {
      // === ИСПРАВЛЕНО: Обрабатываем результат входа ===
      const result = await dispatch(login({ email: values.email, password: values.password }) as any);
      
      // Проверяем, что вход успешен и есть payload
      if (login.fulfilled.match(result)) {
        const userRole = result.payload?.user?.role;

        // Редирект в зависимости от роли
        if (userRole === 'super_admin') {
          router.push('/admin'); // Админ -> в Админку
        } else {
          router.push('/'); // Остальные -> на Главную
        }
      }
    } catch (e) {
      // Ошибки уже обработаны в slice
      console.error('Login failed', e);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f0f2f5' }}>
      <Head><title>Вход | Франчайзинг</title></Head>

      {/* Кнопка назад в левом верхнем углу */}
      <div style={{ position: 'absolute', top: 20, left: 20 }}>
        <BackButton />
      </div>

      <Card title="Вход в систему" style={{ width: 400 }}>
        {error && <Alert message={typeof error === 'string' ? error : 'Ошибка'} type="error" showIcon style={{ marginBottom: 16 }} />}
        <Form onFinish={onFinish} layout="vertical">
          <Form.Item name="email" rules={[{ required: true, message: 'Введите email' }]}>
            <Input prefix={<UserOutlined />} placeholder="Email" size="large" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: 'Введите пароль' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="Пароль" size="large" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block size="large" loading={loading}>Войти</Button>
          </Form.Item>
        </Form>
        <div style={{ textAlign: 'center' }}>
          Нет аккаунта? <Button type="link" onClick={() => router.push('/register')}>Зарегистрироваться</Button>
        </div>
      </Card>
    </div>
  );
};

export default LoginPage;
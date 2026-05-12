// src/pages/login.tsx
import React, { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useRouter } from 'next/router';
import { Card, Form, Input, Button, Alert } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { RootState, AppDispatch } from '@/store';      // <-- импортируем типы
import { login } from '@/store/authSlice';
import Head from 'next/head';
import BackButton from '@/components/Dashboard/BackButton';

const LoginPage: React.FC = () => {
  const router = useRouter();

  // типизируем dispatch – теперь он знает о thunk‑методах
  const dispatch = useDispatch<AppDispatch>();

  const { loading, error, isAuthenticated, user } = useSelector(
    (state: RootState) => state.auth,
  );

  /* -------------------------------------------------
     Если пользователь уже залогинен – редиректим сразу
     ------------------------------------------------- */
  useEffect(() => {
    if (isAuthenticated && user?.role) {
      switch (user.role) {
        case 'super_admin':
          router.replace('/admin');
          break;
        case 'franchiser':
        case 'franchiser_manager':
          router.replace('/franchiser-manager');
          break;
        case 'dealer':
          router.replace('/dealer');
          break;
        case 'salon_manager':
          router.replace('/salon-manager');
          break;
        default:
          router.replace('/');
      }
    }
  }, [isAuthenticated, user?.role, router]);

  /* -------------------------------------------------
     Обработчик формы входа
     ------------------------------------------------- */
  const onFinish = async (values: { email: string; password: string }) => {
    try {
      // dispatch(login) возвращает thunk‑action → unwrap() выбрасывает ошибку,
      // если запрос завершился неудачно, иначе возвращает payload
      const result = await dispatch(
        login({ email: values.email, password: values.password })
      ).unwrap(); // <-- теперь типы работают без ошибок

      const role = result.user?.role;

      if (role === 'super_admin') {
        router.replace('/admin');
      } else if (role === 'franchiser' || role === 'franchiser_manager') {
        router.replace('/franchiser-manager');
      } else if (role === 'dealer') {
        router.replace('/dealer');
      } else if (role === 'salon_manager') {
        router.replace('/salon-manager');
      } else {
        router.replace('/');
      }
    } catch (e) {
      // Ошибки уже записаны в slice → покажутся в <Alert>
      console.error('Login failed', e);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
      }}
    >
      <Head>
        <title>Вход | Франчайзинг</title>
      </Head>

      {/* Кнопка «Назад» в левом верхнем углу */}
      <div style={{ position: 'absolute', top: 20, left: 20 }}>
        <BackButton />
      </div>

      <Card title="Вход в систему" style={{ width: 400 }}>
        {error && (
          <Alert
            message={typeof error === 'string' ? error : 'Ошибка'}
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Form onFinish={onFinish} layout="vertical">
          <Form.Item
            name="email"
            rules={[{ required: true, message: 'Введите email' }]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="Email"
              size="large"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: 'Введите пароль' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Пароль"
              size="large"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              block
              size="large"
              loading={loading}
            >
              Войти
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          Нет аккаунта?{' '}
          <Button type="link" onClick={() => router.push('/register')}>
            Зарегистрироваться
          </Button>
        </div>
      </Card>
    </div>
  );
};

export default LoginPage;

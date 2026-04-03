import React, { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useRouter } from 'next/router';
import { Card, Form, Input, Button, Alert } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined, ShopOutlined } from '@ant-design/icons';
import { RootState } from '../store';
import { register } from '../store/authSlice';
import Head from 'next/head';
import BackButton from '@/components/Dashboard/BackButton'; // <--- Импорт

const RegisterPage: React.FC = () => {
  const router = useRouter();
  const dispatch = useDispatch();
  const { loading, error, isAuthenticated } = useSelector((state: RootState) => state.auth);
  const [form] = Form.useForm();

  useEffect(() => {
    if (isAuthenticated) router.push('/');
  }, [isAuthenticated, router]);

  const onFinish = async (values: any) => {
    const payload = {
      email: values.email,
      password: values.password,
      first_name: values.firstName,
      last_name: values.lastName,
      company_name: values.companyName,
      role: 'user',
      tenant_id: 'default',
    };
    await dispatch((register as any)(payload));
    router.push('/login');
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)' }}>
      <Head><title>Регистрация | Франчайзинг</title></Head>

      {/* Кнопка назад в левом верхнем углу */}
      <div style={{ position: 'absolute', top: 20, left: 20 }}>
        <BackButton />
      </div>

      <Card title="Создать аккаунт" style={{ width: 450 }}>
        {error && <Alert message="Ошибка регистрации" description={typeof error === 'string' ? error : 'Ошибка'} type="error" showIcon style={{ marginBottom: 16 }} />}
        <Form form={form} onFinish={onFinish} layout="vertical">
          <Form.Item name="firstName" label="Имя" rules={[{ required: true }]}><Input prefix={<UserOutlined />} /></Form.Item>
          <Form.Item name="lastName" label="Фамилия" rules={[{ required: true }]}><Input prefix={<UserOutlined />} /></Form.Item>
          <Form.Item name="companyName" label="Название компании" rules={[{ required: true }]}><Input prefix={<ShopOutlined />} /></Form.Item>
          <Form.Item name="email" label="Email" rules={[{ required: true, type: 'email' }]}><Input prefix={<MailOutlined />} /></Form.Item>
          <Form.Item name="password" label="Пароль" rules={[{ required: true, min: 8 }]}><Input.Password prefix={<LockOutlined />} /></Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>Зарегистрироваться</Button>
        </Form>
        <div style={{ textAlign: 'center', marginTop: 16 }}>
          Уже есть аккаунт? <Button type="link" onClick={() => router.push('/login')}>Войти</Button>
        </div>
      </Card>
    </div>
  );
};

export default RegisterPage;
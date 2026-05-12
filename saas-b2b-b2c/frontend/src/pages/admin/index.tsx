import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Layout } from 'antd';
import { RootState } from '../../store';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Header from '@/components/Dashboard/Header';
import SuperAdminDashboard from '@/components/Dashboard/SuperAdminDashboard';

const { Content } = Layout;

const AdminPage: React.FC = () => {
  const router = useRouter();
  const { isAuthenticated, user } = useSelector((state: RootState) => state.auth);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  if (!user) {
    return (
      <Layout style={{ minHeight: '100vh',  }}>
        <Head><title>Загрузка...</title></Head>
        <Header />
        <Content />
      </Layout>
    );
  }

  if (user.role === 'super_admin') {
    return (
      <Layout style={{ minHeight: '100vh',  }}>
        <Head><title>SaaS Платформа - Супер Админ</title></Head>
        <SuperAdminDashboard user={user} />
      </Layout>
    );
  }

  if (user.role === 'dealer') {
    router.push('/dealer');
    return null;
  }

  if (user.role === 'franchiser' || user.role === 'franchiser_manager') {
    router.push('/franchiser-manager');
    return null;
  }

  if (user.role === 'salon_manager') {
    router.push('/salon-manager');
    return null;
  }

  router.push('/');
  return null;
};

export default AdminPage;
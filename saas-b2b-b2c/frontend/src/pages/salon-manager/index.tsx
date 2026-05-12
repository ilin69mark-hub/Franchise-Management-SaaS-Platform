import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Layout } from 'antd';
import { RootState } from '../../store';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Header from '@/components/Dashboard/Header';
import SalonManagerDashboard from '@/components/Dashboard/SalonManagerDashboard';

const { Content } = Layout;

const SalonManagerPage: React.FC = () => {
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

  if (user.role === 'salon_manager') {
    return (
      <Layout style={{ minHeight: '100vh',  }}>
        <Head><title>Личный кабинет менеджера салона</title></Head>
        <Header />
        <Content style={{ padding: 0 }}>
          <SalonManagerDashboard user={user} title="Личный кабинет менеджера салона" />
        </Content>
      </Layout>
    );
  }

  if (user.role === 'super_admin') {
    router.push('/admin');
    return null;
  }

  if (user.role === 'franchiser' || user.role === 'franchiser_manager') {
    router.push('/franchiser-manager');
    return null;
  }

  if (user.role === 'dealer') {
    router.push('/dealer');
    return null;
  }

  router.push('/');
  return null;
};

export default SalonManagerPage;
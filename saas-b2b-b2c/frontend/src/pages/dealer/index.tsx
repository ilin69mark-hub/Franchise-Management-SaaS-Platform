import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Layout } from 'antd';
import { RootState } from '../../store';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Header from '@/components/Dashboard/Header';
import DealerDashboard from '@/components/Dashboard/DealerDashboard';

const { Content } = Layout;

const DealerPage: React.FC = () => {
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
      <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
        <Head><title>Загрузка...</title></Head>
        <Header />
        <Content />
      </Layout>
    );
  }

  if (user.role === 'dealer') {
    return (
      <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
        <Head><title>Личный кабинет дилера</title></Head>
        <Header />
        <Content style={{ padding: 0 }}>
          <DealerDashboard user={user} title="Личный кабинет дилера" />
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

  if (user.role === 'salon_manager') {
    router.push('/salon-manager');
    return null;
  }

  router.push('/');
  return null;
};

export default DealerPage;
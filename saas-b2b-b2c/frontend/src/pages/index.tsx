import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Layout } from 'antd';
import { RootState } from '../store';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Header from '@/components/Dashboard/Header';
import FranchiserDashboard from '@/components/Dashboard/FranchiserDashboard';
import DealerDashboard from '@/components/Dashboard/DealerDashboard';
import SalonManagerDashboard from '@/components/Dashboard/SalonManagerDashboard';

const { Content } = Layout;

const DashboardPage: React.FC = () => {
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

  if (user.role === 'super_admin') {
    router.push('/admin');
    return null;
  }

  if (user.role === 'franchiser' || user.role === 'franchiser_manager') {
    const title = user.role === 'franchiser' ? 'Личный кабинет Франчайзера' : 'Личный кабинет Менеджера';
    return (
      <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
        <Head><title>{title}</title></Head>
        <Header />
        <Content style={{ padding: 0 }}>
          <FranchiserDashboard user={user} title={title} />
        </Content>
      </Layout>
    );
  }

  // Роутинг для Менеджера Салона
  if (user.role === 'salon_manager') {
    return (
      <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
        <Head><title>Личный кабинет  менеджера салона</title></Head>
        <Header />
        <Content style={{ padding: 0 }}>
          <SalonManagerDashboard user={user} title="Личный кабинет менеджера салона" />
        </Content>
      </Layout>
    );
  }

  // Для Дилера
  return (
    <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
      <Head><title>Личный кабинет дилера</title></Head>
      <Header />
      <Content style={{ padding: 0 }}>
        <DealerDashboard user={user} title="Личный кабинет дилера" />
      </Content>
    </Layout>
  );
};

export default DashboardPage;

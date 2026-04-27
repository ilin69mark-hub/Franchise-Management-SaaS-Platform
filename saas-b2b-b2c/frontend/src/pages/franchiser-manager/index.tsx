import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Layout } from 'antd';
import { RootState } from '../../store';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Header from '@/components/Dashboard/Header';
import FranchiserDashboard from '@/components/Dashboard/FranchiserDashboard';

const { Content } = Layout;

const FranchiserManagerPage: React.FC = () => {
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

  if (user.role === 'franchiser') {
    const title = 'Личный кабинет Франчайзера';
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

  if (user.role === 'franchiser_manager') {
    const title = 'Личный кабинет Менеджера';
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

  if (user.role === 'super_admin') {
    router.push('/admin');
    return null;
  }

  if (user.role === 'dealer') {
    router.push('/dealer');
    return null;
  }

  if (user.role === 'salon_manager') {
    router.push('/salon-manager');
    return null;
  }

  router.push('/');
  return null;
};

export default FranchiserManagerPage;
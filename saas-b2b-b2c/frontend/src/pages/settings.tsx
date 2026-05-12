'use client';

import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useRouter } from 'next/router';
import Head from 'next/head';
import { Layout, Row, Col, Card, Switch, Typography, Divider, Button, Breadcrumb } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import Link from 'next/link';
import { RootState } from '@/store';
import { useThemeMode } from '@/components/ThemeProvider';

const { Title, Text } = Typography;
const { Content } = Layout;

const SettingsPage: React.FC = () => {
  const router = useRouter();
  const { isAuthenticated } = useSelector((state: RootState) => state.auth);
  const { theme, toggleTheme } = useThemeMode();

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  return (
    <Layout style={{ minHeight: '100vh',  }}>
      <Head><title>Настройки</title></Head>
      <Content style={{ padding: 24 }}>
        <Row justify="center">
          <Col xs={24} lg={16}>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => router.back()}
              style={{ marginBottom: 16 }}
            >
              Назад
            </Button>
            <Breadcrumb
              style={{ marginBottom: 16 }}
              items={[
                { title: <Link href="/">Главная</Link> },
                { title: 'Настройки' },
              ]}
            />

            <Title level={3}>Настройки</Title>

            <Card title="Оформление">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <Text strong>Тёмная тема</Text>
                  <br />
                  <Text type="secondary">
                    {theme === 'dark' ? 'Сейчас включена тёмная тема' : 'Сейчас включена светлая тема'}
                  </Text>
                </div>
                <Switch
                  checked={theme === 'dark'}
                  onChange={toggleTheme}
                  checkedChildren="🌙"
                  unCheckedChildren="☀️"
                />
              </div>
              <Divider />
              <Text type="secondary">
                Тема применяется мгновенно и сохраняется в браузере между сеансами.
              </Text>
            </Card>
          </Col>
        </Row>
      </Content>
    </Layout>
  );
};

export default SettingsPage;

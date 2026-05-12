'use client';

import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useRouter } from 'next/router';
import Head from 'next/head';
import {
  Layout,
  Row,
  Col,
  Card,
  Avatar,
  Button,
  Form,
  Input,
  Switch,
  Select,
  Tag,
  Skeleton,
  Alert,
  message,
  Typography,
  Space,
  Divider,
  Breadcrumb,
} from 'antd';
import { UserOutlined, SaveOutlined, ReloadOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import Link from 'next/link';
import { RootState } from '@/store';
import { useGetMyProfileQuery, useUpdateProfileMutation } from '@/services/userApi';
import type { UpdateProfileRequest } from '@/types';

const { Title, Text } = Typography;
const { Content } = Layout;
const { TextArea } = Input;

const STATUS_OPTIONS = [
  { value: 'online', label: '🟢 В сети' },
  { value: 'office', label: '🟡 В офисе' },
  { value: 'meeting', label: '🔴 На встрече' },
  { value: 'vacation', label: '⚪ В отпуске' },
  { value: 'dnd', label: '⚫ Не беспокоить' },
];

const ProfilePage: React.FC = () => {
  const router = useRouter();
  const { isAuthenticated } = useSelector((state: RootState) => state.auth);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  const { data: profile, isLoading, isError, refetch } = useGetMyProfileQuery();
  const [updateProfile, { isLoading: isUpdating }] = useUpdateProfileMutation();

  const [mainForm] = Form.useForm();
  const [contactsForm] = Form.useForm();
  const [statusForm] = Form.useForm();

  useEffect(() => {
    if (profile) {
      mainForm.setFieldsValue({
        first_name: profile.first_name,
        last_name: profile.last_name,
        display_name: profile.display_name,
        position: profile.position,
        bio: profile.bio,
        quote: profile.quote,
      });
      contactsForm.setFieldsValue({
        phone: profile.contacts?.phone,
        telegram: profile.contacts?.telegram,
        whatsapp: profile.contacts?.whatsapp,
        working_hours: profile.contacts?.working_hours,
        email_visible: profile.contacts?.email_visible,
        phone_visible: profile.contacts?.phone_visible,
      });
      statusForm.setFieldsValue({
        status: profile.status || 'online',
        available_for_questions: profile.available_for_questions ?? true,
      });
    }
  }, [profile, mainForm, contactsForm, statusForm]);

  const handleMainSave = async (values: Record<string, unknown>) => {
    try {
      await updateProfile(values as UpdateProfileRequest).unwrap();
      message.success('Профиль сохранён');
    } catch {
      message.error('Ошибка сохранения профиля');
    }
  };

  const handleContactsSave = async (values: Record<string, unknown>) => {
    try {
      const flatData: Record<string, unknown> = {
        contacts_email_visible: values.email_visible,
        contacts_phone_visible: values.phone_visible,
        contacts_phone: values.phone,
        contacts_telegram: values.telegram,
        contacts_whatsapp: values.whatsapp,
        contacts_working_hours: values.working_hours,
      };
      await updateProfile(flatData as unknown as UpdateProfileRequest).unwrap();
      message.success('Контакты сохранены');
    } catch {
      message.error('Ошибка сохранения контактов');
    }
  };

  const handleStatusChange = async (changedValues: Record<string, unknown>) => {
    try {
      await updateProfile(changedValues as UpdateProfileRequest).unwrap();
    } catch {
      message.error('Ошибка обновления статуса');
    }
  };

  const getInitials = () => {
    if (!profile) return '?';
    const f = profile.first_name?.[0] || '';
    const l = profile.last_name?.[0] || '';
    return (f + l).toUpperCase() || profile.email?.[0]?.toUpperCase() || '?';
  };

  if (isLoading) {
    return (
      <Layout style={{ minHeight: '100vh' }}>
        <Content style={{ padding: 24 }}>
          <Row justify="center">
            <Col xs={24} lg={16}>
              <Skeleton active paragraph={{ rows: 6 }} />
            </Col>
          </Row>
        </Content>
      </Layout>
    );
  }

  if (isError) {
    return (
      <Layout style={{ minHeight: '100vh' }}>
        <Content style={{ padding: 24 }}>
          <Row justify="center">
            <Col xs={24} lg={16}>
              <Alert
                type="error"
                message="Не удалось загрузить профиль"
                description="Попробуйте обновить страницу"
                action={<Button icon={<ReloadOutlined />} onClick={() => refetch()}>Повторить</Button>}
              />
            </Col>
          </Row>
        </Content>
      </Layout>
    );
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Head><title>Мой профиль</title></Head>
      <Content style={{ padding: 24 }}>
        <Row justify="center">
          <Col xs={24} lg={20}>
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              <div>
                <Button
                  icon={<ArrowLeftOutlined />}
                  onClick={() => router.back()}
                  style={{ marginBottom: 8 }}
                >
                  Назад
                </Button>
                <Breadcrumb
                  style={{ marginTop: 8 }}
                  items={[
                    { title: <Link href="/">Главная</Link> },
                    { title: 'Мой профиль' },
                  ]}
                />
              </div>

              <Title level={3}>Мой профиль</Title>

              <Row gutter={[24, 24]}>
                <Col xs={24} md={8}>
                  <Card title="Аватар">
                    <div style={{ textAlign: 'center' }}>
                      <Avatar
                        size={120}
                        src={profile?.avatar_url}
                        icon={<UserOutlined />}
                        style={{ fontSize: 40, backgroundColor: '#1677ff' }}
                      >
                        {!profile?.avatar_url && getInitials()}
                      </Avatar>
                      <div style={{ marginTop: 16 }}>
                        <Text type="secondary">
                          Загрузка аватара будет доступна позже
                        </Text>
                      </div>
                    </div>
                  </Card>

                  <Card title="Статус и настроение" style={{ marginTop: 24 }}>
                    <Form form={statusForm} layout="vertical" onValuesChange={handleStatusChange}>
                      <Form.Item name="status" label="Текущий статус">
                        <Select options={STATUS_OPTIONS} />
                      </Form.Item>
                      <Form.Item name="available_for_questions" label="Доступен для вопросов" valuePropName="checked">
                        <Switch />
                      </Form.Item>
                      {profile?.achievements && profile.achievements.length > 0 && (
                        <>
                          <Divider />
                          <Text type="secondary">Достижения</Text>
                          <div style={{ marginTop: 8 }}>
                            {profile.achievements.map((a) => (
                              <Tag key={a} color="blue">{a}</Tag>
                            ))}
                          </div>
                        </>
                      )}
                    </Form>
                  </Card>
                </Col>

                <Col xs={24} md={16}>
                  <Card title="Основная информация">
                    <Form
                      form={mainForm}
                      layout="vertical"
                      onFinish={handleMainSave}
                    >
                      <Row gutter={16}>
                        <Col xs={24} md={12}>
                          <Form.Item
                            name="first_name"
                            label="Имя"
                            rules={[{ required: true, min: 2, message: 'Минимум 2 символа' }]}
                          >
                            <Input placeholder="Имя" />
                          </Form.Item>
                        </Col>
                        <Col xs={24} md={12}>
                          <Form.Item
                            name="last_name"
                            label="Фамилия"
                            rules={[{ required: true, min: 2, message: 'Минимум 2 символа' }]}
                          >
                            <Input placeholder="Фамилия" />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Row gutter={16}>
                        <Col xs={24} md={12}>
                          <Form.Item name="display_name" label="Отображаемое имя">
                            <Input placeholder="Иван Петров" />
                          </Form.Item>
                        </Col>
                        <Col xs={24} md={12}>
                          <Form.Item name="position" label="Должность">
                            <Input placeholder="Руководитель отдела" />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Form.Item name="bio" label="О себе" extra="Максимум 500 символов">
                        <TextArea
                          rows={3}
                          maxLength={500}
                          showCount
                          placeholder="Расскажите о себе..."
                        />
                      </Form.Item>
                      <Form.Item name="quote" label="Цитата" extra="Максимум 100 символов">
                        <Input
                          maxLength={100}
                          showCount
                          placeholder="Ваша любимая цитата..."
                        />
                      </Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        icon={<SaveOutlined />}
                        loading={isUpdating}
                      >
                        Сохранить
                      </Button>
                    </Form>
                  </Card>

                  <Card title="Контакты для дилеров" style={{ marginTop: 24 }}>
                    <Form
                      form={contactsForm}
                      layout="vertical"
                      onFinish={handleContactsSave}
                    >
                      <Row gutter={16}>
                        <Col xs={24} md={16}>
                          <Form.Item name="phone" label="Телефон">
                            <Input placeholder="+7 (999) 123-45-67" />
                          </Form.Item>
                        </Col>
                        <Col xs={24} md={8}>
                          <Form.Item name="phone_visible" label="Показывать" valuePropName="checked">
                            <Switch />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Row gutter={16}>
                        <Col xs={24} md={16}>
                          <Form.Item name="telegram" label="Telegram">
                            <Input placeholder="@username" />
                          </Form.Item>
                        </Col>
                        <Col xs={24} md={8}>
                          <Form.Item name="email_visible" label="Email виден" valuePropName="checked">
                            <Switch checked={profile?.contacts?.email_visible} disabled />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Row gutter={16}>
                        <Col xs={24} md={16}>
                          <Form.Item name="whatsapp" label="WhatsApp">
                            <Input placeholder="+7 (999) 123-45-67" />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Row gutter={16}>
                        <Col xs={24}>
                          <Form.Item name="working_hours" label="Часы работы">
                            <Input placeholder="будни 10:00-18:00" />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Button
                        type="primary"
                        htmlType="submit"
                        icon={<SaveOutlined />}
                        loading={isUpdating}
                      >
                        Сохранить контакты
                      </Button>
                    </Form>
                  </Card>
                </Col>
              </Row>
            </Space>
          </Col>
        </Row>
      </Content>
    </Layout>
  );
};

export default ProfilePage;

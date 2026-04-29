import React, { useEffect, useState } from 'react';
import {
  Badge,
  Dropdown,
  List,
  Button,
  Tag,
  Typography,
  Spin,
  message,
  Switch,
  Input,
  Card,
  Row,
  Col,
  Space,
} from 'antd';
import {
  BellOutlined,
  WarningOutlined,
  SettingOutlined,
  SafetyOutlined,
  LineChartOutlined,
  CloseCircleOutlined,
  RightOutlined,
  MailOutlined,
} from '@ant-design/icons';
import { useAlertStore } from '@/store/alertsStore';
import dayjs from 'dayjs';

const { Text, Title } = Typography;

interface AdminAlertsProps {
  compact?: boolean;
}

const AdminAlerts: React.FC<AdminAlertsProps> = ({ compact = false }) => {
  const {
    alerts,
    unreadCount,
    settings,
    isLoading,
    loadAlerts,
    markAsRead,
    updateSettings,
  } = useAlertStore();

  const [showSettings, setShowSettings] = useState(false);

  useEffect(() => {
    loadAlerts();
  }, []);

  const getCategoryIcon = (type: string) => {
    const icons: Record<string, React.ReactNode> = {
      billing: <WarningOutlined style={{ color: '#faad14' }} />,
      technical: <SettingOutlined style={{ color: '#1890ff' }} />,
      security: <SafetyOutlined style={{ color: '#ff4d4f' }} />,
      activity: <LineChartOutlined style={{ color: '#722ed1' }} />,
    };
    return icons[type] || <BellOutlined />;
  };

  const getCategoryLabel = (type: string) => {
    const labels: Record<string, string> = {
      billing: '💰 Биллинг',
      technical: '⚙️ Технические',
      security: '🔒 Безопасность',
      activity: '📊 Активность',
    };
    return labels[type] || type;
  };

  const handleMarkAsRead = async (id: string) => {
    await markAsRead(id);
    message.success('Алерт закрыт');
  };

  const getLink = (link?: string) => {
    if (!link) return null;
    return (
      <Button type="link" size="small" icon={<RightOutlined />}>
        Перейти
      </Button>
    );
  };

  const groupedAlerts = alerts.reduce((acc, alert) => {
    if (!acc[alert.type]) acc[alert.type] = [];
    acc[alert.type].push(alert);
    return acc;
  }, {} as Record<string, typeof alerts>);

  const alertList = (
    <div style={{ width: 400, maxHeight: 500, overflow: 'auto' }}>
      <div style={{ padding: 16, borderBottom: '1px solid #f0f0f0' }}>
        <Space>
          <Title level={5} style={{ margin: 0 }}>Уведомления</Title>
          <Text type="secondary">({unreadCount} новых)</Text>
        </Space>
      </div>

      {Object.keys(groupedAlerts).length === 0 ? (
        <div style={{ padding: 32, textAlign: 'center' }}>
          <Text type="secondary">Нет уведомлений</Text>
        </div>
      ) : (
        Object.entries(groupedAlerts).map(([type, typeAlerts]) => (
          <div key={type}>
            <div style={{ padding: '8px 16px', background: '#fafafa' }}>
              <Text strong>{getCategoryLabel(type)}</Text>
            </div>
            <List
              dataSource={typeAlerts}
              renderItem={(item) => (
                <List.Item
                  style={{ 
                    padding: 12, 
                    background: item.read ? '#fff' : '#f6ffed',
                    opacity: item.read ? 0.6 : 1,
                  }}
                  actions={[
                    getLink(item.link),
                    <Button
                      key="close"
                      type="text"
                      size="small"
                      icon={<CloseCircleOutlined />}
                      onClick={() => handleMarkAsRead(item.id)}
                    />,
                  ]}
                >
                  <List.Item.Meta
                    avatar={getCategoryIcon(item.type)}
                    title={item.title}
                    description={
                      <div>
                        <div>{item.message}</div>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {dayjs(item.timestamp).format('HH:mm DD.MM')}
                        </Text>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </div>
        ))
      )}

      <div style={{ padding: 16, borderTop: '1px solid #f0f0f0' }}>
        <Button type="link" block onClick={() => setShowSettings(true)}>
          Настройки уведомлений
        </Button>
      </div>
    </div>
  );

  const settingsPanel = (
    <Card title="Настройки уведомлений" style={{ width: 400 }}>
      <Title level={5}>Каналы</Title>
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Switch checked={settings?.channels?.includes('in-app')} />
          <Text>In-app</Text>
        </Space>
      </div>
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Switch checked={settings?.channels?.includes('email')} />
          <Text>Email</Text>
        </Space>
      </div>
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Switch checked={settings?.channels?.includes('telegram')} />
          <Text>Telegram</Text>
        </Space>
      </div>

      <Title level={5}>Пороги</Title>
      <Row gutter={16}>
        <Col span={12}>
          <Text>Error Rate (%)</Text>
          <Input
            type="number"
            defaultValue={settings?.thresholds?.errorRate || 1}
            style={{ marginTop: 4 }}
          />
        </Col>
        <Col span={12}>
          <Text>Uptime (%)</Text>
          <Input
            type="number"
            defaultValue={settings?.thresholds?.uptime || 99.9}
            style={{ marginTop: 4 }}
          />
        </Col>
      </Row>

      <Button
        type="primary"
        style={{ marginTop: 16 }}
        onClick={() => {
          message.success('Настройки сохранены');
          setShowSettings(false);
        }}
      >
        Сохранить
      </Button>
    </Card>
  );

  if (isLoading) {
    return <Spin />;
  }

  if (showSettings) {
    return <Dropdown overlay={settingsPanel} open={showSettings} onOpenChange={setShowSettings}>
      <Badge count={unreadCount} size="small">
        <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
      </Badge>
    </Dropdown>;
  }

  return (
    <Dropdown
      dropdownRender={() => alertList}
      trigger={['click']}
      placement="bottomRight"
    >
      <Badge count={unreadCount} size="small" offset={[-5, 5]}>
        <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
      </Badge>
    </Dropdown>
  );
};

export default AdminAlerts;
// src/components/Dashboard/Alerts/DealerAlerts.tsx
import React, { useState, useEffect, useRef, useMemo } from 'react';
import { Badge, Popover, List, Button, Space, Typography, Switch, InputNumber, Card, Empty, Spin, Divider, message, notification, Modal } from 'antd';
import { 
  BellOutlined, 
  DollarOutlined, 
  ShopOutlined, 
  FileTextOutlined, 
  SettingOutlined,
  ArrowRightOutlined,
  CloseOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Text } = Typography;

export type AlertCategory = 'finance' | 'operation' | 'communication';
export type AlertPriority = 'high' | 'medium' | 'low';

export interface Alert {
  id: string;
  category: AlertCategory;
  priority: AlertPriority;
  title: string;
  description: string;
  details?: string;
  createdAt: string;
  isRead: boolean;
  salonId?: string;
  deepLink?: string;
}

export interface AlertSettings {
  fotPercentThreshold: number;
  rentPercentThreshold: number;
  conversionDropPercent: number;
  trafficDropPercent: number;
  stuckDealDays: number;
  stuckDealAmount: number;
  nonLiquidDays: number;
}

const defaultSettings: AlertSettings = {
  fotPercentThreshold: 20,
  rentPercentThreshold: 12,
  conversionDropPercent: 20,
  trafficDropPercent: 30,
  stuckDealDays: 7,
  stuckDealAmount: 200000,
  nonLiquidDays: 90,
};

interface DealerAlertsProps {
  alerts?: Alert[];
  loading?: boolean;
  onAlertClick?: (alert: Alert) => void;
  onMarkAsRead?: (alertId: string) => void;
  onSettingsSave?: (settings: AlertSettings) => Promise<void>;
}

const DealerAlerts: React.FC<DealerAlertsProps> = ({
  alerts: initialAlerts = [],
  loading = false,
  onAlertClick,
  onMarkAsRead,
  onSettingsSave,
}) => {
  const [alerts, setAlerts] = useState<Alert[]>(initialAlerts);
  const [settings, setSettings] = useState<AlertSettings>(defaultSettings);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [filterUnread, setFilterUnread] = useState(false);
  const [wsConnected, setWsConnected] = useState(false);
  const [localUnread, setLocalUnread] = useState<Record<string, boolean>>({});
  const [settingsLoading, setSettingsLoading] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    setAlerts(initialAlerts);
  }, [initialAlerts]);

  useEffect(() => {
    const saved = localStorage.getItem('dealerAlertSettings');
    if (saved) {
      try {
        setSettings(JSON.parse(saved));
      } catch (e) {}
    }
    const savedUnread = localStorage.getItem('dealerUnreadAlerts');
    if (savedUnread) {
      try {
        setLocalUnread(JSON.parse(savedUnread));
      } catch (e) {}
    }
  }, []);

  useEffect(() => {
    const connectWebSocket = () => {
      const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws/alerts';
      try {
        wsRef.current = new WebSocket(wsUrl);
        
        wsRef.current.onopen = () => {
          setWsConnected(true);
        };
        
        wsRef.current.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
          } catch (e) {}
        };
        
        wsRef.current.onclose = () => {
          setWsConnected(false);
          setTimeout(connectWebSocket, 5000);
        };
      } catch (e) {
        console.error('WebSocket error', e);
      }
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const handleWebSocketMessage = (data: { type: string; payload: Alert }) => {
    if (data.type === 'NEW_ALERT') {
      setAlerts(prev => [data.payload, ...prev]);
      setLocalUnread(prev => {
        const updated = { ...prev, [data.payload.id]: true };
        localStorage.setItem('dealerUnreadAlerts', JSON.stringify(updated));
        return updated;
      });
      
      const categoryIcons: Record<string, React.ReactNode> = {
        finance: <DollarOutlined />,
        operation: <ShopOutlined />,
        communication: <FileTextOutlined />,
      };
      
      notification.info({
        message: data.payload.title,
        description: data.payload.description,
        icon: categoryIcons[data.payload.category],
        duration: 5,
      });
    }
  };

  const unreadAlerts = useMemo(() => {
    return alerts.filter(a => localUnread[a.id] !== false);
  }, [alerts, localUnread]);

  const filteredAlerts = useMemo(() => {
    if (filterUnread) {
      return unreadAlerts;
    }
    return alerts;
  }, [alerts, unreadAlerts, filterUnread]);

  const groupedAlerts = useMemo(() => {
    const groups: Record<AlertCategory, Alert[]> = {
      finance: [],
      operation: [],
      communication: [],
    };
    filteredAlerts.forEach(a => groups[a.category].push(a));
    return groups;
  }, [filteredAlerts]);

  const totalUnread = unreadAlerts.length;

  const handleMarkAsRead = (alertId: string) => {
    setLocalUnread(prev => {
      const updated = { ...prev, [alertId]: false };
      localStorage.setItem('dealerUnreadAlerts', JSON.stringify(updated));
      return updated;
    });
    onMarkAsRead?.(alertId);
  };

  const handleSettingsSave = async () => {
    setSettingsLoading(true);
    localStorage.setItem('dealerAlertSettings', JSON.stringify(settings));
    try {
      await onSettingsSave?.(settings);
      message.success('Настройки сохранены');
      setSettingsOpen(false);
    } catch (e) {
      message.error('Ошибка сохранения');
    } finally {
      setSettingsLoading(false);
    }
  };

  const categoryIcons: Record<AlertCategory, React.ReactNode> = {
    finance: <DollarOutlined style={{ color: '#ff4d4f' }} />,
    operation: <ShopOutlined style={{ color: '#fa8c16' }} />,
    communication: <FileTextOutlined style={{ color: '#1890ff' }} />,
  };

  const getCategoryName = (category: AlertCategory) => {
    const names: Record<AlertCategory, string> = {
      finance: 'Финансы',
      operation: 'Операции',
      communication: 'Коммуникации',
    };
    return names[category];
  };

  const alertContent = (
    <div style={{ width: 380, maxHeight: 500, overflow: 'auto' }}>
      <div style={{ padding: '8px 16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Text strong>Алерты</Text>
          {totalUnread > 0 && <Badge count={totalUnread} style={{ marginLeft: 8 }} />}
        </Space>
        <Space>
          <Switch size="small" checked={filterUnread} onChange={setFilterUnread} />
          <Text type="secondary">Непрочитанные</Text>
        </Space>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 24 }}>
          <Spin />
        </div>
      ) : filteredAlerts.length === 0 ? (
        <Empty description="Нет алертов" style={{ padding: 24 }} />
      ) : (
        <>
          {(['finance', 'operation', 'communication'] as AlertCategory[]).map(category => {
            const items = groupedAlerts[category];
            if (items.length === 0) return null;
            return (
              <div key={category}>
                <div style={{ padding: '8px 16px', background: '#fafafa', display: 'flex', alignItems: 'center', gap: 8 }}>
                  {categoryIcons[category]}
                  <Text strong>{getCategoryName(category)}</Text>
                  <Badge count={items.filter(i => localUnread[i.id] !== false).length} />
                </div>
                <List
                  dataSource={items}
                  renderItem={(alert) => (
                    <List.Item
                      style={{ padding: '12px 16px', background: localUnread[alert.id] === false ? 'transparent' : '#fff7e6' }}
                      actions={[
                        <Button key="goto" size="small" type="link" icon={<ArrowRightOutlined />} onClick={() => onAlertClick?.(alert)} />,
                        <Button key="close" size="small" type="text" icon={<CloseOutlined />} onClick={() => handleMarkAsRead(alert.id)} />,
                      ]}
                    >
                      <List.Item.Meta
                        title={<Space><Text strong={localUnread[alert.id] !== false}>{alert.title}</Text></Space>}
                        description={
                          <Space direction="vertical" size={0}>
                            <Text type="secondary" style={{ fontSize: 12 }}>{alert.description}</Text>
                            {alert.details && <Text type="secondary" style={{ fontSize: 11 }}>{alert.details}</Text>}
                            <Text type="secondary" style={{ fontSize: 11 }}>{dayjs(alert.createdAt).format('DD.MM.YYYY HH:mm')}</Text>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              </div>
            );
          })}
        </>
      )}

      <Divider style={{ margin: 0 }} />
      <div style={{ padding: 8, textAlign: 'center' }}>
        <Button type="link" icon={<SettingOutlined />} onClick={() => setSettingsOpen(true)}>
          Настройки алертов
        </Button>
      </div>
    </div>
  );

  return (
    <>
      <Popover 
        content={alertContent} 
        trigger="click" 
        placement="bottomRight"
        overlayStyle={{ padding: 0 }}
      >
        <Badge count={totalUnread} size="small">
          <Button 
            type="text" 
            icon={
              <Space>
                <BellOutlined />
                {wsConnected ? <span style={{ width: 6, height: 6, borderRadius: '50%', background: '#52c41a' }} /> : <span style={{ width: 6, height: 6, borderRadius: '50%', background: '#d9d9d9' }} />}
              </Space>
            } 
            style={{ fontSize: 18 }}
          />
        </Badge>
      </Popover>

      <Modal
        title="Настройки алертов"
        open={settingsOpen}
        onCancel={() => setSettingsOpen(false)}
        footer={null}
        width={450}
      >
        <Card title="⚙️ Настройки порогов алертов" size="small">
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <div>
              <Text>Порог ФОТ (% от выручки):</Text>
              <InputNumber value={settings.fotPercentThreshold} onChange={v => setSettings(s => ({ ...s, fotPercentThreshold: v || 20 }))} min={0} max={100} style={{ width: '100%' }} />
            </div>
            <div>
              <Text>Порог аренды (% от выручки):</Text>
              <InputNumber value={settings.rentPercentThreshold} onChange={v => setSettings(s => ({ ...s, rentPercentThreshold: v || 12 }))} min={0} max={100} style={{ width: '100%' }} />
            </div>
            <div>
              <Text>Порог падения конверсии (%):</Text>
              <InputNumber value={settings.conversionDropPercent} onChange={v => setSettings(s => ({ ...s, conversionDropPercent: v || 20 }))} min={0} max={100} style={{ width: '100%' }} />
            </div>
            <div>
              <Text>Порог падения трафика (%):</Text>
              <InputNumber value={settings.trafficDropPercent} onChange={v => setSettings(s => ({ ...s, trafficDropPercent: v || 30 }))} min={0} max={100} style={{ width: '100%' }} />
            </div>
            <div>
              <Text>Дней зависания сделки:</Text>
              <InputNumber value={settings.stuckDealDays} onChange={v => setSettings(s => ({ ...s, stuckDealDays: v || 7 }))} min={0} max={365} style={{ width: '100%' }} />
            </div>
            <div>
              <Text>Сумма крупной сделки (руб.):</Text>
              <InputNumber value={settings.stuckDealAmount} onChange={v => setSettings(s => ({ ...s, stuckDealAmount: v || 200000 }))} min={0} style={{ width: '100%' }} />
            </div>
            <div>
              <Text>Дней неликвида:</Text>
              <InputNumber value={settings.nonLiquidDays} onChange={v => setSettings(s => ({ ...s, nonLiquidDays: v || 90 }))} min={0} max={365} style={{ width: '100%' }} />
            </div>
            <Button type="primary" onClick={handleSettingsSave} loading={settingsLoading} block>
              Сохранить настройки
            </Button>
          </Space>
        </Card>
      </Modal>
    </>
  );
};

export default DealerAlerts;
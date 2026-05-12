import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Badge, Dropdown, List, Button, Spin, Empty, Tag, Tooltip, message } from 'antd';
import { BellOutlined, WarningOutlined, ExclamationCircleOutlined, InfoCircleOutlined, CloseOutlined, LinkOutlined } from '@ant-design/icons';
import apiClient from '@/api/axiosClient';

interface AlertItem {
  id: string;
  title: string;
  message: string;
  type: string;
  severity: string;
  link: string;
  is_read: boolean;
  created_at: string;
  data?: string;
}

const NotificationBell: React.FC = () => {
  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const wsRef = useRef<WebSocket | null>(null);
  const pollingRef = useRef<NodeJS.Timeout | null>(null);

  const STORAGE_KEY = 'salon_alerts_unread';
  const POLL_INTERVAL = 2 * 60 * 1000; // 2 минуты

  // Загрузка сохраненных непрочитанных из localStorage
  const loadSavedUnread = useCallback(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) {
        const count = parseInt(saved, 10);
        if (!isNaN(count)) {
          setUnreadCount(count);
        }
      }
    } catch (e) {
      console.error('Error loading saved alerts', e);
    }
  }, []);

  // Сохранение непрочитанных в localStorage
  const saveUnread = useCallback((count: number) => {
    try {
      localStorage.setItem(STORAGE_KEY, count.toString());
    } catch (e) {
      console.error('Error saving alerts', e);
    }
  }, []);

  // Загрузка алертов с API
  const fetchAlerts = useCallback(async () => {
    try {
      setLoading(true);
      const res = await apiClient.get('/alerts');
      const alertsData = res.data.alerts || [];
      const count = res.data.unread_count || 0;

      setAlerts(alertsData);
      setUnreadCount(count);
      saveUnread(count);
    } catch (e) {
      console.error('Error fetching alerts', e);
    } finally {
      setLoading(false);
    }
  }, [saveUnread]);

  // Инициализация WebSocket
  const initWebSocket = useCallback(() => {
    const userId = localStorage.getItem('userId') || localStorage.getItem('user_id');
    if (!userId) return;

    const wsUrl = `ws://${window.location.host}/ws/alerts?user_id=${userId}`;

    try {
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('WebSocket connected');
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'alert') {
            const newAlert: AlertItem = {
              id: data.data.id || Date.now().toString(),
              title: data.data.title || 'Новый алерт',
              message: data.data.description || '',
              type: data.data.type || 'info',
              severity: data.data.severity || 'info',
              link: data.data.link || '',
              is_read: false,
              created_at: new Date().toISOString(),
            };

            setAlerts(prev => [newAlert, ...prev]);
            setUnreadCount(prev => {
              const newCount = prev + 1;
              saveUnread(newCount);
              return newCount;
            });
            message.info(newAlert.title);
          }
        } catch (e) {
          console.error('Error parsing WS message', e);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error', error);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected, falling back to polling');
        startPolling();
      };

      wsRef.current = ws;
    } catch (e) {
      console.error('WebSocket init error', e);
      startPolling();
    }
  }, [saveUnread]);

  // Polling fallback
  const startPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
    }

    pollingRef.current = setInterval(() => {
      fetchAlerts();
    }, POLL_INTERVAL);

    fetchAlerts();
  }, [fetchAlerts]);

  useEffect(() => {
    loadSavedUnread();
    fetchAlerts();

    // Пробуем WebSocket, если не работает - будет polling
    initWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
      }
    };
  }, [loadSavedUnread, fetchAlerts, initWebSocket]);

  const handleMarkRead = async (alertId: string) => {
    try {
      await apiClient.patch(`/alerts/${alertId}/read`);

      setAlerts(prev =>
        prev.map(a => (a.id === alertId ? { ...a, is_read: true } : a))
      );
      setUnreadCount(prev => {
        const newCount = Math.max(0, prev - 1);
        saveUnread(newCount);
        return newCount;
      });
    } catch (e) {
      console.error('Error marking alert as read', e);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return '#ff4d4f';
      case 'warning': return '#faad14';
      default: return '#1890ff';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical': return <WarningOutlined style={{ color: '#ff4d4f' }} />;
      case 'warning': return <ExclamationCircleOutlined style={{ color: '#faad14' }} />;
      default: return <InfoCircleOutlined style={{ color: '#1890ff' }} />;
    }
  };

  const getTypeLabel = (type: string) => {
    const types: Record<string, string> = {
      overdue_measurement: 'Просроченный замер',
      abandoned_kp: 'Брошенное КП',
      conversion_drop: 'Падение конверсии',
      traffic_drop: 'Падение трафика',
    };
    return types[type] || type;
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return 'Только что';
    if (diffMins < 60) return `${diffMins} мин назад`;
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)} ч назад`;
    return date.toLocaleDateString('ru-RU');
  };

  const getAlertItem = (item: AlertItem) => (
    <List.Item
      style={{
        padding: '12px',
        background: item.is_read ? '#fff' : '#fff7e6',
        borderLeft: item.severity === 'critical' ? '3px solid #ff4d4f' :
                    item.severity === 'warning' ? '3px solid #faad14' : '3px solid #1890ff',
        opacity: item.is_read ? 0.6 : 1,
      }}
    >
      <List.Item.Meta
        avatar={getSeverityIcon(item.severity)}
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Text strong>{item.title}</Text>
            <Tag color={item.severity === 'critical' ? 'error' : item.severity === 'warning' ? 'warning' : 'blue'}>
              {getTypeLabel(item.type)}
            </Tag>
          </div>
        }
        description={
          <div>
            <div>{item.message}</div>
            <div style={{ fontSize: 11, color: '#999', marginTop: 4 }}>
              {formatTime(item.created_at)}
            </div>
          </div>
        }
      />
      <div style={{ display: 'flex', gap: 8 }}>
        {item.link && (
          <Tooltip title="Перейти">
            <Button
              size="small"
              icon={<LinkOutlined />}
              onClick={() => window.location.href = item.link}
            />
          </Tooltip>
        )}
        {!item.is_read && (
          <Tooltip title="Закрыть">
            <Button
              size="small"
              icon={<CloseOutlined />}
              onClick={() => handleMarkRead(item.id)}
            />
          </Tooltip>
        )}
      </div>
    </List.Item>
  );

  const dropdownContent = (
    <div style={{ width: 360, maxHeight: 500, overflowY: 'auto', background: '#fff', boxShadow: '0 4px 12px rgba(0,0,0,0.15)', borderRadius: 8 }}>
      <div style={{ padding: '12px 16px', fontWeight: 'bold', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between' }}>
        <span>Алерты</span>
        <Tag color="red">{unreadCount}</Tag>
      </div>
      {loading ? (
        <div style={{ padding: 40, textAlign: 'center' }}><Spin /></div>
      ) : alerts.length === 0 ? (
        <Empty description="Нет алертов" style={{ padding: 40 }} />
      ) : (
        <List
          dataSource={alerts}
          renderItem={getAlertItem}
        />
      )}
    </div>
  );

  return (
    <Dropdown
      menu={{ items: [] }}
      popupRender={() => (
        <div style={{ width: 360, maxHeight: 500, overflowY: 'auto', background: '#fff', boxShadow: '0 4px 12px rgba(0,0,0,0.15)', borderRadius: 8 }}>
          <div style={{ padding: '12px 16px', fontWeight: 'bold', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between' }}>
            <span>Алерты</span>
            <Tag color="red">{unreadCount}</Tag>
          </div>
          {loading ? (
            <div style={{ padding: 40, textAlign: 'center' }}><Spin /></div>
          ) : alerts.length === 0 ? (
            <Empty description="Нет алертов" style={{ padding: 40 }} />
          ) : (
            <List
              dataSource={alerts}
              renderItem={getAlertItem}
            />
          )}
        </div>
      )}
      trigger={['click']}
      placement="bottomRight"
    >
      <Badge count={unreadCount} size="small" style={{ cursor: 'pointer', backgroundColor: unreadCount > 0 ? '#ff4d4f' : undefined }}>
        <BellOutlined style={{ fontSize: '18px', cursor: 'pointer', color: unreadCount > 0 ? '#ff4d4f' : undefined }} />
      </Badge>
    </Dropdown>
  );
};

export default NotificationBell;
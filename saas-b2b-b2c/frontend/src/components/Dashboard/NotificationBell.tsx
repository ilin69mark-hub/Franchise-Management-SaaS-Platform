import React from 'react';
import { Badge, Dropdown, List, Button, Spin, Empty } from 'antd';
import { BellOutlined } from '@ant-design/icons';
import { useGetNotificationsQuery, useReadNotificationMutation } from '@/services/api';

const NotificationBell: React.FC = () => {
  // Используем RTK Query хук
  const { data: notifications, isLoading } = useGetNotificationsQuery();
  const [readNotification] = useReadNotificationMutation();

  // Считаем непрочитанные
  const unreadCount = notifications?.filter((n: any) => !n.read_at).length || 0;

  const handleRead = async (id: string) => {
    try {
      await readNotification(id).unwrap();
    } catch (e) {
      console.error('Failed to mark as read', e);
    }
  };

  const menu = (
    <div style={{ width: 300, maxHeight: 400, overflowY: 'auto', background: '#fff', boxShadow: '0 2px 8px rgba(0,0,0,0.15)' }}>
      <div style={{ padding: '8px 12px', fontWeight: 'bold', borderBottom: '1px solid #f0f0f0' }}>
        Уведомления
      </div>
      {isLoading ? (
        <div style={{ padding: 20, textAlign: 'center' }}><Spin /></div>
      ) : !notifications || notifications.length === 0 ? (
        <Empty description="Нет уведомлений" style={{ padding: 20 }} />
      ) : (
        <List
          dataSource={notifications}
          renderItem={(item: any) => (
            <List.Item style={{ padding: '8px 12px', background: item.read_at ? '#fff' : '#e6f7ff' }}>
              <List.Item.Meta
                title={item.title || 'Уведомление'}
                description={item.message || item.content}
              />
              {!item.read_at && (
                <Button size="small" type="link" onClick={() => handleRead(item.id)}>
                  Прочитано
                </Button>
              )}
            </List.Item>
          )}
        />
      )}
    </div>
  );

  return (
    <Dropdown overlay={menu} trigger={['click']} placement="bottomRight">
      <Badge count={unreadCount} size="small" style={{ cursor: 'pointer' }}>
        <BellOutlined style={{ fontSize: '18px', cursor: 'pointer' }} />
      </Badge>
    </Dropdown>
  );
};

export default NotificationBell;

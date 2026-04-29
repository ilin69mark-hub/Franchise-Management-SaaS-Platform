// src/components/Dashboard/MorningBriefDrawer.tsx
import React, { useMemo } from 'react';
import { Drawer, Card, Row, Col, Typography, List, Tag, Space, Button, Divider, Statistic, Empty } from 'antd';
import { WarningOutlined, CheckCircleOutlined, ClockCircleOutlined, FileTextOutlined, PhoneOutlined, ArrowRightOutlined, SendOutlined, SettingOutlined, SmileOutlined, FrownOutlined } from '@ant-design/icons';
import { LineChart, Line, ResponsiveContainer, Tooltip as RechartsTooltip } from 'recharts';
import dayjs from 'dayjs';
import { useTerritoryManagerStore } from '@/store/territoryManagerStore';

const { Text, Title } = Typography;

interface BriefRiskItem {
  dealerName: string;
  percent: number;
  reason: string;
  recommendation: string;
}

interface BriefControlItem {
  dealerName: string;
  issue: string;
  urgency: 'high' | 'medium';
}

interface BriefGoodNews {
  dealerName: string;
  description: string;
}

interface BriefTask {
  id: string;
  type: 'call' | 'meeting' | 'task';
  dealerName: string;
  description: string;
  time?: string;
}

interface MorningBriefDrawerProps {
  open: boolean;
  onClose: () => void;
  userName?: string;
}

const MorningBriefDrawer: React.FC<MorningBriefDrawerProps> = ({ open, onClose, userName = 'Александр' }) => {
  const { summary, dealers } = useTerritoryManagerStore();
  const currentDate = dayjs().format('D MMMM YYYY');

  const overallStatus = useMemo(() => {
    const planPercent = summary?.planCompletionPercent || 87;
    const forecast = summary?.quarterForecastPercent || 94;
    if (planPercent >= 90) {
      return { text: `План территории выполнен на ${planPercent}%. Прогноз закрытия месяца — ${forecast}%.`, type: 'good' as const };
    }
    if (planPercent >= 70) {
      return { text: `План территории выполнен на ${planPercent}%. Прогноз закрытия месяца — ${forecast}%. Отставание некритичное.`, type: 'warning' as const };
    }
    return { text: `Внимание: план выполнен только на ${planPercent}%. Прогноз снижен из-за проблем у нескольких дилеров.`, type: 'critical' as const };
  }, [summary]);

  const riskItems = useMemo((): BriefRiskItem[] => [
    { dealerName: 'МебельЛига', percent: 68, reason: 'падение конверсии в салоне Б (менеджер Иванов)', recommendation: 'Провести аудит работы менеджера' },
    { dealerName: 'Евромебель', percent: 65, reason: 'низкий трафик, 3 недели без продаж', recommendation: 'Проверить маркетинговую активность' },
    { dealerName: 'Диванит Воронеж', percent: 72, reason: 'просроченная дебиторка 250т.р.', recommendation: 'Провести звонок по оплате' },
  ], []);

  const controlItems = useMemo((): BriefControlItem[] => [
    { dealerName: 'Евромебель', issue: '3 несогласованных возврата, один срок истекает сегодня', urgency: 'high' },
    { dealerName: 'Мебель Москва', issue: 'не предоставил еженедельный отчёт (просрочка 2 дня)', urgency: 'medium' },
    { dealerName: 'Салон мебели Казань', issue: 'остаток критичный по модели "Прима" (2 шт.)', urgency: 'medium' },
  ], []);

  const goodNews = useMemo((): BriefGoodNews[] => [
    { dealerName: 'Мебель Москва', description: 'выполнил план на 110% досрочно' },
    { dealerName: 'Салон мебели Казань', description: 'открыл новый салон, первые продажи выше ожиданий' },
    { dealerName: 'Диванит Воронеж', description: 'принял на работу сильного менеджера с рынка' },
  ], []);

  const todayTasks = useMemo((): BriefTask[] => [
    { id: '1', type: 'call', dealerName: 'МебельЛига', description: 'разбор воронки', time: '10:00' },
    { id: '2', type: 'call', dealerName: 'Салон мебели Казань', description: 'утверждение витрины', time: '12:00' },
    { id: '3', type: 'meeting', dealerName: 'Евромебель', description: 'поздравление с выполнением плана', time: '15:00' },
  ], []);

  const sparklineData = useMemo(() => [
    { day: 'Пн', plan: 12, fact: 11 },
    { day: 'Вт', plan: 12, fact: 13 },
    { day: 'Ср', plan: 12, fact: 12 },
    { day: 'Чт', plan: 12, fact: 10 },
    { day: 'Пт', plan: 12, fact: 14 },
    { day: 'Сб', plan: 12, fact: 15 },
    { day: 'Вс', plan: 12, fact: 13 },
  ], []);

  const getGreeting = () => {
    const hour = dayjs().hour();
    if (hour < 12) return 'Доброе утро';
    if (hour < 18) return 'Добрый день';
    return 'Добрый вечер';
  };

  const renderRiskItem = (item: BriefRiskItem) => (
    <List.Item>
      <Space direction="vertical" size={0}>
        <Space>
          <WarningOutlined style={{ color: '#ff4d4f' }} />
          <Text strong>{item.dealerName}</Text>
          <Tag color="red">{item.percent}% плана</Tag>
        </Space>
        <Text type="secondary" style={{ fontSize: 12 }}>{item.reason}</Text>
        <Text type="secondary" style={{ fontSize: 11, fontStyle: 'italic' }}>Рекомендация: {item.recommendation}</Text>
      </Space>
    </List.Item>
  );

  const renderControlItem = (item: BriefControlItem) => (
    <List.Item>
      <Space direction="vertical" size={0}>
        <Space>
          <ClockCircleOutlined style={{ color: item.urgency === 'high' ? '#ff4d4f' : '#fa8c16' }} />
          <Text strong>{item.dealerName}</Text>
          <Tag color={item.urgency === 'high' ? 'red' : 'orange'}>{item.urgency === 'high' ? 'Срочно' : 'На контроле'}</Tag>
        </Space>
        <Text type="secondary" style={{ fontSize: 12 }}>{item.issue}</Text>
      </Space>
    </List.Item>
  );

  const renderGoodNews = (item: BriefGoodNews) => (
    <List.Item>
      <Space>
        <CheckCircleOutlined style={{ color: '#52c41a' }} />
        <Text strong>{item.dealerName}</Text>
        <Text>{item.description}</Text>
      </Space>
    </List.Item>
  );

  const renderTask = (task: BriefTask) => (
    <List.Item>
      <Space>
        {task.type === 'call' && <PhoneOutlined />}
        {task.type === 'meeting' && <ClockCircleOutlined />}
        {task.type === 'task' && <FileTextOutlined />}
        <Text>{task.dealerName}: {task.description}</Text>
        {task.time && <Text type="secondary"> в {task.time}</Text>}
      </Space>
    </List.Item>
  );

  return (
    <Drawer
      title={
        <Space>
          <SmileOutlined />
          <span>{getGreeting()}, {userName}! Сводка по территории на {currentDate}</span>
        </Space>
      }
      placement="right"
      width={450}
      onClose={onClose}
      open={open}
      extra={
        <Button icon={<SettingOutlined />} onClick={() => {}}>
          Настройки
        </Button>
      }
    >
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        <Card size="small">
          <Space direction="vertical" size={0}>
            <Text>
              {overallStatus.text}
            </Text>
          </Space>
        </Card>

        <Divider orientation="left" style={{ margin: '8px 0' }}>
          <Space>
            <WarningOutlined style={{ color: '#ff4d4f' }} />
            <Text strong>Зона риска</Text>
            <Tag color="red">{riskItems.length}</Tag>
          </Space>
        </Divider>
        <List
          size="small"
          dataSource={riskItems}
          renderItem={renderRiskItem}
          locale={{ emptyText: 'Нет дилеров в красной зоне' }}
        />

        <Divider orientation="left" style={{ margin: '8px 0' }}>
          <Space>
            <ClockCircleOutlined style={{ color: '#fa8c16' }} />
            <Text strong>Зона контроля</Text>
          </Space>
        </Divider>
        <List
          size="small"
          dataSource={controlItems}
          renderItem={renderControlItem}
          locale={{ emptyText: 'Нет проблем на контроле' }}
        />

        <Divider orientation="left" style={{ margin: '8px 0' }}>
          <Space>
            <CheckCircleOutlined style={{ color: '#52c41a' }} />
            <Text strong>Хорошие новости</Text>
            <Tag color="green">{goodNews.length}</Tag>
          </Space>
        </Divider>
        {goodNews.length > 0 ? (
          <List size="small" dataSource={goodNews} renderItem={renderGoodNews} />
        ) : (
          <Empty description="Пока нет хороших новостей" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        )}

        <Divider orientation="left" style={{ margin: '8px 0' }}>
          <Space>
            <FileTextOutlined />
            <Text strong>Задачи на сегодня</Text>
            <Tag color="blue">{todayTasks.length}</Tag>
          </Space>
        </Divider>
        <List size="small" dataSource={todayTasks} renderItem={renderTask} />

        <Card size="small" title="План-факт за последние 7 дней">
          <ResponsiveContainer width="100%" height={80}>
            <LineChart data={sparklineData}>
              <Line type="monotone" dataKey="plan" stroke="#d9d9d9" strokeDasharray="3 3" dot={false} />
              <Line type="monotone" dataKey="fact" stroke="#52c41a" strokeWidth={2} dot={{ fill: '#52c41a', r: 3 }} />
              <RechartsTooltip />
            </LineChart>
          </ResponsiveContainer>
        </Card>

        <Divider />

        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" block icon={<ArrowRightOutlined />}>
            Открыть детали
          </Button>
          <Button block icon={<SendOutlined />}>
            Отправить руководителю
          </Button>
        </Space>
      </Space>
    </Drawer>
  );
};

export default MorningBriefDrawer;
// src/components/Dashboard/GoalCard.tsx
import React from 'react';
import { Card, Statistic, Row, Col, Spin, Alert } from 'antd';
import dayjs from 'dayjs';
import { useGetMyGoalQuery } from '@/services/goalApi';
import type { Goal } from '@/types';

interface GoalCardProps {
  /** Дата, к которой относится план (по умолчанию – сегодня) */
  date?: string; // формат 'YYYY-MM-DD'
}

/**
 * Компонент карты личного плана.
 *
 *  - Делает запрос `GET /goals?date=YYYY-MM-DD` (через RTK‑Query).
 *  - Если план найден – выводит KPI.
 *  - Если 404 → выводит сообщение «План не задан».
 *  - При других ошибках выводит `Alert` с текстом ошибки.
 */
const GoalCard: React.FC<GoalCardProps> = ({ date }) => {
  // Дата, по которой ищем план (по умолчанию – сегодня)
  const targetDate = date ?? dayjs().format('YYYY-MM-DD');

  // RTK‑Query‑хук
  const { data, isLoading, error, isError } = useGetMyGoalQuery(targetDate);

  // ---------- Состояния загрузки ----------
  if (isLoading) {
    return <Spin tip="Загрузка плана…" />;
  }

  // ---------- Обработка ошибок ----------
  // Если сервер вернул 404 → план пока не создан
  if (isError && (error as any)?.status === 404) {
    return <Alert type="info" message="План не задан" showIcon />;
  }

  // Любые другие ошибки (401, 500, сеть и т.д.)
  if (isError) {
    const errMsg =
      (error as any)?.data?.error ||
      (error as any)?.message ||
      'Не удалось загрузить план';
    return <Alert type="error" message={errMsg} showIcon />;
  }

  // ---------- Данные есть ----------
  const goal = data as Goal; // data гарантировано не undefined после проверки ошибок

  return (
    <Card
      title={`Мой план на ${dayjs(targetDate).format('DD MMM YYYY')}`}
      bordered
      style={{ marginBottom: 24 }}
    >
      <Row gutter={16}>
        <Col span={12}>
          <Statistic title="Продажи (₽)" value={goal.sales_plan} />
        </Col>
        <Col span={12}>
          <Statistic title="Лиды" value={goal.leads_plan} />
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Statistic title="Звонки" value={goal.calls_plan} />
        </Col>
        <Col span={12}>
          <Statistic title="Встречи" value={goal.meetings_plan} />
        </Col>
      </Row>
    </Card>
  );
};

export default GoalCard;

import React, { useState, useEffect, useCallback } from 'react';
import { Drawer, Card, Typography, Progress, Tag, Row, Col, Button, Spin, Alert, Tooltip } from 'antd';
import { FlagOutlined, WarningOutlined, GiftOutlined, PercentageOutlined, DollarOutlined, ArrowUpOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import dayjs from 'dayjs';

const { Text, Title } = Typography;

interface Promotion {
  id: string;
  name: string;
  condition: string;
  discount_min: number;
  discount_max: number;
  end_date: string;
  is_expiring: boolean;
}

interface TargetPlan {
  total_amount: number;
  current_amount: number;
  percent: number;
}

interface ManagerTargetsData {
  has_targets: boolean;
  plan: TargetPlan | null;
  target_conversion: number;
  current_conversion: number;
  target_extras_percent: number;
  current_extras_percent: number;
  promotions: Promotion[];
  bonus_forecast: number;
  max_bonus: number;
  warning_level: string;
}

interface DealerDirectivesProps {
  user: any;
}

const DealerDirectives: React.FC<DealerDirectivesProps> = ({ user }) => {
  const [open, setOpen] = useState(false);
  const [data, setData] = useState<ManagerTargetsData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchTargets = useCallback(async () => {
    try {
      setLoading(true);
      const date = dayjs().format('YYYY-MM-DD');
      const res = await apiClient.get(`/manager/targets?date=${date}`);
      setData(res.data);
      setError(null);
    } catch (e: any) {
      setError(e?.response?.data?.error || 'Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open && !data) {
      fetchTargets();
    }
  }, [open, data, fetchTargets]);

  const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);

  const getPlanColor = (percent: number) => {
    if (percent >= 80) return '#52c41a';
    if (percent >= 50) return '#faad14';
    return '#ff4d4f';
  };

  const getConversionColor = (current: number, target: number) => {
    if (current >= target) return '#52c41a';
    if (current >= target * 0.7) return '#faad14';
    return '#ff4d4f';
  };

  const getWarningColor = (level: string) => {
    switch (level) {
      case 'red': return '#ff4d4f';
      case 'yellow': return '#faad14';
      default: return '#52c41a';
    }
  };

  return (
    <>
      <Tooltip title="Директивы от дилера">
        <Button
          type="text"
          icon={<FlagOutlined style={{ fontSize: 18, color: data?.has_targets ? '#1890ff' : '#999' }} />}
          onClick={() => setOpen(true)}
          style={{ marginRight: 8 }}
        >
          Директивы
          {data?.warning_level === 'yellow' && (
            <WarningOutlined style={{ color: '#faad14', marginLeft: 4 }} />
          )}
        </Button>
      </Tooltip>

      <Drawer
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <FlagOutlined />
            <span>Директивы от дилера</span>
          </div>
        }
        placement="top"
        height={400}
        onClose={() => setOpen(false)}
        open={open}
      >
        {loading ? (
          <div style={{ textAlign: 'center', padding: 40 }}><Spin size="large" /></div>
        ) : error ? (
          <Alert message="Ошибка" description={error} type="error" action={<a onClick={fetchTargets}>Повторить</a>} />
        ) : !data?.has_targets ? (
          <Alert
            message="План на месяц ещё не утверждён"
            description="Дилер ещё не спустил план продаж. Ожидайте утверждения."
            type="info"
            showIcon
          />
        ) : (
          <div>
            {/* План продаж */}
            <Card title="План продаж на месяц" size="small" style={{ marginBottom: 16 }}>
              <Row gutter={16}>
                <Col span={16}>
                  <div style={{ marginBottom: 8 }}>
                    <Text strong>Выполнение плана: {data?.plan?.percent || 0}%</Text>
                  </div>
                  <Progress
                    percent={data?.plan?.percent || 0}
                    strokeColor={getPlanColor(data?.plan?.percent || 0)}
                    format={(p) => `${formatMoney(data?.plan?.current_amount || 0)} / ${formatMoney(data?.plan?.total_amount || 0)} ₽`}
                  />
                </Col>
                <Col span={8} style={{ textAlign: 'right' }}>
                  <Text type="secondary">Осталось:</Text>
                  <div>
                    <Text strong style={{ fontSize: 18 }}>
                      {formatMoney((data?.plan?.total_amount || 0) - (data?.plan?.current_amount || 0))} ₽
                    </Text>
                  </div>
                </Col>
              </Row>
            </Card>

            {/* Бенчмарки */}
            <Card title="Бенчмарки" size="small" style={{ marginBottom: 16 }}>
              <Row gutter={16}>
                <Col span={12}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <PercentageOutlined style={{ fontSize: 20 }} />
                    <div>
                      <Text type="secondary">Конверсия из замера в договор</Text>
                      <div>
                        <Text strong style={{ color: getConversionColor(data?.current_conversion || 0, data?.target_conversion || 0) }}>
                          {(data?.current_conversion || 0).toFixed(1)}%
                        </Text>
                        <Text type="secondary"> / {(data?.target_conversion || 0).toFixed(0)}%</Text>
                      </div>
                    </div>
                  </div>
                </Col>
                <Col span={12}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <GiftOutlined style={{ fontSize: 20 }} />
                    <div>
                      <Text type="secondary">Доля допов в чеке</Text>
                      <div>
                        <Text strong style={{ color: getConversionColor(data?.current_extras_percent || 0, data?.target_extras_percent || 0) }}>
                          {(data?.current_extras_percent || 0).toFixed(1)}%
                        </Text>
                        <Text type="secondary"> / {(data?.target_extras_percent || 0).toFixed(0)}%</Text>
                      </div>
                    </div>
                  </div>
                </Col>
              </Row>
            </Card>

            {/* Акции */}
            <Card title="Акции" size="small" style={{ marginBottom: 16 }}>
              {data?.promotions?.map((promo) => (
                <div key={promo.id} style={{ marginBottom: 12, padding: 8, background: promo.is_expiring ? '#fff7e6' : '#fafafa', borderRadius: 4, borderLeft: promo.is_expiring ? '3px solid #faad14' : '3px solid #52c41a' }}>
                  <Row justify="space-between" align="middle">
                    <Col>
                      <Text strong>{promo.name}</Text>
                      <br />
                      <Text type="secondary">{promo.condition}</Text>
                    </Col>
                    <Col>
                      <Tag color={promo.is_expiring ? 'warning' : 'success'}>
                        {promo.discount_min === promo.discount_max
                          ? `${promo.discount_min}%`
                          : `${promo.discount_min}-${promo.discount_max}%`}
                      </Tag>
                      {promo.is_expiring && (
                        <WarningOutlined style={{ color: '#faad14', marginLeft: 4 }} />
                      )}
                      <div style={{ textAlign: 'right', marginTop: 4 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          До {dayjs(promo.end_date).format('DD MMM')}
                        </Text>
                      </div>
                    </Col>
                  </Row>
                </div>
              ))}
            </Card>

            {/* Прогноз премии */}
            <Card
              size="small"
              style={{
                background: data.warning_level === 'yellow' ? '#fffbe6' :
                           data.warning_level === 'red' ? '#fff1f0' : '#f6ffed',
              }}
            >
              <Row justify="space-between" align="middle">
                <Col>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <DollarOutlined style={{ fontSize: 20, color: getWarningColor(data.warning_level) }} />
                    <div>
                      <Text>Прогноз личной премии</Text>
                      <div>
                        <Title level={4} style={{ margin: 0, color: getWarningColor(data.warning_level) }}>
                          {formatMoney(data.bonus_forecast)} ₽
                        </Title>
                        <Text type="secondary">из {formatMoney(data.max_bonus)} ₽ максимум</Text>
                      </div>
                    </div>
                  </div>
                </Col>
                <Col>
                  <Progress
                    type="circle"
                    percent={Math.round((data.bonus_forecast / data.max_bonus) * 100)}
                    strokeColor={getWarningColor(data.warning_level)}
                    width={60}
                  />
                </Col>
              </Row>
              {data.warning_level === 'yellow' && (
                <Alert
                  message="Внимание"
                  description="При текущем темпе вы получите менее 50% премии. Ускорьте темп!"
                  type="warning"
                  showIcon
                  icon={<ExclamationCircleOutlined />}
                  style={{ marginTop: 12 }}
                />
              )}
            </Card>
          </div>
        )}
      </Drawer>
    </>
  );
};

export default DealerDirectives;
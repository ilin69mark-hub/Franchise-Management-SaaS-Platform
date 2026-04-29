// src/components/Dashboard/tabs/ProfitTab.tsx
import React, { useState, useMemo } from 'react';
import { Card, Row, Col, Statistic, Typography, Spin, Table, Button, Alert, Space, Tag, Empty, Collapse } from 'antd';
import { RiseOutlined, FallOutlined, WarningOutlined, ReloadOutlined, ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import UnitEconomyCalculator from './UnitEconomyCalculator';
import { useGetUnitTemplatesQuery, useCreateUnitTemplateMutation } from '@/services/api';

const { Text } = Typography;

interface ExpenseBreakdown {
  category: string;
  amount: number;
  percent_of_revenue: number;
  prev_month_amount: number;
}

interface FinanceData {
  revenue: number;
  cogs: number;
  rent: number;
  utilities: number;
  payroll: number;
  taxes: number;
  logistics: number;
  marketing: number;
  defects: number;
  other_expenses: number;
  bonus: number;
  net_profit: number;
  net_profit_forecast: number;
  prev_month_net_profit: number;
  expense_breakdown: ExpenseBreakdown[];
}

interface ProfitTabProps {
  data?: FinanceData;
  loading?: boolean;
  error?: boolean;
  onRetry?: () => void;
  onNavigateToExpenseForm?: () => void;
}

const ProfitTab: React.FC<ProfitTabProps> = ({ data, loading, error, onRetry, onNavigateToExpenseForm }) => {
  const [selectedMonth] = useState(dayjs().format('YYYY-MM'));
  const { data: templatesData, isLoading: templatesLoading } = useGetUnitTemplatesQuery();
  const [createUnitTemplate] = useCreateUnitTemplateMutation();

  const profitChange = useMemo(() => {
    if (!data?.net_profit || !data.prev_month_net_profit) return null;
    return ((data.net_profit - data.prev_month_net_profit) / data.prev_month_net_profit) * 100;
  }, [data?.net_profit, data?.prev_month_net_profit]);

  const marginProfit = (data?.revenue || 0) - (data?.cogs || 0);

  const hasExpenses = useMemo(() => {
    if (!data) return false;
    return data.rent > 0 || data.utilities > 0 || data.payroll > 0 || 
           data.taxes > 0 || data.logistics > 0 || data.marketing > 0 || 
           data.defects > 0 || data.other_expenses > 0;
  }, [data]);

  const expenses = useMemo(() => {
    if (!data) return [];
    return [
      { name: 'Себестоимость', value: data.cogs },
      { name: 'Аренда', value: data.rent },
      { name: 'ФОТ', value: data.payroll },
      { name: 'Налоги', value: data.taxes },
      { name: 'Логистика', value: data.logistics },
      { name: 'Бонусы', value: data.bonus || 0 },
      { name: 'Брак', value: data.defects },
      { name: 'Прочее', value: data.other_expenses },
    ].filter(item => item.value > 0);
  }, [data]);

  const pieData = useMemo(() => {
    if (!data) return [];
    const total = data.rent + data.utilities + data.payroll + data.taxes + data.logistics + data.marketing + data.defects + data.other_expenses;
    if (total === 0) return [];
    return [
      { category: 'Аренда', amount: data.rent, percent: (data.rent / total) * 100 },
      { category: 'Коммуналка', amount: data.utilities, percent: (data.utilities / total) * 100 },
      { category: 'ФОТ', amount: data.payroll, percent: (data.payroll / total) * 100 },
      { category: 'Налоги', amount: data.taxes, percent: (data.taxes / total) * 100 },
      { category: 'Логистика', amount: data.logistics, percent: (data.logistics / total) * 100 },
      { category: 'Маркетинг', amount: data.marketing, percent: (data.marketing / total) * 100 },
      { category: 'Брак', amount: data.defects, percent: (data.defects / total) * 100 },
      { category: 'Прочее', amount: data.other_expenses, percent: (data.other_expenses / total) * 100 },
    ].filter(item => item.amount > 0);
  }, [data]);

  const isRentHigh = data ? (data.rent / data.revenue) * 100 > 12 : false;
  const isPayrollHigh = data ? (data.payroll / data.revenue) * 100 > 20 : false;

  const expenseColumns = [
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => {
        const colors: Record<string, string> = {
          'Аренда': 'blue',
          'Коммуналка': 'cyan',
          'ФОТ': 'purple',
          'Налоги': 'orange',
          'Логистика': 'geekblue',
          'Маркетинг': 'green',
          'Брак': 'red',
          'Прочее': 'gold',
        };
        return <Tag color={colors[category] || 'default'}>{category}</Tag>;
      },
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (val: number) => <Text strong>{val.toLocaleString()} ₽</Text>,
    },
    {
      title: '% от выручки',
      dataIndex: 'percent_of_revenue',
      key: 'percent_of_revenue',
      render: (val: number) => <Text>{val.toFixed(1)}%</Text>,
    },
    {
      title: 'К прошлому месяцу',
      key: 'change',
      render: (_: any, record: ExpenseBreakdown) => {
        const change = record.prev_month_amount > 0 
          ? ((record.amount - record.prev_month_amount) / record.prev_month_amount) * 100 
          : 0;
        if (change === 0) return <Text type="secondary">-</Text>;
        return (
          <Space>
            {change > 0 ? <ArrowUpOutlined style={{ color: '#ff4d4f' }} /> : <ArrowDownOutlined style={{ color: '#52c41a' }} />}
            <Text type={change > 0 ? 'danger' : 'success'}>{Math.abs(change).toFixed(1)}%</Text>
          </Space>
        );
      },
    },
  ];

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert
        message="Ошибка загрузки данных"
        description="Не удалось загрузить финансовые данные. Попробуйте повторить запрос."
        type="error"
        showIcon
        action={
          <Button icon={<ReloadOutlined />} onClick={onRetry}>
            Повторить
          </Button>
        }
      />
    );
  }

  if (!hasExpenses && data) {
    return (
      <Alert
        message="Заполните расходы"
        description="Для расчёта чистой прибыли необходимо внести данные по расходам за текущий месяц."
        type="warning"
        showIcon
        action={
          <Button type="primary" onClick={onNavigateToExpenseForm}>
            Перейти к форме
          </Button>
        }
      />
    );
  }

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card style={{ background: (data?.net_profit || 0) >= 0 ? '#e6f7ff' : '#fff1f0' }}>
            <Statistic
              title={
                <Space>
                  Чистая прибыль (Net Profit)
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    за {dayjs(selectedMonth).format('MMMM YYYY')}
                  </Text>
                </Space>
              }
              value={data?.net_profit || 0}
              precision={0}
              prefix="₽ "
              valueStyle={{ 
                fontSize: 28, 
                color: (data?.net_profit || 0) >= 0 ? '#1890ff' : '#ff4d4f' 
              }}
              suffix={
                profitChange !== null ? (
                  <Text type={profitChange >= 0 ? 'success' : 'danger'} style={{ fontSize: 14 }}>
                    {' '}
                    {profitChange >= 0 ? <RiseOutlined /> : <FallOutlined />}
                    {' '}{Math.abs(profitChange).toFixed(1)}% к прошлому месяцу
                  </Text>
                ) : null
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Валовый оборот"
              value={data?.revenue || 0}
              precision={0}
              prefix="₽ "
              valueStyle={{ fontSize: 20 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Маржинальная прибыль"
              value={marginProfit || 0}
              precision={0}
              prefix="₽ "
              suffix={
                data?.revenue ? (
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {' '}({((marginProfit / data.revenue) * 100).toFixed(1)}%)
                  </Text>
                ) : null
              }
              valueStyle={{ fontSize: 20 }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Прогноз чистой прибыли"
              value={data?.net_profit_forecast || 0}
              precision={0}
              prefix="₽ до конца месяца"
              valueStyle={{ fontSize: 20, color: (data?.net_profit_forecast || 0) >= 0 ? '#1890ff' : '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>

      {(isRentHigh || isPayrollHigh) && (
        <Alert
          message="Внимание: превышены нормативы"
          description={
            <Space direction="vertical">
              {isRentHigh && <Text>Аренда превышает 12% от выручки: {(((data?.rent || 0) / (data?.revenue || 1)) * 100).toFixed(1)}%</Text>}
              {isPayrollHigh && <Text>ФОТ превышает 20% от выручки: {(((data?.payroll || 0) / (data?.revenue || 1)) * 100).toFixed(1)}%</Text>}
            </Space>
          }
          type="error"
          showIcon
          style={{ marginTop: 16 }}
          icon={<WarningOutlined />}
        />
      )}

      <Card title="Путь от выручки к чистой прибыли" style={{ marginTop: 16 }}>
        {expenses.length > 0 ? (
          <Row gutter={[8, 8]}>
            <Col xs={24}>
              <div style={{ padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
                <Text strong>Выручка</Text>
                <Text style={{ float: 'right', color: '#52c41a' }}>{(data?.revenue || 0).toLocaleString()} ₽</Text>
              </div>
            </Col>
            {expenses.map((expense, idx) => {
              const runningTotal = (data?.revenue || 0) - expenses.slice(0, idx + 1).reduce((sum, e) => sum + e.value, 0);
              return (
                <Col xs={24} key={expense.name}>
                  <div style={{ padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
                    <Text type="secondary">- {expense.name}</Text>
                    <Text style={{ float: 'right', color: '#ff4d4f' }}>{expense.value.toLocaleString()} ₽</Text>
                    <div style={{ fontSize: 12, color: '#888' }}>
                      Остаток: {runningTotal.toLocaleString()} ₽
                    </div>
                  </div>
                </Col>
              );
            })}
            <Col xs={24}>
              <div style={{ padding: '12px 0', background: (data?.net_profit || 0) >= 0 ? '#e6f7ff' : '#fff1f0' }}>
                <Text strong>Чистая прибыль</Text>
                <Text style={{ float: 'right', fontSize: 18, color: (data?.net_profit || 0) >= 0 ? '#1890ff' : '#ff4d4f' }}>
                  {(data?.net_profit || 0).toLocaleString()} ₽
                </Text>
              </div>
            </Col>
          </Row>
        ) : (
          <Empty description="Нет данных" />
        )}
      </Card>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="Структура затрат">
            {pieData.length > 0 ? (
              <div>
                {pieData.map(item => (
                  <div key={item.category} style={{ marginBottom: 8 }}>
                    <Row>
                      <Col span={12}>
                        <Tag color={
                          item.category === 'Аренда' ? 'blue' :
                          item.category === 'ФОТ' ? 'purple' :
                          item.category === 'Налоги' ? 'orange' :
                          item.category === 'Логистика' ? 'geekblue' :
                          item.category === 'Брак' ? 'red' : 'default'
                        }>{item.category}</Tag>
                      </Col>
                      <Col span={6}>
                        <Text>{item.amount.toLocaleString()} ₽</Text>
                      </Col>
                      <Col span={6}>
                        <Text type="secondary">{item.percent.toFixed(1)}%</Text>
                      </Col>
                    </Row>
                  </div>
                ))}
              </div>
            ) : (
              <Empty description="Нет данных о затратах" />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Детализация расходов">
            <Table
              dataSource={data?.expense_breakdown || []}
              columns={expenseColumns}
              rowKey="category"
              pagination={false}
              size="small"
              locale={{ emptyText: 'Нет данных о расходах' }}
            />
          </Card>
        </Col>
      </Row>

      <Collapse 
        ghost 
        style={{ marginTop: 16 }}
        items={[
          {
            key: 'unit-calculator',
            label: '🧮 Калькулятор unit-экономики',
            children: templatesLoading ? (
              <Spin />
            ) : (
              <UnitEconomyCalculator
                templates={templatesData || []}
                onSaveTemplate={async (template) => {
                  await createUnitTemplate(template).unwrap();
                }}
              />
            ),
          },
        ]}
      />
    </div>
  );
};

export default ProfitTab;
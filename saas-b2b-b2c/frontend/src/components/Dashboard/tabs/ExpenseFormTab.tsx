// src/components/Dashboard/tabs/ExpenseFormTab.tsx
import React, { useState, useEffect, useMemo } from 'react';
import { 
  Card, Form, InputNumber, Button, Row, Col, DatePicker, Space, 
  Typography, Spin, Alert, Upload, message, Tag, Divider, Popconfirm, Radio 
} from 'antd';
import { 
  SaveOutlined, UploadOutlined, CopyOutlined, FileTextOutlined, 
  WarningOutlined, DeleteOutlined, CalculatorOutlined 
} from '@ant-design/icons';
import dayjs from 'dayjs';
import type { UploadProps } from 'antd';

const { Text, Title } = Typography;
const { RangePicker } = DatePicker;

interface ExpenseRecord {
  id?: string;
  month: string;
  rent: number;
  utilities: number;
  payroll: number;
  taxes: number;
  logistics: number;
  marketing: number;
  defects: number;
  other_expenses: number;
  other_expense_name?: string;
  total: number;
}

interface ExpenseFormTabProps {
  onSave?: (data: ExpenseRecord) => Promise<void>;
  onImport?: (file: File) => Promise<void>;
}

const ExpenseFormTab: React.FC<ExpenseFormTabProps> = ({ onSave, onImport }) => {
  const [form] = Form.useForm();
  const [selectedMonth, setSelectedMonth] = useState(dayjs());
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [importing, setImporting] = useState(false);
  const [initialData, setInitialData] = useState<ExpenseRecord | null>(null);
  const [prevMonthData, setPrevMonthData] = useState<ExpenseRecord | null>(null);
  const [taxType, setTaxType] = useState<'usn' | 'patent' | 'ndfl'>('usn');
  const [hasChanges, setHasChanges] = useState(false);

  const monthStr = selectedMonth.format('YYYY-MM');

  useEffect(() => {
    fetchExpenses();
  }, [selectedMonth]);

  const fetchExpenses = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('accessToken');
      const res = await fetch(`/api/dealer/expenses?month=${monthStr}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();
        setInitialData(data);
        form.setFieldsValue(data);
      } else {
        setInitialData(null);
        form.resetFields();
        const prevRes = await fetch(`/api/dealer/expenses?month=${dayjs(selectedMonth).subtract(1, 'month').format('YYYY-MM')}`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (prevRes.ok) {
          const prevData = await prevRes.json();
          setPrevMonthData(prevData);
        }
      }
    } catch (e) {
      console.error('Error fetching expenses', e);
    } finally {
      setLoading(false);
    }
  };

  const handleValuesChange = () => {
    setHasChanges(true);
  };

  const calculateTax = (revenue: number) => {
    switch (taxType) {
      case 'usn':
        return Math.round(revenue * 0.06);
      case 'patent':
        return 65000;
      case 'ndfl':
        return Math.round(revenue * 0.13);
      default:
        return 0;
    }
  };

  const handleCopyFromPrevMonth = () => {
    if (prevMonthData) {
      form.setFieldsValue({
        rent: prevMonthData.rent,
        utilities: prevMonthData.utilities,
        payroll: prevMonthData.payroll,
        logistics: prevMonthData.logistics,
        marketing: prevMonthData.marketing,
        defects: prevMonthData.defects,
        other_expenses: prevMonthData.other_expenses,
        other_expense_name: prevMonthData.other_expense_name,
      });
      setHasChanges(true);
      message.success('Данные скопированы из прошлого месяца');
    }
  };

  const handleSave = async (values: any) => {
    setSaving(true);
    try {
      const total = 
        (values.rent || 0) + 
        (values.utilities || 0) + 
        (values.payroll || 0) + 
        (values.taxes || 0) + 
        (values.logistics || 0) + 
        (values.marketing || 0) + 
        (values.defects || 0) + 
        (values.other_expenses || 0);

      const payload = {
        month: monthStr,
        rent: values.rent || 0,
        utilities: values.utilities || 0,
        payroll: values.payroll || 0,
        taxes: values.taxes || 0,
        logistics: values.logistics || 0,
        marketing: values.marketing || 0,
        defects: values.defects || 0,
        other_expenses: values.other_expenses || 0,
        other_expense_name: values.other_expense_name || '',
        total,
      };

      if (onSave) {
        await onSave(payload);
      } else {
        const token = localStorage.getItem('accessToken');
        const res = await fetch('/api/dealer/expenses', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify(payload),
        });
        if (res.ok) {
          message.success('Расходы сохранены');
          setInitialData(payload);
          setHasChanges(false);
          fetchExpenses();
        } else {
          message.error('Ошибка сохранения');
        }
      }
    } catch (e) {
      message.error('Ошибка сохранения');
    } finally {
      setSaving(false);
    }
  };

  const handleImport: UploadProps['customRequest'] = async (options) => {
    setImporting(true);
    const { file, onSuccess, onError } = options;
    
    try {
      if (onImport) {
        await onImport(file as File);
        onSuccess?.('ok');
      } else {
        const formData = new FormData();
        formData.append('file', file);
        
        const token = localStorage.getItem('accessToken');
        const res = await fetch('/api/dealer/expenses/import', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
          body: formData,
        });
        
        if (res.ok) {
          const data = await res.json();
          form.setFieldsValue(data);
          setHasChanges(true);
          message.success('Импорт завершён');
          onSuccess?.('ok');
        } else {
          message.error('Ошибка импорта');
          onError?.(new Error('Import failed'));
        }
      }
    } catch (e) {
      message.error('Ошибка импорта');
      onError?.(e as Error);
    } finally {
      setImporting(false);
    }
  };

  const fields = [
    { name: 'rent', label: 'Аренда помещения', prefix: '₽', prevKey: 'rent' },
    { name: 'utilities', label: 'Коммунальные платежи', prefix: '₽', prevKey: 'utilities' },
    { name: 'payroll', label: 'Фонд оплаты труда', prefix: '₽', prevKey: 'payroll' },
    { name: 'logistics', label: 'Логистика и доставка', prefix: '₽', prevKey: 'logistics' },
    { name: 'marketing', label: 'Маркетинг и реклама', prefix: '₽', prevKey: 'marketing' },
    { name: 'defects', label: 'Брак и рекламации', prefix: '₽', prevKey: 'defects' },
    { name: 'other_expenses', label: 'Прочие расходы', prefix: '₽', prevKey: 'other_expenses', hasName: true },
  ];

  const prevMonthFields = ['rent', 'utilities', 'payroll', 'logistics', 'marketing', 'defects', 'other_expenses'];
  const hasPrevMonthData = prevMonthData && prevMonthFields.some(f => (prevMonthData as any)[f] > 0);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
          onValuesChange={handleValuesChange}
          initialValues={{
            month: dayjs(),
            rent: 0,
            utilities: 0,
            payroll: 0,
            taxes: 0,
            logistics: 0,
            marketing: 0,
            defects: 0,
            other_expenses: 0,
          }}
        >
          <Row gutter={[16, 0]} align="middle">
            <Col xs={24} md={12}>
              <Space>
                <Text strong>Месяц:</Text>
                <DatePicker.MonthPicker
                  value={selectedMonth}
                  onChange={(date) => setSelectedMonth(date || dayjs())}
                  format="MMMM YYYY"
                  allowClear={false}
                />
              </Space>
            </Col>
            <Col xs={24} md={12}>
              <Space wrap>
                <Upload
                  customRequest={handleImport}
                  showUploadList={false}
                  accept=".csv,.xlsx,.xls"
                  disabled={importing}
                >
                  <Button icon={<UploadOutlined />} loading={importing}>
                    Импорт из выписки
                  </Button>
                </Upload>
                {hasPrevMonthData && (
                  <Button 
                    icon={<CopyOutlined />} 
                    onClick={handleCopyFromPrevMonth}
                    title="Скопировать из прошлого месяца"
                  >
                    С прошлого месяца
                  </Button>
                )}
              </Space>
            </Col>
          </Row>

          <Divider />

          <Row gutter={[16, 16]}>
            <Col xs={24} md={12}>
              <Form.Item
                name="rent"
                label="Аренда помещения"
                tooltip={prevMonthData?.rent ? `Прошлый месяц: ${prevMonthData.rent.toLocaleString()} ₽` : undefined}
                rules={[{ type: 'number', min: 0, message: 'Сумма должна быть >= 0' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="utilities"
                label="Коммунальные платежи"
                tooltip={prevMonthData?.utilities ? `Прошлый месяц: ${prevMonthData.utilities.toLocaleString()} ₽` : undefined}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="payroll"
                label="Фонд оплаты труда (ФОТ)"
                tooltip={prevMonthData?.payroll ? `Прошлый месяц: ${prevMonthData.payroll.toLocaleString()} ₽` : undefined}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="taxes"
                label={
                  <Space>
                    Налоги
                    <Radio.Group 
                      value={taxType} 
                      onChange={(e) => setTaxType(e.target.value)}
                      size="small"
                    >
                      <Radio.Button value="usn">УСН</Radio.Button>
                      <Radio.Button value="patent">Патент</Radio.Button>
                      <Radio.Button value="ndfl">НДФЛ</Radio.Button>
                    </Radio.Group>
                  </Space>
                }
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="logistics"
                label="Логистика и доставка"
                tooltip={prevMonthData?.logistics ? `Прошлый месяц: ${prevMonthData.logistics.toLocaleString()} ₽` : undefined}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="marketing"
                label="Маркетинг и реклама"
                tooltip={prevMonthData?.marketing ? `Прошлый месяц: ${prevMonthData.marketing.toLocaleString()} ₽` : undefined}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="defects"
                label="Брак и рекламации"
                tooltip={prevMonthData?.defects ? `Прошлый месяц: ${prevMonthData.defects.toLocaleString()} ₽` : undefined}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24} md={12}>
              <Form.Item
                name="other_expenses"
                label="Прочие расходы"
                tooltip={prevMonthData?.other_expenses ? `Прошлый месяц: ${prevMonthData.other_expenses.toLocaleString()} ₽` : undefined}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={(value) => value?.replace(/ /g, '') as any}
                  placeholder="0"
                  min={0}
                />
              </Form.Item>
            </Col>

            <Col xs={24}>
              <Form.Item
                name="other_expense_name"
                label="Наименование прочих расходов"
              >
                <InputNumber
                  style={{ width: '100%' }}
                  placeholder="Например: связь, охрана, канцтовары"
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} md={12}>
              <Card size="small" style={{ background: '#f0f5ff' }}>
                <Space>
                  <CalculatorOutlined />
                  <Text strong>Итого расходов: </Text>
                  <Text strong style={{ fontSize: 18 }}>
                    <Form.Item name="total" noStyle>
                      {(form.getFieldValue('rent') || 0) + 
                       (form.getFieldValue('utilities') || 0) + 
                       (form.getFieldValue('payroll') || 0) + 
                       (form.getFieldValue('taxes') || 0) + 
                       (form.getFieldValue('logistics') || 0) + 
                       (form.getFieldValue('marketing') || 0) + 
                       (form.getFieldValue('defects') || 0) + 
                       (form.getFieldValue('other_expenses') || 0)} ₽
                    </Form.Item>
                  </Text>
                </Space>
              </Card>
            </Col>
            <Col xs={24} md={12}>
              <Space>
                <Popconfirm
                  title="Сохранить изменения?"
                  onConfirm={() => form.submit()}
                  okText="Сохранить"
                  cancelText="Отмена"
                >
                  <Button 
                    type="primary" 
                    icon={<SaveOutlined />} 
                    loading={saving}
                    disabled={!hasChanges}
                  >
                    Сохранить
                  </Button>
                </Popconfirm>
                <Popconfirm
                  title="Очистить форму?"
                  onConfirm={() => {
                    form.resetFields();
                    setHasChanges(false);
                  }}
                  okText="Очистить"
                  cancelText="Отмена"
                >
                  <Button icon={<DeleteOutlined />}>
                    Очистить
                  </Button>
                </Popconfirm>
              </Space>
            </Col>
          </Row>
        </Form>
      </Card>

      {hasPrevMonthData && !initialData && (
        <Alert
          message="Есть данные за прошлый месяц"
          description="Нажмите 'С прошлого месяца' для автозаполнения или заполните форму вручную."
          type="info"
          showIcon
          style={{ marginTop: 16 }}
          icon={<WarningOutlined />}
        />
      )}
    </div>
  );
};

export default ExpenseFormTab;
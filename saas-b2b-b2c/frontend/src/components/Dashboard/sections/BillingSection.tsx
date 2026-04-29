import React, { useEffect, useState } from 'react';
import {
  Card,
  Row,
  Col,
  Table,
  Input,
  Select,
  Button,
  Modal,
  Form,
  Input as InputForm,
  DatePicker,
  Typography,
  Spin,
  Tag,
  Statistic,
  message,
} from 'antd';
import {
  DollarOutlined,
  PlusOutlined,
  ExportOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  SendOutlined,
} from '@ant-design/icons';
import apiClient from '@/api/axiosClient';
import { useBillingStore } from '@/store/billingStore';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { RangePicker } = DatePicker;

const BillingSection: React.FC = () => {
  const {
    invoices,
    settings,
    history,
    period,
    isLoading,
    showInvoiceModal,
    showPaymentModal,
    setInvoices,
    setSettings,
    setHistory,
    setPeriod,
    setLoading,
    setShowInvoiceModal,
    setShowPaymentModal,
  } = useBillingStore();

  const [localLoading, setLocalLoading] = useState(true);
  const [invoiceForm] = Form.useForm();
  const [paymentForm] = Form.useForm();
  const [filters, setFilters] = useState({ status: '', search: '', tenant: '' });

  useEffect(() => {
    fetchData();
  }, [period]);

  const fetchData = async () => {
    setLoading(true);
    setLocalLoading(true);
    try {
      const [invoicesRes, settingsRes, historyRes] = await Promise.all([
        apiClient.get(`/admin/billing/invoices?period=${period}`),
        apiClient.get('/admin/billing/settings'),
        apiClient.get(`/admin/billing/history?months=12`),
      ]);

      setInvoices(invoicesRes.data || []);
      setSettings(settingsRes.data || {});
      setHistory(historyRes.data || []);
    } catch {
      setInvoices([
        {
          id: '1',
          tenant: 'ООО Техно',
          invoiceNumber: 'INV-2026-001',
          period: 'Апрель 2026',
          amount: 40000,
          issuedAt: '2026-04-01',
          paidAt: '2026-04-05',
          status: 'paid',
          paymentMethod: 'auto',
        },
        {
          id: '2',
          tenant: 'АО Бизнес',
          invoiceNumber: 'INV-2026-002',
          period: 'Апрель 2026',
          amount: 150000,
          issuedAt: '2026-04-01',
          status: 'pending',
          paymentMethod: 'manual',
        },
        {
          id: '3',
          tenant: 'ИП Сидоров',
          invoiceNumber: 'INV-2026-003',
          period: 'Март 2026',
          amount: 5000,
          issuedAt: '2026-03-01',
          status: 'overdue',
          paymentMethod: 'manual',
        },
      ]);

      setSettings({
        notifyAfterDays: 3,
        secondNotifyAfterDays: 7,
        suspendAfterDays: 14,
        autoInvoiceDays: 3,
      });

      setHistory([
        { id: '1', date: '2026-04-05', tenant: 'ООО Техно', amount: 40000, type: 'income' },
        { id: '2', date: '2026-03-28', tenant: 'АО Бизнес', amount: 150000, type: 'income' },
        { id: '3', date: '2026-03-15', tenant: 'ИП Сидоров', amount: 5000, type: 'income' },
      ]);
    } finally {
      setLoading(false);
      setLocalLoading(false);
    }
  };

  const handleCreateInvoice = async (values: any) => {
    try {
      await apiClient.post('/admin/billing/invoices', values);
      message.success('Счёт выставлен');
      setShowInvoiceModal(false);
      invoiceForm.resetFields();
      fetchData();
    } catch {
      message.error('Ошибка');
    }
  };

  const handleCreatePayment = async (values: any) => {
    try {
      await apiClient.post('/admin/billing/payments', values);
      message.success('Платёж зачислен');
      setShowPaymentModal(false);
      paymentForm.resetFields();
      fetchData();
    } catch {
      message.error('Ошибка');
    }
  };

  const handleUpdateSettings = async (key: string, value: number) => {
    try {
      await apiClient.put('/admin/billing/settings', { [key]: value });
      message.success('Настройки обновлены');
      fetchData();
    } catch {
      message.error('Ошибка');
    }
  };

  const getStatusTag = (s: string) => {
    const config = {
      paid: { color: 'green', icon: <CheckCircleOutlined />, text: 'Оплачен' },
      pending: { color: 'gold', icon: <ClockCircleOutlined />, text: 'Ожидается' },
      overdue: { color: 'red', icon: <CloseCircleOutlined />, text: 'Просрочен' },
    };
    const c = config[s as keyof typeof config] || config.pending;
    return <Tag icon={c.icon} color={c.color}>{c.text}</Tag>;
  };

  const getPaymentMethodTag = (m: string) => {
    const methods: Record<string, string> = {
      auto: 'Автоплатёж',
      manual: 'Ручной',
      other: 'Другое',
    };
    return <Tag>{methods[m] || m}</Tag>;
  };

  if (localLoading || isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  const filteredInvoices = invoices.filter((inv) => {
    if (filters.status && inv.status !== filters.status) return false;
    if (filters.search && !inv.invoiceNumber.toLowerCase().includes(filters.search.toLowerCase()) && 
        !inv.tenant.toLowerCase().includes(filters.search.toLowerCase())) return false;
    if (filters.tenant && inv.tenant !== filters.tenant) return false;
    return true;
  });

  const invoiceColumns = [
    { title: 'Тенант', dataIndex: 'tenant', key: 'tenant' },
    { title: 'Счёт', dataIndex: 'invoiceNumber', key: 'invoiceNumber' },
    { title: 'Период', dataIndex: 'period', key: 'period' },
    { title: 'Сумма', dataIndex: 'amount', key: 'amount', render: (v: number) => `${v.toLocaleString()} ₽` },
    { title: 'Выставлен', dataIndex: 'issuedAt', key: 'issuedAt', render: (d: string) => dayjs(d).format('DD.MM.YYYY') },
    { title: 'Оплачен', dataIndex: 'paidAt', key: 'paidAt', render: (d: string) => d ? dayjs(d).format('DD.MM.YYYY') : '-' },
    { title: 'Статус', dataIndex: 'status', key: 'status', render: (s: string) => getStatusTag(s) },
    { title: 'Способ', dataIndex: 'paymentMethod', key: 'paymentMethod', render: (m: string) => getPaymentMethodTag(m) },
  ];

  const historyColumns = [
    { title: 'Дата', dataIndex: 'date', key: 'date', render: (d: string) => dayjs(d).format('DD.MM.YYYY') },
    { title: 'Тенант', dataIndex: 'tenant', key: 'tenant' },
    { title: 'Сумма', dataIndex: 'amount', key: 'amount', render: (v: number, r: any) => (
      <span style={{ color: r.type === 'income' ? '#52c41a' : '#ff4d4f' }}>
        {r.type === 'income' ? '+' : '-'}{v.toLocaleString()} ₽
      </span>
    )},
    { title: 'Тип', dataIndex: 'type', key: 'type', render: (t: string) => <Tag color={t === 'income' ? 'green' : 'red'}>{t === 'income' ? 'Поступление' : 'Возврат'}</Tag> },
  ];

  const totalIncome = history.filter(h => h.type === 'income').reduce((acc, h) => acc + h.amount, 0);
  const totalRefunds = history.filter(h => h.type === 'refund').reduce((acc, h) => acc + h.amount, 0);

  return (
    <div>
      <Title level={3}>Биллинг</Title>

      <Card title="Счета и платежи">
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col>
            <Input
              placeholder="Поиск по счёту..."
              value={filters.search}
              onChange={(e) => setFilters({ ...filters, search: e.target.value })}
              style={{ width: 200 }}
              allowClear
            />
          </Col>
          <Col>
            <Select
              placeholder="Статус"
              value={filters.status || undefined}
              onChange={(v) => setFilters({ ...filters, status: v || '' })}
              style={{ width: 150 }}
              allowClear
            >
              <Option value="paid">Оплачен</Option>
              <Option value="pending">Ожидается</Option>
              <Option value="overdue">Просрочен</Option>
            </Select>
          </Col>
          <Col>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setShowInvoiceModal(true)}>
              Выставить ручной счёт
            </Button>
          </Col>
          <Col>
            <Button icon={<ExportOutlined />} onClick={() => setShowPaymentModal(true)}>
              Зачислить оплату
            </Button>
          </Col>
        </Row>
        <Table
          dataSource={filteredInvoices}
          columns={invoiceColumns}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Card title="Настройки биллинга" style={{ marginTop: 16 }}>
        <Row gutter={16}>
          <Col span={8}>
            <Text>Уведомление о просрочке через (дней):</Text>
            <InputForm
              type="number"
              defaultValue={settings?.notifyAfterDays}
              onPressEnter={(e: any) => handleUpdateSettings('notifyAfterDays', parseInt(e.target.value))}
              style={{ marginTop: 4 }}
            />
          </Col>
          <Col span={8}>
            <Text>Повторное уведомление через (дней):</Text>
            <InputForm
              type="number"
              defaultValue={settings?.secondNotifyAfterDays}
              onPressEnter={(e: any) => handleUpdateSettings('secondNotifyAfterDays', parseInt(e.target.value))}
              style={{ marginTop: 4 }}
            />
          </Col>
          <Col span={8}>
            <Text>Автоприостановка через (дней):</Text>
            <InputForm
              type="number"
              defaultValue={settings?.suspendAfterDays}
              onPressEnter={(e: any) => handleUpdateSettings('suspendAfterDays', parseInt(e.target.value))}
              style={{ marginTop: 4 }}
            />
          </Col>
        </Row>
        <Row gutter={16} style={{ marginTop: 16 }}>
          <Col span={8}>
            <Text>Автовыставление за N дней до платежа:</Text>
            <InputForm
              type="number"
              defaultValue={settings?.autoInvoiceDays}
              onPressEnter={(e: any) => handleUpdateSettings('autoInvoiceDays', parseInt(e.target.value))}
              style={{ marginTop: 4 }}
            />
          </Col>
        </Row>
      </Card>

      <Card title="История платежей" style={{ marginTop: 16 }}>
        <Row gutter={16}>
          <Col>
            <Statistic
              title="Всего поступлений"
              value={totalIncome}
              prefix="₽"
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col>
            <Statistic
              title="Возвраты"
              value={totalRefunds}
              prefix="₽"
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Col>
          <Col>
            <RangePicker />
          </Col>
        </Row>
        <Table
          dataSource={history}
          columns={historyColumns}
          rowKey="id"
          pagination={{ pageSize: 10 }}
          style={{ marginTop: 16 }}
        />
      </Card>

      <Modal
        title="Выставить ручной счёт"
        open={showInvoiceModal}
        onCancel={() => setShowInvoiceModal(false)}
        onOk={() => invoiceForm.submit()}
      >
        <Form form={invoiceForm} layout="vertical" onFinish={handleCreateInvoice}>
          <Form.Item name="tenant" label="Тенант" rules={[{ required: true }]}>
            <InputForm />
          </Form.Item>
          <Form.Item name="amount" label="Сумма" rules={[{ required: true }]}>
            <InputForm type="number" />
          </Form.Item>
          <Form.Item name="description" label="Назначение">
            <InputForm />
          </Form.Item>
          <Form.Item name="period" label="Период">
            <InputForm placeholder="Апрель 2026" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Зачислить оплату"
        open={showPaymentModal}
        onCancel={() => setShowPaymentModal(false)}
        onOk={() => paymentForm.submit()}
      >
        <Form form={paymentForm} layout="vertical" onFinish={handleCreatePayment}>
          <Form.Item name="tenant" label="Тенант" rules={[{ required: true }]}>
            <InputForm />
          </Form.Item>
          <Form.Item name="invoiceId" label="Счёт">
            <Select>
              {invoices.filter(i => i.status !== 'paid').map(inv => (
                <Option key={inv.id} value={inv.id}>
                  {inv.invoiceNumber} - {inv.amount.toLocaleString()} ₽
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="amount" label="Сумма" rules={[{ required: true }]}>
            <InputForm type="number" />
          </Form.Item>
          <Form.Item name="receivedAt" label="Дата поступления">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="comment" label="Комментарий">
            <InputForm />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default BillingSection;
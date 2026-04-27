import React, { useEffect } from 'react';
import {
  Modal,
  Form,
  InputNumber,
  Select,
  DatePicker,
  Radio,
} from 'antd';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import type { Goal, Employee } from '@/types';

dayjs.locale('ru');

interface GoalFormModalProps {
  visible: boolean;
  onCancel: () => void;
  onOk: (values: any) => void;
  initialValues?: Partial<Goal>;
  employees?: Employee[];
  assignableRoles: string[];
}

const GoalFormModal: React.FC<GoalFormModalProps> = ({
  visible,
  onCancel,
  onOk,
  initialValues,
  employees = [],
  assignableRoles,
}) => {
  const [form] = Form.useForm();

  const now = dayjs();

  useEffect(() => {
    if (visible) {
      const currentUserId = typeof window !== 'undefined' ? localStorage.getItem('id') : '';
      form.setFieldsValue({
        ...initialValues,
        assignee_id: initialValues?.assignee_id ?? currentUserId,
        role: initialValues?.role ?? '',
        period: initialValues?.period || 'day',
        target_date: initialValues?.target_date ? dayjs(initialValues.target_date) : now,
        start_date: initialValues?.start_date ? dayjs(initialValues.start_date) : now,
        end_date: initialValues?.end_date ? dayjs(initialValues.end_date) : now,
      });
    } else {
      form.resetFields();
    }
  }, [visible, initialValues, form, now]);

  const hasRoles = assignableRoles.length > 0;
  const isEdit = !!initialValues?.id;

  const periodOptions = [
    { label: 'День', value: 'day' },
    { label: 'Неделя', value: 'week' },
    { label: 'Месяц', value: 'month' },
    { label: 'Год', value: 'year' },
    { label: 'Свой период', value: 'custom' },
  ];

  return (
    <Modal
      title={isEdit ? 'Редактировать план' : 'Назначить план'}
      open={visible}
      onCancel={onCancel}
      onOk={async () => {
        try {
          const values = await form.validateFields();
          onOk(values);
        } catch {
          // Ant Design already shows validation errors
        }
      }}
      okButtonProps={{ disabled: !hasRoles }}
      width={600}
    >
      <Form form={form} layout="vertical" initialValues={{ period: 'day' }}>
        <Form.Item
          name="assignee_id"
          label="Получатель (сотрудник)"
          rules={[{ required: true, message: 'Выберите получателя' }]}
        >
          <Select placeholder="Выберите сотрудника">
            {employees.map((e) => (
              <Select.Option key={e.id} value={e.id}>
                {e.first_name} {e.last_name} ({e.role})
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name="role"
          label="Роль получателя"
          rules={[{ required: true, message: 'Выберите роль' }]}
        >
          <Select placeholder="Выберите роль" disabled={!hasRoles}>
            {assignableRoles.map((r) => (
              <Select.Option key={r} value={r}>
                {r}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item name="period" label="Период" initialValue="day">
          <Radio.Group options={periodOptions} optionType="button" />
        </Form.Item>

        <Form.Item
          noStyle
          shouldUpdate={(prev, curr) => prev.period !== curr.period}
        >
          {({ getFieldValue }) => {
            const period = getFieldValue('period');
            if (period === 'day') {
              return (
                <Form.Item name="target_date" label="Дата">
                  <DatePicker 
                    format="DD.MM.YYYY" 
                    style={{ width: '100%' }}
                    placeholder="ДД.ММ.ГГГГ"
                  />
                </Form.Item>
              );
            } else if (period === 'custom') {
              return (
                <Form.Item label="Период">
                  <Form.Item name="start_date" noStyle>
                    <DatePicker
                      format="DD.MM.YYYY"
                      placeholder="ДД.ММ.ГГГГ"
                      style={{ width: '45%' }}
                    />
                  </Form.Item>
                  <span style={{ margin: '0 8px' }}>—</span>
                  <Form.Item name="end_date" noStyle>
                    <DatePicker
                      format="DD.MM.YYYY"
                      placeholder="ДД.ММ.ГГГГ"
                      style={{ width: '45%' }}
                    />
                  </Form.Item>
                </Form.Item>
              );
            } else {
              return (
                <Form.Item name="start_date" label="Дата начала">
                  <DatePicker 
                    format="DD.MM.YYYY" 
                    style={{ width: '100%' }}
                    placeholder="ДД.ММ.ГГГГ"
                  />
                </Form.Item>
              );
            }
          }}
        </Form.Item>

        <Form.Item name="sales_plan" label="Продажи (₽)" rules={[{ required: true, type: 'number' }]}>
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="leads_plan" label="Лиды">
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="calls_plan" label="Звонки">
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="meetings_plan" label="Встречи">
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default GoalFormModal;

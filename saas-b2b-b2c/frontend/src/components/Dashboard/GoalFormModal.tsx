import React, { useEffect } from 'react';
import {
  Modal,
  Form,
  InputNumber,
  Select,
  DatePicker,
} from 'antd';
import dayjs from 'dayjs';
import type { Goal, Employee } from '@/types';

/* ---------- Пропсы ---------- */
interface GoalFormModalProps {
  visible: boolean;
  onCancel: () => void;
  /** Save callback */
  onOk: (values: any) => void;
  /** Если редактируем – передаём объект Goal */
  initialValues?: Partial<Goal>;
  /** Список всех сотрудников (для выбора получателя) */
  employees?: Employee[];
  /** Роли, которые может назначать текущий пользователь */
  assignableRoles: string[];
}

/* ---------- Компонент ---------- */
const GoalFormModal: React.FC<GoalFormModalProps> = ({
  visible,
  onCancel,
  onOk,
  initialValues,
  employees = [],
  assignableRoles,
}) => {
  const [form] = Form.useForm();

  /* ----- При открытии заполняем форму ----- */
  useEffect(() => {
    if (visible) {
      const currentUserId = typeof window !== 'undefined' ? localStorage.getItem('id') : '';
      form.setFieldsValue({
        ...initialValues,
        // Если цель создаётся «для себя», ставим свой ID в качестве получателя
        assignee_id: initialValues?.assignee_id ?? currentUserId,
        role: initialValues?.role ?? '',
        target_date: initialValues?.target_date
          ? dayjs(initialValues.target_date)
          : dayjs(),
      });
    } else {
      form.resetFields();
    }
  }, [visible, initialValues, form]);

  const hasRoles = assignableRoles.length > 0;

  return (
    <Modal
      title={initialValues?.id ? 'Редактировать план' : 'Назначить план'}
      open={visible}
      onCancel={onCancel}
      onOk={async () => {
        try {
          const values = await form.validateFields();
          onOk(values);
        } catch {
          // Ant‑Design уже покажет ошибки валидации
        }
      }}
      okButtonProps={{ disabled: !hasRoles }}
    >
      <Form form={form} layout="vertical">
        {/* ---------- Выбор получателя (список сотрудников) ---------- */}
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

        {/* ---------- Выбор роли получателя ---------- */}
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

        {/* ---------- Плановые показатели ---------- */}
        <Form.Item name="sales_plan" label="Продажи (₽)" rules={[{ required: true, type: 'number' }]}>
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="leads_plan" label="Лиды" rules={[{ type: 'number' }]}>
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="calls_plan" label="Звонки" rules={[{ type: 'number' }]}>
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="meetings_plan" label="Встречи" rules={[{ type: 'number' }]}>
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>

        {/* ---------- Дата плана ---------- */}
        <Form.Item name="target_date" label="Дата плана" rules={[{ required: true }]}>
          <DatePicker format="YYYY-MM-DD" style={{ width: '100%' }} />
        </Form.Item>

        {/* Прогноз удалён – поле не выводим */}
      </Form>
    </Modal>
  );
};

export default GoalFormModal;

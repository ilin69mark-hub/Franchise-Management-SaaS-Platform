import React from 'react';
import { Card, Form, InputNumber, Switch, Radio, Button, Space, Typography, Divider, message } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { useFranchiserAlertStore, FranchiserAlertSettings } from '@/store/franchiserAlertStore';

const { Title, Text } = Typography;

const AlertSettingsPanel: React.FC = () => {
  const { settings, setSettings } = useFranchiserAlertStore();
  const [form] = Form.useForm();

  React.useEffect(() => {
    form.setFieldsValue(settings);
  }, [settings, form]);

  const handleSave = () => {
    form.validateFields().then(values => {
      setSettings(values);
      message.success('Настройки сохранены');
    });
  };

  return (
    <Card>
      <Title level={4}>Настройки уведомлений</Title>
      
      <Form form={form} layout="vertical" initialValues={settings}>
        <Divider>Каналы уведомлений</Divider>
        
        <Form.Item name={['channel', 'inApp']} valuePropName="checked">
          <Switch /> <Text style={{ marginLeft: 8 }}>In-app уведомления</Text>
        </Form.Item>
        
        <Form.Item name={['channel', 'email']} valuePropName="checked">
          <Switch /> <Text style={{ marginLeft: 8 }}>Email дайджест (раз в неделю)</Text>
        </Form.Item>
        
        <Form.Item name={['channel', 'push']} valuePropName="checked">
          <Switch /> <Text style={{ marginLeft: 8 }}>Push-уведомления</Text>
        </Form.Item>

        <Divider>Пороговые значения</Divider>
        
        <Form.Item 
          name={['thresholds', 'criticalForecastPercent']} 
          label="Критический прогноз сети (%)"
          tooltip="Ниже этого значения — критический алерт"
        >
          <InputNumber min={0} max={100} style={{ width: 120 }} />
        </Form.Item>
        
        <Form.Item 
          name={['thresholds', 'dealerChurnPercent']} 
          label="Порог оттока дилеров (%)"
          tooltip="Выше этого значения за квартал — алерт"
        >
          <InputNumber min={0} max={100} style={{ width: 120 }} />
        </Form.Item>
        
        <Form.Item 
          name={['thresholds', 'managerKpiPercent']} 
          label="Порог KPI менеджера (%)"
          tooltip="Ниже этого значения — предупреждение"
        >
          <InputNumber min={0} max={100} style={{ width: 120 }} />
        </Form.Item>

        <Divider />
        
        <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>
          Сохранить настройки
        </Button>
      </Form>
    </Card>
  );
};

export default AlertSettingsPanel;
// src/components/Dashboard/Alerts/AlertSettingsPanel.tsx
import React, { useEffect } from 'react';
import { Card, Row, Col, Typography, Switch, InputNumber, Checkbox, Button, Divider, Space, message, Form } from 'antd';
import { SettingOutlined, SaveOutlined, BellOutlined, MailOutlined, GlobalOutlined } from '@ant-design/icons';
import { useAlertStore, AlertSettings } from '@/store/alertStore';

const { Text, Title } = Typography;

interface AlertSettingsPanelProps {
  onClose?: () => void;
}

const AlertSettingsPanel: React.FC<AlertSettingsPanelProps> = ({ onClose }) => {
  const { settings, setSettings } = useAlertStore();
  const [form] = Form.useForm();

  useEffect(() => {
    form.setFieldsValue(settings);
  }, [settings, form]);

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSettings(values);
      message.success('Настройки сохранены');
      onClose?.();
    } catch (e) {
      message.error('Ошибка сохранения');
    }
  };

  return (
    <Card title={<Space><SettingOutlined />Настройки алертов</Space>}>
      <Form form={form} layout="vertical">
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Card size="small" title="Общие">
              <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                <Text>Включить алерты</Text>
                <Switch 
                  checked={settings.enabled} 
                  onChange={(checked) => setSettings({ enabled: checked })} 
                />
              </Space>
            </Card>
          </Col>

          <Col span={24}>
            <Card size="small" title="Пороговые значения">
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <div>
                  <Text>Порог красной зоны плана (%):</Text>
                  <InputNumber 
                    name="planThreshold" 
                    min={0} 
                    max={100} 
                    style={{ width: '100%', marginTop: 4 }}
                  />
                </div>
                <div>
                  <Text>Порог падения конверсии (%):</Text>
                  <InputNumber 
                    name="conversionDropThreshold" 
                    min={0} 
                    max={100} 
                    style={{ width: '100%', marginTop: 4 }}
                  />
                </div>
                <div>
                  <Text>Порог падения трафика (%):</Text>
                  <InputNumber 
                    name="trafficDropThreshold" 
                    min={0} 
                    max={100} 
                    style={{ width: '100%', marginTop: 4 }}
                  />
                </div>
                <div>
                  <Text>Дней зависшей сделки:</Text>
                  <InputNumber 
                    name="stuckDealDays" 
                    min={0} 
                    max={365} 
                    style={{ width: '100%', marginTop: 4 }}
                  />
                </div>
                <div>
                  <Text>Сумма крупной сделки (руб.):</Text>
                  <InputNumber 
                    name="stuckDealAmount" 
                    min={0} 
                    style={{ width: '100%', marginTop: 4 }}
                  />
                </div>
                <div>
                  <Text>Дней неактивности дилера:</Text>
                  <InputNumber 
                    name="inactiveDealerDays" 
                    min={0} 
                    max={30} 
                    style={{ width: '100%', marginTop: 4 }}
                  />
                </div>
                <div>
                  <Text>Дней неликвида:</Text>
                  <InputNumber 
                    name="nonLiquidDays" 
                    min={0} 
                    max={365} 
                    style={{ width: '100%', marginTop: 4 }}
                  />
                </div>
              </Space>
            </Card>
          </Col>

          <Col span={24}>
            <Card size="small" title="Каналы доставки">
              <Space direction="vertical" style={{ width: '100%' }}>
                <Checkbox 
                  checked={settings.channels.inApp}
                  onChange={(e) => setSettings({ 
                    channels: { ...settings.channels, inApp: e.target.checked } 
                  })}
                >
                  <Space><BellOutlined />В системе (in-app)</Space>
                </Checkbox>
                <Checkbox 
                  checked={settings.channels.emailInstant}
                  onChange={(e) => setSettings({ 
                    channels: { ...settings.channels, emailInstant: e.target.checked } 
                  })}
                >
                  <Space><MailOutlined />Email мгновенно</Space>
                </Checkbox>
                <Checkbox 
                  checked={settings.channels.emailDigest}
                  onChange={(e) => setSettings({ 
                    channels: { ...settings.channels, emailDigest: e.target.checked } 
                  })}
                >
                  <Space><MailOutlined />Email дайджест (раз в день)</Space>
                </Checkbox>
                <Checkbox 
                  checked={settings.channels.push}
                  onChange={(e) => setSettings({ 
                    channels: { ...settings.channels, push: e.target.checked } 
                  })}
                >
                  <Space><GlobalOutlined />Push-уведомления</Space>
                </Checkbox>
              </Space>
            </Card>
          </Col>
        </Row>

        <Divider />

        <Row justify="end">
          <Space>
            <Button onClick={onClose}>Отмена</Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>
              Сохранить
            </Button>
          </Space>
        </Row>
      </Form>
    </Card>
  );
};

export default AlertSettingsPanel;
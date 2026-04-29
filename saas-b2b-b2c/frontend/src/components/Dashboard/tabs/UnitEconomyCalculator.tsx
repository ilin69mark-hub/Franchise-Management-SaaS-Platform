// src/components/Dashboard/tabs/UnitEconomyCalculator.tsx
import React, { useState, useMemo, useEffect } from 'react';
import { Card, Row, Col, InputNumber, Slider, Typography, Button, Space, Input, Select, message, Divider, Tag } from 'antd';
import { SaveOutlined, FileTextOutlined, DeleteOutlined, CalculatorOutlined } from '@ant-design/icons';

const { Text, Title } = Typography;
const { Option } = Select;

export interface UnitEconomyInput {
  salePrice: number;
  purchasePrice: number;
  managerPercent: number;
  deliveryCost: number;
  acquiringPercent: number;
  rentAmortization: number;
  additionalServices: number;
}

export interface UnitEconomyResult {
  marginalProfit: number;
  netProfit: number;
  profitability: number;
  priceAfterDiscount: number;
}

export interface UnitTemplate {
  id?: string;
  name: string;
  category: 'sofa' | 'kitchen' | 'wardrobe' | 'other';
  input: UnitEconomyInput;
}

interface UnitEconomyCalculatorProps {
  templates?: UnitTemplate[];
  onSaveTemplate?: (template: UnitTemplate) => Promise<void>;
}

const defaultInput: UnitEconomyInput = {
  salePrice: 0,
  purchasePrice: 0,
  managerPercent: 5,
  deliveryCost: 0,
  acquiringPercent: 2,
  rentAmortization: 0,
  additionalServices: 0,
};

const UnitEconomyCalculator: React.FC<UnitEconomyCalculatorProps> = ({ templates = [], onSaveTemplate }) => {
  const [input, setInput] = useState<UnitEconomyInput>(defaultInput);
  const [discount, setDiscount] = useState(0);
  const [templateName, setTemplateName] = useState('');
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>('');
  const [saving, setSaving] = useState(false);

  const result: UnitEconomyResult = useMemo(() => {
    const priceAfterDiscount = input.salePrice * (1 - discount / 100);
    const managerBonus = priceAfterDiscount * (input.managerPercent / 100);
    const acquiringFee = priceAfterDiscount * (input.acquiringPercent / 100);
    
    const totalCosts = input.purchasePrice + managerBonus + input.deliveryCost + 
                     acquiringFee + input.rentAmortization + input.additionalServices;
    
    const marginalProfit = priceAfterDiscount - input.purchasePrice;
    const netProfit = marginalProfit - input.deliveryCost - managerBonus - acquiringFee - input.rentAmortization - input.additionalServices;
    const profitability = priceAfterDiscount > 0 ? (netProfit / priceAfterDiscount) * 100 : 0;

    return {
      marginalProfit,
      netProfit,
      profitability,
      priceAfterDiscount,
    };
  }, [input, discount]);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('ru-RU').format(value) + ' ₽';
  };

  const handleInputChange = (field: keyof UnitEconomyInput, value: number) => {
    setInput(prev => ({ ...prev, [field]: value }));
  };

  const handleLoadTemplate = (templateId: string) => {
    const template = templates.find(t => t.id === templateId);
    if (template) {
      setInput(template.input);
      setSelectedTemplateId(templateId);
    }
  };

  const handleSaveTemplate = async () => {
    if (!templateName.trim()) {
      message.error('Введите название шаблона');
      return;
    }
    if (!onSaveTemplate) {
      message.warning('Сохранение шаблонов недоступно');
      return;
    }
    setSaving(true);
    try {
      await onSaveTemplate({
        name: templateName,
        category: 'other',
        input,
      });
      message.success('Шаблон сохранён');
      setTemplateName('');
    } catch (error) {
      message.error('Ошибка сохранения шаблона');
    } finally {
      setSaving(false);
    }
  };

  const getProfitColor = () => {
    if (result.netProfit < 0) return '#ff4d4f';
    if (result.netProfit < input.salePrice * 0.05) return '#fa8c16';
    return '#52c41a';
  };

  const getProfitabilityColor = () => {
    if (result.profitability < 0) return '#ff4d4f';
    if (result.profitability < 5) return '#fa8c16';
    return '#52c41a';
  };

  return (
    <Card title={<><CalculatorOutlined /> Калькулятор unit-экономики</>}>
      <Row gutter={[24, 16]}>
        <Col xs={24} lg={12}>
          <Card type="inner" title="📥 Входящие данные" size="small">
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <div>
                <Text type="secondary">Цена продажи (руб.)</Text>
                <InputNumber
                  style={{ width: '100%', marginTop: 4 }}
                  value={input.salePrice}
                  onChange={v => handleInputChange('salePrice', v || 0)}
                  min={0}
                  placeholder="0"
                  formatter={value => `${value} ₽`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={value => value?.replace(/ ₽/g, '').replace(/ /g, '') as any}
                />
              </div>

              <div>
                <Text type="secondary">Закупочная цена у фабрики (руб.)</Text>
                <InputNumber
                  style={{ width: '100%', marginTop: 4 }}
                  value={input.purchasePrice}
                  onChange={v => handleInputChange('purchasePrice', v || 0)}
                  min={0}
                  placeholder="0"
                  formatter={value => `${value} ₽`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={value => value?.replace(/ ₽/g, '').replace(/ /g, '') as any}
                />
              </div>

              <Row gutter={16}>
                <Col span={12}>
                  <Text type="secondary">Процент менеджеру (%)</Text>
                  <InputNumber
                    style={{ width: '100%', marginTop: 4 }}
                    value={input.managerPercent}
                    onChange={v => handleInputChange('managerPercent', v || 0)}
                    min={0}
                    max={100}
                    placeholder="5"
                  />
                </Col>
                <Col span={12}>
                  <Text type="secondary">Процент за эквайринг (%)</Text>
                  <InputNumber
                    style={{ width: '100%', marginTop: 4 }}
                    value={input.acquiringPercent}
                    onChange={v => handleInputChange('acquiringPercent', v || 0)}
                    min={0}
                    max={100}
                    placeholder="2"
                  />
                </Col>
              </Row>

              <div>
                <Text type="secondary">Стоимость доставки (руб.)</Text>
                <InputNumber
                  style={{ width: '100%', marginTop: 4 }}
                  value={input.deliveryCost}
                  onChange={v => handleInputChange('deliveryCost', v || 0)}
                  min={0}
                  placeholder="0"
                  formatter={value => `${value} ₽`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={value => value?.replace(/ ₽/g, '').replace(/ /g, '') as any}
                />
              </div>

              <div>
                <Text type="secondary">Амортизация аренды на единицу (руб.)</Text>
                <InputNumber
                  style={{ width: '100%', marginTop: 4 }}
                  value={input.rentAmortization}
                  onChange={v => handleInputChange('rentAmortization', v || 0)}
                  min={0}
                  placeholder="0"
                  formatter={value => `${value} ₽`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={value => value?.replace(/ ₽/g, '').replace(/ /g, '') as any}
                />
              </div>

              <div>
                <Text type="secondary">Дополнительные услуги (чехлы, пропитка, сборка) (руб.)</Text>
                <InputNumber
                  style={{ width: '100%', marginTop: 4 }}
                  value={input.additionalServices}
                  onChange={v => handleInputChange('additionalServices', v || 0)}
                  min={0}
                  placeholder="0"
                  formatter={value => `${value} ₽`.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')}
                  parser={value => value?.replace(/ ₽/g, '').replace(/ /g, '') as any}
                />
              </div>
            </Space>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card type="inner" title="📊 Результат" size="small">
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <div style={{ 
                padding: 16, 
                background: result.marginalProfit >= 0 ? '#f6ffed' : '#fff1f0',
                borderRadius: 8,
                border: `1px solid ${result.marginalProfit >= 0 ? '#b7eb8f' : '#ffccc7'}`
              }}>
                <Text type="secondary">Маржинальная прибыль с единицы</Text>
                <Title level={3} style={{ margin: 0, color: result.marginalProfit >= 0 ? '#52c41a' : '#ff4d4f' }}>
                  {formatCurrency(result.marginalProfit)}
                </Title>
              </div>

              <div style={{ 
                padding: 16, 
                background: result.netProfit >= 0 ? '#e6f7ff' : '#fff1f0',
                borderRadius: 8,
                border: `1px solid ${result.netProfit >= 0 ? '#91d5ff' : '#ffccc7'}`
              }}>
                <Text type="secondary">Чистая прибыль с единицы</Text>
                <Title level={2} style={{ margin: 0, color: getProfitColor() }}>
                  {formatCurrency(result.netProfit)}
                </Title>
                {discount > 0 && (
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    (без скидки: {formatCurrency(result.netProfit + (input.salePrice * discount / 100 * (1 - input.managerPercent / 100 - input.acquiringPercent / 100) - input.deliveryCost - input.rentAmortization - input.additionalServices))})
                  </Text>
                )}
              </div>

              <div>
                <Text type="secondary">Рентабельность продажи</Text>
                <Title level={3} style={{ margin: 0, color: getProfitabilityColor() }}>
                  {result.profitability.toFixed(1)}%
                </Title>
              </div>

              <Divider style={{ margin: '8px 0' }} />

              <div>
                <Text type="secondary">Цена после скидки: {formatCurrency(result.priceAfterDiscount)}</Text>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card type="inner" size="small">
            <Text type="secondary" style={{ marginBottom: 8, display: 'block' }}>
              Скидка клиенту: {discount}%
            </Text>
            <Slider
              marks={{
                0: '0%',
                10: '10%',
                20: '20%',
                30: '30%',
              }}
              min={0}
              max={30}
              value={discount}
              onChange={setDiscount}
              tooltip={{ formatter: value => `${value}%` }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} md={12}>
          <Card type="inner" title="📂 Шаблоны" size="small">
            {templates.length > 0 ? (
              <Select
                style={{ width: '100%' }}
                placeholder="Выберите шаблон"
                value={selectedTemplateId || undefined}
                onChange={handleLoadTemplate}
                allowClear
              >
                {templates.map(template => (
                  <Option key={template.id} value={template.id!}>
                    <Tag>{template.category}</Tag> {template.name}
                  </Option>
                ))}
              </Select>
            ) : (
              <Text type="secondary">Нет сохранённых шаблонов</Text>
            )}
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card type="inner" title="💾 Сохранить как шаблон" size="small">
            <Space>
              <Input
                placeholder="Название (диван, кухня, шкаф)"
                value={templateName}
                onChange={e => setTemplateName(e.target.value)}
                style={{ width: 200 }}
              />
              <Button 
                type="primary" 
                icon={<SaveOutlined />}
                onClick={handleSaveTemplate}
                loading={saving}
              >
                Сохранить
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </Card>
  );
};

export default UnitEconomyCalculator;
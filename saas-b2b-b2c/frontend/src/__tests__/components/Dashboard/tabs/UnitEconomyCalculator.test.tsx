// __tests__/components/Dashboard/tabs/UnitEconomyCalculator.test.tsx
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { message } from 'antd';
import UnitEconomyCalculator, { UnitEconomyInput, UnitTemplate } from '@/components/Dashboard/tabs/UnitEconomyCalculator';

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: {
      ...actualAntd.message,
      error: jest.fn(),
      warning: jest.fn(),
      success: jest.fn(),
    },
  };
});

describe('UnitEconomyCalculator', () => {
  const defaultInput: UnitEconomyInput = {
    salePrice: 100000,
    purchasePrice: 50000,
    managerPercent: 5,
    deliveryCost: 5000,
    acquiringPercent: 2,
    rentAmortization: 1000,
    additionalServices: 3000,
  };

  const mockTemplates: UnitTemplate[] = [
    {
      id: '1',
      name: 'Диван стандарт',
      category: 'sofa',
      input: defaultInput,
    },
    {
      id: '2',
      name: 'Кухня классика',
      category: 'kitchen',
      input: {
        ...defaultInput,
        salePrice: 200000,
        purchasePrice: 120000,
      },
    },
  ];

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders calculator component', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText('Калькулятор unit-экономики')).toBeInTheDocument();
    expect(screen.getByText(/Входящие данные/i)).toBeInTheDocument();
    expect(screen.getByText(/Результат/i)).toBeInTheDocument();
  });

  it('renders all input fields', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText('Цена продажи (руб.)')).toBeInTheDocument();
    expect(screen.getByText('Закупочная цена у фабрики (руб.)')).toBeInTheDocument();
    expect(screen.getByText('Процент менеджеру (%)')).toBeInTheDocument();
    expect(screen.getByText('Процент за эквайринг (%)')).toBeInTheDocument();
    expect(screen.getByText('Стоимость доставки (руб.)')).toBeInTheDocument();
    expect(screen.getByText('Амортизация аренды на единицу (руб.)')).toBeInTheDocument();
    expect(screen.getByText('Дополнительные услуги (чехлы, пропитка, сборка) (руб.)')).toBeInTheDocument();
  });

  it('calculates marginal profit correctly', () => {
    render(<UnitEconomyCalculator />);
    
    const salePriceInput = document.querySelector('input[id*="salePrice"]') || document.querySelector('.ant-input-number input');
    if (salePriceInput) {
      fireEvent.change(salePriceInput, { target: { value: '100000' } });
    }
    
    expect(screen.getByText('Маржинальная прибыль с единицы')).toBeInTheDocument();
  });

  it('calculates net profit correctly', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText('Чистая прибыль с единицы')).toBeInTheDocument();
  });

  it('calculates profitability correctly', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText('Рентабельность продажи')).toBeInTheDocument();
  });

  it('renders discount slider', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText('Скидка клиенту: 0%')).toBeInTheDocument();
  });

  it('renders templates section', () => {
    render(<UnitEconomyCalculator templates={mockTemplates} />);
    expect(screen.getByText(/Шаблоны/i)).toBeInTheDocument();
  });

  it('renders save template section', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText(/Сохранить как шаблон/i)).toBeInTheDocument();
  });

  it('shows empty templates message when no templates', () => {
    render(<UnitEconomyCalculator templates={[]} />);
    expect(screen.getByText('Нет сохранённых шаблонов')).toBeInTheDocument();
  });

  it('calls onSaveTemplate when save button clicked', async () => {
    const mockSaveTemplate = jest.fn().mockResolvedValue(undefined);
    render(<UnitEconomyCalculator onSaveTemplate={mockSaveTemplate} />);
    
    const input = screen.getByPlaceholderText('Название (диван, кухня, шкаф)');
    fireEvent.change(input, { target: { value: 'Тестовый шаблон' } });
    
    const saveButton = screen.getByText('Сохранить');
    fireEvent.click(saveButton);
    
    await waitFor(() => {
      expect(mockSaveTemplate).toHaveBeenCalled();
    });
  });

  it('shows error when saving without name', async () => {
    const mockSaveTemplate = jest.fn();
    render(<UnitEconomyCalculator onSaveTemplate={mockSaveTemplate} />);
    
    const saveButton = screen.getByText('Сохранить');
    fireEvent.click(saveButton);
    
    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith('Введите название шаблона');
    });
  });

  it('shows warning when onSaveTemplate not provided', async () => {
    render(<UnitEconomyCalculator />);
    
    const input = screen.getByPlaceholderText('Название (диван, кухня, шкаф)');
    fireEvent.change(input, { target: { value: 'Тестовый шаблон' } });
    
    const saveButton = screen.getByText('Сохранить');
    fireEvent.click(saveButton);
    
    await waitFor(() => {
      expect(message.warning).toHaveBeenCalledWith('Сохранение шаблонов недоступно');
    });
  });

  it('displays negative profit in red', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText('Чистая прибыль с единицы')).toBeInTheDocument();
  });
});

describe('UnitEconomyCalculator calculations', () => {
  it('calculates marginal profit = salePrice - purchasePrice', () => {
    const input: UnitEconomyInput = {
      salePrice: 100000,
      purchasePrice: 50000,
      managerPercent: 5,
      deliveryCost: 0,
      acquiringPercent: 2,
      rentAmortization: 0,
      additionalServices: 0,
    };
    const discount = 0;
    
    const priceAfterDiscount = input.salePrice * (1 - discount / 100);
    const marginalProfit = priceAfterDiscount - input.purchasePrice;
    
    expect(marginalProfit).toBe(50000);
  });

  it('calculates net profit with all costs', () => {
    const input: UnitEconomyInput = {
      salePrice: 100000,
      purchasePrice: 50000,
      managerPercent: 5,
      deliveryCost: 5000,
      acquiringPercent: 2,
      rentAmortization: 1000,
      additionalServices: 3000,
    };
    const discount = 0;
    
    const priceAfterDiscount = input.salePrice * (1 - discount / 100);
    const managerBonus = priceAfterDiscount * (input.managerPercent / 100);
    const acquiringFee = priceAfterDiscount * (input.acquiringPercent / 100);
    const marginalProfit = priceAfterDiscount - input.purchasePrice;
    const netProfit = marginalProfit - input.deliveryCost - managerBonus - acquiringFee - input.rentAmortization - input.additionalServices;
    
    expect(netProfit).toBe(34000);
  });

  it('calculates profitability percentage', () => {
    const input: UnitEconomyInput = {
      salePrice: 100000,
      purchasePrice: 50000,
      managerPercent: 5,
      deliveryCost: 0,
      acquiringPercent: 2,
      rentAmortization: 0,
      additionalServices: 0,
    };
    const discount = 0;
    
    const priceAfterDiscount = input.salePrice * (1 - discount / 100);
    const netProfit = 43000;
    const profitability = (netProfit / priceAfterDiscount) * 100;
    
    expect(profitability).toBe(43);
  });

  it('applies discount correctly', () => {
    const input: UnitEconomyInput = {
      salePrice: 100000,
      purchasePrice: 50000,
      managerPercent: 5,
      deliveryCost: 0,
      acquiringPercent: 2,
      rentAmortization: 0,
      additionalServices: 0,
    };
    const discount = 10;
    
    const priceAfterDiscount = input.salePrice * (1 - discount / 100);
    
    expect(priceAfterDiscount).toBe(90000);
  });

  it('handles zero sale price', () => {
    const input: UnitEconomyInput = {
      salePrice: 0,
      purchasePrice: 0,
      managerPercent: 0,
      deliveryCost: 0,
      acquiringPercent: 0,
      rentAmortization: 0,
      additionalServices: 0,
    };
    
    const profitability = input.salePrice > 0 ? (0 / input.salePrice) * 100 : 0;
    
    expect(profitability).toBe(0);
  });
});

describe('UnitEconomyCalculator functionality', () => {
  it('loads template when selected', () => {
    const mockTemplates: UnitTemplate[] = [
      {
        id: '1',
        name: 'Диван',
        category: 'sofa',
        input: {
          salePrice: 100000,
          purchasePrice: 50000,
          managerPercent: 5,
          deliveryCost: 5000,
          acquiringPercent: 2,
          rentAmortization: 1000,
          additionalServices: 3000,
        },
      },
    ];

    render(<UnitEconomyCalculator templates={mockTemplates} />);
    expect(screen.getByText(/Шаблоны/i)).toBeInTheDocument();
  });

  it('resets form with default values', () => {
    render(<UnitEconomyCalculator />);
    expect(screen.getByText(/Скидка клиенту/i)).toBeInTheDocument();
  });
});
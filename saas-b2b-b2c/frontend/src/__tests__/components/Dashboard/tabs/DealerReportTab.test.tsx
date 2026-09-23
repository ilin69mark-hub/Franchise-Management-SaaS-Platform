import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import DealerReportTab from '@/components/Dashboard/tabs/DealerReportTab';

jest.mock('antd', () => {
  const actualAntd = jest.requireActual('antd');
  return {
    ...actualAntd,
    message: { ...actualAntd.message, success: jest.fn(), error: jest.fn() },
  };
});

const reportData = {
  dealerName: 'ООО Тест',
  period: 'month' as const,
  startDate: '01.09.2026',
  endDate: '30.09.2026',
  generatedAt: '30.09.2026',
  planFact: { plan: 1000000, fact: 850000, percent: 85 },
  funnel: { traffic: 1000, consultation: 500, measurement: 300, kp: 200, contract: 100, payment: 80 },
  inventory: [{ id: '1', name: 'Диван Бостон', stock: 5, value: 250000 }],
  returns: { count: 2, amount: 30000 },
  lostSales: { total: 150000, topReasons: [{ reason: 'Нет в наличии', amount: 100000 }, { reason: 'Дорого', amount: 50000 }] },
  marketingBudget: { total: 100000, used: 30000, remaining: 70000 },
  comment: 'Тестовый комментарий',
};

const reportHistory = [
  { id: '1', period: 'Сентябрь 2026', generatedAt: '30.09.2026', sentAt: '01.10.2026', status: 'sent' as const },
  { id: '2', period: 'Август 2026', generatedAt: '31.08.2026', status: 'draft' as const },
];

describe('DealerReportTab', () => {
  it('рендерит генерацию отчёта с контролами', () => {
    render(<DealerReportTab />);
    expect(screen.getByText('📊 Генерация отчёта для бренда')).toBeInTheDocument();
    expect(screen.getByText('Период отчёта:')).toBeInTheDocument();
    expect(screen.getByText('Месяц')).toBeInTheDocument();
    expect(screen.getByText('Дата:')).toBeInTheDocument();
    expect(screen.getByText('Комментарий дилера:')).toBeInTheDocument();
    expect(screen.getByText('Сформировать отчёт')).toBeInTheDocument();
    expect(screen.getByText('История пуста')).toBeInTheDocument();
    expect(screen.getByText("Нажмите 'Сформировать отчёт' для предпросмотра")).toBeInTheDocument();
  });

  it('показывает loading', () => {
    const { container } = render(<DealerReportTab loading />);
    expect(container.querySelector('.ant-spin')).toBeInTheDocument();
  });

  it('отображает предпросмотр отчёта со всеми блоками', () => {
    render(<DealerReportTab reportData={reportData} reportHistory={reportHistory} />);
    expect(screen.getByText('ОТЧЁТ ДИЛЕРА')).toBeInTheDocument();
    expect(screen.getByText('ООО Тест')).toBeInTheDocument();
    expect(screen.getAllByText(/01\.09\.2026/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/30\.09\.2026/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('📈 Блок 1: План-факт')).toBeInTheDocument();
    expect(screen.getAllByText('Факт').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('🔽 Блок 2: Воронка продаж')).toBeInTheDocument();
    expect(screen.getByText('Трафик:')).toBeInTheDocument();
    expect(screen.getByText('📦 Блок 3: Товарные остатки (топ-10)')).toBeInTheDocument();
    expect(screen.getByText('Диван Бостон')).toBeInTheDocument();
    expect(screen.getByText('↩️ Блок 4: Возвраты и рекламации')).toBeInTheDocument();
    expect(screen.getByText('💸 Блок 5: Упущенная прибыль')).toBeInTheDocument();
    expect(screen.getByText(/Нет в наличии/)).toBeInTheDocument();
    expect(screen.getByText('💰 Блок 6: Маркетинговый бюджет')).toBeInTheDocument();
    expect(screen.getByText('💬 Блок 7: Комментарий дилера')).toBeInTheDocument();
    expect(screen.getByText('Тестовый комментарий')).toBeInTheDocument();
    expect(screen.getByText('История отчётов')).toBeInTheDocument();
    expect(screen.getByText('Сентябрь 2026')).toBeInTheDocument();
  });

  it('показывает fallback комментария когда его нет', () => {
    render(<DealerReportTab reportData={{ ...reportData, comment: undefined }} />);
    expect(screen.getByText('Нет комментария')).toBeInTheDocument();
  });

  it('генерирует отчёт и открывает модалку', async () => {
    jest.useFakeTimers();
    render(<DealerReportTab reportData={reportData} />);
    fireEvent.click(screen.getByText('Сформировать отчёт'));
    jest.advanceTimersByTime(1500);
    await waitFor(() => expect(screen.getByText('Предпросмотр отчёта')).toBeInTheDocument());
    jest.useRealTimers();
  });

  it('скачивает PDF', async () => {
    render(<DealerReportTab reportData={reportData} />);
    fireEvent.click(screen.getByText('Скачать PDF'));
    await waitFor(() => expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('PDF скачивается...'));
  });

  it('отправляет отчёт успешно', async () => {
    const onSend = jest.fn().mockResolvedValue(undefined);
    render(<DealerReportTab reportData={reportData} onSend={onSend} />);
    fireEvent.click(screen.getByText('Отправить менеджеру'));
    await waitFor(() => expect(onSend).toHaveBeenCalledWith(reportData));
    await waitFor(() => expect(jest.requireMock('antd').message.success).toHaveBeenCalledWith('Отчёт отправлен'));
  });

  it('обрабатывает ошибку отправки', async () => {
    const onSend = jest.fn().mockRejectedValue(new Error('fail'));
    render(<DealerReportTab reportData={reportData} onSend={onSend} />);
    fireEvent.click(screen.getByText('Отправить менеджеру'));
    await waitFor(() => expect(jest.requireMock('antd').message.error).toHaveBeenCalledWith('Ошибка отправки'));
  });

  it('меняет период и комментарий', () => {
    render(<DealerReportTab />);
    const textarea = screen.getByPlaceholderText('Добавьте комментарий...');
    fireEvent.change(textarea, { target: { value: 'Новый коммент' } });
    expect(textarea).toHaveValue('Новый коммент');
  });

  it('отображает историю с sentAt', () => {
    render(<DealerReportTab reportHistory={reportHistory} />);
    expect(screen.getByText('Сентябрь 2026')).toBeInTheDocument();
    expect(screen.getByText('Август 2026')).toBeInTheDocument();
  });
});

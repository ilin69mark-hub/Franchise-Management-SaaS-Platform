import React from 'react';
import { ComposedChart, Line, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, ReferenceLine, Area, AreaChart } from 'recharts';

interface DataPoint {
  month: string;
  plan: number;
  fact: number | null;
  forecast: number | null;
  cumulativePlan: number;
  cumulativeFact: number;
}

const mockData: DataPoint[] = [
  { month: 'Янв', plan: 3500000, fact: 3200000, forecast: null, cumulativePlan: 3500000, cumulativeFact: 3200000 },
  { month: 'Фев', plan: 3500000, fact: 2900000, forecast: null, cumulativePlan: 7000000, cumulativeFact: 6100000 },
  { month: 'Мар', plan: 3500000, fact: 3800000, forecast: null, cumulativePlan: 10500000, cumulativeFact: 9900000 },
  { month: 'Апр', plan: 3500000, fact: 3400000, forecast: null, cumulativePlan: 14000000, cumulativeFact: 13300000 },
  { month: 'Май', plan: 3500000, fact: 3100000, forecast: null, cumulativePlan: 17500000, cumulativeFact: 16400000 },
  { month: 'Июн', plan: 3500000, fact: null, forecast: 3300000, cumulativePlan: 21000000, cumulativeFact: 19700000 },
  { month: 'Июл', plan: 3500000, fact: null, forecast: 3400000, cumulativePlan: 24500000, cumulativeFact: 23100000 },
  { month: 'Авг', plan: 3500000, fact: null, forecast: 3500000, cumulativePlan: 28000000, cumulativeFact: 26600000 },
];

const ReportPlanFactChart: React.FC = () => {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <AreaChart data={mockData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
        <XAxis dataKey="month" tick={{ fontSize: 11 }} />
        <YAxis 
          tick={{ fontSize: 11 }} 
          tickFormatter={(v) => `${(v / 1000000).toFixed(1)}M`}
        />
        <Tooltip 
          formatter={(value: number) => value ? `${value.toLocaleString()} ₽` : '-'}
          contentStyle={{ borderRadius: 4 }}
        />
        <Legend />
        
        <Area
          type="monotone"
          dataKey="cumulativePlan"
          stroke="transparent"
          fill="#f0f0f0"
          fillOpacity={0.3}
          name="План накоп."
        />
        
        <Line
          type="monotone"
          dataKey="cumulativePlan"
          stroke="#8c8c8c"
          strokeWidth={2}
          strokeDasharray="5 5"
          dot={false}
          name="План"
        />
        
        <Line
          type="monotone"
          dataKey="cumulativeFact"
          stroke="#52c41a"
          strokeWidth={2}
          dot={{ r: 4, fill: '#52c41a' }}
          connectNulls={false}
          name="Факт"
        />
        
        <Line
          type="monotone"
          dataKey="cumulativeFact"
          stroke="#1890ff"
          strokeWidth={2}
          strokeDasharray="5 5"
          dot={{ r: 4, fill: '#1890ff' }}
          connectNulls={false}
          name="Прогноз"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
};

export default ReportPlanFactChart;
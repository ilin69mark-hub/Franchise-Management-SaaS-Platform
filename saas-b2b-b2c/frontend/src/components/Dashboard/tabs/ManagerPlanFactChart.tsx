import React from 'react';
import { ComposedChart, Line, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, ReferenceLine } from 'recharts';

interface DataPoint {
  month: string;
  plan: number;
  fact: number;
}

interface ManagerPlanFactChartProps {
  data: DataPoint[];
}

const ManagerPlanFactChart: React.FC<ManagerPlanFactChartProps> = ({ data }) => {
  if (!data || data.length === 0) {
    return <div style={{ textAlign: 'center', color: '#999' }}>Нет данных</div>;
  }

  const maxValue = Math.max(...data.map(d => Math.max(d.plan, d.fact)));

  return (
    <ResponsiveContainer width="100%" height={200}>
      <ComposedChart data={data} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
        <XAxis dataKey="month" tick={{ fontSize: 11 }} />
        <YAxis 
          tick={{ fontSize: 11 }} 
          tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`}
          domain={[0, maxValue * 1.1]}
        />
        <Tooltip 
          formatter={(value: number) => [`${value.toLocaleString()} ₽`, '']}
          contentStyle={{ borderRadius: 4 }}
        />
        <Legend />
        <Bar dataKey="plan" fill="#f0f0f0" name="План" />
        <Bar dataKey="fact" fill="#52c41a" name="Факт" />
        <Line type="monotone" dataKey="fact" stroke="#52c41a" strokeWidth={2} dot={{ r: 3 }} />
      </ComposedChart>
    </ResponsiveContainer>
  );
};

export default ManagerPlanFactChart;
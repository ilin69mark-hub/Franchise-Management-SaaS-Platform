import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';

interface DataPoint {
  month: string;
  kpi: number;
}

interface ManagerKpiChartProps {
  data: DataPoint[];
}

const ManagerKpiChart: React.FC<ManagerKpiChartProps> = ({ data }) => {
  if (!data || data.length === 0) {
    return <div style={{ textAlign: 'center', color: '#999' }}>Нет данных</div>;
  }

  return (
    <ResponsiveContainer width="100%" height={200}>
      <LineChart data={data} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
        <XAxis dataKey="month" tick={{ fontSize: 11 }} />
        <YAxis domain={[0, 100]} tick={{ fontSize: 11 }} tickFormatter={(v) => `${v}%`} />
        <Tooltip 
          formatter={(value: number) => [`${value}%`, 'KPI']}
          contentStyle={{ borderRadius: 4 }}
        />
        <ReferenceLine y={90} stroke="#52c41a" strokeDasharray="3 3" />
        <ReferenceLine y={75} stroke="#fa8c16" strokeDasharray="3 3" />
        <Line 
          type="monotone" 
          dataKey="kpi" 
          stroke="#1890ff" 
          strokeWidth={2}
          dot={{ r: 4, fill: '#1890ff' }}
          activeDot={{ r: 6 }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
};

export default ManagerKpiChart;
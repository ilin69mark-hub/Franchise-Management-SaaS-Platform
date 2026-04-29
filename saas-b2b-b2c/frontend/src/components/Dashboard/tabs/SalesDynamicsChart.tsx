import React from 'react';
import { Card, Radio, Space } from 'antd';
import { 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  Legend, 
  ResponsiveContainer,
  Area,
  ComposedChart,
  ReferenceLine,
} from 'recharts';

interface SalesData {
  month: string;
  plan: number;
  fact: number | null;
  forecast: number | null;
}

interface SalesDynamicsChartProps {
  data: SalesData[];
  view?: 'monthly' | 'quarterly' | 'cumulative';
}

const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    return (
      <Card size="small" style={{ background: '#fff' }}>
        <Space direction="vertical" size={0}>
          <strong>{label}</strong>
          {payload.map((p: any) => (
            <div key={p.name} style={{ color: p.color }}>
              {p.name}: {p.value?.toLocaleString()} ₽
            </div>
          ))}
        </Space>
      </Card>
    );
  }
  return null;
};

const SalesDynamicsChart: React.FC<SalesDynamicsChartProps> = ({ data, view = 'monthly' }) => {
  const processedData = React.useMemo(() => {
    if (view === 'cumulative') {
      let cumPlan = 0;
      let cumFact = 0;
      return data.map(d => {
        cumPlan += d.plan;
        cumFact += d.fact || 0;
        return {
          ...d,
          plan: cumPlan,
          fact: d.fact ? cumFact : null,
          forecast: d.forecast ? cumPlan : null,
        };
      });
    }
    return data;
  }, [data, view]);

  const minValue = Math.min(
    ...processedData.map(d => Math.min(d.plan, d.fact || d.plan, d.forecast || d.plan))
  );
  const maxValue = Math.max(
    ...processedData.map(d => Math.max(d.plan, d.fact || 0, d.forecast || 0))
  );

  return (
    <div>
      <Radio.Group defaultValue={view} style={{ marginBottom: 16 }}>
        <Radio.Button value="monthly">Помесячно</Radio.Button>
        <Radio.Button value="quarterly">Поквартально</Radio.Button>
        <Radio.Button value="cumulative">Накопленным итогом</Radio.Button>
      </Radio.Group>

      <ResponsiveContainer width="100%" height={350}>
        <ComposedChart data={processedData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          <XAxis 
            dataKey="month" 
            tick={{ fontSize: 12 }}
            axisLine={{ stroke: '#d9d9d9' }}
          />
          <YAxis 
            tick={{ fontSize: 12 }}
            axisLine={{ stroke: '#d9d9d9' }}
            tickFormatter={(v) => `${(v / 1000000).toFixed(1)}M`}
            domain={[0, maxValue * 1.1]}
          />
          <Tooltip content={<CustomTooltip />} />
          <Legend />

          <Area
            type="monotone"
            dataKey="plan"
            stroke="transparent"
            fill="#f0f0f0"
            fillOpacity={0.3}
            name="План сети"
          />

          <Line
            type="dash"
            dataKey="plan"
            stroke="#8c8c8c"
            strokeWidth={2}
            strokeDasharray="5 5"
            dot={false}
            name="План сети"
          />

          <Line
            type="monotone"
            dataKey="fact"
            stroke="#52c41a"
            strokeWidth={2}
            dot={{ r: 4, fill: '#52c41a' }}
            connectNulls={false}
            name="Факт сети"
          />

          <Line
            type="dash"
            dataKey="forecast"
            stroke="#1890ff"
            strokeWidth={2}
            strokeDasharray="5 5"
            dot={{ r: 4, fill: '#1890ff' }}
            connectNulls={false}
            name="Прогноз"
          />
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
};

export default SalesDynamicsChart;
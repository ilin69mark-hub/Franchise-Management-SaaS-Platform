import React from 'react';
import { Card, Table, Tag, Typography } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface RegionData {
  region: string;
  dealers: number;
  population: number;
  density: 'high' | 'medium' | 'low' | 'none';
}

const mockRegions: RegionData[] = [
  { region: 'Центр (Москва, СПб)', dealers: 8, population: 25000000, density: 'high' },
  { region: 'Приволжье', dealers: 5, population: 28000000, density: 'medium' },
  { region: 'Юг', dealers: 4, population: 20000000, density: 'medium' },
  { region: 'Урал', dealers: 3, population: 12000000, density: 'low' },
  { region: 'Сибирь', dealers: 2, population: 15000000, density: 'low' },
  { region: 'Дальний Восток', dealers: 1, population: 8000000, density: 'none' },
  { region: 'Северо-Запад', dealers: 1, population: 10000000, density: 'none' },
];

const GeographyMap: React.FC = () => {
  const columns = [
    {
      title: 'Регион',
      dataIndex: 'region',
      key: 'region',
      render: (r: string) => (
        <Text>
          <GlobalOutlined /> {r}
        </Text>
      ),
    },
    {
      title: 'Дилеров',
      dataIndex: 'dealers',
      key: 'dealers',
    },
    {
      title: 'Население',
      dataIndex: 'population',
      key: 'population',
      render: (v: number) => (v / 1000000).toFixed(1) + 'M',
    },
    {
      title: 'Плотность',
      dataIndex: 'density',
      key: 'density',
      render: (d: 'high' | 'medium' | 'low' | 'none') => (
        <Tag 
          color={d === 'high' ? 'green' : d === 'medium' ? 'orange' : d === 'low' ? 'red' : 'default'}
        >
          {d === 'high' ? 'Высокая' : d === 'medium' ? 'Средняя' : d === 'low' ? 'Низкая' : 'Нет'}
        </Tag>
      ),
    },
  ];

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Text type="secondary">
          Карта присутствия дилеров по регионам России
        </Text>
      </Card>
      <Table
        dataSource={mockRegions}
        columns={columns}
        rowKey="region"
        pagination={false}
        size="small"
      />
    </div>
  );
};

export default GeographyMap;
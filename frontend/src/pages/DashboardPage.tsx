// Dashboard ringkas — contoh kartu statistik dari data server (React Query).

import { TeamOutlined, UserAddOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { Card, Col, Row, Skeleton, Statistic, Typography } from 'antd';
import { daftarUsers } from '../api/users';

export default function DashboardPage() {
  // Cukup total dari meta — per_page kecil agar payload hemat.
  const { data, isLoading } = useQuery({
    queryKey: ['users', { page: 1, per_page: 1 }],
    queryFn: () => daftarUsers({ page: 1, per_page: 1 }),
  });

  return (
    <>
      <Typography.Title level={4}>Dashboard</Typography.Title>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            {isLoading ? (
              <Skeleton active paragraph={false} />
            ) : (
              <Statistic title="Total Pengguna" value={data?.total ?? 0} prefix={<TeamOutlined />} />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic title="Modul Terpasang" value={1} prefix={<UserAddOutlined />} suffix="/ ∞" />
          </Card>
        </Col>
      </Row>
    </>
  );
}

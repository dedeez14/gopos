// Halaman Pengguna — contoh lengkap pola server-state dengan TanStack Query:
// pencarian + paginasi server-side, cache per queryKey, dan tabel AntD.

import { useState } from 'react';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { Card, Input, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { daftarUsers, type User } from '../api/users';

const WARNA_ROLE: Record<User['role'], string> = {
  OWNER: 'gold',
  MANAGER: 'blue',
  KASIR: 'green',
};

const kolom: ColumnsType<User> = [
  { title: 'Nama', dataIndex: 'nama' },
  { title: 'Email', dataIndex: 'email' },
  {
    title: 'Peran',
    dataIndex: 'role',
    width: 120,
    render: (role: User['role']) => <Tag color={WARNA_ROLE[role]}>{role}</Tag>,
  },
  {
    title: 'Status',
    dataIndex: 'aktif',
    width: 110,
    render: (aktif: boolean) =>
      aktif ? <Tag color="success">Aktif</Tag> : <Tag color="default">Nonaktif</Tag>,
  },
  {
    title: 'Dibuat',
    dataIndex: 'dibuat_pada',
    width: 180,
    render: (iso: string) => new Date(iso).toLocaleString('id-ID'),
  },
];

export default function UsersPage() {
  const [cari, setCari] = useState('');
  const [halaman, setHalaman] = useState(1);
  const perHalaman = 10;

  // queryKey memuat seluruh parameter — ganti cari/halaman otomatis refetch,
  // dan hasil lama tetap tampil selagi memuat (placeholderData).
  const { data, isFetching } = useQuery({
    queryKey: ['users', { q: cari, page: halaman, per_page: perHalaman }],
    queryFn: () => daftarUsers({ q: cari, page: halaman, per_page: perHalaman }),
    placeholderData: keepPreviousData,
  });

  return (
    <Card>
      <Typography.Title level={4}>Pengguna</Typography.Title>

      <Input.Search
        placeholder="Cari nama atau email…"
        allowClear
        style={{ maxWidth: 320, marginBottom: 16 }}
        onSearch={(nilai) => {
          setHalaman(1);
          setCari(nilai);
        }}
      />

      <Table<User>
        rowKey="id"
        columns={kolom}
        dataSource={data?.rows}
        loading={isFetching}
        pagination={{
          current: halaman,
          pageSize: perHalaman,
          total: data?.total ?? 0,
          onChange: setHalaman,
          showTotal: (total) => `${total} pengguna`,
        }}
      />
    </Card>
  );
}

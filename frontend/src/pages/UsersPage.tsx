// Halaman Pengguna — CRUD lengkap: tabel + modal tambah/ubah (nama, email,
// role, aktif, dan GANTI SANDI — kosongkan bila tidak diganti).

import { useState } from 'react';
import { PlusOutlined } from '@ant-design/icons';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Button, Card, Form, Input, Modal, Select, Space, Switch, Table, Tag,
  Typography, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { api, bukaAmplop, pesanError, type Amplop } from '../api/client';
import { daftarUsers, type User } from '../api/users';

const WARNA_ROLE: Record<User['role'], string> = {
  OWNER: 'gold',
  MANAGER: 'blue',
  KASIR: 'green',
};

interface FormUser {
  nama: string;
  email: string;
  password?: string;
  role: User['role'];
  aktif: boolean;
}

async function simpanUser(id: number | null, v: FormUser): Promise<User> {
  const payload = { ...v, password: v.password || undefined };
  const res = id
    ? await api.put<Amplop<User>>(`/users/${id}`, payload)
    : await api.post<Amplop<User>>('/users', payload);
  return bukaAmplop(res.data);
}

export default function UsersPage() {
  const qc = useQueryClient();
  const [form] = Form.useForm<FormUser>();
  const [cari, setCari] = useState('');
  const [halaman, setHalaman] = useState(1);
  const [modalBuka, setModalBuka] = useState(false);
  const [sedangEdit, setSedangEdit] = useState<User | null>(null);
  const perHalaman = 10;

  const { data, isFetching } = useQuery({
    queryKey: ['users', { q: cari, page: halaman, per_page: perHalaman }],
    queryFn: () => daftarUsers({ q: cari, page: halaman, per_page: perHalaman }),
    placeholderData: keepPreviousData,
  });

  const simpan = useMutation({
    mutationFn: (v: FormUser) => simpanUser(sedangEdit?.id ?? null, v),
    onSuccess: () => {
      message.success(sedangEdit ? 'Pengguna diperbarui.' : 'Pengguna ditambahkan.');
      setModalBuka(false);
      qc.invalidateQueries({ queryKey: ['users'] });
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const bukaTambah = () => {
    setSedangEdit(null);
    form.resetFields();
    form.setFieldsValue({ role: 'KASIR', aktif: true });
    setModalBuka(true);
  };

  const bukaEdit = (u: User) => {
    setSedangEdit(u);
    form.setFieldsValue({ nama: u.nama, email: u.email, role: u.role, aktif: u.aktif, password: '' });
    setModalBuka(true);
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
      title: 'Aksi',
      width: 90,
      render: (_, u) => (
        <Button size="small" onClick={() => bukaEdit(u)}>
          Ubah
        </Button>
      ),
    },
  ];

  return (
    <Card>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Pengguna
        </Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={bukaTambah}>
          Tambah
        </Button>
      </Space>

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

      <Modal
        title={sedangEdit ? `Ubah — ${sedangEdit.email}` : 'Tambah Pengguna'}
        open={modalBuka}
        onCancel={() => setModalBuka(false)}
        onOk={() => form.submit()}
        confirmLoading={simpan.isPending}
        okText="Simpan"
        cancelText="Batal"
        destroyOnClose
      >
        <Form<FormUser> form={form} layout="vertical" onFinish={(v) => simpan.mutate(v)}>
          <Form.Item name="nama" label="Nama" rules={[{ required: true, message: 'Nama wajib diisi.' }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="email"
            label="Email (untuk masuk)"
            rules={[
              { required: true, message: 'Email wajib diisi.' },
              { type: 'email', message: 'Format email tidak sah.' },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="password"
            label={sedangEdit ? 'Sandi baru (kosongkan bila tidak diganti)' : 'Sandi'}
            rules={[
              { required: !sedangEdit, message: 'Sandi wajib diisi.' },
              { min: 8, message: 'Minimal 8 karakter.' },
            ]}
          >
            <Input.Password placeholder="min. 8 karakter" />
          </Form.Item>
          <Space size="large" style={{ width: '100%' }}>
            <Form.Item name="role" label="Peran" rules={[{ required: true }]} style={{ minWidth: 160 }}>
              <Select
                options={[
                  { value: 'OWNER', label: 'Owner (akses penuh)' },
                  { value: 'MANAGER', label: 'Manager' },
                  { value: 'KASIR', label: 'Kasir' },
                ]}
              />
            </Form.Item>
            <Form.Item name="aktif" label="Aktif" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </Card>
  );
}

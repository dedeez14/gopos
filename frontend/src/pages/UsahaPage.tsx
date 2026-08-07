// Usaha (Merchant) — khusus SUPERADMIN. Daftar tenant + buat usaha baru
// sekaligus akun owner-nya + suspend/aktifkan.

import { useState } from 'react';
import { PlusOutlined } from '@ant-design/icons';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Button, Card, Form, Input, Modal, Space, Switch, Table, Tag, Typography, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { api, bukaAmplop, pesanError, type Amplop } from '../api/client';

interface Usaha {
  id: number;
  kode: string;
  nama: string;
  aktif: boolean;
  midtrans_aktif: boolean;
  dibuat: string;
}

interface FormUsaha {
  nama: string;
  kode?: string;
  owner_nama: string;
  owner_email: string;
  owner_password: string;
}

async function daftarUsaha(q: string, page: number) {
  const res = await api.get<Amplop<Usaha[]>>('/usahas', { params: { q, page, per_page: 10 } });
  return { rows: bukaAmplop(res.data), total: res.data.meta?.total ?? 0 };
}

export default function UsahaPage() {
  const qc = useQueryClient();
  const [form] = Form.useForm<FormUsaha>();
  const [cari, setCari] = useState('');
  const [halaman, setHalaman] = useState(1);
  const [modalBuka, setModalBuka] = useState(false);

  const { data, isFetching } = useQuery({
    queryKey: ['usahas', { q: cari, page: halaman }],
    queryFn: () => daftarUsaha(cari, halaman),
    placeholderData: keepPreviousData,
  });

  const segarkan = () => qc.invalidateQueries({ queryKey: ['usahas'] });

  const buat = useMutation({
    mutationFn: (v: FormUsaha) => api.post('/usahas', v).then((r) => bukaAmplop(r.data as Amplop<Usaha>)),
    onSuccess: () => {
      message.success('Usaha dibuat — owner sudah bisa masuk.');
      setModalBuka(false);
      form.resetFields();
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const toggle = useMutation({
    mutationFn: (v: { id: number; aktif: boolean }) =>
      api.patch(`/usahas/${v.id}`, { aktif: v.aktif }),
    onSuccess: (_r, v) => {
      message.success(v.aktif ? 'Usaha diaktifkan.' : 'Usaha di-suspend — penggunanya tak bisa masuk.');
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const toggleMidtrans = useMutation({
    mutationFn: (v: { id: number; on: boolean }) =>
      api.patch(`/usahas/${v.id}`, { midtrans_aktif: v.on }),
    onSuccess: (_r, v) => {
      message.success(v.on ? 'Modul Midtrans diaktifkan untuk usaha ini.' : 'Modul Midtrans dimatikan.');
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const kolom: ColumnsType<Usaha> = [
    { title: 'Kode', dataIndex: 'kode', width: 140 },
    { title: 'Nama Usaha', dataIndex: 'nama' },
    {
      title: 'Dibuat',
      dataIndex: 'dibuat',
      width: 170,
      render: (iso: string) => new Date(iso).toLocaleDateString('id-ID'),
    },
    {
      title: 'Status',
      dataIndex: 'aktif',
      width: 160,
      render: (aktif: boolean, u) => (
        <Space>
          {aktif ? <Tag color="success">Aktif</Tag> : <Tag color="red">Suspended</Tag>}
          <Switch
            size="small"
            checked={aktif}
            loading={toggle.isPending}
            onChange={(v) => toggle.mutate({ id: u.id, aktif: v })}
          />
        </Space>
      ),
    },
    {
      title: 'Modul Midtrans',
      dataIndex: 'midtrans_aktif',
      width: 150,
      render: (on: boolean, u) => (
        <Space>
          <Switch
            size="small"
            checked={on}
            loading={toggleMidtrans.isPending}
            onChange={(v) => toggleMidtrans.mutate({ id: u.id, on: v })}
          />
          <span style={{ fontSize: 12, color: 'var(--muted)' }}>{on ? 'aktif' : 'nonaktif'}</span>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Usaha / Merchant
        </Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalBuka(true)}>
          Buat Usaha
        </Button>
      </Space>

      <Input.Search
        placeholder="Cari nama/kode usaha…"
        allowClear
        style={{ maxWidth: 320, marginBottom: 16 }}
        onSearch={(v) => {
          setHalaman(1);
          setCari(v);
        }}
      />

      <Table<Usaha>
        rowKey="id"
        columns={kolom}
        dataSource={data?.rows}
        loading={isFetching}
        pagination={{
          current: halaman,
          pageSize: 10,
          total: data?.total ?? 0,
          onChange: setHalaman,
          showTotal: (t) => `${t} usaha`,
        }}
      />

      <Modal
        title="Buat Usaha Baru"
        open={modalBuka}
        onCancel={() => setModalBuka(false)}
        onOk={() => form.submit()}
        confirmLoading={buat.isPending}
        okText="Buat"
        cancelText="Batal"
        destroyOnClose
      >
        <Form<FormUsaha> form={form} layout="vertical" onFinish={(v) => buat.mutate(v)}>
          <Typography.Text type="secondary">Data usaha</Typography.Text>
          <Space.Compact block style={{ marginTop: 8 }}>
            <Form.Item name="nama" label="Nama Usaha" style={{ width: '65%' }} rules={[{ required: true, message: 'Nama usaha wajib.' }]}>
              <Input placeholder="mis. Warung Bu Sari" />
            </Form.Item>
            <Form.Item name="kode" label="Kode (opsional)" style={{ width: '35%' }}>
              <Input placeholder="otomatis" />
            </Form.Item>
          </Space.Compact>

          <Typography.Text type="secondary">Akun pemilik (langsung bisa login)</Typography.Text>
          <Form.Item name="owner_nama" label="Nama Owner" style={{ marginTop: 8 }} rules={[{ required: true, message: 'Nama owner wajib.' }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="owner_email"
            label="Email Owner"
            rules={[
              { required: true, message: 'Email wajib.' },
              { type: 'email', message: 'Format email tidak sah.' },
            ]}
          >
            <Input placeholder="owner@usaha.id" />
          </Form.Item>
          <Form.Item
            name="owner_password"
            label="Sandi Owner"
            rules={[
              { required: true, message: 'Sandi wajib.' },
              { min: 8, message: 'Minimal 8 karakter.' },
            ]}
          >
            <Input.Password placeholder="min. 8 karakter" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}

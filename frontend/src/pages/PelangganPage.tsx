// Pelanggan — daftar + quick add (nama + WA) seperti kartu di layar
// Inventory Tuléh; tautan WhatsApp langsung dari nomor ternormalisasi.

import { useState } from 'react';
import { WhatsAppOutlined } from '@ant-design/icons';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Card, Form, Input, Space, Table, Tag, Typography, message } from 'antd';
import { daftarPelanggan, quickPelanggan, type Pelanggan } from '../api/laporan';
import { pesanError } from '../api/client';

export default function PelangganPage() {
  const qc = useQueryClient();
  const [form] = Form.useForm<{ nama: string; telepon?: string }>();
  const [cari, setCari] = useState('');
  const [halaman, setHalaman] = useState(1);

  const { data, isFetching } = useQuery({
    queryKey: ['pelanggan', { q: cari, page: halaman }],
    queryFn: () => daftarPelanggan(cari, halaman),
    placeholderData: keepPreviousData,
  });

  const tambah = useMutation({
    mutationFn: (v: { nama: string; telepon?: string }) => quickPelanggan(v.nama, v.telepon ?? ''),
    onSuccess: () => {
      message.success('Pelanggan tersimpan.');
      form.resetFields();
      qc.invalidateQueries({ queryKey: ['pelanggan'] });
    },
    onError: (e) => message.error(pesanError(e)),
  });

  return (
    <Card>
      <Typography.Title level={4}>Pelanggan</Typography.Title>

      <Card size="small" title="Quick Add (nama + WhatsApp)" style={{ marginBottom: 16 }}>
        <Form form={form} layout="inline" onFinish={(v) => tambah.mutate(v)}>
          <Form.Item
            name="nama"
            rules={[{ required: true, message: 'Nama wajib diisi.' }]}
            style={{ minWidth: 200 }}
          >
            <Input placeholder="Nama pelanggan" />
          </Form.Item>
          <Form.Item name="telepon" style={{ minWidth: 180 }}>
            <Input placeholder="No. WA (mis. 0812…)" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={tambah.isPending}>
            Simpan
          </Button>
        </Form>
      </Card>

      <Input.Search
        placeholder="Cari nama/telepon…"
        allowClear
        style={{ maxWidth: 320, marginBottom: 16 }}
        onSearch={(v) => {
          setHalaman(1);
          setCari(v);
        }}
      />

      <Table<Pelanggan>
        rowKey="id"
        loading={isFetching}
        dataSource={data?.rows}
        pagination={{
          current: halaman,
          pageSize: 10,
          total: data?.total ?? 0,
          onChange: setHalaman,
          showTotal: (t) => `${t} pelanggan`,
        }}
        columns={[
          { title: 'Nama', dataIndex: 'nama' },
          {
            title: 'WhatsApp',
            width: 200,
            render: (_, p) =>
              p.wa_link ? (
                <Space>
                  <span>{p.telepon}</span>
                  <Button
                    size="small"
                    icon={<WhatsAppOutlined />}
                    href={p.wa_link}
                    target="_blank"
                    style={{ color: '#25D366' }}
                  />
                </Space>
              ) : (
                <Tag>tanpa nomor</Tag>
              ),
          },
          { title: 'Catatan', dataIndex: 'catatan' },
        ]}
      />
    </Card>
  );
}

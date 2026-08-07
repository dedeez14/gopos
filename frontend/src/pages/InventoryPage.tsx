// Inventory — paritas layar Tuléh: kartu Tambah Stok & Opname berdampingan +
// tabel Riwayat Perubahan Stok (MASUK/OPNAME/JUAL/BATAL, satu buku).

import { useState } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Button, Card, Col, Form, Input, InputNumber, Row, Select, Table, Tag,
  Typography, message,
} from 'antd';
import { daftarProduk } from '../api/produk';
import { opname, riwayatStok, stokMasuk, type StokLog } from '../api/inventory';
import { pesanError } from '../api/client';

const WARNA_JENIS: Record<StokLog['jenis'], string> = {
  MASUK: 'green',
  OPNAME: 'orange',
  JUAL: 'blue',
  BATAL: 'purple',
};

export default function InventoryPage() {
  const qc = useQueryClient();
  const [formMasuk] = Form.useForm<{ produk_id: number; jumlah: number; keterangan?: string }>();
  const [formOpname] = Form.useForm<{ produk_id: number; stok_fisik: number; keterangan?: string }>();
  const [halaman, setHalaman] = useState(1);

  // Dropdown produk kelola-stok (maks 100 — cukup utk pilihan; cari via ketik).
  const { data: produk } = useQuery({
    queryKey: ['produk-stok'],
    queryFn: () => daftarProduk({ per_page: 100, termasuk_nonaktif: undefined }),
  });
  const opsiProduk = produk?.rows
    .filter((p) => p.kelola_stok)
    .map((p) => ({ value: p.id, label: `${p.nama} (stok: ${p.stok})` }));

  const { data: riwayat, isFetching } = useQuery({
    queryKey: ['riwayat-stok', halaman],
    queryFn: () => riwayatStok(halaman),
    placeholderData: keepPreviousData,
  });

  const segarkan = () => {
    qc.invalidateQueries({ queryKey: ['riwayat-stok'] });
    qc.invalidateQueries({ queryKey: ['produk-stok'] });
    qc.invalidateQueries({ queryKey: ['produk'] });
  };

  const masuk = useMutation({
    mutationFn: (v: { produk_id: number; jumlah: number; keterangan?: string }) =>
      stokMasuk(v.produk_id, v.jumlah, v.keterangan),
    onSuccess: () => {
      message.success('Stok ditambahkan.');
      formMasuk.resetFields();
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const koreksi = useMutation({
    mutationFn: (v: { produk_id: number; stok_fisik: number; keterangan?: string }) =>
      opname(v.produk_id, v.stok_fisik, v.keterangan),
    onSuccess: (log) => {
      message.success(`Opname tercatat (selisih ${log.jumlah > 0 ? '+' : ''}${log.jumlah}).`);
      formOpname.resetFields();
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  return (
    <>
      <Typography.Title level={4}>Inventory</Typography.Title>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="Tambah Stok (restock)" size="small">
            <Form form={formMasuk} layout="vertical" onFinish={(v) => masuk.mutate(v)}>
              <Form.Item name="produk_id" label="Produk" rules={[{ required: true, message: 'Pilih produk.' }]}>
                <Select showSearch optionFilterProp="label" options={opsiProduk} placeholder="Pilih produk…" />
              </Form.Item>
              <Form.Item name="jumlah" label="Jumlah masuk" rules={[{ required: true, message: 'Isi jumlah.' }]}>
                <InputNumber min={0.001} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="keterangan" label="Keterangan (opsional)">
                <Input placeholder="mis. Belanja mingguan" />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={masuk.isPending} block>
                Tambah Stok
              </Button>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="Opname (koreksi hitung fisik)" size="small">
            <Form form={formOpname} layout="vertical" onFinish={(v) => koreksi.mutate(v)}>
              <Form.Item name="produk_id" label="Produk" rules={[{ required: true, message: 'Pilih produk.' }]}>
                <Select showSearch optionFilterProp="label" options={opsiProduk} placeholder="Pilih produk…" />
              </Form.Item>
              <Form.Item
                name="stok_fisik"
                label="Stok fisik hasil hitung"
                rules={[{ required: true, message: 'Isi hasil hitung.' }]}
              >
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="keterangan" label="Keterangan (opsional)">
                <Input placeholder="mis. Stok opname akhir bulan" />
              </Form.Item>
              <Button htmlType="submit" loading={koreksi.isPending} block danger>
                Catat Opname
              </Button>
            </Form>
          </Card>
        </Col>
      </Row>

      <Card title="Riwayat Perubahan Stok" size="small">
        <Table<StokLog>
          rowKey="id"
          size="small"
          loading={isFetching}
          dataSource={riwayat?.rows}
          pagination={{
            current: halaman,
            pageSize: 10,
            total: riwayat?.total ?? 0,
            onChange: setHalaman,
          }}
          columns={[
            {
              title: 'Waktu',
              dataIndex: 'waktu',
              width: 160,
              render: (iso: string) => new Date(iso).toLocaleString('id-ID'),
            },
            { title: 'Item', dataIndex: 'produk' },
            {
              title: 'Jenis',
              dataIndex: 'jenis',
              width: 100,
              render: (j: StokLog['jenis']) => <Tag color={WARNA_JENIS[j]}>{j}</Tag>,
            },
            {
              title: 'Jumlah',
              dataIndex: 'jumlah',
              width: 90,
              align: 'right',
              render: (v: number) => (
                <span style={{ color: v >= 0 ? '#3f8600' : '#cf1322' }}>
                  {v > 0 ? `+${v}` : v}
                </span>
              ),
            },
            { title: 'Stok Akhir', dataIndex: 'stok_sesudah', width: 100, align: 'right' },
            { title: 'Sesi Kasir', dataIndex: 'sesi', width: 170, render: (s: string | null) => s ?? '—' },
            { title: 'Keterangan', dataIndex: 'keterangan' },
          ]}
        />
      </Card>
    </>
  );
}

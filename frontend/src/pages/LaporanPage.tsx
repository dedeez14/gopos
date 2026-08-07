// Laporan Keuangan — paritas layar Tuléh: 3 kartu (Omzet / Pengeluaran /
// Laba), input pengeluaran dua kolom, riwayat pengeluaran, produk terlaris.

import { useState } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Button, Card, Col, DatePicker, Form, Input, InputNumber, Popconfirm, Row,
  Statistic, Table, Typography, message,
} from 'antd';
import dayjs, { type Dayjs } from 'dayjs';
import {
  catatPengeluaran, daftarPengeluaran, hapusPengeluaran, keuangan,
  produkTerlaris, type Pengeluaran,
} from '../api/laporan';
import { pesanError } from '../api/client';

const rupiah = (n: number) => `Rp${n.toLocaleString('id-ID')}`;

export default function LaporanPage() {
  const qc = useQueryClient();
  const [form] = Form.useForm<{ keterangan: string; nominal: number }>();
  const [bulan, setBulan] = useState<Dayjs>(dayjs());
  const [halaman, setHalaman] = useState(1);
  const bulanStr = bulan.format('YYYY-MM');

  const { data: ringkas } = useQuery({
    queryKey: ['keuangan', bulanStr],
    queryFn: () => keuangan(bulanStr),
  });
  const { data: pengeluaran, isFetching } = useQuery({
    queryKey: ['pengeluaran', bulanStr, halaman],
    queryFn: () => daftarPengeluaran(bulanStr, halaman),
    placeholderData: keepPreviousData,
  });
  const { data: terlaris } = useQuery({
    queryKey: ['terlaris'],
    queryFn: () => produkTerlaris(30, 10),
  });

  const segarkan = () => {
    qc.invalidateQueries({ queryKey: ['keuangan'] });
    qc.invalidateQueries({ queryKey: ['pengeluaran'] });
  };

  const catat = useMutation({
    mutationFn: (v: { keterangan: string; nominal: number }) =>
      catatPengeluaran(v.keterangan, v.nominal),
    onSuccess: () => {
      message.success('Pengeluaran tercatat.');
      form.resetFields();
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const hapus = useMutation({
    mutationFn: hapusPengeluaran,
    onSuccess: segarkan,
    onError: (e) => message.error(pesanError(e)),
  });

  return (
    <>
      <Row align="middle" justify="space-between" style={{ marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Laporan Keuangan
        </Typography.Title>
        <DatePicker
          picker="month"
          value={bulan}
          allowClear={false}
          onChange={(v) => {
            if (v) {
              setBulan(v);
              setHalaman(1);
            }
          }}
        />
      </Row>

      {/* 3 kartu — bahasa uang sederhana: masuk, keluar, sisa. */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} md={8}>
          <Card>
            <Statistic
              title={`Omzet ${bulanStr} (${ringkas?.jumlah_trx ?? 0} transaksi)`}
              value={ringkas?.omzet ?? 0}
              formatter={(v) => rupiah(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card>
            <Statistic
              title="Total Pengeluaran"
              value={ringkas?.pengeluaran ?? 0}
              formatter={(v) => rupiah(Number(v))}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card>
            <Statistic
              title="Laba Bersih"
              value={ringkas?.laba ?? 0}
              formatter={(v) => rupiah(Number(v))}
              valueStyle={{ color: (ringkas?.laba ?? 0) >= 0 ? '#3f8600' : '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={14}>
          <Card title="Input Pengeluaran Baru" size="small" style={{ marginBottom: 16 }}>
            {/* Dua kolom persis layar Tuléh: untuk apa + berapa. */}
            <Form form={form} layout="inline" onFinish={(v) => catat.mutate(v)}>
              <Form.Item
                name="keterangan"
                rules={[{ required: true, message: 'Keterangan wajib diisi.' }]}
                style={{ flex: 1, minWidth: 220 }}
              >
                <Input placeholder="Untuk apa? (mis. Bayar listrik)" />
              </Form.Item>
              <Form.Item name="nominal" rules={[{ required: true, message: 'Nominal wajib diisi.' }]}>
                <InputNumber min={1} placeholder="Nominal" style={{ width: 160 }} />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={catat.isPending}>
                Catat
              </Button>
            </Form>
          </Card>

          <Card title={`Pengeluaran ${bulanStr}`} size="small">
            <Table<Pengeluaran>
              rowKey="id"
              size="small"
              loading={isFetching}
              dataSource={pengeluaran?.rows}
              pagination={{
                current: halaman,
                pageSize: 10,
                total: pengeluaran?.total ?? 0,
                onChange: setHalaman,
              }}
              columns={[
                { title: 'Tanggal', dataIndex: 'tanggal', width: 110 },
                { title: 'Keterangan', dataIndex: 'keterangan' },
                {
                  title: 'Nominal',
                  dataIndex: 'nominal',
                  align: 'right',
                  width: 130,
                  render: rupiah,
                },
                {
                  title: '',
                  width: 80,
                  render: (_, p) => (
                    <Popconfirm title="Hapus catatan ini?" onConfirm={() => hapus.mutate(p.id)}>
                      <Button size="small" danger>
                        Hapus
                      </Button>
                    </Popconfirm>
                  ),
                },
              ]}
            />
          </Card>
        </Col>

        <Col xs={24} lg={10}>
          <Card title="Produk Terlaris (30 hari)" size="small">
            <Table
              rowKey="produk_id"
              size="small"
              pagination={false}
              dataSource={terlaris}
              columns={[
                { title: 'Produk', dataIndex: 'nama' },
                { title: 'Terjual', dataIndex: 'terjual', align: 'right', width: 90 },
                {
                  title: 'Omzet',
                  dataIndex: 'omzet',
                  align: 'right',
                  width: 130,
                  render: rupiah,
                },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </>
  );
}

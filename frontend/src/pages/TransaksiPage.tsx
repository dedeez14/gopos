// Riwayat Transaksi — tabel + Drawer struk lengkap (baris item, diskon,
// pajak, kembalian). Filter rentang tanggal server-side.

import { useState } from 'react';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import {
  Card, DatePicker, Descriptions, Divider, Drawer, Table, Tag, Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { Dayjs } from 'dayjs';
import { daftarTransaksi, type Struk } from '../api/transaksi';

const rupiah = (n: number) => `Rp${n.toLocaleString('id-ID')}`;

const WARNA_BAYAR: Record<Struk['tipe_pembayaran'], string> = {
  TUNAI: 'green',
  TRANSFER: 'blue',
  QRIS: 'purple',
};

export default function TransaksiPage() {
  const [halaman, setHalaman] = useState(1);
  const [rentang, setRentang] = useState<[Dayjs | null, Dayjs | null] | null>(null);
  const [detail, setDetail] = useState<Struk | null>(null);
  const perHalaman = 10;

  const { data, isFetching } = useQuery({
    queryKey: ['transaksi', { page: halaman, rentang: rentang?.map((d) => d?.format('YYYY-MM-DD')) }],
    queryFn: () =>
      daftarTransaksi({
        page: halaman,
        per_page: perHalaman,
        tanggal_dari: rentang?.[0]?.format('YYYY-MM-DD'),
        tanggal_sampai: rentang?.[1]?.format('YYYY-MM-DD'),
      }),
    placeholderData: keepPreviousData,
  });

  const kolom: ColumnsType<Struk> = [
    { title: 'Nomor', dataIndex: 'nomor', width: 190 },
    {
      title: 'Waktu',
      dataIndex: 'tanggal',
      width: 170,
      render: (iso: string) => new Date(iso).toLocaleString('id-ID'),
    },
    { title: 'Kasir', dataIndex: 'kasir' },
    {
      title: 'Bayar',
      dataIndex: 'tipe_pembayaran',
      width: 110,
      render: (t: Struk['tipe_pembayaran']) => <Tag color={WARNA_BAYAR[t]}>{t}</Tag>,
    },
    {
      title: 'Total',
      dataIndex: 'grand_total',
      width: 140,
      align: 'right',
      render: (v: number) => <strong>{rupiah(v)}</strong>,
    },
  ];

  return (
    <Card>
      <Typography.Title level={4}>Riwayat Transaksi</Typography.Title>

      <DatePicker.RangePicker
        style={{ marginBottom: 16 }}
        onChange={(v) => {
          setHalaman(1);
          setRentang(v);
        }}
      />

      <Table<Struk>
        rowKey="id"
        columns={kolom}
        dataSource={data?.rows}
        loading={isFetching}
        onRow={(row) => ({ onClick: () => setDetail(row), style: { cursor: 'pointer' } })}
        pagination={{
          current: halaman,
          pageSize: perHalaman,
          total: data?.total ?? 0,
          onChange: setHalaman,
          showTotal: (t) => `${t} transaksi`,
        }}
      />

      <Drawer
        title={detail?.nomor}
        open={Boolean(detail)}
        onClose={() => setDetail(null)}
        width={480}
      >
        {detail && (
          <>
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Waktu">
                {new Date(detail.tanggal).toLocaleString('id-ID')}
              </Descriptions.Item>
              <Descriptions.Item label="Kasir">{detail.kasir}</Descriptions.Item>
              <Descriptions.Item label="Pembayaran">
                <Tag color={WARNA_BAYAR[detail.tipe_pembayaran]}>{detail.tipe_pembayaran}</Tag>
              </Descriptions.Item>
              {detail.catatan && (
                <Descriptions.Item label="Catatan">{detail.catatan}</Descriptions.Item>
              )}
            </Descriptions>

            <Divider style={{ margin: '12px 0' }} />

            <Table<Struk['items'][number]>
              rowKey={(_, i) => String(i)}
              size="small"
              pagination={false}
              columns={[
                { title: 'Item', dataIndex: 'nama_produk' },
                {
                  title: 'Qty',
                  width: 70,
                  render: (_, it) => `${it.kuantitas} ${it.satuan}`,
                },
                {
                  title: 'Subtotal',
                  dataIndex: 'subtotal',
                  width: 110,
                  align: 'right',
                  render: rupiah,
                },
              ]}
              dataSource={detail.items}
            />

            <Divider style={{ margin: '12px 0' }} />

            <Descriptions column={1} size="small">
              <Descriptions.Item label="Subtotal">{rupiah(detail.subtotal)}</Descriptions.Item>
              <Descriptions.Item label="Total Diskon">
                −{rupiah(detail.total_diskon)}
              </Descriptions.Item>
              <Descriptions.Item label="Pajak">{rupiah(detail.total_pajak)}</Descriptions.Item>
              <Descriptions.Item label="Grand Total">
                <strong>{rupiah(detail.grand_total)}</strong>
              </Descriptions.Item>
              <Descriptions.Item label="Dibayar">{rupiah(detail.dibayar)}</Descriptions.Item>
              <Descriptions.Item label="Kembalian">{rupiah(detail.kembalian)}</Descriptions.Item>
            </Descriptions>
          </>
        )}
      </Drawer>
    </Card>
  );
}

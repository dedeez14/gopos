// Halaman Produk — pola CRUD lengkap: useQuery (list) + useMutation
// (simpan/nonaktif) + invalidateQueries agar tabel selalu segar.

import { useState } from 'react';
import { PlusOutlined, StarFilled } from '@ant-design/icons';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Button, Card, DatePicker, Form, Input, InputNumber, Modal, Popconfirm,
  Select, Space, Switch, Table, Tag, Typography, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs, { type Dayjs } from 'dayjs';
import {
  buatProduk, daftarKategori, daftarProduk, nonaktifkanProduk, ubahProduk,
  type Produk, type SimpanProdukPayload,
} from '../api/produk';
import { pesanError } from '../api/client';

const rupiah = (n: number) => `Rp${n.toLocaleString('id-ID')}`;

// Bentuk nilai form (RangePicker menghasilkan Dayjs, bukan string).
interface FormProduk {
  nama: string;
  kode?: string;
  barcode?: string;
  tipe: 'PRODUK' | 'JASA';
  satuan?: string;
  harga_beli: number;
  harga_jual: number;
  harga_promo?: number | null;
  periode_promo?: [Dayjs | null, Dayjs | null];
  favorit: boolean;
  kategori_id?: number;
  aktif: boolean;
}

function kePayload(v: FormProduk): SimpanProdukPayload {
  return {
    nama: v.nama,
    kode: v.kode || undefined,
    barcode: v.barcode || undefined,
    tipe: v.tipe,
    satuan: v.satuan || undefined,
    harga_beli: v.harga_beli ?? 0,
    harga_jual: v.harga_jual,
    harga_promo: v.harga_promo ?? null,
    promo_mulai: v.periode_promo?.[0]?.format('YYYY-MM-DD'),
    promo_selesai: v.periode_promo?.[1]?.format('YYYY-MM-DD'),
    favorit: v.favorit,
    kategori_id: v.kategori_id,
    aktif: v.aktif,
  };
}

export default function ProdukPage() {
  const qc = useQueryClient();
  const [form] = Form.useForm<FormProduk>();
  const [cari, setCari] = useState('');
  const [halaman, setHalaman] = useState(1);
  const [modalBuka, setModalBuka] = useState(false);
  const [sedangEdit, setSedangEdit] = useState<Produk | null>(null);
  const perHalaman = 10;

  const { data, isFetching } = useQuery({
    queryKey: ['produk', { q: cari, page: halaman }],
    queryFn: () => daftarProduk({ q: cari, page: halaman, per_page: perHalaman, termasuk_nonaktif: 1 }),
    placeholderData: keepPreviousData,
  });

  const { data: kategori } = useQuery({ queryKey: ['kategori'], queryFn: daftarKategori });

  const segarkan = () => qc.invalidateQueries({ queryKey: ['produk'] });

  const simpan = useMutation({
    mutationFn: (v: FormProduk) =>
      sedangEdit ? ubahProduk(sedangEdit.id, kePayload(v)) : buatProduk(kePayload(v)),
    onSuccess: () => {
      message.success(sedangEdit ? 'Produk diperbarui.' : 'Produk tersimpan.');
      setModalBuka(false);
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const nonaktif = useMutation({
    mutationFn: nonaktifkanProduk,
    onSuccess: () => {
      message.success('Produk dinonaktifkan.');
      segarkan();
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const bukaTambah = () => {
    setSedangEdit(null);
    form.resetFields();
    form.setFieldsValue({ tipe: 'PRODUK', favorit: false, aktif: true, harga_beli: 0 });
    setModalBuka(true);
  };

  const bukaEdit = (p: Produk) => {
    setSedangEdit(p);
    form.setFieldsValue({
      nama: p.nama,
      kode: p.kode,
      barcode: p.barcode ?? undefined,
      tipe: p.tipe,
      satuan: p.satuan,
      harga_beli: p.harga_beli ?? 0,
      harga_jual: p.harga_jual,
      harga_promo: p.harga_promo,
      periode_promo:
        p.promo_mulai || p.promo_selesai
          ? [p.promo_mulai ? dayjs(p.promo_mulai) : null, p.promo_selesai ? dayjs(p.promo_selesai) : null]
          : undefined,
      favorit: p.favorit,
      kategori_id: p.kategori_id ?? undefined,
      aktif: p.aktif,
    });
    setModalBuka(true);
  };

  const kolom: ColumnsType<Produk> = [
    {
      title: 'Nama',
      dataIndex: 'nama',
      render: (nama: string, p) => (
        <Space>
          {p.favorit && <StarFilled style={{ color: '#faad14' }} />}
          <span>{nama}</span>
        </Space>
      ),
    },
    { title: 'Kode', dataIndex: 'kode', width: 130 },
    {
      title: 'Tipe',
      dataIndex: 'tipe',
      width: 90,
      render: (t: Produk['tipe']) => <Tag color={t === 'JASA' ? 'purple' : 'blue'}>{t}</Tag>,
    },
    {
      title: 'Harga',
      width: 180,
      render: (_, p) =>
        p.promo_aktif ? (
          <Space direction="vertical" size={0}>
            <Typography.Text delete type="secondary" className="uang">
              {rupiah(p.harga_jual)}
            </Typography.Text>
            <Tag color="red" className="uang">
              {rupiah(p.harga_efektif)} PROMO
            </Tag>
          </Space>
        ) : (
          <span className="uang">{rupiah(p.harga_jual)}</span>
        ),
    },
    {
      title: 'Stok',
      dataIndex: 'stok',
      width: 90,
      render: (stok: number, p) => (p.kelola_stok ? stok : <Tag>jasa</Tag>),
    },
    {
      title: 'Status',
      dataIndex: 'aktif',
      width: 100,
      render: (aktif: boolean) =>
        aktif ? <Tag color="success">Aktif</Tag> : <Tag>Nonaktif</Tag>,
    },
    {
      title: 'Aksi',
      width: 160,
      render: (_, p) => (
        <Space>
          <Button size="small" onClick={() => bukaEdit(p)}>
            Ubah
          </Button>
          {p.aktif && (
            <Popconfirm
              title="Nonaktifkan produk ini?"
              okText="Ya"
              cancelText="Batal"
              onConfirm={() => nonaktif.mutate(p.id)}
            >
              <Button size="small" danger>
                Nonaktifkan
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Produk & Jasa
        </Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={bukaTambah}>
          Tambah
        </Button>
      </Space>

      <Input.Search
        placeholder="Cari nama/kode/barcode…"
        allowClear
        style={{ maxWidth: 320, marginBottom: 16 }}
        onSearch={(v) => {
          setHalaman(1);
          setCari(v);
        }}
      />

      <Table<Produk>
        rowKey="id"
        columns={kolom}
        dataSource={data?.rows}
        loading={isFetching}
        pagination={{
          current: halaman,
          pageSize: perHalaman,
          total: data?.total ?? 0,
          onChange: setHalaman,
          showTotal: (t) => `${t} item`,
        }}
      />

      <Modal
        title={sedangEdit ? 'Ubah Produk' : 'Tambah Produk'}
        open={modalBuka}
        onCancel={() => setModalBuka(false)}
        onOk={() => form.submit()}
        confirmLoading={simpan.isPending}
        okText="Simpan"
        cancelText="Batal"
        destroyOnClose
      >
        <Form<FormProduk> form={form} layout="vertical" onFinish={(v) => simpan.mutate(v)}>
          <Form.Item name="nama" label="Nama" rules={[{ required: true, message: 'Nama wajib diisi.' }]}>
            <Input />
          </Form.Item>
          <Space.Compact block>
            <Form.Item name="kode" label="Kode (kosong = otomatis)" style={{ width: '50%' }}>
              <Input />
            </Form.Item>
            <Form.Item name="barcode" label="Barcode" style={{ width: '50%' }}>
              <Input />
            </Form.Item>
          </Space.Compact>
          <Space.Compact block>
            <Form.Item name="tipe" label="Tipe" style={{ width: '34%' }} rules={[{ required: true }]}>
              <Select
                options={[
                  { value: 'PRODUK', label: 'Produk' },
                  { value: 'JASA', label: 'Jasa' },
                ]}
              />
            </Form.Item>
            <Form.Item name="satuan" label="Satuan" style={{ width: '33%' }}>
              <Input placeholder="pcs" />
            </Form.Item>
            <Form.Item name="kategori_id" label="Kategori" style={{ width: '33%' }}>
              <Select
                allowClear
                placeholder="Tanpa kategori"
                options={kategori?.map((k) => ({ value: k.id, label: k.nama }))}
              />
            </Form.Item>
          </Space.Compact>
          <Space.Compact block>
            <Form.Item name="harga_beli" label="Harga Beli" style={{ width: '50%' }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item
              name="harga_jual"
              label="Harga Jual"
              style={{ width: '50%' }}
              rules={[{ required: true, message: 'Harga jual wajib diisi.' }]}
            >
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Space.Compact>
          <Space.Compact block>
            <Form.Item name="harga_promo" label="Harga Promo (kosong = tanpa promo)" style={{ width: '45%' }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="periode_promo" label="Periode Promo" style={{ width: '55%' }}>
              <DatePicker.RangePicker allowEmpty={[true, true]} style={{ width: '100%' }} />
            </Form.Item>
          </Space.Compact>
          <Space size="large">
            <Form.Item name="favorit" label="Favorit" valuePropName="checked">
              <Switch />
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

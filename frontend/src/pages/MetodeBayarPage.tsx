// Metode Pembayaran — cara bayar dasar merchant (bank / e-wallet / QRIS
// statis). Lapisan yang SELALU ada; dibaca aplikasi kasir untuk menampilkan
// ke mana pelanggan membayar. CRUD tabel + modal dengan field kondisional
// per jenis (BANK/EWALLET: nomor + atas nama; QRIS: gambar QR + pratinjau).

import { useState } from 'react';
import { BankOutlined, PlusOutlined, QrcodeOutlined, WalletOutlined } from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Alert,
  Button,
  Card,
  Form,
  Image,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { pesanError } from '../api/client';
import {
  daftarMetode,
  hapusMetode,
  simpanMetode,
  type InputMetode,
  type JenisMetode,
  type MetodeBayar,
} from '../api/metodePembayaran';
import { auth } from '../lib/auth';

const META: Record<JenisMetode, { warna: string; ikon: React.ReactNode; label: string }> = {
  BANK: { warna: 'blue', ikon: <BankOutlined />, label: 'Transfer Bank' },
  EWALLET: { warna: 'purple', ikon: <WalletOutlined />, label: 'E-Wallet' },
  QRIS: { warna: 'green', ikon: <QrcodeOutlined />, label: 'QRIS Statis' },
};

export default function MetodeBayarPage() {
  const qc = useQueryClient();
  const [form] = Form.useForm<InputMetode>();
  const [modalBuka, setModalBuka] = useState(false);
  const [sedangEdit, setSedangEdit] = useState<MetodeBayar | null>(null);
  const bisaKelola = ['OWNER', 'MANAGER', 'SUPERADMIN'].includes(auth.role());
  const jenis = Form.useWatch('jenis', form) as JenisMetode | undefined;

  const { data, isFetching } = useQuery({ queryKey: ['metode-bayar'], queryFn: () => daftarMetode(false) });

  const simpan = useMutation({
    mutationFn: (v: InputMetode) => simpanMetode(sedangEdit?.id ?? null, v),
    onSuccess: () => {
      message.success(sedangEdit ? 'Metode diperbarui.' : 'Metode ditambahkan.');
      setModalBuka(false);
      qc.invalidateQueries({ queryKey: ['metode-bayar'] });
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const hapus = useMutation({
    mutationFn: (id: number) => hapusMetode(id),
    onSuccess: () => {
      message.success('Metode dihapus.');
      qc.invalidateQueries({ queryKey: ['metode-bayar'] });
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const bukaTambah = () => {
    setSedangEdit(null);
    form.resetFields();
    form.setFieldsValue({ jenis: 'BANK', aktif: true, urutan: (data?.length ?? 0) + 1 } as InputMetode);
    setModalBuka(true);
  };

  const bukaEdit = (m: MetodeBayar) => {
    setSedangEdit(m);
    form.setFieldsValue(m);
    setModalBuka(true);
  };

  const kolom: ColumnsType<MetodeBayar> = [
    {
      title: 'Jenis',
      dataIndex: 'jenis',
      width: 150,
      render: (j: JenisMetode) => (
        <Tag color={META[j].warna} icon={META[j].ikon}>
          {META[j].label}
        </Tag>
      ),
    },
    { title: 'Nama', dataIndex: 'nama', render: (t: string) => <b>{t}</b> },
    {
      title: 'Detail',
      render: (_, m) =>
        m.jenis === 'QRIS' ? (
          m.gambar_url ? (
            <Image src={m.gambar_url} height={40} style={{ borderRadius: 6 }} />
          ) : (
            <Tag>tanpa gambar</Tag>
          )
        ) : (
          <span>
            <span className="uang">{m.nomor}</span>
            {m.atas_nama ? <span style={{ color: 'var(--muted)' }}> · {m.atas_nama}</span> : null}
          </span>
        ),
    },
    {
      title: 'Status',
      dataIndex: 'aktif',
      width: 110,
      render: (aktif: boolean) =>
        aktif ? <Tag color="success">Aktif</Tag> : <Tag color="default">Nonaktif</Tag>,
    },
    ...(bisaKelola
      ? [
          {
            title: 'Aksi',
            width: 150,
            render: (_: unknown, m: MetodeBayar) => (
              <Space>
                <Button size="small" onClick={() => bukaEdit(m)}>
                  Ubah
                </Button>
                <Popconfirm
                  title="Hapus metode ini?"
                  description="Riwayat transaksi tetap aman."
                  okText="Hapus"
                  cancelText="Batal"
                  onConfirm={() => hapus.mutate(m.id)}
                >
                  <Button size="small" danger>
                    Hapus
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]
      : []),
  ];

  return (
    <Card>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 6 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Metode Pembayaran
        </Typography.Title>
        {bisaKelola && (
          <Button type="primary" icon={<PlusOutlined />} onClick={bukaTambah}>
            Tambah
          </Button>
        )}
      </Space>
      <Typography.Paragraph type="secondary" style={{ fontSize: 13 }}>
        Cara bayar non-tunai milik Anda sendiri — selalu tersedia di kasir, terlepas dari gateway
        pembayaran. Tunai tidak perlu diatur di sini.
      </Typography.Paragraph>

      {!bisaKelola && (
        <Alert
          style={{ marginBottom: 14 }}
          type="info"
          showIcon
          message="Hanya bisa dilihat"
          description="Perubahan metode pembayaran khusus Owner & Manager."
        />
      )}

      <Table<MetodeBayar>
        rowKey="id"
        columns={kolom}
        dataSource={data}
        loading={isFetching}
        pagination={false}
      />

      <Modal
        title={sedangEdit ? `Ubah — ${sedangEdit.nama}` : 'Tambah Metode Pembayaran'}
        open={modalBuka}
        onCancel={() => setModalBuka(false)}
        onOk={() => form.submit()}
        confirmLoading={simpan.isPending}
        okText="Simpan"
        cancelText="Batal"
        destroyOnClose
      >
        <Form<InputMetode> form={form} layout="vertical" onFinish={(v) => simpan.mutate(v)}>
          <Space size="large" style={{ width: '100%' }} align="start">
            <Form.Item name="jenis" label="Jenis" rules={[{ required: true }]} style={{ minWidth: 180 }}>
              <Select
                options={(['BANK', 'EWALLET', 'QRIS'] as JenisMetode[]).map((j) => ({
                  value: j,
                  label: (
                    <span>
                      {META[j].ikon} {META[j].label}
                    </span>
                  ),
                }))}
              />
            </Form.Item>
            <Form.Item name="urutan" label="Urutan" tooltip="Urutan tampil di kasir (kecil dulu).">
              <InputNumber min={0} step={1} style={{ width: 100 }} />
            </Form.Item>
          </Space>

          <Form.Item
            name="nama"
            label="Nama / Label"
            rules={[{ required: true, message: 'Nama wajib diisi.' }]}
          >
            <Input placeholder={jenis === 'QRIS' ? 'mis. QRIS Toko' : 'mis. BCA, OVO'} />
          </Form.Item>

          {jenis === 'QRIS' ? (
            <>
              <Form.Item
                name="gambar_url"
                label="URL Gambar QR"
                rules={[
                  { required: true, message: 'Gambar QR wajib diisi.' },
                  { type: 'url', message: 'Harus berupa tautan (https://…).' },
                ]}
              >
                <Input placeholder="https://…/qr.png" />
              </Form.Item>
              <Form.Item shouldUpdate={(a, b) => a.gambar_url !== b.gambar_url} noStyle>
                {() => {
                  const url = form.getFieldValue('gambar_url');
                  return url ? (
                    <div style={{ marginBottom: 12 }}>
                      <Image src={url} height={120} style={{ borderRadius: 8 }} />
                    </div>
                  ) : null;
                }}
              </Form.Item>
            </>
          ) : (
            <Space size="large" style={{ width: '100%' }} align="start">
              <Form.Item
                name="nomor"
                label={jenis === 'EWALLET' ? 'Nomor HP / Akun' : 'Nomor Rekening'}
                rules={[{ required: true, message: 'Nomor wajib diisi.' }]}
                style={{ minWidth: 200 }}
              >
                <Input placeholder={jenis === 'EWALLET' ? '0812…' : '1234567890'} />
              </Form.Item>
              <Form.Item
                name="atas_nama"
                label="Atas Nama"
                rules={[{ required: true, message: 'Atas nama wajib diisi.' }]}
                style={{ minWidth: 200 }}
              >
                <Input placeholder="Nama pemilik" />
              </Form.Item>
            </Space>
          )}

          <Form.Item name="instruksi" label="Instruksi (opsional)">
            <Input.TextArea rows={2} placeholder="mis. Konfirmasi ke kasir setelah transfer." />
          </Form.Item>
          <Form.Item name="aktif" label="Aktif" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}

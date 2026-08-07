// Pengaturan usaha — profil toko, struk, pajak & perilaku kasir. Semua nilai
// DINAMIS (tak ada yang di-hardcode). Elemen tanda tangan halaman: pratinjau
// STRUK yang hidup — apa pun yang diketik langsung tampak seperti kertas yang
// akan tercetak, karena struk adalah artefak paling nyata dari sebuah kasir.

import { useEffect } from 'react';
import {
  PrinterOutlined,
  SaveOutlined,
  ShopOutlined,
  PercentageOutlined,
} from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Alert,
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Segmented,
  Skeleton,
  Switch,
  Typography,
  message,
} from 'antd';
import { pesanError } from '../api/client';
import { ambilPengaturan, simpanPengaturan, type Pengaturan } from '../api/pengaturan';
import { auth } from '../lib/auth';

const anom = (n: number, simbol: string) => `${simbol}${Math.round(n).toLocaleString('id-ID')}`;

// ── Pratinjau struk (kertas termal) — mencerminkan nilai form secara live ──
function PreviewStruk({ v }: { v: Partial<Pengaturan> }) {
  const simbol = v.mata_uang?.trim() || 'Rp';
  const lebar = v.ukuran_kertas === '58mm' ? 232 : 300;

  const items = [
    { nama: 'Kopi Hitam', qty: 2, harga: 5000 },
    { nama: 'Es Teh Manis', qty: 1, harga: 4000 },
  ];
  const subtotal = items.reduce((s, i) => s + i.qty * i.harga, 0);
  const pajak = v.pajak_aktif ? (subtotal * (v.pajak_persen ?? 0)) / 100 : 0;
  const totalKasar = subtotal + pajak;
  const bulat = v.pembulatan && v.pembulatan > 0 ? v.pembulatan : 0;
  const total = bulat ? Math.round(totalKasar / bulat) * bulat : totalKasar;

  return (
    <div style={{ display: 'flex', justifyContent: 'center' }}>
      <div
        style={{
          width: lebar,
          maxWidth: '100%',
          background: '#fff',
          color: '#1c1c1c',
          fontFamily: '"Space Grotesk", ui-monospace, "SFMono-Regular", monospace',
          fontSize: 12,
          lineHeight: 1.55,
          padding: '20px 18px 24px',
          borderRadius: 4,
          boxShadow: '0 18px 40px -22px rgba(12,59,42,0.55)',
          // Tepi kertas robek atas & bawah — zigzag halus.
          backgroundImage:
            'linear-gradient(-45deg, transparent 8px, #fff 0), linear-gradient(45deg, transparent 8px, #fff 0)',
          backgroundPosition: 'left top, right top',
          backgroundSize: '16px 16px',
          backgroundRepeat: 'repeat-x',
          position: 'relative',
        }}
      >
        {v.struk_header?.trim() && (
          <div style={{ textAlign: 'center', letterSpacing: 1, textTransform: 'uppercase', fontSize: 10, color: '#6b6b6b' }}>
            {v.struk_header}
          </div>
        )}

        {v.tampilkan_logo && (
          <div
            style={{
              width: 44,
              height: 44,
              margin: '4px auto 8px',
              borderRadius: 10,
              display: 'grid',
              placeItems: 'center',
              background: '#0C3B2A',
              color: '#fff',
              fontWeight: 700,
              fontFamily: '"Plus Jakarta Sans", sans-serif',
            }}
          >
            {(v.nama_toko || 'T').trim().charAt(0).toUpperCase()}
          </div>
        )}

        <div style={{ textAlign: 'center', fontWeight: 700, fontSize: 15, marginTop: 2 }}>
          {v.nama_toko?.trim() || 'Nama Toko'}
        </div>
        {v.alamat?.trim() && (
          <div style={{ textAlign: 'center', fontSize: 10.5, color: '#6b6b6b' }}>{v.alamat}</div>
        )}
        {v.telepon?.trim() && (
          <div style={{ textAlign: 'center', fontSize: 10.5, color: '#6b6b6b' }}>{v.telepon}</div>
        )}

        <div style={{ borderTop: '1px dashed #bbb', margin: '10px 0' }} />

        {items.map((i) => (
          <div key={i.nama} style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span>
              {i.nama} <span style={{ color: '#8a8a8a' }}>×{i.qty}</span>
            </span>
            <span>{anom(i.qty * i.harga, simbol)}</span>
          </div>
        ))}

        <div style={{ borderTop: '1px dashed #bbb', margin: '10px 0' }} />

        <Baris label="Subtotal" nilai={anom(subtotal, simbol)} />
        {v.pajak_aktif && (
          <Baris label={`Pajak ${v.pajak_persen ?? 0}%`} nilai={anom(pajak, simbol)} />
        )}
        {bulat > 0 && total !== totalKasar && (
          <Baris label="Pembulatan" nilai={anom(total - totalKasar, simbol)} redup />
        )}
        <div style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 700, fontSize: 14, marginTop: 4 }}>
          <span>TOTAL</span>
          <span>{anom(total, simbol)}</span>
        </div>

        <div style={{ borderTop: '1px dashed #bbb', margin: '10px 0' }} />
        <div style={{ textAlign: 'center', fontSize: 11, whiteSpace: 'pre-wrap' }}>
          {v.struk_footer?.trim() || 'Terima kasih 🙏'}
        </div>
        <div style={{ textAlign: 'center', fontSize: 9, color: '#b3b3b3', marginTop: 8, letterSpacing: 1 }}>
          — Tuléh POS —
        </div>
      </div>
    </div>
  );
}

function Baris({ label, nilai, redup }: { label: string; nilai: string; redup?: boolean }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', color: redup ? '#9a9a9a' : undefined }}>
      <span>{label}</span>
      <span>{nilai}</span>
    </div>
  );
}

export default function PengaturanPage() {
  const qc = useQueryClient();
  const [form] = Form.useForm<Pengaturan>();
  const bisaKelola = ['OWNER', 'MANAGER', 'SUPERADMIN'].includes(auth.role());

  const { data, isLoading } = useQuery({ queryKey: ['pengaturan'], queryFn: ambilPengaturan });

  useEffect(() => {
    if (data) form.setFieldsValue(data);
  }, [data, form]);

  // Nilai live untuk pratinjau (fallback ke data awal saat form belum tersentuh).
  const v = { ...data, ...Form.useWatch([], form) } as Partial<Pengaturan>;

  const simpan = useMutation({
    mutationFn: (p: Pengaturan) => simpanPengaturan(p),
    onSuccess: (baru) => {
      message.success('Pengaturan disimpan.');
      qc.setQueryData(['pengaturan'], baru);
    },
    onError: (e) => message.error(pesanError(e)),
  });

  if (isLoading) {
    return (
      <Card>
        <Skeleton active paragraph={{ rows: 8 }} />
      </Card>
    );
  }

  return (
    <>
      <Typography.Title level={4} style={{ marginBottom: 2 }}>
        Pengaturan Usaha
      </Typography.Title>
      <Typography.Text type="secondary">
        Semua nilai di sini dinamis — mengubahnya langsung memengaruhi struk & aplikasi kasir.
      </Typography.Text>

      {!bisaKelola && (
        <Alert
          style={{ margin: '14px 0' }}
          type="info"
          showIcon
          message="Hanya bisa dilihat"
          description="Perubahan pengaturan usaha khusus Owner & Manager."
        />
      )}

      <Row gutter={[20, 20]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={14}>
          <Form
            form={form}
            layout="vertical"
            disabled={!bisaKelola}
            initialValues={data}
            onFinish={(vals) => simpan.mutate(vals)}
          >
            <Card
              size="small"
              title={
                <span>
                  <ShopOutlined /> Profil Toko
                </span>
              }
              style={{ marginBottom: 16 }}
            >
              <Form.Item
                name="nama_toko"
                label="Nama toko"
                rules={[{ required: true, message: 'Nama toko wajib diisi.' }]}
              >
                <Input placeholder="mis. Warung Bahagia" />
              </Form.Item>
              <Form.Item name="alamat" label="Alamat">
                <Input.TextArea rows={2} placeholder="Alamat toko (tampil di struk)" />
              </Form.Item>
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item name="telepon" label="Telepon">
                    <Input placeholder="0812…" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name="email" label="Email" rules={[{ type: 'email', message: 'Email tidak valid.' }]}>
                    <Input placeholder="toko@email.com" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={16}>
                  <Form.Item
                    name="logo_url"
                    label="URL Logo"
                    rules={[{ type: 'url', message: 'Harus berupa tautan (https://…).' }]}
                  >
                    <Input placeholder="https://…/logo.png" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="mata_uang" label="Simbol mata uang" rules={[{ required: true }]}>
                    <Input placeholder="Rp" />
                  </Form.Item>
                </Col>
              </Row>
            </Card>

            <Card
              size="small"
              title={
                <span>
                  <PrinterOutlined /> Struk
                </span>
              }
              style={{ marginBottom: 16 }}
            >
              <Form.Item name="struk_header" label="Teks header (opsional)">
                <Input placeholder="mis. Struk Pembelian" />
              </Form.Item>
              <Form.Item name="struk_footer" label="Teks footer">
                <Input.TextArea rows={2} placeholder="Terima kasih telah berbelanja 🙏" />
              </Form.Item>
              <Row gutter={12} align="middle">
                <Col span={14}>
                  <Form.Item name="ukuran_kertas" label="Ukuran kertas">
                    <Segmented
                      options={[
                        { label: '58 mm', value: '58mm' },
                        { label: '80 mm', value: '80mm' },
                      ]}
                    />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item name="tampilkan_logo" label="Tampilkan logo" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>
            </Card>

            <Card
              size="small"
              title={
                <span>
                  <PercentageOutlined /> Pajak & Perilaku
                </span>
              }
              style={{ marginBottom: 16 }}
            >
              <Row gutter={12} align="middle">
                <Col span={8}>
                  <Form.Item name="pajak_aktif" label="Kenakan pajak" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="pajak_persen" label="Persen pajak (%)">
                    <InputNumber min={0} max={100} step={1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="pembulatan"
                    label="Pembulatan"
                    tooltip="0 = tanpa. mis. 100 = bulatkan total ke kelipatan 100."
                  >
                    <InputNumber min={0} max={100000} step={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </Card>

            {bisaKelola && (
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={simpan.isPending} size="large">
                Simpan Pengaturan
              </Button>
            )}
          </Form>
        </Col>

        <Col xs={24} lg={10}>
          <div style={{ position: 'sticky', top: 90 }}>
            <Typography.Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 10 }}>
              PRATINJAU STRUK · {v.ukuran_kertas === '58mm' ? '58 mm' : '80 mm'}
            </Typography.Text>
            <PreviewStruk v={v} />
          </div>
        </Col>
      </Row>
    </>
  );
}

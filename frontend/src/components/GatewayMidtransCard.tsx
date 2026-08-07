// Gateway Midtrans (lapisan 2 pembayaran) — kartu konfigurasi server/client key
// MILIK MERCHANT. Muncul hanya bila platform mengaktifkan modul; server key
// dikirim server dalam bentuk BERTOPENG (hint 4 digit), input dikosongkan =
// pertahankan yang tersimpan.

import { useEffect } from 'react';
import { CheckCircleOutlined, LockOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  Segmented,
  Skeleton,
  Space,
  Switch,
  Tag,
  Typography,
  message,
} from 'antd';
import { pesanError } from '../api/client';
import { ambilGateway, simpanGateway, type KonfigMidtrans, type SimpanGateway } from '../api/gateway';

type FormGateway = Omit<SimpanGateway, 'server_key'> & { server_key?: string };

export default function GatewayMidtransCard() {
  const qc = useQueryClient();
  const [form] = Form.useForm<FormGateway>();
  const { data, isLoading } = useQuery({ queryKey: ['gateway-midtrans'], queryFn: ambilGateway });

  useEffect(() => {
    if (data) {
      form.setFieldsValue({
        aktif: data.aktif,
        env: data.env,
        merchant_id: data.merchant_id,
        client_key: data.client_key,
        server_key: '',
      });
    }
  }, [data, form]);

  const simpan = useMutation({
    mutationFn: (v: FormGateway) => simpanGateway({ ...v, server_key: v.server_key ?? '' }),
    onSuccess: (baru) => {
      message.success('Konfigurasi Midtrans disimpan.');
      qc.setQueryData(['gateway-midtrans'], baru);
      form.setFieldValue('server_key', '');
    },
    onError: (e) => message.error(pesanError(e)),
  });

  const judul = (
    <Space>
      <ThunderboltOutlined style={{ color: 'var(--turmeric)' }} />
      <span>Gateway Midtrans</span>
      <Tag color="default" style={{ fontWeight: 400 }}>
        QRIS dinamis
      </Tag>
    </Space>
  );

  if (isLoading) {
    return (
      <Card title={judul} style={{ marginBottom: 18 }}>
        <Skeleton active paragraph={{ rows: 3 }} />
      </Card>
    );
  }

  const k = data as KonfigMidtrans;

  // Platform belum mengaktifkan modul → tak ada form, cukup pemberitahuan.
  if (!k.platform_aktif) {
    return (
      <Card title={judul} style={{ marginBottom: 18 }}>
        <Alert
          type="info"
          showIcon
          message="Modul Midtrans belum diaktifkan"
          description="Pembayaran QRIS dinamis via Midtrans belum tersedia untuk usaha Anda. Hubungi pengelola platform untuk mengaktifkannya. Sementara itu, gunakan metode dasar di bawah."
        />
      </Card>
    );
  }

  return (
    <Card
      title={judul}
      style={{ marginBottom: 18 }}
      extra={
        k.siap ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            Siap dipakai
          </Tag>
        ) : (
          <Tag color="warning">Belum siap</Tag>
        )
      }
    >
      <Typography.Paragraph type="secondary" style={{ fontSize: 13, marginTop: -4 }}>
        <LockOutlined /> Server key Anda disimpan terenkripsi dan tidak pernah ditampilkan kembali.
        Client key boleh dilihat (memang dipakai aplikasi).
      </Typography.Paragraph>

      <Form<FormGateway> form={form} layout="vertical" onFinish={(v) => simpan.mutate(v)}>
        <Space size="large" style={{ width: '100%' }} align="start" wrap>
          <Form.Item name="aktif" label="Aktifkan Midtrans" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="env" label="Lingkungan">
            <Segmented
              options={[
                { label: 'Sandbox', value: 'sandbox' },
                { label: 'Production', value: 'production' },
              ]}
            />
          </Form.Item>
        </Space>

        <Form.Item name="merchant_id" label="Merchant ID">
          <Input placeholder="mis. G123456789" />
        </Form.Item>
        <Form.Item name="client_key" label="Client Key (publik)">
          <Input placeholder="Mid-client-…" />
        </Form.Item>
        <Form.Item
          name="server_key"
          label="Server Key (rahasia)"
          extra={
            k.server_key_terisi
              ? `Tersimpan (${k.server_key_hint}). Kosongkan bila tidak diganti.`
              : 'Belum ada — tempel server key Midtrans Anda.'
          }
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder={k.server_key_terisi ? '••••  (biarkan kosong jika tidak diganti)' : 'SB-Mid-server-…'}
            autoComplete="new-password"
          />
        </Form.Item>

        <Button type="primary" htmlType="submit" loading={simpan.isPending}>
          Simpan Konfigurasi
        </Button>
      </Form>
    </Card>
  );
}

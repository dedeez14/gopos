// Masuk — dua panel: kiri identitas (pine + anyaman), kanan formulir.
// Panel kiri menyusut jadi kop ramping di layar kecil (lihat index.css).

import { LockOutlined, MailOutlined, RiseOutlined } from '@ant-design/icons';
import { useMutation } from '@tanstack/react-query';
import { Alert, Button, Form, Input, Typography } from 'antd';
import { useNavigate } from 'react-router-dom';
import Brand from '../components/Brand';
import { login } from '../api/users';
import { pesanError } from '../api/client';

interface FormMasuk {
  email: string;
  password: string;
}

export default function LoginPage() {
  const navigate = useNavigate();

  const mutasi = useMutation({
    mutationFn: (v: FormMasuk) => login(v.email, v.password),
    onSuccess: () => navigate('/', { replace: true }),
  });

  return (
    <div className="login-wrap">
      {/* ── Panel identitas ── */}
      <div className="login-brand">
        <Brand flush />

        <div>
          <h1 className="login-tag">
            Kasir pintar untuk <span>usaha Anda.</span>
          </h1>
          <p className="login-sub">
            Kelola produk, transaksi, stok, dan laporan dalam satu tempat yang rapi — dari mana pun,
            kapan pun.
          </p>

          {/* Ilustrasi produk (angka contoh) */}
          <div className="peek" style={{ marginTop: 32 }}>
            <div className="peek-label">Omzet hari ini</div>
            <div className="peek-value">Rp2.480.000</div>
            <div className="peek-delta">
              <RiseOutlined /> 12% dari kemarin
            </div>
            <div className="peek-bars">
              {[40, 62, 48, 80, 55, 92, 70].map((h, i) => (
                <i key={i} style={{ height: `${h}%` }} />
              ))}
            </div>
          </div>
        </div>

        <div style={{ color: 'rgba(255,255,255,0.45)', fontSize: 12.5 }}>
          © {new Date().getFullYear()} Tuléh POS
        </div>
      </div>

      {/* ── Panel formulir ── */}
      <div className="login-form">
        <div className="login-form-inner">
          <Typography.Title level={3} style={{ marginBottom: 4 }}>
            Masuk
          </Typography.Title>
          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 24 }}>
            Gunakan email dan kata sandi akun Anda.
          </Typography.Text>

          {mutasi.isError && (
            <Alert
              type="error"
              showIcon
              style={{ marginBottom: 18 }}
              message={pesanError(mutasi.error)}
            />
          )}

          <Form<FormMasuk> layout="vertical" onFinish={(v) => mutasi.mutate(v)} requiredMark={false}>
            <Form.Item
              name="email"
              label="Email"
              rules={[
                { required: true, message: 'Email wajib diisi.' },
                { type: 'email', message: 'Format email tidak sah.' },
              ]}
            >
              <Input prefix={<MailOutlined />} placeholder="nama@usaha.id" autoFocus size="large" />
            </Form.Item>

            <Form.Item
              name="password"
              label="Kata Sandi"
              rules={[{ required: true, message: 'Kata sandi wajib diisi.' }]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="••••••••" size="large" />
            </Form.Item>

            <Button
              type="primary"
              htmlType="submit"
              block
              size="large"
              loading={mutasi.isPending}
              style={{ marginTop: 4 }}
            >
              Masuk
            </Button>
          </Form>

          <p className="login-legal">Butuh akun? Hubungi pengelola usaha Anda.</p>
        </div>
      </div>
    </div>
  );
}

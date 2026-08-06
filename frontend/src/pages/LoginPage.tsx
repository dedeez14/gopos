// Halaman masuk: Form AntD → useMutation(login) → simpan token → dashboard.

import { LockOutlined, MailOutlined } from '@ant-design/icons';
import { useMutation } from '@tanstack/react-query';
import { Alert, Button, Card, Form, Input, Typography } from 'antd';
import { useNavigate } from 'react-router-dom';
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
    <div
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        background: '#f0f2f5',
      }}
    >
      <Card style={{ width: 380 }}>
        <Typography.Title level={3} style={{ textAlign: 'center' }}>
          Tuléh Admin
        </Typography.Title>

        {mutasi.isError && (
          <Alert
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
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
            <Input prefix={<MailOutlined />} placeholder="admin@tuleh.local" autoFocus />
          </Form.Item>

          <Form.Item
            name="password"
            label="Kata Sandi"
            rules={[{ required: true, message: 'Kata sandi wajib diisi.' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="••••••••" />
          </Form.Item>

          <Button type="primary" htmlType="submit" block loading={mutasi.isPending}>
            Masuk
          </Button>
        </Form>
      </Card>
    </div>
  );
}

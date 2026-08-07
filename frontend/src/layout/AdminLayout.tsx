// Kerangka panel: Sider pine (menu) + Header (konteks + peran + keluar) +
// Content bertekstur. Di layar kecil, Sider berubah jadi Drawer (hamburger).

import { useState } from 'react';
import {
  ApartmentOutlined,
  BarChartOutlined,
  ContactsOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  LogoutOutlined,
  MenuOutlined,
  SettingOutlined,
  ShoppingOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { Button, Drawer, Grid, Layout, Menu, Typography } from 'antd';
import type { ItemType } from 'antd/es/menu/interface';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import Brand from '../components/Brand';
import { logout } from '../api/users';
import { auth } from '../lib/auth';

const { Sider, Header, Content } = Layout;

// Label per rute — dipakai untuk judul header (konteks halaman).
const JUDUL: Record<string, string> = {
  '/': 'Dashboard',
  '/usaha': 'Usaha / Merchant',
  '/produk': 'Produk & Jasa',
  '/transaksi': 'Transaksi',
  '/inventory': 'Inventory',
  '/laporan': 'Laporan',
  '/pelanggan': 'Pelanggan',
  '/users': 'Pengguna',
  '/pengaturan': 'Pengaturan Usaha',
};

export default function AdminLayout() {
  const navigate = useNavigate();
  const lokasi = useLocation();
  const layar = Grid.useBreakpoint();
  const mobile = !layar.lg;
  const [drawer, setDrawer] = useState(false);

  const profil = auth.profil();

  const keluar = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  const menu: ItemType[] = [
    { key: '/', icon: <DashboardOutlined />, label: 'Dashboard' },
    ...(auth.role() === 'SUPERADMIN'
      ? [{ key: '/usaha', icon: <ApartmentOutlined />, label: 'Usaha (Merchant)' }]
      : []),
    { key: '/produk', icon: <ShoppingOutlined />, label: 'Produk & Jasa' },
    { key: '/transaksi', icon: <FileTextOutlined />, label: 'Transaksi' },
    { key: '/inventory', icon: <DatabaseOutlined />, label: 'Inventory' },
    { key: '/laporan', icon: <BarChartOutlined />, label: 'Laporan' },
    { key: '/pelanggan', icon: <ContactsOutlined />, label: 'Pelanggan' },
    { key: '/users', icon: <TeamOutlined />, label: 'Pengguna' },
    ...(['OWNER', 'MANAGER', 'SUPERADMIN'].includes(auth.role())
      ? [{ key: '/pengaturan', icon: <SettingOutlined />, label: 'Pengaturan' }]
      : []),
  ];

  const isiNav = (
    <>
      <Brand />
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[lokasi.pathname]}
        onClick={({ key }) => {
          navigate(key);
          setDrawer(false);
        }}
        items={menu}
        style={{ marginTop: 4 }}
      />
      <div style={{ flex: 1 }} />
      <div className="sider-foot">Tuléh POS · v0.1</div>
    </>
  );

  const inisial = (profil.nama || 'U').trim().charAt(0).toUpperCase();

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {!mobile && (
        <Sider width={244} style={{ position: 'sticky', top: 0, height: '100vh' }}>
          <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>{isiNav}</div>
        </Sider>
      )}

      {mobile && (
        <Drawer
          placement="left"
          width={244}
          open={drawer}
          onClose={() => setDrawer(false)}
          closable={false}
          styles={{ body: { padding: 0, display: 'flex', flexDirection: 'column' } }}
        >
          {isiNav}
        </Drawer>
      )}

      <Layout>
        <Header
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 12,
            position: 'sticky',
            top: 0,
            zIndex: 10,
            borderBottom: '1px solid var(--line)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
            {mobile && (
              <Button type="text" icon={<MenuOutlined />} onClick={() => setDrawer(true)} />
            )}
            <Typography.Title level={5} style={{ margin: 0, whiteSpace: 'nowrap' }}>
              {JUDUL[lokasi.pathname] ?? 'Panel'}
            </Typography.Title>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div className="role-chip" style={mobile ? { padding: 3 } : undefined}>
              <span className="role-avatar">{inisial}</span>
              {!mobile && (
                <span style={{ lineHeight: 1.1 }}>
                  <span style={{ display: 'block', fontWeight: 600, fontSize: 13 }}>
                    {profil.nama || 'Pengguna'}
                  </span>
                  <span style={{ fontSize: 11, color: 'var(--muted)' }}>{profil.role || '—'}</span>
                </span>
              )}
            </div>
            <Button icon={<LogoutOutlined />} onClick={keluar}>
              {!mobile && 'Keluar'}
            </Button>
          </div>
        </Header>

        <Content className="app-canvas" style={{ padding: mobile ? 16 : 28 }}>
          <div style={{ maxWidth: 1200, margin: '0 auto' }}>
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
}

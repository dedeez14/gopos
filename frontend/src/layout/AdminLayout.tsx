// Kerangka halaman admin: Sider (menu) + Header (identitas & keluar) +
// Content. Semua halaman privat dirender lewat <Outlet/> di dalam layout ini.

import { BarChartOutlined, ContactsOutlined, DashboardOutlined, DatabaseOutlined, FileTextOutlined, LogoutOutlined, ShoppingOutlined, TeamOutlined } from '@ant-design/icons';
import { Button, Layout, Menu, Typography, theme } from 'antd';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { logout } from '../api/users';

const { Sider, Header, Content } = Layout;

export default function AdminLayout() {
  const navigate = useNavigate();
  const lokasi = useLocation();
  const { token } = theme.useToken();

  const keluar = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth={0}>
        <div style={{ color: '#fff', fontWeight: 700, fontSize: 18, padding: '16px 24px' }}>
          Tuléh Admin
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[lokasi.pathname]}
          onClick={({ key }) => navigate(key)}
          items={[
            { key: '/', icon: <DashboardOutlined />, label: 'Dashboard' },
            { key: '/produk', icon: <ShoppingOutlined />, label: 'Produk & Jasa' },
            { key: '/transaksi', icon: <FileTextOutlined />, label: 'Transaksi' },
            { key: '/inventory', icon: <DatabaseOutlined />, label: 'Inventory' },
            { key: '/laporan', icon: <BarChartOutlined />, label: 'Laporan' },
            { key: '/pelanggan', icon: <ContactsOutlined />, label: 'Pelanggan' },
            { key: '/users', icon: <TeamOutlined />, label: 'Pengguna' },
          ]}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            background: token.colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            paddingInline: 24,
          }}
        >
          <Typography.Text strong>Panel Admin Tuléh POS</Typography.Text>
          <Button icon={<LogoutOutlined />} onClick={keluar}>
            Keluar
          </Button>
        </Header>

        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}

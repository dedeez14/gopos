// Dashboard — hero omzet bulan ini (uang = pahlawan) + kartu ringkas.
// Semua angka dari server (React Query); tahan bila laporan tak diizinkan.

import {
  ContactsOutlined,
  RiseOutlined,
  ShoppingOutlined,
  TeamOutlined,
  WalletOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { Card, Col, Row, Skeleton, Statistic, Typography } from 'antd';
import type { ReactNode } from 'react';
import { keuangan } from '../api/laporan';
import { daftarPelanggan } from '../api/laporan';
import { daftarProduk } from '../api/produk';
import { daftarUsers } from '../api/users';
import { auth } from '../lib/auth';
import TrenOmzet from '../components/TrenOmzet';

const rupiah = (n: number) => `Rp${n.toLocaleString('id-ID')}`;

function StatCard({
  judul,
  nilai,
  ikon,
  warna,
  memuat,
}: {
  judul: string;
  nilai: ReactNode;
  ikon: ReactNode;
  warna: string;
  memuat?: boolean;
}) {
  return (
    <Card styles={{ body: { padding: 20 } }}>
      {memuat ? (
        <Skeleton active paragraph={false} />
      ) : (
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <span
            style={{
              width: 46,
              height: 46,
              borderRadius: 13,
              display: 'grid',
              placeItems: 'center',
              background: `${warna}1A`,
              color: warna,
              fontSize: 20,
              flex: 'none',
            }}
          >
            {ikon}
          </span>
          <Statistic title={judul} value={nilai as number} />
        </div>
      )}
    </Card>
  );
}

export default function DashboardPage() {
  const profil = auth.profil();
  const bisaLaporan = ['OWNER', 'MANAGER', 'SUPERADMIN'].includes(profil.role);

  const keu = useQuery({
    queryKey: ['dash-keuangan'],
    queryFn: () => keuangan(),
    enabled: bisaLaporan,
    retry: false,
  });
  const produk = useQuery({ queryKey: ['dash-produk'], queryFn: () => daftarProduk({ per_page: 1 }) });
  const pelanggan = useQuery({ queryKey: ['dash-pelanggan'], queryFn: () => daftarPelanggan('', 1) });
  const users = useQuery({ queryKey: ['dash-users'], queryFn: () => daftarUsers({ per_page: 1 }) });

  const jam = new Date().getHours();
  const sapa = jam < 11 ? 'Selamat pagi' : jam < 15 ? 'Selamat siang' : jam < 19 ? 'Selamat sore' : 'Selamat malam';

  return (
    <>
      <Typography.Title level={4} style={{ marginBottom: 2 }}>
        {sapa}
        {profil.nama ? `, ${profil.nama.split(' ')[0]}` : ''} 👋
      </Typography.Title>
      <Typography.Text type="secondary">Ringkasan usaha Anda hari ini.</Typography.Text>

      <Row gutter={[18, 18]} style={{ marginTop: 18 }}>
        {/* ── Hero omzet ── */}
        {bisaLaporan && (
          <Col xs={24} lg={14}>
            <Card
              styles={{ body: { padding: 26 } }}
              style={{
                border: 'none',
                color: '#fff',
                background:
                  'radial-gradient(120% 120% at 90% 10%, #17915f 0%, #0c3b2a 62%)',
              }}
            >
              {keu.isLoading ? (
                <Skeleton active paragraph={{ rows: 2 }} />
              ) : (
                <>
                  <div
                    style={{
                      textTransform: 'uppercase',
                      letterSpacing: '0.06em',
                      fontSize: 12.5,
                      color: 'rgba(255,255,255,0.72)',
                    }}
                  >
                    Omzet bulan ini · {keu.data?.bulan ?? ''}
                  </div>
                  <div
                    className="uang"
                    style={{ fontSize: 42, fontWeight: 700, margin: '6px 0 10px' }}
                  >
                    {rupiah(keu.data?.omzet ?? 0)}
                  </div>
                  <div style={{ display: 'flex', gap: 26, flexWrap: 'wrap' }}>
                    <div>
                      <div style={{ color: 'rgba(255,255,255,0.7)', fontSize: 12.5 }}>
                        Laba bersih
                      </div>
                      <div className="uang" style={{ fontSize: 18, fontWeight: 600, color: 'var(--turmeric-soft)' }}>
                        {rupiah(keu.data?.laba ?? 0)}
                      </div>
                    </div>
                    <div>
                      <div style={{ color: 'rgba(255,255,255,0.7)', fontSize: 12.5 }}>Transaksi</div>
                      <div className="uang" style={{ fontSize: 18, fontWeight: 600 }}>
                        {(keu.data?.jumlah_trx ?? 0).toLocaleString('id-ID')}
                      </div>
                    </div>
                    <div>
                      <div style={{ color: 'rgba(255,255,255,0.7)', fontSize: 12.5 }}>Pengeluaran</div>
                      <div className="uang" style={{ fontSize: 18, fontWeight: 600 }}>
                        {rupiah(keu.data?.pengeluaran ?? 0)}
                      </div>
                    </div>
                  </div>
                </>
              )}
            </Card>
          </Col>
        )}

        {/* ── Kartu ringkas ── */}
        <Col xs={24} sm={12} lg={bisaLaporan ? 10 : 8}>
          <StatCard
            judul="Produk & Jasa"
            nilai={produk.data?.total ?? 0}
            ikon={<ShoppingOutlined />}
            warna="#147A54"
            memuat={produk.isLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={bisaLaporan ? 7 : 8}>
          <StatCard
            judul="Pelanggan"
            nilai={pelanggan.data?.total ?? 0}
            ikon={<ContactsOutlined />}
            warna="#E0952B"
            memuat={pelanggan.isLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={bisaLaporan ? 7 : 8}>
          <StatCard
            judul="Pengguna"
            nilai={users.data?.total ?? 0}
            ikon={<TeamOutlined />}
            warna="#3C6E8F"
            memuat={users.isLoading}
          />
        </Col>

        {!bisaLaporan && (
          <Col xs={24}>
            <Card styles={{ body: { padding: 20 } }}>
              <Typography.Text type="secondary">
                <WalletOutlined /> Ringkasan keuangan hanya tersedia untuk Owner & Manager.
              </Typography.Text>
            </Card>
          </Col>
        )}
      </Row>

      {bisaLaporan && (
        <div style={{ marginTop: 18 }}>
          <TrenOmzet />
        </div>
      )}

      {bisaLaporan && (
        <Typography.Paragraph type="secondary" style={{ marginTop: 16, fontSize: 13 }}>
          <RiseOutlined /> Lihat rincian di menu <b>Laporan</b> — omzet harian, produk terlaris, dan
          pengeluaran.
        </Typography.Paragraph>
      )}
    </>
  );
}

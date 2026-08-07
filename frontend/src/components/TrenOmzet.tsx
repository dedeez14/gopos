// Grafik tren omzet 14 hari — area SVG dependency-free, responsif (viewBox),
// dengan hover tooltip. Motif & palet menyatu dgn sistem "kertas grafik uang":
// area jade→transparan, garis jade, penanda puncak turmeric.

import { useMemo, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, Skeleton, Typography } from 'antd';
import dayjs from 'dayjs';
import { penjualanHarian } from '../api/laporan';

const rupiah = (n: number) => `Rp${n.toLocaleString('id-ID')}`;
const rupiahRingkas = (n: number) =>
  n >= 1_000_000 ? `${(n / 1_000_000).toFixed(1)}jt` : n >= 1000 ? `${Math.round(n / 1000)}rb` : `${n}`;

const W = 720;
const H = 240;
const PAD = { t: 18, r: 14, b: 30, l: 14 };
const HARI = 14;

export default function TrenOmzet() {
  const wrapRef = useRef<HTMLDivElement>(null);
  const [aktif, setAktif] = useState<number | null>(null);

  const dari = dayjs().subtract(HARI - 1, 'day').format('YYYY-MM-DD');
  const sampai = dayjs().format('YYYY-MM-DD');

  const { data, isLoading } = useQuery({
    queryKey: ['tren-omzet', dari, sampai],
    queryFn: () => penjualanHarian(dari, sampai),
  });

  // Isi 14 hari penuh — endpoint hanya mengembalikan hari yang ada penjualan.
  const titik = useMemo(() => {
    const peta = new Map((data ?? []).map((r) => [r.tanggal, r]));
    return Array.from({ length: HARI }, (_, i) => {
      const tgl = dayjs()
        .subtract(HARI - 1 - i, 'day')
        .format('YYYY-MM-DD');
      const r = peta.get(tgl);
      return { tgl, omzet: r?.omzet ?? 0, trx: r?.jumlah_trx ?? 0 };
    });
  }, [data]);

  const maks = Math.max(1, ...titik.map((t) => t.omzet));
  const puncak = titik.reduce((a, t, i) => (t.omzet > titik[a].omzet ? i : a), 0);

  const plotW = W - PAD.l - PAD.r;
  const plotH = H - PAD.t - PAD.b;
  const x = (i: number) => PAD.l + (titik.length === 1 ? plotW / 2 : (plotW * i) / (titik.length - 1));
  const y = (v: number) => PAD.t + plotH - (plotH * v) / maks;

  const garis = titik.map((t, i) => `${i === 0 ? 'M' : 'L'}${x(i).toFixed(1)},${y(t.omzet).toFixed(1)}`).join(' ');
  const area = `${garis} L${x(titik.length - 1).toFixed(1)},${(PAD.t + plotH).toFixed(1)} L${x(0).toFixed(1)},${(
    PAD.t + plotH
  ).toFixed(1)} Z`;

  const onMove = (e: React.MouseEvent) => {
    const rect = wrapRef.current?.getBoundingClientRect();
    if (!rect) return;
    const ratio = (e.clientX - rect.left) / rect.width;
    setAktif(Math.max(0, Math.min(titik.length - 1, Math.round(ratio * (titik.length - 1)))));
  };

  const t = aktif != null ? titik[aktif] : null;

  return (
    <Card
      title="Tren omzet · 14 hari terakhir"
      styles={{ body: { padding: '8px 12px 4px' } }}
    >
      {isLoading ? (
        <Skeleton active paragraph={{ rows: 3 }} />
      ) : (
        <div
          ref={wrapRef}
          style={{ position: 'relative' }}
          onMouseMove={onMove}
          onMouseLeave={() => setAktif(null)}
        >
          <svg viewBox={`0 0 ${W} ${H}`} width="100%" role="img" aria-label="Grafik omzet harian" style={{ display: 'block' }}>
            <defs>
              <linearGradient id="areaOmzet" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#147A54" stopOpacity="0.28" />
                <stop offset="100%" stopColor="#147A54" stopOpacity="0.02" />
              </linearGradient>
            </defs>

            {/* Gridline horizontal (0, ½, maks) */}
            {[0, 0.5, 1].map((g) => (
              <g key={g}>
                <line
                  x1={PAD.l}
                  x2={W - PAD.r}
                  y1={y(maks * g)}
                  y2={y(maks * g)}
                  stroke="#E4EAE4"
                  strokeDasharray={g === 0 ? '0' : '3 4'}
                />
                <text x={W - PAD.r} y={y(maks * g) - 4} textAnchor="end" fontSize="10.5" fill="#9AA89F">
                  {rupiahRingkas(maks * g)}
                </text>
              </g>
            ))}

            <path d={area} fill="url(#areaOmzet)" />
            <path d={garis} fill="none" stroke="#147A54" strokeWidth="2.4" strokeLinejoin="round" strokeLinecap="round" />

            {/* Garis pemandu + dot aktif saat hover */}
            {t && (
              <line x1={x(aktif!)} x2={x(aktif!)} y1={PAD.t} y2={PAD.t + plotH} stroke="#0C3B2A" strokeOpacity="0.18" />
            )}

            {/* Dot: puncak turmeric, aktif pine, sisanya samar */}
            {titik.map((p, i) => (
              <circle
                key={i}
                cx={x(i)}
                cy={y(p.omzet)}
                r={i === aktif ? 5 : i === puncak ? 4 : 2.5}
                fill={i === aktif ? '#0C3B2A' : i === puncak ? '#E0952B' : '#147A54'}
                stroke="#fff"
                strokeWidth={i === aktif || i === puncak ? 2 : 0}
              />
            ))}

            {/* Label tanggal — tiap 3 hari + hari ini */}
            {titik.map((p, i) =>
              i % 3 === 0 || i === titik.length - 1 ? (
                <text key={i} x={x(i)} y={H - 10} textAnchor="middle" fontSize="10.5" fill="#9AA89F">
                  {dayjs(p.tgl).format('D/M')}
                </text>
              ) : null,
            )}
          </svg>

          {/* Tooltip mengambang */}
          {t && (
            <div
              style={{
                position: 'absolute',
                top: 6,
                left: `${(aktif! / (titik.length - 1)) * 100}%`,
                transform: `translateX(${aktif! > titik.length / 2 ? '-108%' : '8%'})`,
                background: '#0C3B2A',
                color: '#fff',
                borderRadius: 10,
                padding: '8px 12px',
                pointerEvents: 'none',
                boxShadow: '0 8px 24px -10px rgba(12,59,42,0.6)',
                whiteSpace: 'nowrap',
              }}
            >
              <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.7)' }}>
                {dayjs(t.tgl).format('dddd, D MMM')}
              </div>
              <div className="uang" style={{ fontSize: 16, fontWeight: 700 }}>
                {rupiah(t.omzet)}
              </div>
              <div style={{ fontSize: 11, color: 'var(--turmeric-soft)' }}>{t.trx} transaksi</div>
            </div>
          )}

          <Typography.Text type="secondary" style={{ fontSize: 12, display: 'block', padding: '2px 2px 8px' }}>
            Arahkan kursor untuk melihat rincian per hari.
          </Typography.Text>
        </div>
      )}
    </Card>
  );
}

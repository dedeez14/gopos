// API laporan keuangan, pengeluaran, pelanggan.

import { api, bukaAmplop, type Amplop } from './client';

export interface RingkasanKeuangan {
  bulan: string;
  omzet: number;
  per_tipe: Record<string, number>;
  jumlah_trx: number;
  pengeluaran: number;
  laba: number;
}

export interface PenjualanHarian {
  tanggal: string;
  jumlah_trx: number;
  omzet: number;
  diskon: number;
  pajak: number;
}

export interface ProdukTerlaris {
  produk_id: number;
  nama: string;
  terjual: number;
  omzet: number;
}

export interface Pengeluaran {
  id: number;
  tanggal: string;
  keterangan: string;
  nominal: number;
  pencatat: string;
}

export interface Pelanggan {
  id: number;
  nama: string;
  telepon: string | null;
  wa_link: string | null;
  email: string | null;
  catatan: string;
  aktif: boolean;
}

export async function keuangan(bulan?: string): Promise<RingkasanKeuangan> {
  const res = await api.get<Amplop<RingkasanKeuangan>>('/laporan/keuangan', { params: { bulan } });
  return bukaAmplop(res.data);
}

export async function penjualanHarian(dari?: string, sampai?: string): Promise<PenjualanHarian[]> {
  const res = await api.get<Amplop<PenjualanHarian[]>>('/laporan/penjualan-harian', {
    params: { dari, sampai },
  });
  return bukaAmplop(res.data);
}

export async function produkTerlaris(hari = 30, limit = 10): Promise<ProdukTerlaris[]> {
  const res = await api.get<Amplop<ProdukTerlaris[]>>('/laporan/produk-terlaris', { params: { hari, limit } });
  return bukaAmplop(res.data);
}

export async function daftarPengeluaran(bulan: string, page = 1) {
  const res = await api.get<Amplop<Pengeluaran[]>>('/pengeluaran', { params: { bulan, page, per_page: 10 } });
  return { rows: bukaAmplop(res.data), total: res.data.meta?.total ?? 0 };
}

export async function catatPengeluaran(keterangan: string, nominal: number): Promise<Pengeluaran> {
  const res = await api.post<Amplop<Pengeluaran>>('/pengeluaran', { keterangan, nominal });
  return bukaAmplop(res.data);
}

export async function hapusPengeluaran(id: number): Promise<void> {
  await api.delete(`/pengeluaran/${id}`);
}

export async function daftarPelanggan(q: string, page = 1) {
  const res = await api.get<Amplop<Pelanggan[]>>('/pelanggan', { params: { q, page, per_page: 10 } });
  return { rows: bukaAmplop(res.data), total: res.data.meta?.total ?? 0 };
}

export async function quickPelanggan(nama: string, telepon: string): Promise<Pelanggan> {
  const res = await api.post<Amplop<Pelanggan>>('/pelanggan/quick', { nama, telepon });
  return bukaAmplop(res.data);
}

// API riwayat transaksi kasir (read-only di admin panel — checkout dilakukan
// aplikasi kasir).

import { api, bukaAmplop, type Amplop } from './client';

export interface ItemStruk {
  nama_produk: string;
  satuan: string;
  harga: number;
  kuantitas: number;
  diskon_persen: number;
  pajak_persen: number;
  subtotal: number;
}

export interface Struk {
  id: number;
  nomor: string;
  tanggal: string;
  status: string;
  kasir: string;
  tipe_pembayaran: 'TUNAI' | 'TRANSFER' | 'QRIS';
  subtotal: number;
  total_diskon: number;
  diskon_persen: number;
  diskon_nominal: number;
  total_pajak: number;
  pembulatan: number;
  grand_total: number;
  dibayar: number;
  kembalian: number;
  catatan: string;
  items: ItemStruk[];
}

export interface DaftarTransaksiParams {
  page?: number;
  per_page?: number;
  tanggal_dari?: string;
  tanggal_sampai?: string;
}

export async function daftarTransaksi(params: DaftarTransaksiParams) {
  const res = await api.get<Amplop<Struk[]>>('/transaksi', { params });
  return { rows: bukaAmplop(res.data), total: res.data.meta?.total ?? 0 };
}

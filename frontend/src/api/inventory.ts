// API inventory (stok masuk, opname, riwayat) + batal transaksi.

import { api, bukaAmplop, type Amplop } from './client';
import type { Struk } from './transaksi';

export interface StokLog {
  id: number;
  produk: string;
  jenis: 'MASUK' | 'OPNAME' | 'JUAL' | 'BATAL';
  jumlah: number;
  stok_sesudah: number;
  keterangan: string;
  sesi: string | null;
  pencatat: string;
  waktu: string;
}

export async function stokMasuk(produk_id: number, jumlah: number, keterangan?: string): Promise<StokLog> {
  const res = await api.post<Amplop<StokLog>>('/inventory/stok-masuk', { produk_id, jumlah, keterangan });
  return bukaAmplop(res.data);
}

export async function opname(produk_id: number, stok_fisik: number, keterangan?: string): Promise<StokLog> {
  const res = await api.post<Amplop<StokLog>>('/inventory/opname', { produk_id, stok_fisik, keterangan });
  return bukaAmplop(res.data);
}

export async function riwayatStok(page = 1) {
  const res = await api.get<Amplop<StokLog[]>>('/inventory/riwayat', { params: { page, per_page: 10 } });
  return { rows: bukaAmplop(res.data), total: res.data.meta?.total ?? 0 };
}

export async function batalTransaksi(id: number): Promise<Struk> {
  const res = await api.post<Amplop<Struk>>(`/transaksi/${id}/batal`);
  return bukaAmplop(res.data);
}

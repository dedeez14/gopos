// API katalog produk & kategori.

import { api, bukaAmplop, type Amplop } from './client';

export interface Produk {
  id: number;
  kode: string;
  nama: string;
  barcode: string | null;
  tipe: 'PRODUK' | 'JASA';
  satuan: string;
  harga_beli: number | null;
  harga_jual: number;
  harga_promo: number | null;
  promo_aktif: boolean;
  promo_mulai: string | null;
  promo_selesai: string | null;
  harga_efektif: number;
  favorit: boolean;
  kelola_stok: boolean;
  stok: number;
  kategori: string | null;
  kategori_id: number | null;
  aktif: boolean;
}

export interface Kategori {
  id: number;
  nama: string;
}

export interface DaftarProdukParams {
  q?: string;
  page?: number;
  per_page?: number;
  termasuk_nonaktif?: 1;
}

// Payload simpan (create/update) — bentuk yang diminta backend.
export interface SimpanProdukPayload {
  nama: string;
  kode?: string;
  barcode?: string;
  tipe: 'PRODUK' | 'JASA';
  satuan?: string;
  harga_beli: number;
  harga_jual: number;
  harga_promo: number | null;
  promo_mulai?: string;
  promo_selesai?: string;
  favorit: boolean;
  kelola_stok?: boolean;
  kategori_id?: number;
  aktif: boolean;
}

export async function daftarProduk(params: DaftarProdukParams) {
  const res = await api.get<Amplop<Produk[]>>('/produk', { params });
  return { rows: bukaAmplop(res.data), total: res.data.meta?.total ?? 0 };
}

export async function buatProduk(payload: SimpanProdukPayload): Promise<Produk> {
  const res = await api.post<Amplop<Produk>>('/produk', payload);
  return bukaAmplop(res.data);
}

export async function ubahProduk(id: number, payload: SimpanProdukPayload): Promise<Produk> {
  const res = await api.put<Amplop<Produk>>(`/produk/${id}`, payload);
  return bukaAmplop(res.data);
}

export async function nonaktifkanProduk(id: number): Promise<void> {
  await api.delete(`/produk/${id}`);
}

export async function daftarKategori(): Promise<Kategori[]> {
  const res = await api.get<Amplop<Kategori[]>>('/kategori');
  return bukaAmplop(res.data);
}

export async function buatKategori(nama: string): Promise<Kategori> {
  const res = await api.post<Amplop<Kategori>>('/kategori', { nama });
  return bukaAmplop(res.data);
}

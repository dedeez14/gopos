// API pengaturan usaha — profil toko, struk, pajak. Singleton per usaha.

import { api, bukaAmplop, type Amplop } from './client';

export interface Pengaturan {
  nama_toko: string;
  alamat: string;
  telepon: string;
  email: string;
  logo_url: string;
  mata_uang: string;
  struk_header: string;
  struk_footer: string;
  ukuran_kertas: string; // "58mm" | "80mm"
  tampilkan_logo: boolean;
  pajak_persen: number;
  pajak_aktif: boolean;
  pembulatan: number;
}

export async function ambilPengaturan(): Promise<Pengaturan> {
  const res = await api.get<Amplop<Pengaturan>>('/pengaturan');
  return bukaAmplop(res.data);
}

export async function simpanPengaturan(p: Pengaturan): Promise<Pengaturan> {
  const res = await api.put<Amplop<Pengaturan>>('/pengaturan', p);
  return bukaAmplop(res.data);
}

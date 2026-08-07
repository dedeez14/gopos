// API metode bayar dasar merchant (bank / e-wallet / QRIS statis).

import { api, bukaAmplop, type Amplop } from './client';

export type JenisMetode = 'BANK' | 'EWALLET' | 'QRIS';

export interface MetodeBayar {
  id: number;
  jenis: JenisMetode;
  nama: string;
  nomor: string;
  atas_nama: string;
  gambar_url: string;
  instruksi: string;
  urutan: number;
  aktif: boolean;
}

export type InputMetode = Omit<MetodeBayar, 'id'>;

export async function daftarMetode(hanyaAktif = false): Promise<MetodeBayar[]> {
  const res = await api.get<Amplop<MetodeBayar[]>>('/metode-bayar', {
    params: hanyaAktif ? { aktif: 1 } : undefined,
  });
  return bukaAmplop(res.data);
}

export async function simpanMetode(id: number | null, v: InputMetode): Promise<MetodeBayar> {
  const res = id
    ? await api.put<Amplop<MetodeBayar>>(`/metode-bayar/${id}`, v)
    : await api.post<Amplop<MetodeBayar>>('/metode-bayar', v);
  return bukaAmplop(res.data);
}

export async function hapusMetode(id: number): Promise<void> {
  await api.delete(`/metode-bayar/${id}`);
}

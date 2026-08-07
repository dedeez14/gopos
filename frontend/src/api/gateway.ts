// API gateway pembayaran Midtrans (lapisan 2). Server key TAK pernah plaintext
// dari server — hanya hint 4 digit + flag terisi.

import { api, bukaAmplop, type Amplop } from './client';

export interface KonfigMidtrans {
  platform_aktif: boolean;
  aktif: boolean;
  env: 'sandbox' | 'production';
  merchant_id: string;
  client_key: string;
  server_key_hint: string;
  server_key_terisi: boolean;
  siap: boolean;
}

export interface SimpanGateway {
  aktif: boolean;
  env: 'sandbox' | 'production';
  merchant_id: string;
  client_key: string;
  server_key: string; // kosong = jangan ubah yang tersimpan
}

export async function ambilGateway(): Promise<KonfigMidtrans> {
  const res = await api.get<Amplop<KonfigMidtrans>>('/gateway/midtrans');
  return bukaAmplop(res.data);
}

export async function simpanGateway(v: SimpanGateway): Promise<KonfigMidtrans> {
  const res = await api.put<Amplop<KonfigMidtrans>>('/gateway/midtrans', v);
  return bukaAmplop(res.data);
}

export interface TagihanQris {
  id: number;
  order_id: string;
  nominal: number;
  status: 'PENDING' | 'PAID' | 'EXPIRED' | 'FAILED';
  qr_string: string;
  qr_url: string;
  kedaluwarsa: string;
}

// Buat tagihan QRIS — dipakai aplikasi kasir; di panel dipakai "tes koneksi".
export async function buatQris(nominal: number): Promise<TagihanQris> {
  const res = await api.post<Amplop<TagihanQris>>('/pembayaran/qris', { nominal });
  return bukaAmplop(res.data);
}

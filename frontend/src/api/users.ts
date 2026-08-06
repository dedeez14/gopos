// API pengguna + autentikasi — fungsi murni yang dikonsumsi React Query.

import { api, bukaAmplop, type Amplop } from './client';
import { auth, type UserRingkas } from '../lib/auth';

export interface User {
  id: number;
  nama: string;
  email: string;
  role: 'OWNER' | 'MANAGER' | 'KASIR';
  aktif: boolean;
  dibuat_pada: string;
  diubah_pada: string;
}

export interface DaftarUsersParams {
  q?: string;
  page?: number;
  per_page?: number;
}

export interface HasilDaftarUsers {
  rows: User[];
  total: number;
}

export async function daftarUsers(params: DaftarUsersParams): Promise<HasilDaftarUsers> {
  const res = await api.get<Amplop<User[]>>('/users', { params });
  return { rows: bukaAmplop(res.data), total: res.data.meta?.total ?? 0 };
}

export interface LoginResult {
  user: UserRingkas;
}

export async function login(email: string, password: string): Promise<LoginResult> {
  const res = await api.post<
    Amplop<{ access_token: string; refresh_token: string; user: UserRingkas }>
  >('/auth/login', { email, password });
  const data = bukaAmplop(res.data);
  auth.simpan(data.access_token, data.refresh_token);
  return { user: data.user };
}

export async function logout(): Promise<void> {
  const refresh = auth.refresh();
  if (refresh) {
    // Best-effort: sesi lokal tetap dibersihkan walau server tak terjangkau.
    await api.post('/auth/logout', { refresh_token: refresh }).catch(() => undefined);
  }
  auth.hapus();
}

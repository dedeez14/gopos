// Klien HTTP tunggal untuk seluruh app.
//
// Dua interceptor:
//  1. Request  : sisipkan Authorization: Bearer <access>.
//  2. Response : saat 401, coba SEKALI menukar refresh token → ulangi request
//     yang gagal. Bila refresh ikut gagal → bersihkan sesi & lempar ke /login.
//
// Amplop backend selalu { success, data, meta, message, errors } — helper
// bukaAmplop mengekstrak data + melempar Error(message) saat gagal, supaya
// React Query menerima bentuk yang bersih.

import axios, { AxiosError, type AxiosRequestConfig } from 'axios';
import { auth } from '../lib/auth';

export interface Amplop<T> {
  success: boolean;
  data: T;
  meta: { current_page: number; per_page: number; total: number; last_page: number } | null;
  message: string;
  errors: Record<string, string> | null;
}

export const api = axios.create({ baseURL: '/api/v1' });

api.interceptors.request.use((config) => {
  const access = auth.access();
  if (access) config.headers.Authorization = `Bearer ${access}`;
  return config;
});

let refreshSedangJalan: Promise<void> | null = null;

api.interceptors.response.use(undefined, async (error: AxiosError<Amplop<unknown>>) => {
  const asli = error.config as (AxiosRequestConfig & { _ulang?: boolean }) | undefined;

  const bisaRefresh =
    error.response?.status === 401 && asli && !asli._ulang && auth.refresh() && !asli.url?.includes('/auth/');

  if (!bisaRefresh) throw error;

  // Satu refresh untuk banyak request yang gagal bersamaan.
  refreshSedangJalan ??= axios
    .post<Amplop<{ access_token: string; refresh_token: string }>>('/api/v1/auth/refresh', {
      refresh_token: auth.refresh(),
    })
    .then((res) => auth.simpan(res.data.data.access_token, res.data.data.refresh_token))
    .catch((e) => {
      auth.hapus();
      window.location.href = '/login';
      throw e;
    })
    .finally(() => {
      refreshSedangJalan = null;
    });

  await refreshSedangJalan;
  asli._ulang = true;
  return api(asli);
});

/** Ekstrak data dari amplop; gagal → Error dengan pesan server (Indonesia). */
export function bukaAmplop<T>(amplop: Amplop<T>): T {
  if (!amplop.success) throw new Error(amplop.message || 'Terjadi kesalahan.');
  return amplop.data;
}

/** Ambil pesan error yang layak tampil dari AxiosError. */
export function pesanError(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const amplop = err.response?.data as Amplop<unknown> | undefined;
    if (amplop?.message) return amplop.message;
  }
  return 'Tidak dapat menghubungi server.';
}

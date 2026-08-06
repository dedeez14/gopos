// Penyimpanan token di satu tempat — komponen lain tidak menyentuh
// localStorage langsung (Explicit > Implicit).

const KUNCI_ACCESS = 'tuleh.access';
const KUNCI_REFRESH = 'tuleh.refresh';

export interface UserRingkas {
  id: number;
  nama: string;
  email: string;
  role: string;
}

export const auth = {
  simpan(access: string, refresh: string): void {
    localStorage.setItem(KUNCI_ACCESS, access);
    localStorage.setItem(KUNCI_REFRESH, refresh);
  },
  access: (): string | null => localStorage.getItem(KUNCI_ACCESS),
  refresh: (): string | null => localStorage.getItem(KUNCI_REFRESH),
  hapus(): void {
    localStorage.removeItem(KUNCI_ACCESS);
    localStorage.removeItem(KUNCI_REFRESH);
  },
  sudahMasuk: (): boolean => Boolean(localStorage.getItem(KUNCI_ACCESS)),
};

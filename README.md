# Tuléh Server — Backend Mandiri Tuléh POS

Boilerplate produksi untuk memisahkan backend Tuléh POS dari MOVERA (Laravel)
**secara bertahap** (strangler pattern). Backend **Go (Echo + GORM/PostgreSQL
+ Redis)** dengan Clean Architecture; Admin Panel **React (Vite + TypeScript
+ Ant Design + TanStack Query)**.

## Struktur

```
backend/
├── cmd/api/main.go                  # entry point: config → infra → rakit → serve
├── internal/
│   ├── config/                      # SEMUA env dibaca di satu tempat
│   ├── domain/                      # ★ PUSAT: entitas + kontrak + RBAC + error domain
│   │   └── user.go                  #   (tidak meng-import apa pun dari luar)
│   ├── usecase/                     # aturan aplikasi (tanpa HTTP, tanpa GORM)
│   ├── repository/
│   │   ├── postgres/                # implementasi UserRepository (GORM)
│   │   └── redis/                   # implementasi TokenRepository (refresh token)
│   └── delivery/http/               # handler (+komentar swaggo) & router
│       └── middleware/              # JWT, RBAC (ButuhIzin), CORS/Secure/RateLimit
├── pkg/                             # utilitas lintas-fitur (DRY)
│   ├── respond/                     # amplop {success,data,meta,message,errors} + paginasi
│   ├── apperror/                    # error domain → kode HTTP (satu tempat)
│   └── validasi/                    # validator v10 + pesan Indonesia per field
frontend/
└── src/{api,pages,layout,lib}       # klien axios+refresh, halaman AntD, React Query
```

**Arah dependency: selalu ke dalam.** `delivery → usecase → domain` dan
`repository → domain`. Domain tidak tahu Echo/GORM/Redis. Menambah fitur baru =
ulangi pola `user`: entitas+kontrak di `domain/`, logika di `usecase/`,
implementasi DB di `repository/postgres/`, endpoint di `delivery/http/`.

## Menjalankan Backend

Prasyarat: Go ≥1.23, PostgreSQL ≥14, Redis ≥6.

```bash
cd backend
cp .env.example .env                          # sesuaikan
createdb tuleh_pos                            # atau lewat psql

go mod tidy                                   # unduh dependensi (wajib pertama kali)
set -a; source .env; set +a
go run ./cmd/api                              # http://localhost:8081
```

- **Migrasi DB**: otomatis saat boot (`AutoMigrate` dari struct domain).
  AutoMigrate hanya MENAMBAH kolom/tabel — perubahan destruktif (rename/drop)
  tulis migrasi manual.
- **Admin pertama**: dibuat otomatis saat tabel kosong dari `ADMIN_EMAIL` /
  `ADMIN_PASS` (role OWNER). Ganti sandinya segera.
- **Swagger**: `go install github.com/swaggo/swag/cmd/swag@latest && swag init
  -g cmd/api/main.go`, lalu aktifkan baris `_ "github.com/tuleh-pos/server/docs"`
  di `main.go` → buka `http://localhost:8081/swagger/index.html`.
- Di production, `APP_ENV=production` menolak boot bila `JWT_SECRET`/
  `ADMIN_PASS` masih nilai bawaan.

Uji cepat:

```bash
curl -s localhost:8081/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"admin@tuleh.local","password":"admin1234"}'
# → {success:true, data:{access_token, refresh_token, user...}}
```

## Menjalankan Frontend

```bash
cd frontend
npm install
npm run dev            # http://localhost:5173 (proxy /api → :8081)
```

Masuk dengan kredensial admin di atas. Interceptor axios menangani refresh
token otomatis saat access token kedaluwarsa (401 → refresh → ulangi request).

## Konvensi yang Dijaga

- **Amplop respons** identik dengan kontrak Tuléh yang sudah rilis:
  `{success, data, meta, message, errors}` — klien Flutter tidak boleh bisa
  membedakan server Go dari Laravel. Selalu balas lewat `pkg/respond`.
- **RBAC**: tambah izin di `domain.RolePermissions`, gerbang di router lewat
  `middleware.ButuhIzin(domain.Perm…)`. Fail-closed.
- **Error**: usecase mengembalikan error domain; JANGAN menulis kode status
  di usecase — pemetaan hidup di `pkg/apperror`.
- **Semua teks ke pengguna Bahasa Indonesia**; komentar kode juga.
- **Tanpa `interface{}`/`any` kecuali di batas serialisasi (amplop JSON).**

## Strategi Pemisahan Bertahap dari MOVERA (peta jalan)

Prinsip: **Laravel tetap system-of-record sampai sebuah domain benar-benar
pindah.** Satu domain pindah tuntas (endpoint + data + cutover) sebelum mulai
domain berikutnya. Jangan dual-write dua arah berkepanjangan.

| Fase | Isi | Kriteria selesai |
|---|---|---|
| **0 — Fondasi** (repo ini) | Auth JWT+refresh, RBAC, users, amplop kompatibel | `go build` hijau, admin panel login |
| **1 — Koeksistensi** | Deploy di port sendiri; reverse proxy (Caddy) rutekan path terpilih `/api/pos/v1/...` ke Go, sisanya tetap Laravel | Satu endpoint read-only (mis. `GET /produk` katalog) dilayani Go dari replika data, klien tak menyadari |
| **2 — Domain read** | Pindahkan endpoint BACA per domain (produk, pelanggan, laporan) + sinkron data MySQL→PostgreSQL (CDC/ETL berkala) | Paritas respons diverifikasi diff otomatis Laravel vs Go |
| **3 — Domain write** | Pindahkan TULIS per domain, mulai yang mandiri (hold, pengeluaran) sebelum yang berkait (checkout+stok+jurnal); cutover per domain dengan feature flag di proxy | Laravel berhenti menulis tabel domain tsb |
| **4 — Pemangkasan** | Matikan rute Laravel yang sudah pindah; Laravel menyusut jadi ERP saja | Tuléh POS 100% dilayani Go |

Catatan penting untuk fase 3: integrasi stok FIFO + jurnal otomatis adalah
bagian TERSULIT (hidup di ERP) — pertahankan di Laravel paling lama, atau
ekspos sebagai API internal yang dipanggil Go, sampai ada padanannya.

## Keamanan yang sudah terpasang

CORS whitelist origin, secure headers, body limit 2MB, rate limit per-IP
(ketat khusus login/refresh), bcrypt untuk sandi, JWT HS256 (tolak alg lain),
refresh token opaque di Redis dengan **rotasi** (dipakai ulang = mati),
query ber-parameter (GORM) anti SQL-injection, validasi ketat per field.

package domain

import (
	"context"
	"errors"
	"time"
)

// Pelanggan — kontak pembeli untuk struk, riwayat, dan (kelak) membership.
// Telepon disimpan ternormalisasi format 62… supaya tautan WhatsApp langsung
// jadi dan pencarian tidak terpecah antara 08… dan 62….
type Pelanggan struct {
	ID        uint    `gorm:"primaryKey"`
	Nama      string  `gorm:"size:150;not null;index"`
	Telepon   *string `gorm:"size:20;index"`
	Email     *string `gorm:"size:150"`
	Catatan   string  `gorm:"size:255"`
	Aktif     bool    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrTeleponTerpakai = errors.New("nomor telepon sudah terdaftar pada pelanggan lain")

type FilterPelanggan struct {
	Cari    string // nama/telepon
	Halaman int
	PerHal  int
}

type PelangganRepository interface {
	Simpan(ctx context.Context, p *Pelanggan) error
	Perbarui(ctx context.Context, p *Pelanggan) error
	CariByID(ctx context.Context, id uint) (*Pelanggan, error)
	CariByTelepon(ctx context.Context, telepon string) (*Pelanggan, error)
	Daftar(ctx context.Context, f FilterPelanggan) ([]Pelanggan, int64, error)
	Nonaktifkan(ctx context.Context, id uint) error
}

// ─────────────────────────────────────────────────────────────────── hold

// Hold — keranjang yang diparkir kasir ("pelanggan ambil barang dulu").
// Payload OPAQUE milik aplikasi kasir: server menyimpan, mendaftar, dan
// menghapus — tidak menafsirkan isinya, supaya perubahan bentuk keranjang di
// app tidak butuh migrasi server. Tidak menyentuh uang/stok.
type Hold struct {
	ID        uint   `gorm:"primaryKey"`
	Label     string `gorm:"size:80"`
	Payload   []byte `gorm:"type:jsonb;not null"`
	UserID    uint   `gorm:"index;not null"`
	User      *User
	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrHoldPenuh = errors.New("jumlah hold sudah maksimal. Selesaikan atau hapus hold lama dulu")

type HoldRepository interface {
	Simpan(ctx context.Context, h *Hold) error
	Daftar(ctx context.Context) ([]Hold, error) // terbaru dulu, lintas kasir
	CariByID(ctx context.Context, id uint) (*Hold, error)
	Hapus(ctx context.Context, id uint) error
	Jumlah(ctx context.Context) (int64, error)
}

// ─────────────────────────────────────────────────────────── pengeluaran

// Pengeluaran operasional dari kasir (listrik, wifi, beli gula…). SENGAJA
// bukan jurnal akuntansi: dua kolom "keluar berapa, untuk apa" — bahan kartu
// Laba = Omzet − Pengeluaran di laporan keuangan sederhana.
type Pengeluaran struct {
	ID         uint      `gorm:"primaryKey"`
	Tanggal    time.Time `gorm:"type:date;not null;index"`
	Keterangan string    `gorm:"size:255;not null"`
	Nominal    float64   `gorm:"type:numeric(15,2);not null"`
	UserID     uint      `gorm:"index"`
	User       *User
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type FilterPengeluaran struct {
	Bulan   string // "2026-08"; kosong = semua
	Halaman int
	PerHal  int
}

type PengeluaranRepository interface {
	Simpan(ctx context.Context, p *Pengeluaran) error
	Daftar(ctx context.Context, f FilterPengeluaran) ([]Pengeluaran, int64, error)
	CariByID(ctx context.Context, id uint) (*Pengeluaran, error)
	Hapus(ctx context.Context, id uint) error
	TotalBulan(ctx context.Context, bulan string) (float64, error)
}

// ─────────────────────────────────────────────────────────────── laporan

// Hasil agregat laporan — struct eksplisit, bukan map bebas.
type PenjualanHarian struct {
	Tanggal   string  `json:"tanggal"`
	JumlahTrx int64   `json:"jumlah_trx"`
	Omzet     float64 `json:"omzet"`
	Diskon    float64 `json:"diskon"`
	Pajak     float64 `json:"pajak"`
}

type ProdukTerlaris struct {
	ProdukID uint    `json:"produk_id"`
	Nama     string  `json:"nama"`
	Terjual  float64 `json:"terjual"`
	Omzet    float64 `json:"omzet"`
}

type LaporanRepository interface {
	PenjualanHarian(ctx context.Context, dari, sampai string) ([]PenjualanHarian, error)
	ProdukTerlaris(ctx context.Context, hari, limit int) ([]ProdukTerlaris, error)
	// OmzetBulan: total + rincian per tipe pembayaran + jumlah transaksi.
	OmzetBulan(ctx context.Context, bulan string) (total float64, perTipe map[string]float64, jumlahTrx int64, err error)
}

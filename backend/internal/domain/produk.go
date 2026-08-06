package domain

import (
	"context"
	"time"
)

// ─────────────────────────────────────────────────────────── entitas produk

// TipeProduk di penyimpanan: BARANG | JASA. Kontrak API memakai PRODUK|JASA
// (kompatibel klien Tuléh rilisan) — pemetaan terjadi di lapisan delivery,
// domain tidak perlu tahu.
type TipeProduk string

const (
	TipeBarang TipeProduk = "BARANG"
	TipeJasa   TipeProduk = "JASA"
)

// Kategori produk — sederhana: penanda pengelompokan katalog.
type Kategori struct {
	ID        uint   `gorm:"primaryKey"`
	Nama      string `gorm:"size:100;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Produk adalah item katalog kasir.
//
// Catatan stok: kolom Stok adalah saldo sederhana (bukan lapisan FIFO) —
// cukup untuk fase mandiri awal. Saat domain inventory dipindah dari ERP,
// kolom ini digantikan tabel pergerakan; JANGAN menambah logika akuntansi
// di atasnya.
type Produk struct {
	ID           uint       `gorm:"primaryKey"`
	Kode         string     `gorm:"size:30;uniqueIndex;not null"`
	Nama         string     `gorm:"size:150;not null;index"`
	Barcode      *string    `gorm:"size:60;index"`
	Tipe         TipeProduk `gorm:"size:10;not null;default:BARANG"`
	Satuan       string     `gorm:"size:30;not null;default:pcs"`
	HargaBeli    float64    `gorm:"type:numeric(15,2);not null;default:0"`
	HargaJual    float64    `gorm:"type:numeric(15,2);not null;default:0"`
	HargaPromo   *float64   `gorm:"type:numeric(15,2)"`
	PromoMulai   *time.Time `gorm:"type:date"`
	PromoSelesai *time.Time `gorm:"type:date"`
	// Kolom bool TANPA tag default GORM — disengaja: dengan tag default,
	// GORM MENGHILANGKAN nilai zero (false) saat INSERT sehingga default DB
	// yang menang dan aturan "JASA selalu kelola_stok=false" dikhianati
	// diam-diam. Tanpa tag, GORM selalu menulis nilai eksplisit.
	Favorit      bool    `gorm:"not null"`
	KelolaStok   bool    `gorm:"not null"`
	Stok         float64 `gorm:"type:numeric(15,3);not null;default:0"`
	KategoriID   *uint   `gorm:"index"`
	Kategori     *Kategori
	Aktif        bool `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PromoAktif: harga promo terisi DAN tanggal `pada` berada dalam periode
// (batas kosong = tanpa batas). Perbandingan per-TANGGAL, bukan per-detik —
// promo "s/d 31 Agustus" berlaku sampai 31 Agustus habis.
func (p *Produk) PromoAktif(pada time.Time) bool {
	if p.HargaPromo == nil {
		return false
	}
	tgl := pada.Format("2006-01-02")
	if p.PromoMulai != nil && p.PromoMulai.Format("2006-01-02") > tgl {
		return false
	}
	if p.PromoSelesai != nil && p.PromoSelesai.Format("2006-01-02") < tgl {
		return false
	}
	return true
}

// HargaEfektif = harga yang harus dipakai kasir pada tanggal tsb.
func (p *Produk) HargaEfektif(pada time.Time) float64 {
	if p.PromoAktif(pada) {
		return *p.HargaPromo
	}
	return p.HargaJual
}

// ─────────────────────────────────────────────────────────── kontrak (port)

// FilterProduk — parameter listing katalog, eksplisit per field.
type FilterProduk struct {
	Cari             string // nama/kode/barcode (LIKE)
	KategoriID       uint   // 0 = semua
	Tipe             TipeProduk
	HanyaFavorit     bool
	HanyaPromo       bool // promo AKTIF hari ini
	TermasukNonaktif bool // layar manajemen; katalog kasir selalu false
	Halaman          int
	PerHal           int
}

type ProdukRepository interface {
	Simpan(ctx context.Context, p *Produk) error
	Perbarui(ctx context.Context, p *Produk) error
	CariByID(ctx context.Context, id uint) (*Produk, error)
	CariByKode(ctx context.Context, kode string) (*Produk, error)
	Daftar(ctx context.Context, f FilterProduk) (produk []Produk, total int64, err error)
	// Nonaktifkan, BUKAN hapus baris: produk bisa sudah dirujuk transaksi.
	Nonaktifkan(ctx context.Context, id uint) error
}

type KategoriRepository interface {
	Simpan(ctx context.Context, k *Kategori) error
	Daftar(ctx context.Context) ([]Kategori, error)
	CariByID(ctx context.Context, id uint) (*Kategori, error)
	Hapus(ctx context.Context, id uint) error
}

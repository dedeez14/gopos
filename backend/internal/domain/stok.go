package domain

import (
	"context"
	"errors"
	"time"
)

// StokLog — SATU buku untuk semua pergerakan stok: MASUK (belanja/restock),
// OPNAME (koreksi hitung fisik), JUAL (checkout), BATAL (void mengembalikan).
// Checkout & batal MENULIS log ini di transaksi DB yang sama dengan mutasi
// stoknya — riwayat tidak pernah bolong.
type StokLog struct {
	ID          uint `gorm:"primaryKey"`
	ProdukID    uint `gorm:"index;not null"`
	Produk      *Produk
	Jenis       string  `gorm:"size:10;not null;index"`      // MASUK | OPNAME | JUAL | BATAL
	Jumlah      float64 `gorm:"type:numeric(15,3);not null"` // delta bertanda: masuk +, jual −
	StokSesudah float64 `gorm:"type:numeric(15,3);not null"`
	Keterangan  string  `gorm:"size:255"`
	SesiKasirID *uint   `gorm:"index"`
	SesiKasir   *SesiKasir
	UserID      uint `gorm:"index"`
	User        *User
	CreatedAt   time.Time
}

const (
	StokMasuk  = "MASUK"
	StokOpname = "OPNAME"
	StokJual   = "JUAL"
	StokBatal  = "BATAL"
)

var ErrProdukTanpaStok = errors.New("produk ini tidak mengelola stok (jasa/non-stok)")

type FilterStokLog struct {
	ProdukID uint // 0 = semua
	Jenis    string
	Halaman  int
	PerHal   int
}

type StokRepository interface {
	// Masuk menambah stok sebesar delta (+) dan mencatat log — satu transaksi.
	Masuk(ctx context.Context, produkID uint, delta float64, log *StokLog) error
	// SetAbsolut (opname): stok di-SET ke nilai fisik; log.Jumlah = selisih.
	SetAbsolut(ctx context.Context, produkID uint, stokBaru float64, log *StokLog) error
	Riwayat(ctx context.Context, f FilterStokLog) ([]StokLog, int64, error)
}

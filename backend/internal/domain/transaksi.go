package domain

import (
	"context"
	"errors"
	"time"
)

// Transaksi penjualan kasir (checkout). Uang tercatat saat checkout — tak
// ada status menggantung di v1 (batal/refund menyusul sebagai dokumen
// terpisah, bukan mutasi diam-diam).
type Transaksi struct {
	ID             uint   `gorm:"primaryKey"`
	UsahaID        uint   `gorm:"index;not null;default:0"`
	Nomor          string `gorm:"size:30;uniqueIndex;not null"`
	SesiKasirID    uint   `gorm:"index;not null"`
	UserID         uint   `gorm:"index;not null"`
	User           *User
	PelangganID    *uint `gorm:"index"`
	Pelanggan      *Pelanggan
	IdempotencyKey *string `gorm:"size:80;uniqueIndex"`
	Tanggal        time.Time
	Subtotal       float64 `gorm:"type:numeric(15,2);not null"`
	TotalDiskon    float64 `gorm:"type:numeric(15,2);not null"` // gabungan item + transaksi
	DiskonPersen   float64 `gorm:"type:numeric(5,2);not null"`  // diskon keseluruhan (%)
	DiskonNominal  float64 `gorm:"type:numeric(15,2);not null"` // diskon keseluruhan (Rp, sudah di-cap)
	TotalPajak     float64 `gorm:"type:numeric(15,2);not null"`
	GrandTotal     float64 `gorm:"type:numeric(15,2);not null"`
	Dibayar        float64 `gorm:"type:numeric(15,2);not null"`
	Kembalian      float64 `gorm:"type:numeric(15,2);not null"`
	TipePembayaran string  `gorm:"size:10;not null"` // TUNAI | TRANSFER | QRIS
	Status         string  `gorm:"size:12;not null;index"`
	Catatan        string  `gorm:"size:500"`
	Items          []TransaksiItem
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TransaksiItem menyimpan SNAPSHOT nama/satuan produk — riwayat & struk tak
// boleh berubah walau produknya kelak di-rename atau dinonaktifkan.
type TransaksiItem struct {
	ID           uint    `gorm:"primaryKey"`
	TransaksiID  uint    `gorm:"index;not null"`
	ProdukID     uint    `gorm:"index;not null"`
	NamaProduk   string  `gorm:"size:150;not null"`
	Satuan       string  `gorm:"size:30;not null"`
	Harga        float64 `gorm:"type:numeric(15,2);not null"`
	Kuantitas    float64 `gorm:"type:numeric(15,3);not null"`
	DiskonPersen float64 `gorm:"type:numeric(5,2);not null"`
	PajakPersen  float64 `gorm:"type:numeric(5,2);not null"`
	Subtotal     float64 `gorm:"type:numeric(15,2);not null"` // setelah diskon item + pajak item
}

const (
	TrxSelesai    = "SELESAI"
	TrxDibatalkan = "DIBATALKAN"
	TipeTunai     = "TUNAI"
	TipeTransfer  = "TRANSFER"
	TipeQris      = "QRIS"
)

var (
	ErrProdukNonaktif    = errors.New("produk sudah dinonaktifkan")
	ErrTrxSudahBatal     = errors.New("transaksi sudah dibatalkan")
	ErrTrxBukanMilik     = errors.New("transaksi ini bukan milik Anda")
	ErrSesiTrxSudahTutup = errors.New("transaksi dari sesi yang sudah ditutup tidak bisa dibatalkan")
)

type FilterTransaksi struct {
	SesiKasirID  uint // 0 = semua
	TanggalDari  string
	TanggalAkhir string
	Halaman      int
	PerHal       int
}

type TransaksiRepository interface {
	// Checkout menyimpan transaksi + item DAN memotong stok produk kelola_stok
	// dalam SATU transaksi database — setengah jadi tidak boleh ada.
	// potongStok: produkID → kuantitas.
	Checkout(ctx context.Context, t *Transaksi, potongStok map[uint]float64) error
	CariByID(ctx context.Context, id uint) (*Transaksi, error)
	CariByIdempotency(ctx context.Context, key string) (*Transaksi, error)
	Daftar(ctx context.Context, f FilterTransaksi) ([]Transaksi, int64, error)
	// Batalkan: set status DIBATALKAN + kembalikan stok + log BATAL — satu
	// transaksi DB. kembalikan: produkID → kuantitas.
	Batalkan(ctx context.Context, t *Transaksi, kembalikan map[uint]float64) error
	// TotalSesi: agregat sesi utk rekap tutup — total per tipe pembayaran.
	TotalSesi(ctx context.Context, sesiID uint) (map[string]float64, error)
}

package domain

import (
	"context"
	"errors"
	"time"
)

// MetodePembayaran — cara bayar NON-TUNAI yang dikonfigurasi merchant sendiri:
// lapisan "dasar" yang SELALU ada terlepas dari gateway Midtrans (transfer
// bank, e-wallet, atau QRIS statis berupa gambar QR milik merchant).
//
// Ini data DISPLAY untuk aplikasi kasir (menampilkan ke mana pelanggan
// membayar) — BUKAN FK transaksi: tipe_pembayaran transaksi tetap enum
// TUNAI/TRANSFER/QRIS, jadi metode boleh dihapus tanpa merusak riwayat.
type MetodePembayaran struct {
	ID        uint   `gorm:"primaryKey"`
	UsahaID   uint   `gorm:"index;not null;default:0"`
	Jenis     string `gorm:"size:10;not null"`  // BANK | EWALLET | QRIS
	Nama      string `gorm:"size:80;not null"`  // label: "BCA", "OVO", "QRIS Toko"
	Nomor     string `gorm:"size:60;not null"`  // no. rekening / no. HP (BANK/EWALLET)
	AtasNama  string `gorm:"size:120;not null"` // pemilik rekening/akun
	GambarURL string `gorm:"size:255;not null"` // QR statis (QRIS)
	Instruksi string `gorm:"size:255;not null"` // catatan bayar opsional
	Urutan    int    `gorm:"not null;default:0"`
	Aktif     bool   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	BayarBank    = "BANK"
	BayarEwallet = "EWALLET"
	BayarQris    = "QRIS"
)

var (
	ErrJenisBayarTakDikenal = errors.New("jenis metode pembayaran tidak dikenal")
	// ErrDataBayarKurang: BANK/EWALLET wajib nomor + atas nama; QRIS wajib gambar.
	ErrDataBayarKurang = errors.New("data metode pembayaran belum lengkap untuk jenis ini")
)

// MetodePembayaranRepository — di-scope usaha_id seperti data operasional lain.
type MetodePembayaranRepository interface {
	Simpan(ctx context.Context, m *MetodePembayaran) error
	Perbarui(ctx context.Context, m *MetodePembayaran) error
	CariByID(ctx context.Context, id uint) (*MetodePembayaran, error)
	// Daftar; hanyaAktif=true dipakai aplikasi kasir (sembunyikan yang nonaktif).
	Daftar(ctx context.Context, hanyaAktif bool) ([]MetodePembayaran, error)
	Hapus(ctx context.Context, id uint) error
}

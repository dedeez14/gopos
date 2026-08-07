package domain

import (
	"context"
	"errors"
	"time"
)

// GatewayMidtrans — konfigurasi gateway pembayaran Midtrans MILIK MERCHANT
// (server key masing-masing usaha). Singleton per usaha. Ini "lapisan kedua"
// pembayaran, di balik SAKLAR PLATFORM Usaha.MidtransAktif.
//
// ServerKeyEnc disimpan TERENKRIPSI (AES-GCM, lihat pkg/rahasia) — server key
// TIDAK PERNAH keluar plaintext ke klien. ClientKey publik (memang dipakai SDK
// klien untuk Snap/QRIS), jadi boleh dikembalikan apa adanya.
type GatewayMidtrans struct {
	ID           uint   `gorm:"primaryKey"`
	UsahaID      uint   `gorm:"uniqueIndex;not null;default:0"`
	ServerKeyEnc string `gorm:"size:255;not null"` // AES-GCM base64; kosong = belum diisi
	ClientKey    string `gorm:"size:120;not null"` // publik
	MerchantID   string `gorm:"size:60;not null"`
	Env          string `gorm:"size:12;not null"` // "sandbox" | "production"
	Aktif        bool   `gorm:"not null"`         // saklar MERCHANT (di atas saklar platform)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ErrModulMidtransMati: merchant mencoba mengatur gateway padahal platform
// belum mengaktifkan modul untuk usahanya.
var ErrModulMidtransMati = errors.New("modul Midtrans belum diaktifkan platform untuk usaha ini")

// GatewayMidtransRepository — singleton per usaha (di-scope usaha_id).
type GatewayMidtransRepository interface {
	CariByUsaha(ctx context.Context) (*GatewayMidtrans, error)
	Simpan(ctx context.Context, g *GatewayMidtrans) error
	Perbarui(ctx context.Context, g *GatewayMidtrans) error
}

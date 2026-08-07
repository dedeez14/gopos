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

var (
	// ErrModulMidtransMati: merchant mencoba mengatur/memakai gateway padahal
	// platform belum mengaktifkan modul untuk usahanya.
	ErrModulMidtransMati = errors.New("modul Midtrans belum diaktifkan platform untuk usaha ini")
	// ErrGatewayBelumSiap: modul aktif tapi konfigurasi belum lengkap/aktif
	// (server key kosong atau saklar merchant mati).
	ErrGatewayBelumSiap = errors.New("gateway Midtrans belum siap — aktifkan dan isi server key")
	// ErrGatewayUpstream: kegagalan saat memanggil API Midtrans (jaringan atau
	// status_code non-2xx di body). Dibungkus di atas error asli.
	ErrGatewayUpstream = errors.New("gateway pembayaran menolak permintaan")
	// ErrNominalTakSah: nominal tagihan QRIS harus > 0.
	ErrNominalTakSah = errors.New("nominal tagihan tidak sah")
)

// HasilQris — hasil charge QRIS dari Midtrans (bentuk domain, bebas HTTP).
type HasilQris struct {
	OrderID        string
	TransactionID  string
	QrString       string // payload QRIS (bisa dirender jadi QR oleh klien)
	QrURL          string // URL gambar QR (action generate-qr-code)
	StatusMentah   string // transaction_status mentah dari Midtrans
	KedaluwarsaISO string // expiry_time mentah
}

// GatewayCharger — PORT ke penyedia pembayaran (Midtrans). Implementasi konkret
// (HTTP) hidup di internal/gateway/midtrans; domain hanya tahu kontraknya.
type GatewayCharger interface {
	ChargeQris(ctx context.Context, serverKey, env, orderID string, nominal int64) (HasilQris, error)
	StatusTransaksi(ctx context.Context, serverKey, env, orderID string) (statusMentah string, err error)
}

// GatewayMidtransRepository — singleton per usaha (di-scope usaha_id).
type GatewayMidtransRepository interface {
	CariByUsaha(ctx context.Context) (*GatewayMidtrans, error)
	Simpan(ctx context.Context, g *GatewayMidtrans) error
	Perbarui(ctx context.Context, g *GatewayMidtrans) error
}

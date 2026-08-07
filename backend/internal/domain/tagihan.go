package domain

import (
	"context"
	"time"
)

// TagihanQris — satu "tagihan" QRIS dinamis Midtrans (payment intent). Dibuat
// kasir saat pelanggan memilih bayar QRIS; statusnya di-POLL sampai lunas.
// Di-scope usaha_id seperti data operasional lain.
//
// TransaksiID diisi kelak saat tagihan LUNAS dipakai menutup checkout
// (single-use) — integrasi itu fase berikutnya; untuk sekarang tagihan berdiri
// sendiri sebagai bukti pembayaran.
type TagihanQris struct {
	ID            uint   `gorm:"primaryKey"`
	UsahaID       uint   `gorm:"index;not null;default:0"`
	OrderID       string `gorm:"size:80;uniqueIndex;not null"` // order_id unik yang kita kirim ke Midtrans
	Nominal       int64  `gorm:"not null"`                     // rupiah (IDR tanpa desimal)
	Status        string `gorm:"size:10;not null;index"`       // PENDING | PAID | EXPIRED | FAILED
	QrString      string `gorm:"type:text;not null"`
	QrURL         string `gorm:"size:500;not null"`
	MidtransTrxID string `gorm:"size:80;not null"`
	Kedaluwarsa   string `gorm:"size:40;not null"` // expiry_time mentah dari Midtrans
	TransaksiID   *uint  `gorm:"index"`            // terisi saat dipakai checkout (fase depan)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName dipatok eksplisit — hindari ambiguitas pluralisasi GORM pada
// kata "Qris" (yang berakhiran 's').
func (TagihanQris) TableName() string { return "tagihan_qris" }

const (
	QrisPending = "PENDING"
	QrisPaid    = "PAID"
	QrisExpired = "EXPIRED"
	QrisFailed  = "FAILED"
)

type TagihanQrisRepository interface {
	Simpan(ctx context.Context, t *TagihanQris) error
	Perbarui(ctx context.Context, t *TagihanQris) error
	CariByID(ctx context.Context, id uint) (*TagihanQris, error)
}

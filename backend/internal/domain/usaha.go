package domain

import (
	"context"
	"errors"
	"time"
)

// Usaha — tenant: satu merchant/bisnis. SATU deployment melayani banyak
// usaha; SEMUA data operasional ber-kolom UsahaID dan SELALU di-scope.
//
// Mekanisme scoping (fail-closed by design):
//   - Middleware JWT menaruh usaha_id klaim ke context.Context request.
//   - Repository postgres menambahkan WHERE usaha_id = ? dari context, dan
//     mengisi UsahaID saat Create.
//   - Lupa set → UsahaID 0 → tidak cocok dengan scope siapa pun → data
//     "hilang" (ketahuan saat uji), BUKAN bocor ke usaha lain.
type Usaha struct {
	ID        uint   `gorm:"primaryKey"`
	Kode      string `gorm:"size:30;uniqueIndex;not null"`
	Nama      string `gorm:"size:150;not null"`
	Aktif     bool   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrKodeUsahaTerpakai = errors.New("kode usaha sudah dipakai")
	ErrUsahaNonaktif     = errors.New("usaha ini sedang dinonaktifkan. Hubungi pengelola platform")
)

type FilterUsaha struct {
	Cari    string
	Halaman int
	PerHal  int
}

// UsahaRepository — tabel PLATFORM: sengaja TIDAK di-scope usaha_id.
// Gerbangnya izin usaha.kelola (SUPERADMIN) di route.
type UsahaRepository interface {
	Simpan(ctx context.Context, u *Usaha) error
	Perbarui(ctx context.Context, u *Usaha) error
	CariByID(ctx context.Context, id uint) (*Usaha, error)
	CariByKode(ctx context.Context, kode string) (*Usaha, error)
	Daftar(ctx context.Context, f FilterUsaha) ([]Usaha, int64, error)
}

type kunciUsahaCtx struct{}

// DenganUsaha menitipkan usaha aktif ke context (dipanggil middleware JWT).
func DenganUsaha(ctx context.Context, usahaID uint) context.Context {
	return context.WithValue(ctx, kunciUsahaCtx{}, usahaID)
}

// UsahaDari membaca usaha aktif; 0 = tak ada (fail-closed di scope query).
func UsahaDari(ctx context.Context) uint {
	if id, ok := ctx.Value(kunciUsahaCtx{}).(uint); ok {
		return id
	}
	return 0
}

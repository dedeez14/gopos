package domain

import (
	"context"
	"errors"
	"time"
)

// SesiKasir — sesi kerja kasir: uang tercatat hanya di dalam sesi terbuka.
// Aturan inti (ditegakkan usecase): SATU sesi BUKA per pengguna; tutup sesi
// menghitung kas sistem = kas awal + seluruh penjualan TUNAI sesi tsb, lalu
// selisihnya terhadap kas fisik yang dihitung kasir.
type SesiKasir struct {
	ID          uint   `gorm:"primaryKey"`
	Nomor       string `gorm:"size:30;uniqueIndex;not null"`
	UserID      uint   `gorm:"index;not null"`
	User        *User
	Status      string  `gorm:"size:10;not null;index"` // BUKA | TUTUP
	KasAwal     float64 `gorm:"type:numeric(15,2);not null"`
	KasAkhir    *float64
	KasSistem   *float64 // kas awal + tunai — diisi saat tutup
	Selisih     *float64 // kas akhir − kas sistem
	Catatan     string   `gorm:"size:255"`
	DibukaPada  time.Time
	DitutupPada *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	SesiBuka  = "BUKA"
	SesiTutup = "TUTUP"
)

var (
	ErrSesiSudahBuka  = errors.New("masih ada sesi kasir yang terbuka — tutup dulu sebelum membuka yang baru")
	ErrSesiBelumBuka  = errors.New("belum ada sesi kasir yang terbuka. Buka sesi terlebih dahulu")
	ErrSesiBukanMilik = errors.New("sesi ini bukan milik Anda")
	ErrSesiSudahTutup = errors.New("sesi sudah ditutup")
)

type FilterSesi struct {
	UserID  uint // 0 = semua (layar manajemen)
	Status  string
	Halaman int
	PerHal  int
}

type SesiRepository interface {
	Simpan(ctx context.Context, s *SesiKasir) error
	Perbarui(ctx context.Context, s *SesiKasir) error
	CariByID(ctx context.Context, id uint) (*SesiKasir, error)
	// AktifMilik: sesi BUKA milik user; ErrTidakDitemukan bila tak ada.
	AktifMilik(ctx context.Context, userID uint) (*SesiKasir, error)
	Daftar(ctx context.Context, f FilterSesi) ([]SesiKasir, int64, error)
}

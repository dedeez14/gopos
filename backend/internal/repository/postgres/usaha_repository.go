package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// UsahaRepository — tabel PLATFORM: sengaja tanpa scope usaha_id.
// Gerbangnya izin usaha.kelola (SUPERADMIN) di route.
type UsahaRepository struct {
	db *gorm.DB
}

func NewUsahaRepository(db *gorm.DB) *UsahaRepository {
	return &UsahaRepository{db: db}
}

func (r *UsahaRepository) Simpan(ctx context.Context, u *domain.Usaha) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UsahaRepository) Perbarui(ctx context.Context, u *domain.Usaha) error {
	return r.db.WithContext(ctx).Model(u).Select("*").Omit("id", "created_at").Updates(u).Error
}

func (r *UsahaRepository) CariByID(ctx context.Context, id uint) (*domain.Usaha, error) {
	var u domain.Usaha
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &u, err
}

func (r *UsahaRepository) CariByKode(ctx context.Context, kode string) (*domain.Usaha, error) {
	var u domain.Usaha
	err := r.db.WithContext(ctx).Where("kode = ?", kode).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &u, err
}

func (r *UsahaRepository) Daftar(ctx context.Context, f domain.FilterUsaha) ([]domain.Usaha, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Usaha{})
	if f.Cari != "" {
		pola := "%" + f.Cari + "%"
		q = q.Where("nama ILIKE ? OR kode ILIKE ?", pola, pola)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.Usaha
	err := q.Order("id DESC").
		Offset((f.Halaman - 1) * f.PerHal).Limit(f.PerHal).Find(&rows).Error
	return rows, total, err
}

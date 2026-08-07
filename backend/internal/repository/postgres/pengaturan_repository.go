package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// PengaturanRepository — singleton per usaha (di-scope usaha_id lewat skop()).
type PengaturanRepository struct {
	db *gorm.DB
}

func NewPengaturanRepository(db *gorm.DB) *PengaturanRepository {
	return &PengaturanRepository{db: db}
}

func (r *PengaturanRepository) CariByUsaha(ctx context.Context) (*domain.Pengaturan, error) {
	var p domain.Pengaturan
	err := skop(ctx, r.db).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &p, err
}

func (r *PengaturanRepository) Simpan(ctx context.Context, p *domain.Pengaturan) error {
	isiUsaha(ctx, &p.UsahaID)
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PengaturanRepository) Perbarui(ctx context.Context, p *domain.Pengaturan) error {
	// Select("*") supaya nilai false/0 (pajak nonaktif, pembulatan 0) ikut
	// tersimpan — Updates biasa akan melewati field zero-value.
	return r.db.WithContext(ctx).Model(p).Where("usaha_id = ?", domain.UsahaDari(ctx)).
		Select("*").Omit("id", "created_at", "usaha_id").Updates(p).Error
}

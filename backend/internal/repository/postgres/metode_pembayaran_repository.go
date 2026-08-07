package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// MetodePembayaranRepository — di-scope usaha_id lewat skop().
type MetodePembayaranRepository struct {
	db *gorm.DB
}

func NewMetodePembayaranRepository(db *gorm.DB) *MetodePembayaranRepository {
	return &MetodePembayaranRepository{db: db}
}

func (r *MetodePembayaranRepository) Simpan(ctx context.Context, m *domain.MetodePembayaran) error {
	isiUsaha(ctx, &m.UsahaID)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *MetodePembayaranRepository) Perbarui(ctx context.Context, m *domain.MetodePembayaran) error {
	return r.db.WithContext(ctx).Model(m).Where("usaha_id = ?", domain.UsahaDari(ctx)).
		Select("*").Omit("id", "created_at", "usaha_id").Updates(m).Error
}

func (r *MetodePembayaranRepository) CariByID(ctx context.Context, id uint) (*domain.MetodePembayaran, error) {
	var m domain.MetodePembayaran
	err := skop(ctx, r.db).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &m, err
}

func (r *MetodePembayaranRepository) Daftar(ctx context.Context, hanyaAktif bool) ([]domain.MetodePembayaran, error) {
	q := skop(ctx, r.db).Model(&domain.MetodePembayaran{})
	if hanyaAktif {
		q = q.Where("aktif = ?", true)
	}
	var rows []domain.MetodePembayaran
	err := q.Order("urutan ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *MetodePembayaranRepository) Hapus(ctx context.Context, id uint) error {
	return skop(ctx, r.db).Delete(&domain.MetodePembayaran{}, id).Error
}

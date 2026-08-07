package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// TagihanQrisRepository — di-scope usaha_id lewat skop().
type TagihanQrisRepository struct {
	db *gorm.DB
}

func NewTagihanQrisRepository(db *gorm.DB) *TagihanQrisRepository {
	return &TagihanQrisRepository{db: db}
}

func (r *TagihanQrisRepository) Simpan(ctx context.Context, t *domain.TagihanQris) error {
	isiUsaha(ctx, &t.UsahaID)
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TagihanQrisRepository) Perbarui(ctx context.Context, t *domain.TagihanQris) error {
	return r.db.WithContext(ctx).Model(t).Where("usaha_id = ?", domain.UsahaDari(ctx)).
		Select("*").Omit("id", "created_at", "usaha_id", "order_id").Updates(t).Error
}

func (r *TagihanQrisRepository) CariByID(ctx context.Context, id uint) (*domain.TagihanQris, error) {
	var t domain.TagihanQris
	err := skop(ctx, r.db).First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &t, err
}

package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// GatewayMidtransRepository — singleton per usaha (di-scope usaha_id).
type GatewayMidtransRepository struct {
	db *gorm.DB
}

func NewGatewayMidtransRepository(db *gorm.DB) *GatewayMidtransRepository {
	return &GatewayMidtransRepository{db: db}
}

func (r *GatewayMidtransRepository) CariByUsaha(ctx context.Context) (*domain.GatewayMidtrans, error) {
	var g domain.GatewayMidtrans
	err := skop(ctx, r.db).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &g, err
}

func (r *GatewayMidtransRepository) Simpan(ctx context.Context, g *domain.GatewayMidtrans) error {
	isiUsaha(ctx, &g.UsahaID)
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *GatewayMidtransRepository) Perbarui(ctx context.Context, g *domain.GatewayMidtrans) error {
	return r.db.WithContext(ctx).Model(g).Where("usaha_id = ?", domain.UsahaDari(ctx)).
		Select("*").Omit("id", "created_at", "usaha_id").Updates(g).Error
}

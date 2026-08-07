package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// skop = WithContext + WHERE usaha_id dari context request (diisi middleware
// JWT). Fail-closed: tanpa usaha di context nilainya 0 dan tidak cocok
// dengan baris mana pun — data tidak pernah bocor lintas usaha.
func skop(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx).Where("usaha_id = ?", domain.UsahaDari(ctx))
}

// isiUsaha melengkapi UsahaID entitas dari context bila belum di-set
// (seed/sinkron men-set eksplisit; jalur API memakai context).
func isiUsaha(ctx context.Context, usahaID *uint) {
	if *usahaID == 0 {
		*usahaID = domain.UsahaDari(ctx)
	}
}

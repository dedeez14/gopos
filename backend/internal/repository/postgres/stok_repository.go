package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// StokRepository — mutasi stok + log dalam SATU transaksi DB.
// UPDATE … RETURNING stok memberi angka sesudah-mutasi yang atomik (dua
// pencatat bersamaan tidak saling menimpa hitungan log-nya).
type StokRepository struct {
	db *gorm.DB
}

func NewStokRepository(db *gorm.DB) *StokRepository {
	return &StokRepository{db: db}
}

func (r *StokRepository) Masuk(ctx context.Context, produkID uint, delta float64, log *domain.StokLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sesudah float64
		if err := tx.Raw(`UPDATE produks SET stok = stok + ?, updated_at = NOW() WHERE id = ? RETURNING stok`,
			delta, produkID).Scan(&sesudah).Error; err != nil {
			return err
		}
		log.StokSesudah = sesudah
		return tx.Create(log).Error
	})
}

func (r *StokRepository) SetAbsolut(ctx context.Context, produkID uint, stokBaru float64, log *domain.StokLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE produks SET stok = ?, updated_at = NOW() WHERE id = ?`,
			stokBaru, produkID).Error; err != nil {
			return err
		}
		log.StokSesudah = stokBaru
		return tx.Create(log).Error
	})
}

func (r *StokRepository) Riwayat(ctx context.Context, f domain.FilterStokLog) ([]domain.StokLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.StokLog{})
	if f.ProdukID != 0 {
		q = q.Where("produk_id = ?", f.ProdukID)
	}
	if f.Jenis != "" {
		q = q.Where("jenis = ?", f.Jenis)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.StokLog
	err := q.Preload("Produk").Preload("User").Preload("SesiKasir").Order("id DESC").
		Offset((f.Halaman - 1) * f.PerHal).Limit(f.PerHal).Find(&rows).Error
	return rows, total, err
}

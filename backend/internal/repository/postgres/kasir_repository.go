package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// ─────────────────────────────────────────────────────────────── sesi kasir

type SesiRepository struct {
	db *gorm.DB
}

func NewSesiRepository(db *gorm.DB) *SesiRepository {
	return &SesiRepository{db: db}
}

func (r *SesiRepository) Simpan(ctx context.Context, s *domain.SesiKasir) error {
	// Nomor cantik butuh id → create + update nomor dalam SATU transaksi DB.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(s).Error; err != nil {
			return err
		}
		s.Nomor = fmt.Sprintf("SK-%s-%05d", s.DibukaPada.Format("20060102"), s.ID)
		return tx.Model(s).Update("nomor", s.Nomor).Error
	})
}

func (r *SesiRepository) Perbarui(ctx context.Context, s *domain.SesiKasir) error {
	return r.db.WithContext(ctx).Model(s).Select("*").Omit("id", "created_at").Updates(s).Error
}

func (r *SesiRepository) CariByID(ctx context.Context, id uint) (*domain.SesiKasir, error) {
	var s domain.SesiKasir
	err := r.db.WithContext(ctx).Preload("User").First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &s, err
}

func (r *SesiRepository) AktifMilik(ctx context.Context, userID uint) (*domain.SesiKasir, error) {
	var s domain.SesiKasir
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.SesiBuka).
		Order("id DESC").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &s, err
}

func (r *SesiRepository) Daftar(ctx context.Context, f domain.FilterSesi) ([]domain.SesiKasir, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.SesiKasir{})
	if f.UserID != 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.SesiKasir
	err := q.Preload("User").Order("id DESC").
		Offset((f.Halaman - 1) * f.PerHal).Limit(f.PerHal).Find(&rows).Error

	return rows, total, err
}

// ─────────────────────────────────────────────────────────────── transaksi

type TransaksiRepository struct {
	db *gorm.DB
}

func NewTransaksiRepository(db *gorm.DB) *TransaksiRepository {
	return &TransaksiRepository{db: db}
}

// Checkout: transaksi + item + potong stok + nomor final — SATU transaksi DB.
// Gagal di mana pun = batal semua; tidak pernah ada penjualan tanpa potongan
// stok atau sebaliknya.
func (r *TransaksiRepository) Checkout(ctx context.Context, t *domain.Transaksi, potongStok map[uint]float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		t.Nomor = fmt.Sprintf("POS-%s-%06d", t.Tanggal.Format("20060102"), t.ID)
		if err := tx.Model(t).Update("nomor", t.Nomor).Error; err != nil {
			return err
		}
		for produkID, qty := range potongStok {
			// Ekspresi atomik (stok = stok − qty) — dua kasir menjual barang
			// yang sama bersamaan tidak saling menimpa.
			if err := tx.Model(&domain.Produk{}).Where("id = ?", produkID).
				Update("stok", gorm.Expr("stok - ?", qty)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *TransaksiRepository) CariByID(ctx context.Context, id uint) (*domain.Transaksi, error) {
	var t domain.Transaksi
	err := r.db.WithContext(ctx).Preload("Items").Preload("User").Preload("Pelanggan").First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &t, err
}

func (r *TransaksiRepository) CariByIdempotency(ctx context.Context, key string) (*domain.Transaksi, error) {
	var t domain.Transaksi
	err := r.db.WithContext(ctx).Preload("Items").Preload("User").Preload("Pelanggan").
		Where("idempotency_key = ?", key).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &t, err
}

func (r *TransaksiRepository) Daftar(ctx context.Context, f domain.FilterTransaksi) ([]domain.Transaksi, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Transaksi{})
	if f.SesiKasirID != 0 {
		q = q.Where("sesi_kasir_id = ?", f.SesiKasirID)
	}
	if f.TanggalDari != "" {
		q = q.Where("tanggal::date >= ?", f.TanggalDari)
	}
	if f.TanggalAkhir != "" {
		q = q.Where("tanggal::date <= ?", f.TanggalAkhir)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.Transaksi
	err := q.Preload("Items").Preload("User").Preload("Pelanggan").Order("id DESC").
		Offset((f.Halaman - 1) * f.PerHal).Limit(f.PerHal).Find(&rows).Error

	return rows, total, err
}

func (r *TransaksiRepository) TotalSesi(ctx context.Context, sesiID uint) (map[string]float64, error) {
	var baris []struct {
		TipePembayaran string
		Total          float64
	}
	err := r.db.WithContext(ctx).Model(&domain.Transaksi{}).
		Select("tipe_pembayaran, COALESCE(SUM(grand_total),0) as total").
		Where("sesi_kasir_id = ? AND status = ?", sesiID, domain.TrxSelesai).
		Group("tipe_pembayaran").Scan(&baris).Error
	if err != nil {
		return nil, err
	}

	total := map[string]float64{}
	for _, b := range baris {
		total[b.TipePembayaran] = b.Total
	}
	return total, nil
}

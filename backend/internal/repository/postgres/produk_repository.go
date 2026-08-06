package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// ProdukRepository memenuhi domain.ProdukRepository di atas GORM/PostgreSQL.
type ProdukRepository struct {
	db *gorm.DB
}

func NewProdukRepository(db *gorm.DB) *ProdukRepository {
	return &ProdukRepository{db: db}
}

func (r *ProdukRepository) Simpan(ctx context.Context, p *domain.Produk) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ProdukRepository) Perbarui(ctx context.Context, p *domain.Produk) error {
	// Save + Select("*") supaya kolom pointer yang di-nil-kan (harga_promo,
	// periode, barcode, kategori) IKUT ditulis NULL — perilaku "cabut promo".
	return r.db.WithContext(ctx).Model(p).Select("*").Omit("id", "created_at").Updates(p).Error
}

func (r *ProdukRepository) CariByID(ctx context.Context, id uint) (*domain.Produk, error) {
	var p domain.Produk
	err := r.db.WithContext(ctx).Preload("Kategori").First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &p, err
}

func (r *ProdukRepository) CariByKode(ctx context.Context, kode string) (*domain.Produk, error) {
	var p domain.Produk
	err := r.db.WithContext(ctx).Where("kode = ?", kode).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &p, err
}

func (r *ProdukRepository) Daftar(ctx context.Context, f domain.FilterProduk) ([]domain.Produk, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Produk{})

	if !f.TermasukNonaktif {
		q = q.Where("aktif = ?", true)
	}
	if f.Tipe != "" {
		q = q.Where("tipe = ?", f.Tipe)
	}
	if f.KategoriID != 0 {
		q = q.Where("kategori_id = ?", f.KategoriID)
	}
	if f.HanyaFavorit {
		q = q.Where("favorit = ?", true)
	}
	if f.HanyaPromo {
		hariIni := time.Now().Format("2006-01-02")
		q = q.Where("harga_promo IS NOT NULL").
			Where("(promo_mulai IS NULL OR promo_mulai <= ?)", hariIni).
			Where("(promo_selesai IS NULL OR promo_selesai >= ?)", hariIni)
	}
	if f.Cari != "" {
		pola := "%" + f.Cari + "%"
		q = q.Where("nama ILIKE ? OR kode ILIKE ? OR barcode ILIKE ?", pola, pola, pola)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var produk []domain.Produk
	err := q.Preload("Kategori").Order("nama ASC").
		Offset((f.Halaman - 1) * f.PerHal).Limit(f.PerHal).
		Find(&produk).Error

	return produk, total, err
}

func (r *ProdukRepository) Nonaktifkan(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&domain.Produk{}).
		Where("id = ?", id).Update("aktif", false).Error
}

// ─────────────────────────────────────────────────────────────── kategori

type KategoriRepository struct {
	db *gorm.DB
}

func NewKategoriRepository(db *gorm.DB) *KategoriRepository {
	return &KategoriRepository{db: db}
}

func (r *KategoriRepository) Simpan(ctx context.Context, k *domain.Kategori) error {
	return r.db.WithContext(ctx).Create(k).Error
}

func (r *KategoriRepository) Daftar(ctx context.Context) ([]domain.Kategori, error) {
	var rows []domain.Kategori
	err := r.db.WithContext(ctx).Order("nama ASC").Find(&rows).Error
	return rows, err
}

func (r *KategoriRepository) CariByID(ctx context.Context, id uint) (*domain.Kategori, error) {
	var k domain.Kategori
	err := r.db.WithContext(ctx).First(&k, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &k, err
}

func (r *KategoriRepository) Hapus(ctx context.Context, id uint) error {
	// Produk yang memakai kategori ini dilepas dulu (SET NULL manual) supaya
	// tidak ada FK menggantung.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Produk{}).Where("kategori_id = ?", id).
			Update("kategori_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Kategori{}, id).Error
	})
}

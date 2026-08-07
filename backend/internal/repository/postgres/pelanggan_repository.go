package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// ─────────────────────────────────────────────────────────────── pelanggan

type PelangganRepository struct {
	db *gorm.DB
}

func NewPelangganRepository(db *gorm.DB) *PelangganRepository {
	return &PelangganRepository{db: db}
}

func (r *PelangganRepository) Simpan(ctx context.Context, p *domain.Pelanggan) error {
	isiUsaha(ctx, &p.UsahaID)
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PelangganRepository) Perbarui(ctx context.Context, p *domain.Pelanggan) error {
	return r.db.WithContext(ctx).Model(p).Where("usaha_id = ?", domain.UsahaDari(ctx)).
		Select("*").Omit("id", "created_at", "usaha_id").Updates(p).Error
}

func (r *PelangganRepository) CariByID(ctx context.Context, id uint) (*domain.Pelanggan, error) {
	var p domain.Pelanggan
	err := skop(ctx, r.db).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &p, err
}

func (r *PelangganRepository) CariByTelepon(ctx context.Context, telepon string) (*domain.Pelanggan, error) {
	var p domain.Pelanggan
	err := skop(ctx, r.db).Where("telepon = ?", telepon).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &p, err
}

func (r *PelangganRepository) Daftar(ctx context.Context, f domain.FilterPelanggan) ([]domain.Pelanggan, int64, error) {
	q := skop(ctx, r.db).Model(&domain.Pelanggan{}).Where("aktif = ?", true)
	if f.Cari != "" {
		pola := "%" + f.Cari + "%"
		q = q.Where("nama ILIKE ? OR telepon ILIKE ?", pola, pola)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.Pelanggan
	err := q.Order("nama ASC").
		Offset((f.Halaman - 1) * f.PerHal).Limit(f.PerHal).Find(&rows).Error
	return rows, total, err
}

func (r *PelangganRepository) Nonaktifkan(ctx context.Context, id uint) error {
	return skop(ctx, r.db).Model(&domain.Pelanggan{}).
		Where("id = ?", id).Update("aktif", false).Error
}

// ─────────────────────────────────────────────────────────────────── hold

type HoldRepository struct {
	db *gorm.DB
}

func NewHoldRepository(db *gorm.DB) *HoldRepository {
	return &HoldRepository{db: db}
}

func (r *HoldRepository) Simpan(ctx context.Context, h *domain.Hold) error {
	isiUsaha(ctx, &h.UsahaID)
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *HoldRepository) Daftar(ctx context.Context) ([]domain.Hold, error) {
	var rows []domain.Hold
	err := skop(ctx, r.db).Preload("User").Order("id DESC").Limit(50).Find(&rows).Error
	return rows, err
}

func (r *HoldRepository) CariByID(ctx context.Context, id uint) (*domain.Hold, error) {
	var h domain.Hold
	err := skop(ctx, r.db).First(&h, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &h, err
}

func (r *HoldRepository) Hapus(ctx context.Context, id uint) error {
	return skop(ctx, r.db).Delete(&domain.Hold{}, id).Error
}

func (r *HoldRepository) Jumlah(ctx context.Context) (int64, error) {
	var n int64
	err := skop(ctx, r.db).Model(&domain.Hold{}).Count(&n).Error
	return n, err
}

// ─────────────────────────────────────────────────────────── pengeluaran

type PengeluaranRepository struct {
	db *gorm.DB
}

func NewPengeluaranRepository(db *gorm.DB) *PengeluaranRepository {
	return &PengeluaranRepository{db: db}
}

func (r *PengeluaranRepository) Simpan(ctx context.Context, p *domain.Pengeluaran) error {
	isiUsaha(ctx, &p.UsahaID)
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PengeluaranRepository) Daftar(ctx context.Context, f domain.FilterPengeluaran) ([]domain.Pengeluaran, int64, error) {
	q := skop(ctx, r.db).Model(&domain.Pengeluaran{})
	if f.Bulan != "" {
		q = q.Where("to_char(tanggal, 'YYYY-MM') = ?", f.Bulan)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.Pengeluaran
	err := q.Preload("User").Order("tanggal DESC, id DESC").
		Offset((f.Halaman - 1) * f.PerHal).Limit(f.PerHal).Find(&rows).Error
	return rows, total, err
}

func (r *PengeluaranRepository) CariByID(ctx context.Context, id uint) (*domain.Pengeluaran, error) {
	var p domain.Pengeluaran
	err := skop(ctx, r.db).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &p, err
}

func (r *PengeluaranRepository) Hapus(ctx context.Context, id uint) error {
	return skop(ctx, r.db).Delete(&domain.Pengeluaran{}, id).Error
}

func (r *PengeluaranRepository) TotalBulan(ctx context.Context, bulan string) (float64, error) {
	var total float64
	err := skop(ctx, r.db).Model(&domain.Pengeluaran{}).
		Select("COALESCE(SUM(nominal),0)").
		Where("to_char(tanggal, 'YYYY-MM') = ?", bulan).Scan(&total).Error
	return total, err
}

// ─────────────────────────────────────────────────────────────── laporan

// LaporanRepository — agregat SQL murni di atas tabel transaksi.
// Nama tabel mengikuti GORM: transaksis / transaksi_items.
type LaporanRepository struct {
	db *gorm.DB
}

func NewLaporanRepository(db *gorm.DB) *LaporanRepository {
	return &LaporanRepository{db: db}
}

func (r *LaporanRepository) PenjualanHarian(ctx context.Context, dari, sampai string) ([]domain.PenjualanHarian, error) {
	var rows []domain.PenjualanHarian
	err := r.db.WithContext(ctx).Raw(`
		SELECT tanggal::date::text AS tanggal,
		       COUNT(*)            AS jumlah_trx,
		       COALESCE(SUM(grand_total),0)  AS omzet,
		       COALESCE(SUM(total_diskon),0) AS diskon,
		       COALESCE(SUM(total_pajak),0)  AS pajak
		FROM transaksis
		WHERE status = ? AND usaha_id = ? AND tanggal::date BETWEEN ? AND ?
		GROUP BY tanggal::date ORDER BY tanggal::date`,
		domain.TrxSelesai, domain.UsahaDari(ctx), dari, sampai).Scan(&rows).Error
	return rows, err
}

func (r *LaporanRepository) ProdukTerlaris(ctx context.Context, hari, limit int) ([]domain.ProdukTerlaris, error) {
	var rows []domain.ProdukTerlaris
	// Group per produk_id + NAMA SNAPSHOT — produk yang di-rename muncul
	// sesuai nama saat terjual; itu jujur untuk laporan historis.
	err := r.db.WithContext(ctx).Raw(`
		SELECT i.produk_id, i.nama_produk AS nama,
		       COALESCE(SUM(i.kuantitas),0) AS terjual,
		       COALESCE(SUM(i.subtotal),0)  AS omzet
		FROM transaksi_items i
		JOIN transaksis t ON t.id = i.transaksi_id
		WHERE t.status = ? AND t.usaha_id = ? AND t.tanggal >= NOW() - make_interval(days => ?)
		GROUP BY i.produk_id, i.nama_produk
		ORDER BY terjual DESC LIMIT ?`,
		domain.TrxSelesai, domain.UsahaDari(ctx), hari, limit).Scan(&rows).Error
	return rows, err
}

func (r *LaporanRepository) OmzetBulan(ctx context.Context, bulan string) (float64, map[string]float64, int64, error) {
	var baris []struct {
		TipePembayaran string
		Total          float64
		Jumlah         int64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT tipe_pembayaran, COALESCE(SUM(grand_total),0) AS total, COUNT(*) AS jumlah
		FROM transaksis
		WHERE status = ? AND usaha_id = ? AND to_char(tanggal, 'YYYY-MM') = ?
		GROUP BY tipe_pembayaran`,
		domain.TrxSelesai, domain.UsahaDari(ctx), bulan).Scan(&baris).Error
	if err != nil {
		return 0, nil, 0, err
	}

	var total float64
	var jumlah int64
	perTipe := map[string]float64{}
	for _, b := range baris {
		perTipe[b.TipePembayaran] = b.Total
		total += b.Total
		jumlah += b.Jumlah
	}
	return total, perTipe, jumlah, nil
}

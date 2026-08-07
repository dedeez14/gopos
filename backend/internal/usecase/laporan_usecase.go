package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

// LaporanUsecase — agregat baca untuk layar laporan. Laba di sini adalah
// KAS SEDERHANA (omzet − pengeluaran), bukan akuntansi; pembukuan penuh
// tetap urusan ERP.
type LaporanUsecase struct {
	laporan     domain.LaporanRepository
	pengeluaran domain.PengeluaranRepository
}

func NewLaporanUsecase(laporan domain.LaporanRepository, pengeluaran domain.PengeluaranRepository) *LaporanUsecase {
	return &LaporanUsecase{laporan: laporan, pengeluaran: pengeluaran}
}

// RingkasanKeuangan — bahan kartu "Omzet / Pengeluaran / Laba" + rincian.
type RingkasanKeuangan struct {
	Bulan       string             `json:"bulan"`
	Omzet       float64            `json:"omzet"`
	PerTipe     map[string]float64 `json:"per_tipe"`
	JumlahTrx   int64              `json:"jumlah_trx"`
	Pengeluaran float64            `json:"pengeluaran"`
	Laba        float64            `json:"laba"`
}

// Keuangan(bulan "2026-08"); kosong = bulan berjalan. Bulan divalidasi
// bentuknya di handler; di sini cukup dinormalkan.
func (uc *LaporanUsecase) Keuangan(ctx context.Context, bulan string) (*RingkasanKeuangan, error) {
	bulan = strings.TrimSpace(bulan)
	if bulan == "" {
		bulan = time.Now().Format("2006-01")
	}

	omzet, perTipe, jumlah, err := uc.laporan.OmzetBulan(ctx, bulan)
	if err != nil {
		return nil, err
	}
	keluar, err := uc.pengeluaran.TotalBulan(ctx, bulan)
	if err != nil {
		return nil, err
	}

	return &RingkasanKeuangan{
		Bulan: bulan, Omzet: omzet, PerTipe: perTipe, JumlahTrx: jumlah,
		Pengeluaran: keluar, Laba: omzet - keluar,
	}, nil
}

// PenjualanHarian; rentang kosong = 14 hari terakhir.
func (uc *LaporanUsecase) PenjualanHarian(ctx context.Context, dari, sampai string) ([]domain.PenjualanHarian, error) {
	if dari == "" {
		dari = time.Now().AddDate(0, 0, -13).Format("2006-01-02")
	}
	if sampai == "" {
		sampai = time.Now().Format("2006-01-02")
	}
	return uc.laporan.PenjualanHarian(ctx, dari, sampai)
}

func (uc *LaporanUsecase) ProdukTerlaris(ctx context.Context, hari, limit int) ([]domain.ProdukTerlaris, error) {
	if hari < 1 || hari > 365 {
		hari = 30
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return uc.laporan.ProdukTerlaris(ctx, hari, limit)
}

// ─────────────────────────────────────────────────────────── pengeluaran

type PengeluaranUsecase struct {
	repo domain.PengeluaranRepository
}

func NewPengeluaranUsecase(repo domain.PengeluaranRepository) *PengeluaranUsecase {
	return &PengeluaranUsecase{repo: repo}
}

func (uc *PengeluaranUsecase) Catat(ctx context.Context, userID uint, tanggal time.Time, keterangan string, nominal float64) (*domain.Pengeluaran, error) {
	p := &domain.Pengeluaran{
		Tanggal: tanggal, Keterangan: strings.TrimSpace(keterangan),
		Nominal: nominal, UserID: userID,
	}
	if err := uc.repo.Simpan(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (uc *PengeluaranUsecase) Daftar(ctx context.Context, f domain.FilterPengeluaran) ([]domain.Pengeluaran, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.repo.Daftar(ctx, f)
}

func (uc *PengeluaranUsecase) Hapus(ctx context.Context, id uint) error {
	if _, err := uc.repo.CariByID(ctx, id); err != nil {
		return err
	}
	return uc.repo.Hapus(ctx, id)
}

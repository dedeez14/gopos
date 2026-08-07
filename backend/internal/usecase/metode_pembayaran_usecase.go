package usecase

import (
	"context"
	"strings"

	"github.com/tuleh-pos/server/internal/domain"
)

// MetodePembayaranUsecase — CRUD metode bayar dasar merchant (bank/e-wallet/
// QRIS statis). Validasi kelengkapan per-jenis ditegakkan di sini.
type MetodePembayaranUsecase struct {
	repo domain.MetodePembayaranRepository
}

func NewMetodePembayaranUsecase(repo domain.MetodePembayaranRepository) *MetodePembayaranUsecase {
	return &MetodePembayaranUsecase{repo: repo}
}

type InputMetode struct {
	Jenis     string
	Nama      string
	Nomor     string
	AtasNama  string
	GambarURL string
	Instruksi string
	Urutan    int
	Aktif     bool
}

func (uc *MetodePembayaranUsecase) Daftar(ctx context.Context, hanyaAktif bool) ([]domain.MetodePembayaran, error) {
	return uc.repo.Daftar(ctx, hanyaAktif)
}

func (uc *MetodePembayaranUsecase) Buat(ctx context.Context, in InputMetode) (*domain.MetodePembayaran, error) {
	m, err := rakitMetode(&domain.MetodePembayaran{}, in)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Simpan(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (uc *MetodePembayaranUsecase) Perbarui(ctx context.Context, id uint, in InputMetode) (*domain.MetodePembayaran, error) {
	m, err := uc.repo.CariByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := rakitMetode(m, in); err != nil {
		return nil, err
	}
	if err := uc.repo.Perbarui(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (uc *MetodePembayaranUsecase) Hapus(ctx context.Context, id uint) error {
	if _, err := uc.repo.CariByID(ctx, id); err != nil {
		return err
	}
	return uc.repo.Hapus(ctx, id)
}

// rakitMetode menyalin input ke entitas + menegakkan kelengkapan per-jenis.
// Menerima entitas (baru/lama) supaya dipakai Buat maupun Perbarui.
func rakitMetode(m *domain.MetodePembayaran, in InputMetode) (*domain.MetodePembayaran, error) {
	jenis := strings.ToUpper(strings.TrimSpace(in.Jenis))
	switch jenis {
	case domain.BayarBank, domain.BayarEwallet:
		if strings.TrimSpace(in.Nomor) == "" || strings.TrimSpace(in.AtasNama) == "" {
			return nil, domain.ErrDataBayarKurang
		}
	case domain.BayarQris:
		if strings.TrimSpace(in.GambarURL) == "" {
			return nil, domain.ErrDataBayarKurang
		}
	default:
		return nil, domain.ErrJenisBayarTakDikenal
	}

	m.Jenis = jenis
	m.Nama = strings.TrimSpace(in.Nama)
	m.Nomor = strings.TrimSpace(in.Nomor)
	m.AtasNama = strings.TrimSpace(in.AtasNama)
	m.GambarURL = strings.TrimSpace(in.GambarURL)
	m.Instruksi = strings.TrimSpace(in.Instruksi)
	m.Urutan = in.Urutan
	m.Aktif = in.Aktif
	return m, nil
}

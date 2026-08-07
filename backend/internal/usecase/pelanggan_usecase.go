package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/tuleh-pos/server/internal/domain"
)

// PelangganUsecase — kontak pembeli. Quick-add (nama + WA) adalah jalur
// utama kasir; telepon dinormalkan ke 62… dan dipakai sebagai kunci dedup:
// nomor yang sudah ada mengembalikan pelanggan LAMA, bukan galat — kasir tak
// perlu tahu pelanggan pernah terdaftar.
type PelangganUsecase struct {
	repo domain.PelangganRepository
}

func NewPelangganUsecase(repo domain.PelangganRepository) *PelangganUsecase {
	return &PelangganUsecase{repo: repo}
}

var hanyaDigit = regexp.MustCompile(`\D`)

// NormalisasiTelepon: buang non-digit; "08…" → "628…"; "8…" → "628…".
// Kosong/terlalu pendek → "" (dianggap tanpa telepon).
func NormalisasiTelepon(t string) string {
	d := hanyaDigit.ReplaceAllString(t, "")
	switch {
	case len(d) < 8:
		return ""
	case strings.HasPrefix(d, "0"):
		return "62" + d[1:]
	case strings.HasPrefix(d, "62"):
		return d
	case strings.HasPrefix(d, "8"):
		return "62" + d
	default:
		return d
	}
}

// Quick menambahkan pelanggan dari kasir. Telepon yang sudah terdaftar →
// kembalikan pelanggan lama (idempoten dari sudut pandang kasir).
func (uc *PelangganUsecase) Quick(ctx context.Context, nama, telepon string) (*domain.Pelanggan, bool, error) {
	tel := NormalisasiTelepon(telepon)
	if tel != "" {
		if lama, err := uc.repo.CariByTelepon(ctx, tel); err == nil {
			return lama, false, nil
		} else if !errors.Is(err, domain.ErrTidakDitemukan) {
			return nil, false, err
		}
	}

	p := &domain.Pelanggan{Nama: strings.TrimSpace(nama), Aktif: true}
	if tel != "" {
		p.Telepon = &tel
	}
	if err := uc.repo.Simpan(ctx, p); err != nil {
		return nil, false, err
	}
	return p, true, nil
}

type InputPelanggan struct {
	Nama    string
	Telepon string
	Email   string
	Catatan string
	Aktif   bool
}

func (uc *PelangganUsecase) Perbarui(ctx context.Context, id uint, in InputPelanggan) (*domain.Pelanggan, error) {
	p, err := uc.repo.CariByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tel := NormalisasiTelepon(in.Telepon)
	if tel != "" {
		if lain, err := uc.repo.CariByTelepon(ctx, tel); err == nil && lain.ID != p.ID {
			return nil, domain.ErrTeleponTerpakai
		} else if err != nil && !errors.Is(err, domain.ErrTidakDitemukan) {
			return nil, err
		}
		p.Telepon = &tel
	} else {
		p.Telepon = nil
	}

	p.Nama = strings.TrimSpace(in.Nama)
	if e := strings.TrimSpace(in.Email); e != "" {
		p.Email = &e
	} else {
		p.Email = nil
	}
	p.Catatan = strings.TrimSpace(in.Catatan)
	p.Aktif = in.Aktif

	if err := uc.repo.Perbarui(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (uc *PelangganUsecase) Ambil(ctx context.Context, id uint) (*domain.Pelanggan, error) {
	return uc.repo.CariByID(ctx, id)
}

func (uc *PelangganUsecase) Daftar(ctx context.Context, f domain.FilterPelanggan) ([]domain.Pelanggan, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.repo.Daftar(ctx, f)
}

func (uc *PelangganUsecase) Nonaktifkan(ctx context.Context, id uint) error {
	if _, err := uc.repo.CariByID(ctx, id); err != nil {
		return err
	}
	return uc.repo.Nonaktifkan(ctx, id)
}

// ─────────────────────────────────────────────────────────────────── hold

const maksHold = 50

type HoldUsecase struct {
	repo domain.HoldRepository
}

func NewHoldUsecase(repo domain.HoldRepository) *HoldUsecase {
	return &HoldUsecase{repo: repo}
}

func (uc *HoldUsecase) Simpan(ctx context.Context, userID uint, label string, payload []byte) (*domain.Hold, error) {
	if n, err := uc.repo.Jumlah(ctx); err != nil {
		return nil, err
	} else if n >= maksHold {
		return nil, domain.ErrHoldPenuh
	}

	h := &domain.Hold{Label: strings.TrimSpace(label), Payload: payload, UserID: userID}
	if err := uc.repo.Simpan(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (uc *HoldUsecase) Daftar(ctx context.Context) ([]domain.Hold, error) {
	return uc.repo.Daftar(ctx)
}

func (uc *HoldUsecase) Hapus(ctx context.Context, id uint) error {
	if _, err := uc.repo.CariByID(ctx, id); err != nil {
		return err
	}
	return uc.repo.Hapus(ctx, id)
}

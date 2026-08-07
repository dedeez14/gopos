package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/tuleh-pos/server/internal/domain"
)

// PengaturanUsecase — aturan bisnis konfigurasi usaha (profil toko, struk,
// pajak). Singleton per usaha; dibuat default saat pertama diakses.
type PengaturanUsecase struct {
	repo  domain.PengaturanRepository
	usaha domain.UsahaRepository // untuk prefill nama toko dari nama usaha
}

func NewPengaturanUsecase(repo domain.PengaturanRepository, usaha domain.UsahaRepository) *PengaturanUsecase {
	return &PengaturanUsecase{repo: repo, usaha: usaha}
}

// InputPengaturan — field yang boleh diubah dari panel (eksplisit, bukan map).
type InputPengaturan struct {
	NamaToko      string
	Alamat        string
	Telepon       string
	Email         string
	LogoURL       string
	MataUang      string
	StrukHeader   string
	StrukFooter   string
	UkuranKertas  string
	TampilkanLogo bool
	PajakPersen   float64
	PajakAktif    bool
	Pembulatan    int
}

// Ambil mengembalikan pengaturan usaha aktif, membuat baris default bila belum
// ada — pemanggil (kasir/panel) SELALU dapat objek yang bisa dipakai.
func (uc *PengaturanUsecase) Ambil(ctx context.Context) (*domain.Pengaturan, error) {
	p, err := uc.repo.CariByUsaha(ctx)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, domain.ErrTidakDitemukan) {
		return nil, err
	}

	// Belum ada → buat default, prefill nama toko dari nama usaha.
	nama := ""
	if u, e := uc.usaha.CariByID(ctx, domain.UsahaDari(ctx)); e == nil {
		nama = u.Nama
	}
	def := domain.PengaturanDefault(domain.UsahaDari(ctx), nama)
	if err := uc.repo.Simpan(ctx, &def); err != nil {
		// Balapan first-access (dua request bersamaan) → unik usaha_id
		// dilanggar; baca ulang yang barusan dibuat request lain.
		if p2, e := uc.repo.CariByUsaha(ctx); e == nil {
			return p2, nil
		}
		return nil, err
	}
	return &def, nil
}

// Perbarui menerapkan input ke pengaturan (membuat default dulu bila perlu),
// menormalkan nilai yang di luar rentang, lalu menyimpan.
func (uc *PengaturanUsecase) Perbarui(ctx context.Context, in InputPengaturan) (*domain.Pengaturan, error) {
	p, err := uc.Ambil(ctx)
	if err != nil {
		return nil, err
	}

	p.NamaToko = strings.TrimSpace(in.NamaToko)
	p.Alamat = strings.TrimSpace(in.Alamat)
	p.Telepon = strings.TrimSpace(in.Telepon)
	p.Email = strings.TrimSpace(in.Email)
	p.LogoURL = strings.TrimSpace(in.LogoURL)
	p.MataUang = strings.TrimSpace(in.MataUang)
	p.StrukHeader = strings.TrimSpace(in.StrukHeader)
	p.StrukFooter = strings.TrimSpace(in.StrukFooter)
	p.UkuranKertas = in.UkuranKertas
	p.TampilkanLogo = in.TampilkanLogo
	p.PajakPersen = in.PajakPersen
	p.PajakAktif = in.PajakAktif
	p.Pembulatan = in.Pembulatan

	// Jaring pengaman — handler sudah memvalidasi, tapi domain tetap
	// menegakkan rentang yang masuk akal (defense in depth).
	if p.MataUang == "" {
		p.MataUang = "Rp"
	}
	if p.UkuranKertas != "58mm" && p.UkuranKertas != "80mm" {
		p.UkuranKertas = "80mm"
	}
	if p.PajakPersen < 0 {
		p.PajakPersen = 0
	}
	if p.PajakPersen > 100 {
		p.PajakPersen = 100
	}
	if p.Pembulatan < 0 {
		p.Pembulatan = 0
	}

	if err := uc.repo.Perbarui(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/tuleh-pos/server/internal/domain"
)

// UsahaUsecase — manajemen tenant, level PLATFORM (SUPERADMIN saja lewat
// gerbang route). Membuat usaha selalu SEKALIGUS membuat akun OWNER-nya —
// usaha tanpa pemilik tidak berguna dan menyulitkan onboarding.
type UsahaUsecase struct {
	usahas domain.UsahaRepository
	users  domain.UserRepository
}

func NewUsahaUsecase(usahas domain.UsahaRepository, users domain.UserRepository) *UsahaUsecase {
	return &UsahaUsecase{usahas: usahas, users: users}
}

type InputUsahaBaru struct {
	Nama          string
	Kode          string // kosong = digenerate dari waktu
	OwnerNama     string
	OwnerEmail    string
	OwnerPassword string
}

// Buat membuat usaha + owner. Email owner diperiksa DULU (global unik)
// supaya tidak ada usaha yatim bila email ternyata sudah terpakai.
func (uc *UsahaUsecase) Buat(ctx context.Context, in InputUsahaBaru) (*domain.Usaha, *domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.OwnerEmail))
	if _, err := uc.users.CariByEmail(ctx, email); err == nil {
		return nil, nil, domain.ErrEmailTerpakai
	} else if !errors.Is(err, domain.ErrTidakDitemukan) {
		return nil, nil, err
	}

	kode := strings.ToUpper(strings.TrimSpace(in.Kode))
	if kode == "" {
		kode = fmt.Sprintf("U-%d", time.Now().UnixNano()/1e6)
	}
	if _, err := uc.usahas.CariByKode(ctx, kode); err == nil {
		return nil, nil, domain.ErrKodeUsahaTerpakai
	} else if !errors.Is(err, domain.ErrTidakDitemukan) {
		return nil, nil, err
	}

	usaha := &domain.Usaha{Kode: kode, Nama: strings.TrimSpace(in.Nama), Aktif: true}
	if err := uc.usahas.Simpan(ctx, usaha); err != nil {
		return nil, nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	owner := &domain.User{
		UsahaID: usaha.ID, // eksplisit — BUKAN usaha si superadmin dari context
		Nama:    strings.TrimSpace(in.OwnerNama), Email: email,
		PasswordHash: string(hash), Role: domain.RoleOwner, Aktif: true,
	}
	if err := uc.users.Simpan(ctx, owner); err != nil {
		return nil, nil, err
	}

	return usaha, owner, nil
}

func (uc *UsahaUsecase) Daftar(ctx context.Context, f domain.FilterUsaha) ([]domain.Usaha, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.usahas.Daftar(ctx, f)
}

type InputUbahUsaha struct {
	Nama          string
	Aktif         *bool
	MidtransAktif *bool // saklar platform modul Midtrans
}

// Perbarui — ganti nama / SUSPEND (aktif=false menolak login seluruh
// penggunanya, lihat AuthUsecase).
func (uc *UsahaUsecase) Perbarui(ctx context.Context, id uint, in InputUbahUsaha) (*domain.Usaha, error) {
	u, err := uc.usahas.CariByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n := strings.TrimSpace(in.Nama); n != "" {
		u.Nama = n
	}
	if in.Aktif != nil {
		u.Aktif = *in.Aktif
	}
	if in.MidtransAktif != nil {
		u.MidtransAktif = *in.MidtransAktif
	}
	if err := uc.usahas.Perbarui(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

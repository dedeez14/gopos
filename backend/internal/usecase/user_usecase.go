// Package usecase berisi ATURAN APLIKASI: orkestrasi entitas domain lewat
// kontrak repository. Tidak tahu HTTP, tidak tahu GORM — hanya domain.
package usecase

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/tuleh-pos/server/internal/domain"
)

// UserUsecase menangani CRUD pengguna. Konstruktor menerima KONTRAK
// (interface domain), bukan implementasi — inilah yang membuat usecase bisa
// diuji dengan repository palsu tanpa database.
type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// InputUser adalah data masuk pembuatan/pembaruan — eksplisit per field,
// bukan menerima entitas mentah dari luar.
type InputUser struct {
	Nama     string
	Email    string
	Password string // kosong saat update = tidak mengganti sandi
	Role     domain.Role
	Aktif    bool
}

// Daftar mengembalikan pengguna berpaginasi. Batas per-halaman dijaga DI SINI
// (bukan di handler) supaya aturan berlaku dari jalur mana pun.
func (uc *UserUsecase) Daftar(ctx context.Context, f domain.FilterUser) ([]domain.User, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.repo.Daftar(ctx, f)
}

func (uc *UserUsecase) Ambil(ctx context.Context, id uint) (*domain.User, error) {
	return uc.repo.CariByID(ctx, id)
}

// Buat membuat pengguna baru. Email dinormalkan (lowercase) dan diperiksa
// unik lebih dulu supaya error yang keluar bermakna, bukan pelanggaran
// constraint mentah dari database.
func (uc *UserUsecase) Buat(ctx context.Context, in InputUser) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	if _, err := uc.repo.CariByEmail(ctx, email); err == nil {
		return nil, domain.ErrEmailTerpakai
	} else if !errors.Is(err, domain.ErrTidakDitemukan) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Nama:         strings.TrimSpace(in.Nama),
		Email:        email,
		PasswordHash: string(hash),
		Role:         in.Role,
		Aktif:        in.Aktif,
	}
	if err := uc.repo.Simpan(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Perbarui mengubah data pengguna; sandi hanya diganti bila diisi.
func (uc *UserUsecase) Perbarui(ctx context.Context, id uint, in InputUser) (*domain.User, error) {
	u, err := uc.repo.CariByID(ctx, id)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email != u.Email {
		if _, err := uc.repo.CariByEmail(ctx, email); err == nil {
			return nil, domain.ErrEmailTerpakai
		} else if !errors.Is(err, domain.ErrTidakDitemukan) {
			return nil, err
		}
	}

	u.Nama = strings.TrimSpace(in.Nama)
	u.Email = email
	u.Role = in.Role
	u.Aktif = in.Aktif
	if in.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		u.PasswordHash = string(hash)
	}

	if err := uc.repo.Perbarui(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *UserUsecase) Hapus(ctx context.Context, id uint) error {
	if _, err := uc.repo.CariByID(ctx, id); err != nil {
		return err
	}
	return uc.repo.Hapus(ctx, id)
}

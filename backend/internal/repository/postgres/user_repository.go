// Package postgres berisi implementasi konkret kontrak repository domain di
// atas GORM/PostgreSQL. Hanya lapisan ini yang boleh menyentuh GORM — error
// GORM diterjemahkan ke error domain sebelum naik ke usecase.
package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tuleh-pos/server/internal/domain"
)

// UserRepository memenuhi domain.UserRepository.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Simpan(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) Perbarui(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *UserRepository) Hapus(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

func (r *UserRepository) CariByID(ctx context.Context, id uint) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &u, err
}

func (r *UserRepository) CariByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTidakDitemukan
	}
	return &u, err
}

// Daftar memakai query ber-parameter (?) — GORM men-escape nilainya; JANGAN
// pernah merangkai input pengguna ke string SQL (pintu SQL injection).
func (r *UserRepository) Daftar(ctx context.Context, f domain.FilterUser) ([]domain.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.User{})
	if f.Cari != "" {
		pola := "%" + f.Cari + "%"
		q = q.Where("nama ILIKE ? OR email ILIKE ?", pola, pola)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []domain.User
	err := q.Order("id DESC").
		Offset((f.Halaman - 1) * f.PerHal).
		Limit(f.PerHal).
		Find(&users).Error

	return users, total, err
}

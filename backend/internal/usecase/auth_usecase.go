package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/tuleh-pos/server/internal/domain"
)

// AuthUsecase menangani login/refresh/logout dengan skema dua token:
//   - Access token  : JWT HS256 umur pendek — stateless, diverifikasi middleware.
//   - Refresh token : string acak opaque umur panjang — hidup di Redis, jadi
//     BISA DICABUT (logout paksa, ganti sandi) tanpa menunggu kedaluwarsa.
//
// Rotasi: setiap refresh menerbitkan pasangan baru dan menghapus refresh lama
// — token curian yang dipakai ulang otomatis mati.
type AuthUsecase struct {
	users     domain.UserRepository
	tokens    domain.TokenRepository
	jwtSecret []byte
	accessTTL time.Duration
	refresTTL time.Duration
}

func NewAuthUsecase(
	users domain.UserRepository,
	tokens domain.TokenRepository,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		users: users, tokens: tokens,
		jwtSecret: []byte(jwtSecret),
		accessTTL: accessTTL, refresTTL: refreshTTL,
	}
}

// Klaim adalah isi JWT access token. Sengaja minimal: id + role cukup untuk
// RBAC; data profil diambil dari DB saat dibutuhkan.
type Klaim struct {
	UserID uint        `json:"uid"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// Login memverifikasi kredensial → TokenPair. Pesan error SAMA untuk "email
// tak ada" dan "sandi salah" (ErrKredensialSalah) supaya tidak membocorkan
// email mana yang terdaftar.
func (uc *AuthUsecase) Login(ctx context.Context, email, sandi string) (*domain.TokenPair, *domain.User, error) {
	u, err := uc.users.CariByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, nil, domain.ErrKredensialSalah
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(sandi)) != nil {
		return nil, nil, domain.ErrKredensialSalah
	}
	if !u.Aktif {
		return nil, nil, domain.ErrUserNonaktif
	}

	pair, err := uc.terbitkan(ctx, u)
	return pair, u, err
}

// Refresh menukar refresh token sah dengan pasangan token BARU (rotasi).
func (uc *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	userID, err := uc.tokens.AmbilRefresh(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrTokenTidakSah
	}
	u, err := uc.users.CariByID(ctx, userID)
	if err != nil || !u.Aktif {
		return nil, domain.ErrTokenTidakSah
	}

	// Rotasi: token lama mati sebelum yang baru lahir.
	if err := uc.tokens.HapusRefresh(ctx, refreshToken); err != nil {
		return nil, err
	}
	return uc.terbitkan(ctx, u)
}

// Logout mencabut refresh token; access token dibiarkan kedaluwarsa sendiri
// (umurnya pendek — itulah alasan access dibuat pendek).
func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	return uc.tokens.HapusRefresh(ctx, refreshToken)
}

// VerifikasiAccess memvalidasi JWT dan mengembalikan klaimnya — dipakai
// middleware. Menolak algoritma selain HS256 (mencegah serangan alg=none).
func (uc *AuthUsecase) VerifikasiAccess(tokenString string) (*Klaim, error) {
	klaim := &Klaim{}
	tok, err := jwt.ParseWithClaims(tokenString, klaim, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrTokenTidakSah
		}
		return uc.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return nil, domain.ErrTokenTidakSah
	}
	return klaim, nil
}

func (uc *AuthUsecase) terbitkan(ctx context.Context, u *domain.User) (*domain.TokenPair, error) {
	sekarang := time.Now()
	klaim := &Klaim{
		UserID: u.ID,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Email,
			IssuedAt:  jwt.NewNumericDate(sekarang),
			ExpiresAt: jwt.NewNumericDate(sekarang.Add(uc.accessTTL)),
		},
	}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, klaim).SignedString(uc.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh = 32 byte acak kriptografis → hex 64 karakter.
	mentah := make([]byte, 32)
	if _, err := rand.Read(mentah); err != nil {
		return nil, err
	}
	refresh := hex.EncodeToString(mentah)
	if err := uc.tokens.SimpanRefresh(ctx, refresh, u.ID, uc.refresTTL); err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		KedaluwarsaS: int64(uc.accessTTL.Seconds()),
	}, nil
}

// Package domain adalah PUSAT arsitektur — entitas bisnis + kontrak
// (interface) yang harus dipenuhi lapisan luar.
//
// Aturan dependency Clean Architecture: domain TIDAK BOLEH meng-import
// usecase/repository/delivery/framework apa pun (tidak ada echo, gorm, redis
// di sini). Lapisan luar yang menyesuaikan diri ke kontrak ini, bukan
// sebaliknya.
package domain

import (
	"context"
	"errors"
	"time"
)

// ─────────────────────────────────────────────────────────── entitas & peran

// Role adalah peran pengguna. Tipe string sendiri (bukan string mentah)
// supaya salah ketik ketahuan compiler di seluruh kode.
type Role string

const (
	RoleOwner   Role = "OWNER"   // pemilik usaha — akses penuh
	RoleManager Role = "MANAGER" // pengelola — manajemen tanpa urusan platform
	RoleKasir   Role = "KASIR"   // operasional kasir — akses paling sempit
)

// Permission adalah izin granular yang diperiksa middleware RBAC.
// Konvensi penamaan: "<sumberdaya>.<aksi>".
type Permission string

const (
	PermUserLihat  Permission = "users.lihat"
	PermUserKelola Permission = "users.kelola"
	PermLaporan    Permission = "laporan.lihat"
	PermKasir      Permission = "kasir.operasi"
)

// RolePermissions memetakan peran → izin. SATU sumber kebenaran RBAC;
// middleware dan UI sama-sama membacanya. Fail-closed: peran yang tidak
// terdaftar tidak punya izin apa pun.
var RolePermissions = map[Role][]Permission{
	RoleOwner:   {PermUserLihat, PermUserKelola, PermLaporan, PermKasir},
	RoleManager: {PermUserLihat, PermLaporan, PermKasir},
	RoleKasir:   {PermKasir},
}

// Punya mengecek apakah peran memiliki izin tertentu.
func (r Role) Punya(p Permission) bool {
	for _, izin := range RolePermissions[r] {
		if izin == p {
			return true
		}
	}
	return false
}

// User adalah entitas pengguna. PasswordHash tidak pernah keluar dari
// server — serialisasi ke klien memakai DTO di lapisan delivery, bukan
// struct ini langsung.
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Nama         string    `gorm:"size:150;not null"`
	Email        string    `gorm:"size:150;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:100;not null"`
	Role         Role      `gorm:"size:20;not null;default:KASIR"`
	Aktif        bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TokenPair adalah hasil autentikasi: access pendek (JWT) + refresh panjang
// (opaque, hidup di Redis dan bisa dicabut kapan pun).
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	KedaluwarsaS int64 // detik sisa umur access token — dipakai klien utk jadwal refresh
}

// ─────────────────────────────────────────────────────────── error domain

// Error domain didefinisikan di sini supaya usecase bisa mengembalikan error
// bermakna TANPA tahu HTTP; lapisan delivery yang menerjemahkan ke kode
// status (lihat pkg/apperror).
var (
	ErrTidakDitemukan  = errors.New("data tidak ditemukan")
	ErrEmailTerpakai   = errors.New("email sudah terdaftar")
	ErrKredensialSalah = errors.New("email atau kata sandi salah")
	ErrUserNonaktif    = errors.New("akun dinonaktifkan")
	ErrTokenTidakSah   = errors.New("token tidak sah atau kedaluwarsa")
)

// ─────────────────────────────────────────────────────────── kontrak (port)

// FilterUser adalah parameter listing yang eksplisit — bukan map bebas —
// supaya pemanggil tahu persis apa yang bisa difilter.
type FilterUser struct {
	Cari    string // cocokkan nama/email (LIKE)
	Halaman int    // mulai 1
	PerHal  int    // maks 100, dijaga usecase
}

// UserRepository adalah kontrak penyimpanan pengguna. Implementasi konkret
// (GORM/PostgreSQL) hidup di internal/repository/postgres.
type UserRepository interface {
	Simpan(ctx context.Context, u *User) error
	Perbarui(ctx context.Context, u *User) error
	Hapus(ctx context.Context, id uint) error
	CariByID(ctx context.Context, id uint) (*User, error)
	CariByEmail(ctx context.Context, email string) (*User, error)
	Daftar(ctx context.Context, f FilterUser) (users []User, total int64, err error)
}

// TokenRepository adalah kontrak penyimpanan refresh token (Redis).
// Refresh token bersifat opaque: nilainya acak, artinya hanya ada di server.
type TokenRepository interface {
	SimpanRefresh(ctx context.Context, token string, userID uint, ttl time.Duration) error
	// AmbilRefresh mengembalikan userID pemilik token; ErrTokenTidakSah bila tak ada.
	AmbilRefresh(ctx context.Context, token string) (uint, error)
	HapusRefresh(ctx context.Context, token string) error
}

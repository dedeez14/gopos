// Package config membaca SELURUH konfigurasi dari environment di satu tempat.
//
// Aturan rumah: tidak ada os.Getenv() liar di file lain — semua nilai lewat
// struct Config ini supaya programmer baru cukup membuka satu file untuk tahu
// knob apa saja yang tersedia (Explicit > Implicit).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config memuat semua setelan runtime. Nilai default dipilih aman untuk
// pengembangan lokal; produksi WAJIB mengisi minimal DB_*, REDIS_ADDR,
// JWT_SECRET, dan ADMIN_*.
type Config struct {
	AppEnv   string // "local" | "production"
	HTTPPort string

	// PostgreSQL
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	// Redis (refresh token + cache)
	RedisAddr string
	RedisPass string
	RedisDB   int

	// Auth
	JWTSecret      string
	AccessTokenTTL time.Duration // umur pendek (menit)
	RefreshTTL     time.Duration // umur panjang (hari), disimpan di Redis

	// Bootstrap admin pertama (dibuat hanya bila tabel users kosong)
	AdminNama  string
	AdminEmail string
	AdminPass  string

	// Keamanan HTTP
	CORSOrigins []string // daftar origin admin panel, dipisah koma
	RateRPS     float64  // request/detik per IP (limiter umum)

	// AdminDist: path build admin panel (frontend/dist). Terisi → server ini
	// juga menyajikan panelnya (satu origin, tanpa urusan CORS). Kosong = API saja.
	AdminDist string
}

// Muat membaca env → Config. Gagal keras (error, bukan diam) bila nilai wajib
// kosong di mode production — salah konfigurasi harus ketahuan saat boot,
// bukan saat request pertama.
func Muat() (*Config, error) {
	c := &Config{
		AppEnv:   ambil("APP_ENV", "local"),
		HTTPPort: ambil("HTTP_PORT", "8081"),

		DBHost: ambil("DB_HOST", "127.0.0.1"),
		DBPort: ambil("DB_PORT", "5432"),
		DBUser: ambil("DB_USER", "tuleh"),
		DBPass: ambil("DB_PASS", "tuleh"),
		DBName: ambil("DB_NAME", "tuleh_pos"),

		RedisAddr: ambil("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass: ambil("REDIS_PASS", ""),
		RedisDB:   ambilInt("REDIS_DB", 1),

		JWTSecret:      ambil("JWT_SECRET", "ganti-di-produksi"),
		AccessTokenTTL: time.Duration(ambilInt("ACCESS_TTL_MENIT", 15)) * time.Minute,
		RefreshTTL:     time.Duration(ambilInt("REFRESH_TTL_HARI", 30)) * 24 * time.Hour,

		AdminNama:  ambil("ADMIN_NAMA", "Admin Tuléh"),
		AdminEmail: ambil("ADMIN_EMAIL", "admin@tuleh.local"),
		AdminPass:  ambil("ADMIN_PASS", "admin1234"),

		CORSOrigins: strings.Split(ambil("CORS_ORIGINS", "http://localhost:5173"), ","),
		RateRPS:     float64(ambilInt("RATE_RPS", 20)),
		AdminDist:   ambil("ADMIN_DIST", ""),
	}

	if c.AppEnv == "production" {
		if c.JWTSecret == "ganti-di-produksi" {
			return nil, fmt.Errorf("JWT_SECRET wajib diisi di production")
		}
		if c.AdminPass == "admin1234" {
			return nil, fmt.Errorf("ADMIN_PASS wajib diganti di production")
		}
	}

	return c, nil
}

// DSN merangkai string koneksi PostgreSQL untuk GORM.
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName)
}

func ambil(kunci, bawaan string) string {
	if v := strings.TrimSpace(os.Getenv(kunci)); v != "" {
		return v
	}
	return bawaan
}

func ambilInt(kunci string, bawaan int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(kunci))); err == nil {
		return v
	}
	return bawaan
}

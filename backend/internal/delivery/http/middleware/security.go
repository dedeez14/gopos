package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/tuleh-pos/server/pkg/respond"
)

// Keamanan merangkai middleware keamanan standar dalam SATU pemasangan di
// main.go — supaya tidak ada instance yang lupa terpasang saat menambah
// entry point baru.
//
//   - CORS       : hanya origin admin panel yang diizinkan (bukan "*").
//   - Secure     : header pelindung (X-Frame-Options, nosniff, XSS, HSTS oleh
//     reverse proxy TLS di depan).
//   - BodyLimit  : tolak payload raksasa sebelum menyentuh handler.
//   - RateLimiter: per-IP, in-memory — pelindung brute force & flood. Untuk
//     multi-instance, ganti store ke Redis.
//
// Catatan sanitasi input: SQL injection dicegah oleh query ber-parameter di
// repository (GORM); XSS dicegah karena API ini murni JSON (tidak merender
// HTML) + validasi ketat per field — bukan oleh "global sanitizer" ajaib.
func Keamanan(origins []string, rps float64) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		echomw.CORSWithConfig(echomw.CORSConfig{
			AllowOrigins: origins,
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
			AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType},
		}),
		echomw.SecureWithConfig(echomw.SecureConfig{
			XSSProtection:      "1; mode=block",
			ContentTypeNosniff: "nosniff",
			XFrameOptions:      "DENY",
		}),
		echomw.BodyLimit("2M"),
		echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
			Store: echomw.NewRateLimiterMemoryStore(rate.Limit(rps)),
			ErrorHandler: func(c echo.Context, _ error) error {
				return respond.Gagal(c, http.StatusTooManyRequests, "Terlalu banyak permintaan. Coba lagi sebentar.", nil)
			},
			DenyHandler: func(c echo.Context, _ string, _ error) error {
				return respond.Gagal(c, http.StatusTooManyRequests, "Terlalu banyak permintaan. Coba lagi sebentar.", nil)
			},
		}),
	}
}

// RateKetat adalah limiter khusus endpoint sensitif (login/refresh) — jauh
// lebih ketat dari limiter umum. Laju 1 rps/IP menahan brute force; burst 5
// memberi ruang alur sah yang beruntun (login → refresh → coba ulang) tanpa
// 429 palsu.
func RateKetat() echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(1),
			Burst:     5,
			ExpiresIn: 3 * time.Minute,
		}),
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return respond.Gagal(c, http.StatusTooManyRequests, "Terlalu banyak percobaan. Tunggu sebentar.", nil)
		},
	})
}

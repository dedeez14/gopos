package http

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	appmw "github.com/tuleh-pos/server/internal/delivery/http/middleware"
)

// DaftarkanRute memasang seluruh endpoint. SATU file peta rute — programmer
// baru cukup membuka file ini untuk melihat permukaan API + gerbangnya.
func DaftarkanRute(
	e *echo.Echo,
	authUC *usecase.AuthUsecase,
	userUC *usecase.UserUsecase,
) {
	authH := NewAuthHandler(authUC)
	userH := NewUserHandler(userUC)

	// Dokumentasi Swagger (hasil `swag init`): /swagger/index.html
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := e.Group("/api/v1")

	// ── Publik ───────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	auth.POST("/login", authH.Login, appmw.RateKetat())
	auth.POST("/refresh", authH.Refresh, appmw.RateKetat())

	// ── Ber-auth (JWT) ───────────────────────────────────────────────────
	privat := v1.Group("", appmw.JWT(authUC))
	privat.POST("/auth/logout", authH.Logout)

	// RBAC per-endpoint: baca cukup users.lihat, tulis butuh users.kelola.
	privat.GET("/users", userH.Daftar, appmw.ButuhIzin(domain.PermUserLihat))
	privat.GET("/users/:id", userH.Ambil, appmw.ButuhIzin(domain.PermUserLihat))
	privat.POST("/users", userH.Buat, appmw.ButuhIzin(domain.PermUserKelola))
	privat.PUT("/users/:id", userH.Perbarui, appmw.ButuhIzin(domain.PermUserKelola))
	privat.DELETE("/users/:id", userH.Hapus, appmw.ButuhIzin(domain.PermUserKelola))
}

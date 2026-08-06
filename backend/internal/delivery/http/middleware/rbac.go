package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/pkg/respond"
)

// ButuhIzin adalah gerbang RBAC: dipasang SETELAH middleware JWT, menolak 403
// bila peran pengguna tidak memiliki izin yang diminta.
//
// Sumber kebenaran izin = domain.RolePermissions (satu peta, fail-closed:
// peran tak terdaftar berarti tanpa izin). Contoh pemakaian di router:
//
//	admin.POST("/users", h.Buat, middleware.ButuhIzin(domain.PermUserKelola))
func ButuhIzin(izin domain.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(CtxRole).(domain.Role)
			if !ok || !role.Punya(izin) {
				return respond.Gagal(c, http.StatusForbidden,
					"Anda tidak memiliki izin untuk aksi ini.", nil)
			}
			return next(c)
		}
	}
}

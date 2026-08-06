// Package middleware berisi gerbang lintas-endpoint: JWT, RBAC, dan keamanan
// HTTP. Middleware TIDAK berisi logika bisnis — hanya memutuskan lolos/tolak
// lalu menitipkan identitas ke context.
package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/respond"
)

// Kunci context — konstanta bertipe supaya tidak ada string ajaib tersebar.
const (
	CtxUserID = "auth.user_id"
	CtxRole   = "auth.role"
)

// JWT memverifikasi header "Authorization: Bearer <access>" lalu menitipkan
// user_id + role ke context untuk dipakai handler & RBAC.
func JWT(auth *usecase.AuthUsecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const prefix = "Bearer "
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			if !strings.HasPrefix(header, prefix) {
				return respond.Gagal(c, http.StatusUnauthorized, "Autentikasi diperlukan.", nil)
			}

			klaim, err := auth.VerifikasiAccess(strings.TrimPrefix(header, prefix))
			if err != nil {
				return respond.Gagal(c, http.StatusUnauthorized, "Sesi tidak sah atau kedaluwarsa. Silakan masuk kembali.", nil)
			}

			c.Set(CtxUserID, klaim.UserID)
			c.Set(CtxRole, klaim.Role)
			return next(c)
		}
	}
}

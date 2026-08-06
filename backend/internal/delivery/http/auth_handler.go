// Package http adalah lapisan delivery: menerjemahkan HTTP ↔ usecase.
// Handler TIDAK berisi logika bisnis — hanya bind+validasi request, panggil
// usecase, dan bungkus hasil ke amplop respons.
package http

import (
	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

type AuthHandler struct {
	auth *usecase.AuthUsecase
}

func NewAuthHandler(auth *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// LoginRequest — tag validate diproses pkg/validasi (pesan Indonesia).
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"admin@tuleh.local"`
	Password string `json:"password" validate:"required,min=6" example:"admin1234"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponse adalah DTO keluar autentikasi.
type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"` // detik
	User         UserResponse `json:"user"`
}

// Login godoc
//
//	@Summary		Masuk (email + kata sandi)
//	@Description	Mengembalikan access token (pendek) dan refresh token (panjang).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		LoginRequest	true	"Kredensial"
//	@Success		200		{object}	respond.Amplop{data=TokenResponse}
//	@Failure		401		{object}	respond.Amplop
//	@Failure		422		{object}	respond.Amplop
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}

	pair, user, err := h.auth.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}

	return respond.Sukses(c, TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.KedaluwarsaS,
		User:         keUserResponse(user),
	}, nil, "Berhasil masuk.")
}

// Refresh godoc
//
//	@Summary		Tukar refresh token dengan pasangan token baru (rotasi)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	respond.Amplop{data=TokenResponse}
//	@Failure		401		{object}	respond.Amplop
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}

	pair, err := h.auth.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}

	return respond.Sukses(c, TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.KedaluwarsaS,
	}, nil, "Token diperbarui.")
}

// Logout godoc
//
//	@Summary	Keluar (cabut refresh token)
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		RefreshRequest	true	"Refresh token yang dicabut"
//	@Success	200		{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	var req RefreshRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	if err := h.auth.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Berhasil keluar.")
}

// bindDanValidasi menggabungkan bind + validasi. SENGAJA tidak menulis
// respons apa pun — error yang dikembalikan (echo.HTTPError / GagalValidasi)
// diterjemahkan SEKALI di HTTPErrorHandler global (main.go). Menulis respons
// di helper lalu return nil membuat handler lanjut jalan dan menulis respons
// kedua — bug yang pernah terjadi, jangan diulang.
func bindDanValidasi(c echo.Context, tujuan any) error {
	if err := c.Bind(tujuan); err != nil {
		return echo.NewHTTPError(400, "Payload tidak dapat dibaca.")
	}
	return c.Validate(tujuan) // nil, atau *validasi.GagalValidasi
}

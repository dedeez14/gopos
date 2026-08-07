package http

import (
	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

// GatewayHandler — konfigurasi Midtrans merchant (lapisan kedua pembayaran).
// KONFIG (server key) khusus Owner/Manager; STATUS ringkas untuk aplikasi kasir.
type GatewayHandler struct {
	uc *usecase.GatewayUsecase
}

func NewGatewayHandler(uc *usecase.GatewayUsecase) *GatewayHandler {
	return &GatewayHandler{uc: uc}
}

type GatewayResponse struct {
	PlatformAktif   bool   `json:"platform_aktif"`
	Aktif           bool   `json:"aktif"`
	Env             string `json:"env"`
	MerchantID      string `json:"merchant_id"`
	ClientKey       string `json:"client_key"`
	ServerKeyHint   string `json:"server_key_hint"`
	ServerKeyTerisi bool   `json:"server_key_terisi"`
	Siap            bool   `json:"siap"`
}

func keGatewayResponse(k *usecase.KonfigMidtrans) GatewayResponse {
	return GatewayResponse{
		PlatformAktif: k.PlatformAktif, Aktif: k.Aktif, Env: k.Env,
		MerchantID: k.MerchantID, ClientKey: k.ClientKey,
		ServerKeyHint: k.ServerKeyHint, ServerKeyTerisi: k.ServerKeyTerisi, Siap: k.Siap,
	}
}

type SimpanGatewayRequest struct {
	Aktif      *bool  `json:"aktif" validate:"required"`
	Env        string `json:"env" validate:"required,oneof=sandbox production"`
	MerchantID string `json:"merchant_id" validate:"omitempty,max=60"`
	ClientKey  string `json:"client_key" validate:"omitempty,max=120"`
	ServerKey  string `json:"server_key" validate:"omitempty,max=120"` // kosong = tak diubah
}

// AmbilKonfig godoc
//
//	@Summary	Konfigurasi Midtrans merchant (server key BERTOPENG)
//	@Tags		gateway
//	@Produce	json
//	@Success	200	{object}	respond.Amplop{data=GatewayResponse}
//	@Security	BearerAuth
//	@Router		/gateway/midtrans [get]
func (h *GatewayHandler) AmbilKonfig(c echo.Context) error {
	k, err := h.uc.Ambil(c.Request().Context())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keGatewayResponse(k), nil, "")
}

// SimpanKonfig godoc
//
//	@Summary		Simpan konfigurasi Midtrans (Owner/Manager)
//	@Description	Ditolak (403) bila modul Midtrans belum diaktifkan platform. server_key dikosongkan → key tersimpan dipertahankan.
//	@Tags			gateway
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		SimpanGatewayRequest	true	"Konfigurasi"
//	@Success		200		{object}	respond.Amplop{data=GatewayResponse}
//	@Failure		403		{object}	respond.Amplop	"Modul Midtrans belum aktif"
//	@Security		BearerAuth
//	@Router			/gateway/midtrans [put]
func (h *GatewayHandler) SimpanKonfig(c echo.Context) error {
	var req SimpanGatewayRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	k, err := h.uc.Perbarui(c.Request().Context(), usecase.InputGateway{
		Aktif: *req.Aktif, Env: req.Env, MerchantID: req.MerchantID,
		ClientKey: req.ClientKey, ServerKey: req.ServerKey,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keGatewayResponse(k), nil, "Konfigurasi Midtrans disimpan.")
}

type StatusGatewayResponse struct {
	Siap      bool   `json:"siap"`
	Env       string `json:"env"`
	ClientKey string `json:"client_key"`
}

// Status godoc
//
//	@Summary		Status Midtrans untuk aplikasi kasir (tanpa server key)
//	@Description	siap=true → tampilkan opsi QRIS dinamis; client_key aman dipakai SDK klien.
//	@Tags			gateway
//	@Produce		json
//	@Success		200	{object}	respond.Amplop{data=StatusGatewayResponse}
//	@Security		BearerAuth
//	@Router			/gateway/midtrans/status [get]
func (h *GatewayHandler) Status(c echo.Context) error {
	s, err := h.uc.Status(c.Request().Context())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, StatusGatewayResponse{Siap: s.Siap, Env: s.Env, ClientKey: s.ClientKey}, nil, "")
}

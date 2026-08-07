package http

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

// QrisHandler — pembayaran QRIS dinamis (Midtrans). Alat kasir: buat tagihan
// lalu poll statusnya sampai lunas. Digerbang PermKasir; kesiapan gateway &
// saklar platform ditegakkan di usecase.
type QrisHandler struct {
	uc *usecase.TagihanUsecase
}

func NewQrisHandler(uc *usecase.TagihanUsecase) *QrisHandler {
	return &QrisHandler{uc: uc}
}

type TagihanResponse struct {
	ID          uint   `json:"id"`
	OrderID     string `json:"order_id"`
	Nominal     int64  `json:"nominal"`
	Status      string `json:"status"`
	QrString    string `json:"qr_string"`
	QrURL       string `json:"qr_url"`
	Kedaluwarsa string `json:"kedaluwarsa"`
}

func keTagihanResponse(t *domain.TagihanQris) TagihanResponse {
	return TagihanResponse{
		ID: t.ID, OrderID: t.OrderID, Nominal: t.Nominal, Status: t.Status,
		QrString: t.QrString, QrURL: t.QrURL, Kedaluwarsa: t.Kedaluwarsa,
	}
}

type BuatQrisRequest struct {
	Nominal int64 `json:"nominal" validate:"required,gt=0"`
}

// BuatQris godoc
//
//	@Summary		Buat tagihan QRIS dinamis (Midtrans)
//	@Description	Butuh gateway SIAP (modul aktif platform + saklar merchant + server key). Kegagalan Midtrans → 502.
//	@Tags			qris
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		BuatQrisRequest	true	"Nominal (rupiah)"
//	@Success		201		{object}	respond.Amplop{data=TagihanResponse}
//	@Failure		409		{object}	respond.Amplop	"Gateway belum siap"
//	@Failure		502		{object}	respond.Amplop	"Midtrans menolak"
//	@Security		BearerAuth
//	@Router			/pembayaran/qris [post]
func (h *QrisHandler) BuatQris(c echo.Context) error {
	var req BuatQrisRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	t, err := h.uc.BuatQris(c.Request().Context(), req.Nominal)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keTagihanResponse(t), "Tagihan QRIS dibuat.")
}

// CekQris godoc
//
//	@Summary	Poll status tagihan QRIS (PENDING → PAID/EXPIRED/FAILED)
//	@Tags		qris
//	@Produce	json
//	@Param		id	path		int	true	"ID tagihan"
//	@Success	200	{object}	respond.Amplop{data=TagihanResponse}
//	@Security	BearerAuth
//	@Router		/pembayaran/qris/{id} [get]
func (h *QrisHandler) CekQris(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	t, err := h.uc.CekStatus(c.Request().Context(), uint(id))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keTagihanResponse(t), nil, "")
}

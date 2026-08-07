package http

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

// MetodePembayaranHandler — metode bayar dasar merchant. BACA terbuka kasir
// (aplikasi menampilkan pilihan bayar); TULIS khusus Owner/Manager.
type MetodePembayaranHandler struct {
	uc *usecase.MetodePembayaranUsecase
}

func NewMetodePembayaranHandler(uc *usecase.MetodePembayaranUsecase) *MetodePembayaranHandler {
	return &MetodePembayaranHandler{uc: uc}
}

type MetodeResponse struct {
	ID        uint   `json:"id"`
	Jenis     string `json:"jenis"`
	Nama      string `json:"nama"`
	Nomor     string `json:"nomor"`
	AtasNama  string `json:"atas_nama"`
	GambarURL string `json:"gambar_url"`
	Instruksi string `json:"instruksi"`
	Urutan    int    `json:"urutan"`
	Aktif     bool   `json:"aktif"`
}

func keMetodeResponse(m *domain.MetodePembayaran) MetodeResponse {
	return MetodeResponse{
		ID: m.ID, Jenis: m.Jenis, Nama: m.Nama, Nomor: m.Nomor,
		AtasNama: m.AtasNama, GambarURL: m.GambarURL, Instruksi: m.Instruksi,
		Urutan: m.Urutan, Aktif: m.Aktif,
	}
}

type SimpanMetodeRequest struct {
	Jenis     string `json:"jenis" validate:"required,oneof=BANK EWALLET QRIS"`
	Nama      string `json:"nama" validate:"required,min=1,max=80"`
	Nomor     string `json:"nomor" validate:"omitempty,max=60"`
	AtasNama  string `json:"atas_nama" validate:"omitempty,max=120"`
	GambarURL string `json:"gambar_url" validate:"omitempty,url,max=255"`
	Instruksi string `json:"instruksi" validate:"omitempty,max=255"`
	Urutan    int    `json:"urutan" validate:"gte=0"`
	Aktif     *bool  `json:"aktif" validate:"required"`
}

func (r SimpanMetodeRequest) keInput() usecase.InputMetode {
	return usecase.InputMetode{
		Jenis: r.Jenis, Nama: r.Nama, Nomor: r.Nomor, AtasNama: r.AtasNama,
		GambarURL: r.GambarURL, Instruksi: r.Instruksi, Urutan: r.Urutan, Aktif: *r.Aktif,
	}
}

// Daftar godoc
//
//	@Summary	Daftar metode bayar dasar (bank/e-wallet/QRIS statis)
//	@Tags		metode-bayar
//	@Produce	json
//	@Param		aktif	query		bool	false	"true = hanya yang aktif (dipakai aplikasi kasir)"
//	@Success	200		{object}	respond.Amplop{data=[]MetodeResponse}
//	@Security	BearerAuth
//	@Router		/metode-bayar [get]
func (h *MetodePembayaranHandler) Daftar(c echo.Context) error {
	hanyaAktif := c.QueryParam("aktif") == "true" || c.QueryParam("aktif") == "1"
	rows, err := h.uc.Daftar(c.Request().Context(), hanyaAktif)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	hasil := make([]MetodeResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, keMetodeResponse(&rows[i]))
	}
	return respond.Sukses(c, hasil, nil, "")
}

// Buat godoc
//
//	@Summary	Tambah metode bayar (Owner/Manager)
//	@Tags		metode-bayar
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		SimpanMetodeRequest	true	"Metode"
//	@Success	201		{object}	respond.Amplop{data=MetodeResponse}
//	@Security	BearerAuth
//	@Router		/metode-bayar [post]
func (h *MetodePembayaranHandler) Buat(c echo.Context) error {
	var req SimpanMetodeRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	m, err := h.uc.Buat(c.Request().Context(), req.keInput())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keMetodeResponse(m), "Metode pembayaran ditambahkan.")
}

// Perbarui godoc
//
//	@Summary	Ubah metode bayar (Owner/Manager)
//	@Tags		metode-bayar
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"ID"
//	@Param		payload	body		SimpanMetodeRequest	true	"Metode"
//	@Success	200		{object}	respond.Amplop{data=MetodeResponse}
//	@Security	BearerAuth
//	@Router		/metode-bayar/{id} [put]
func (h *MetodePembayaranHandler) Perbarui(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	var req SimpanMetodeRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	m, err := h.uc.Perbarui(c.Request().Context(), uint(id), req.keInput())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keMetodeResponse(m), nil, "Metode pembayaran diperbarui.")
}

// Hapus godoc
//
//	@Summary	Hapus metode bayar (riwayat transaksi tetap aman)
//	@Tags		metode-bayar
//	@Produce	json
//	@Param		id	path		int	true	"ID"
//	@Success	200	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/metode-bayar/{id} [delete]
func (h *MetodePembayaranHandler) Hapus(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	if err := h.uc.Hapus(c.Request().Context(), uint(id)); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Metode pembayaran dihapus.")
}

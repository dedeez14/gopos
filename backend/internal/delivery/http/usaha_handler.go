package http

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

// UsahaHandler — manajemen tenant (SUPERADMIN, digerbang usaha.kelola).
type UsahaHandler struct {
	usahas *usecase.UsahaUsecase
}

func NewUsahaHandler(usahas *usecase.UsahaUsecase) *UsahaHandler {
	return &UsahaHandler{usahas: usahas}
}

type UsahaResponse struct {
	ID            uint   `json:"id"`
	Kode          string `json:"kode"`
	Nama          string `json:"nama"`
	Aktif         bool   `json:"aktif"`
	MidtransAktif bool   `json:"midtrans_aktif"`
	Dibuat        string `json:"dibuat"`
}

func keUsahaResponse(u *domain.Usaha) UsahaResponse {
	return UsahaResponse{
		ID: u.ID, Kode: u.Kode, Nama: u.Nama, Aktif: u.Aktif,
		MidtransAktif: u.MidtransAktif,
		Dibuat:        u.CreatedAt.Format(time.RFC3339),
	}
}

type BuatUsahaRequest struct {
	Nama          string `json:"nama" validate:"required,min=2,max=150"`
	Kode          string `json:"kode" validate:"omitempty,max=30"`
	OwnerNama     string `json:"owner_nama" validate:"required,min=2,max=150"`
	OwnerEmail    string `json:"owner_email" validate:"required,email,max=150"`
	OwnerPassword string `json:"owner_password" validate:"required,min=8,max=72"`
}

type UbahUsahaRequest struct {
	Nama          string `json:"nama" validate:"omitempty,min=2,max=150"`
	Aktif         *bool  `json:"aktif"`
	MidtransAktif *bool  `json:"midtrans_aktif"`
}

// Daftar godoc
//
//	@Summary	Daftar usaha/merchant (SUPERADMIN)
//	@Tags		usaha
//	@Produce	json
//	@Param		q		query		string	false	"Cari nama/kode"
//	@Param		page	query		int		false	"Halaman"
//	@Success	200		{object}	respond.Amplop{data=[]UsahaResponse,meta=respond.Meta}
//	@Security	BearerAuth
//	@Router		/usahas [get]
func (h *UsahaHandler) Daftar(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))

	rows, total, err := h.usahas.Daftar(c.Request().Context(), domain.FilterUsaha{
		Cari: c.QueryParam("q"), Halaman: halaman, PerHal: perHal,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	if halaman < 1 {
		halaman = 1
	}
	if perHal < 1 || perHal > 100 {
		perHal = 20
	}
	hasil := make([]UsahaResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, keUsahaResponse(&rows[i]))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(halaman, perHal, total), "")
}

// Buat godoc
//
//	@Summary		Buat usaha BARU + akun owner-nya sekaligus (SUPERADMIN)
//	@Description	Owner langsung bisa login dan mengelola usahanya (produk, kasir, dll).
//	@Tags			usaha
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		BuatUsahaRequest	true	"Usaha + owner"
//	@Success		201		{object}	respond.Amplop{data=UsahaResponse}
//	@Failure		409		{object}	respond.Amplop	"Email owner / kode usaha sudah dipakai"
//	@Security		BearerAuth
//	@Router			/usahas [post]
func (h *UsahaHandler) Buat(c echo.Context) error {
	var req BuatUsahaRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	usaha, _, err := h.usahas.Buat(c.Request().Context(), usecase.InputUsahaBaru{
		Nama: req.Nama, Kode: req.Kode,
		OwnerNama: req.OwnerNama, OwnerEmail: req.OwnerEmail, OwnerPassword: req.OwnerPassword,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keUsahaResponse(usaha), "Usaha dibuat — owner sudah bisa masuk.")
}

// Perbarui godoc
//
//	@Summary	Ubah nama / SUSPEND usaha (aktif=false menolak login penggunanya)
//	@Tags		usaha
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"ID usaha"
//	@Param		payload	body		UbahUsahaRequest	true	"Perubahan"
//	@Success	200		{object}	respond.Amplop{data=UsahaResponse}
//	@Security	BearerAuth
//	@Router		/usahas/{id} [patch]
func (h *UsahaHandler) Perbarui(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	var req UbahUsahaRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	u, err := h.usahas.Perbarui(c.Request().Context(), uint(id), usecase.InputUbahUsaha{
		Nama: req.Nama, Aktif: req.Aktif, MidtransAktif: req.MidtransAktif,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keUsahaResponse(u), nil, "Usaha diperbarui.")
}

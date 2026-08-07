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

type InventoryHandler struct {
	inventory *usecase.InventoryUsecase
}

func NewInventoryHandler(inventory *usecase.InventoryUsecase) *InventoryHandler {
	return &InventoryHandler{inventory: inventory}
}

type StokLogResponse struct {
	ID          uint    `json:"id"`
	Produk      string  `json:"produk"`
	Jenis       string  `json:"jenis"`
	Jumlah      float64 `json:"jumlah"`
	StokSesudah float64 `json:"stok_sesudah"`
	Keterangan  string  `json:"keterangan"`
	Sesi        *string `json:"sesi"`
	Pencatat    string  `json:"pencatat"`
	Waktu       string  `json:"waktu"`
}

func keStokLogResponse(l *domain.StokLog) StokLogResponse {
	res := StokLogResponse{
		ID: l.ID, Jenis: l.Jenis, Jumlah: l.Jumlah, StokSesudah: l.StokSesudah,
		Keterangan: l.Keterangan, Waktu: l.CreatedAt.Format(time.RFC3339),
	}
	if l.Produk != nil {
		res.Produk = l.Produk.Nama
	}
	if l.User != nil {
		res.Pencatat = l.User.Nama
	}
	if l.SesiKasir != nil {
		res.Sesi = &l.SesiKasir.Nomor
	}
	return res
}

type StokMasukRequest struct {
	ProdukID   uint    `json:"produk_id" validate:"required"`
	Jumlah     float64 `json:"jumlah" validate:"required,gt=0"`
	Keterangan string  `json:"keterangan" validate:"omitempty,max=255"`
}

type OpnameRequest struct {
	ProdukID   uint     `json:"produk_id" validate:"required"`
	StokFisik  *float64 `json:"stok_fisik" validate:"required,min=0"`
	Keterangan string   `json:"keterangan" validate:"omitempty,max=255"`
}

// StokMasuk godoc
//
//	@Summary	Tambah stok (restock) — Owner/Manager
//	@Tags		inventory
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		StokMasukRequest	true	"Produk + jumlah masuk"
//	@Success	201		{object}	respond.Amplop{data=StokLogResponse}
//	@Failure	409		{object}	respond.Amplop	"Produk non-stok (jasa)"
//	@Security	BearerAuth
//	@Router		/inventory/stok-masuk [post]
func (h *InventoryHandler) StokMasuk(c echo.Context) error {
	var req StokMasukRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	log, err := h.inventory.Masuk(c.Request().Context(), userID(c), req.ProdukID, req.Jumlah, req.Keterangan)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keStokLogResponse(log), "Stok ditambahkan.")
}

// Opname godoc
//
//	@Summary		Opname (koreksi hitung fisik) — Owner/Manager
//	@Description	stok_fisik = hasil hitung; server menyimpan SELISIH-nya di riwayat.
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		OpnameRequest	true	"Produk + stok fisik"
//	@Success		201		{object}	respond.Amplop{data=StokLogResponse}
//	@Security		BearerAuth
//	@Router			/inventory/opname [post]
func (h *InventoryHandler) Opname(c echo.Context) error {
	var req OpnameRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	log, err := h.inventory.Opname(c.Request().Context(), userID(c), req.ProdukID, *req.StokFisik, req.Keterangan)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keStokLogResponse(log), "Opname tercatat.")
}

// Riwayat godoc
//
//	@Summary	Riwayat pergerakan stok (MASUK/OPNAME/JUAL/BATAL, satu buku)
//	@Tags		inventory
//	@Produce	json
//	@Param		produk_id	query		int		false	"Filter produk"
//	@Param		jenis		query		string	false	"MASUK|OPNAME|JUAL|BATAL"
//	@Param		page		query		int		false	"Halaman"
//	@Success	200			{object}	respond.Amplop{data=[]StokLogResponse,meta=respond.Meta}
//	@Security	BearerAuth
//	@Router		/inventory/riwayat [get]
func (h *InventoryHandler) Riwayat(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))
	produkID, _ := strconv.ParseUint(c.QueryParam("produk_id"), 10, 64)

	rows, total, err := h.inventory.Riwayat(c.Request().Context(), domain.FilterStokLog{
		ProdukID: uint(produkID), Jenis: c.QueryParam("jenis"),
		Halaman: halaman, PerHal: perHal,
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
	hasil := make([]StokLogResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, keStokLogResponse(&rows[i]))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(halaman, perHal, total), "")
}

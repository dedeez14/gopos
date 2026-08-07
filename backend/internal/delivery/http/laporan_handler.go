package http

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

// LaporanHandler — laporan + pengeluaran + pelanggan + hold. Satu file untuk
// layar "manajemen ringan" Tuléh: Laporan Keuangan, Customer, parkir keranjang.
type LaporanHandler struct {
	laporan     *usecase.LaporanUsecase
	pengeluaran *usecase.PengeluaranUsecase
	pelanggan   *usecase.PelangganUsecase
	hold        *usecase.HoldUsecase
}

func NewLaporanHandler(
	laporan *usecase.LaporanUsecase,
	pengeluaran *usecase.PengeluaranUsecase,
	pelanggan *usecase.PelangganUsecase,
	hold *usecase.HoldUsecase,
) *LaporanHandler {
	return &LaporanHandler{laporan: laporan, pengeluaran: pengeluaran, pelanggan: pelanggan, hold: hold}
}

// ─────────────────────────────────────────────────────────────── laporan

// Keuangan godoc
//
//	@Summary	Ringkasan keuangan bulan (omzet, per tipe bayar, pengeluaran, laba)
//	@Tags		laporan
//	@Produce	json
//	@Param		bulan	query		string	false	"YYYY-MM (kosong = bulan berjalan)"
//	@Success	200		{object}	respond.Amplop{data=usecase.RingkasanKeuangan}
//	@Security	BearerAuth
//	@Router		/laporan/keuangan [get]
func (h *LaporanHandler) Keuangan(c echo.Context) error {
	r, err := h.laporan.Keuangan(c.Request().Context(), c.QueryParam("bulan"))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, r, nil, "")
}

// PenjualanHarian godoc
//
//	@Summary	Omzet per hari (bahan grafik; default 14 hari terakhir)
//	@Tags		laporan
//	@Produce	json
//	@Param		dari	query		string	false	"YYYY-MM-DD"
//	@Param		sampai	query		string	false	"YYYY-MM-DD"
//	@Success	200		{object}	respond.Amplop{data=[]domain.PenjualanHarian}
//	@Security	BearerAuth
//	@Router		/laporan/penjualan-harian [get]
func (h *LaporanHandler) PenjualanHarian(c echo.Context) error {
	rows, err := h.laporan.PenjualanHarian(c.Request().Context(), c.QueryParam("dari"), c.QueryParam("sampai"))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	if rows == nil {
		rows = []domain.PenjualanHarian{}
	}
	return respond.Sukses(c, rows, nil, "")
}

// ProdukTerlaris godoc
//
//	@Summary	Produk paling laku (dari snapshot item transaksi)
//	@Tags		laporan
//	@Produce	json
//	@Param		hari	query		int	false	"Rentang hari (default 30, maks 365)"
//	@Param		limit	query		int	false	"Jumlah baris (default 10, maks 50)"
//	@Success	200		{object}	respond.Amplop{data=[]domain.ProdukTerlaris}
//	@Security	BearerAuth
//	@Router		/laporan/produk-terlaris [get]
func (h *LaporanHandler) ProdukTerlaris(c echo.Context) error {
	hari, _ := strconv.Atoi(c.QueryParam("hari"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	rows, err := h.laporan.ProdukTerlaris(c.Request().Context(), hari, limit)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	if rows == nil {
		rows = []domain.ProdukTerlaris{}
	}
	return respond.Sukses(c, rows, nil, "")
}

// ─────────────────────────────────────────────────────────── pengeluaran

type PengeluaranResponse struct {
	ID         uint    `json:"id"`
	Tanggal    string  `json:"tanggal"`
	Keterangan string  `json:"keterangan"`
	Nominal    float64 `json:"nominal"`
	Pencatat   string  `json:"pencatat"`
}

func kePengeluaranResponse(p *domain.Pengeluaran) PengeluaranResponse {
	res := PengeluaranResponse{
		ID: p.ID, Tanggal: p.Tanggal.Format("2006-01-02"),
		Keterangan: p.Keterangan, Nominal: p.Nominal,
	}
	if p.User != nil {
		res.Pencatat = p.User.Nama
	}
	return res
}

type CatatPengeluaranRequest struct {
	Tanggal    string  `json:"tanggal" validate:"omitempty,datetime=2006-01-02"` // kosong = hari ini
	Keterangan string  `json:"keterangan" validate:"required,min=3,max=255"`
	Nominal    float64 `json:"nominal" validate:"required,gt=0"`
}

// CatatPengeluaran godoc
//
//	@Summary	Catat pengeluaran operasional (dua kolom: untuk apa, berapa)
//	@Tags		laporan
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		CatatPengeluaranRequest	true	"Pengeluaran"
//	@Success	201		{object}	respond.Amplop{data=PengeluaranResponse}
//	@Security	BearerAuth
//	@Router		/pengeluaran [post]
func (h *LaporanHandler) CatatPengeluaran(c echo.Context) error {
	var req CatatPengeluaranRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	tanggal := time.Now()
	if t, err := time.Parse("2006-01-02", req.Tanggal); err == nil {
		tanggal = t
	}
	p, err := h.pengeluaran.Catat(c.Request().Context(), userID(c), tanggal, req.Keterangan, req.Nominal)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, kePengeluaranResponse(p), "Pengeluaran tercatat.")
}

// DaftarPengeluaran godoc
//
//	@Summary	Daftar pengeluaran (filter bulan)
//	@Tags		laporan
//	@Produce	json
//	@Param		bulan	query		string	false	"YYYY-MM"
//	@Param		page	query		int		false	"Halaman"
//	@Success	200		{object}	respond.Amplop{data=[]PengeluaranResponse,meta=respond.Meta}
//	@Security	BearerAuth
//	@Router		/pengeluaran [get]
func (h *LaporanHandler) DaftarPengeluaran(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))

	rows, total, err := h.pengeluaran.Daftar(c.Request().Context(), domain.FilterPengeluaran{
		Bulan: c.QueryParam("bulan"), Halaman: halaman, PerHal: perHal,
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
	hasil := make([]PengeluaranResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, kePengeluaranResponse(&rows[i]))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(halaman, perHal, total), "")
}

// HapusPengeluaran godoc
//
//	@Summary	Hapus catatan pengeluaran
//	@Tags		laporan
//	@Produce	json
//	@Param		id	path		int	true	"ID"
//	@Success	200	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/pengeluaran/{id} [delete]
func (h *LaporanHandler) HapusPengeluaran(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	if err := h.pengeluaran.Hapus(c.Request().Context(), uint(id)); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Pengeluaran dihapus.")
}

// ─────────────────────────────────────────────────────────────── pelanggan

type PelangganResponse struct {
	ID      uint    `json:"id"`
	Nama    string  `json:"nama"`
	Telepon *string `json:"telepon"`
	WaLink  *string `json:"wa_link"`
	Email   *string `json:"email"`
	Catatan string  `json:"catatan"`
	Aktif   bool    `json:"aktif"`
}

func kePelangganResponse(p *domain.Pelanggan) PelangganResponse {
	res := PelangganResponse{
		ID: p.ID, Nama: p.Nama, Telepon: p.Telepon, Email: p.Email,
		Catatan: p.Catatan, Aktif: p.Aktif,
	}
	if p.Telepon != nil {
		wa := "https://wa.me/" + *p.Telepon
		res.WaLink = &wa
	}
	return res
}

type QuickPelangganRequest struct {
	Nama    string `json:"nama" validate:"required,min=2,max=150"`
	Telepon string `json:"telepon" validate:"omitempty,max=25"`
}

// QuickPelanggan godoc
//
//	@Summary		Quick add pelanggan (nama + WA) — alat kasir
//	@Description	Telepon dinormalkan ke 62…; nomor yang sudah terdaftar mengembalikan pelanggan LAMA (bukan galat).
//	@Tags			pelanggan
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		QuickPelangganRequest	true	"Nama + telepon"
//	@Success		201		{object}	respond.Amplop{data=PelangganResponse}
//	@Security		BearerAuth
//	@Router			/pelanggan/quick [post]
func (h *LaporanHandler) QuickPelanggan(c echo.Context) error {
	var req QuickPelangganRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	p, baru, err := h.pelanggan.Quick(c.Request().Context(), req.Nama, req.Telepon)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	pesan := "Pelanggan tersimpan."
	if !baru {
		pesan = "Nomor sudah terdaftar — memakai pelanggan yang ada."
	}
	return respond.Dibuat(c, kePelangganResponse(p), pesan)
}

type SimpanPelangganRequest struct {
	Nama    string `json:"nama" validate:"required,min=2,max=150"`
	Telepon string `json:"telepon" validate:"omitempty,max=25"`
	Email   string `json:"email" validate:"omitempty,email,max=150"`
	Catatan string `json:"catatan" validate:"omitempty,max=255"`
	Aktif   *bool  `json:"aktif" validate:"required"`
}

// DaftarPelanggan godoc
//
//	@Summary	Daftar pelanggan (cari nama/telepon)
//	@Tags		pelanggan
//	@Produce	json
//	@Param		q		query		string	false	"Cari"
//	@Param		page	query		int		false	"Halaman"
//	@Success	200		{object}	respond.Amplop{data=[]PelangganResponse,meta=respond.Meta}
//	@Security	BearerAuth
//	@Router		/pelanggan [get]
func (h *LaporanHandler) DaftarPelanggan(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))

	rows, total, err := h.pelanggan.Daftar(c.Request().Context(), domain.FilterPelanggan{
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
	hasil := make([]PelangganResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, kePelangganResponse(&rows[i]))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(halaman, perHal, total), "")
}

// PerbaruiPelanggan godoc
//
//	@Summary	Ubah pelanggan (Owner/Manager)
//	@Tags		pelanggan
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int						true	"ID"
//	@Param		payload	body		SimpanPelangganRequest	true	"Data pelanggan"
//	@Success	200		{object}	respond.Amplop{data=PelangganResponse}
//	@Failure	409		{object}	respond.Amplop	"Telepon dipakai pelanggan lain"
//	@Security	BearerAuth
//	@Router		/pelanggan/{id} [put]
func (h *LaporanHandler) PerbaruiPelanggan(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	var req SimpanPelangganRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	p, err := h.pelanggan.Perbarui(c.Request().Context(), uint(id), usecase.InputPelanggan{
		Nama: req.Nama, Telepon: req.Telepon, Email: req.Email,
		Catatan: req.Catatan, Aktif: *req.Aktif,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, kePelangganResponse(p), nil, "Pelanggan diperbarui.")
}

// HapusPelanggan godoc
//
//	@Summary	Nonaktifkan pelanggan (riwayat transaksi terjaga)
//	@Tags		pelanggan
//	@Produce	json
//	@Param		id	path		int	true	"ID"
//	@Success	200	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/pelanggan/{id} [delete]
func (h *LaporanHandler) HapusPelanggan(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	if err := h.pelanggan.Nonaktifkan(c.Request().Context(), uint(id)); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Pelanggan dinonaktifkan.")
}

// ─────────────────────────────────────────────────────────────────── hold

type HoldResponse struct {
	ID      uint            `json:"id"`
	Label   string          `json:"label"`
	Payload json.RawMessage `json:"payload" swaggertype:"object"`
	Kasir   string          `json:"kasir"`
	Dibuat  string          `json:"dibuat"`
}

func keHoldResponse(h *domain.Hold) HoldResponse {
	res := HoldResponse{
		ID: h.ID, Label: h.Label, Payload: json.RawMessage(h.Payload),
		Dibuat: h.CreatedAt.Format(time.RFC3339),
	}
	if h.User != nil {
		res.Kasir = h.User.Nama
	}
	return res
}

type SimpanHoldRequest struct {
	Label   string          `json:"label" validate:"omitempty,max=80"`
	Payload json.RawMessage `json:"payload" validate:"required" swaggertype:"object"`
}

// SimpanHold godoc
//
//	@Summary		Parkir keranjang (hold) — tanpa efek uang/stok
//	@Description	Payload opaque milik aplikasi kasir (maks 64KB, maks 50 hold). Resume = muat payload → checkout biasa → hapus hold.
//	@Tags			pelanggan
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		SimpanHoldRequest	true	"Label + keranjang"
//	@Success		201		{object}	respond.Amplop{data=HoldResponse}
//	@Failure		409		{object}	respond.Amplop	"Hold penuh"
//	@Security		BearerAuth
//	@Router			/hold [post]
func (h *LaporanHandler) SimpanHold(c echo.Context) error {
	var req SimpanHoldRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	if len(req.Payload) > 65536 {
		return respond.Gagal(c, 422, "Payload hold terlalu besar (maks 64KB).", nil)
	}
	hold, err := h.hold.Simpan(c.Request().Context(), userID(c), req.Label, req.Payload)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keHoldResponse(hold), "Transaksi ditunda.")
}

// DaftarHold godoc
//
//	@Summary	Daftar hold (lintas kasir — hold kasir A bisa dilanjutkan kasir B)
//	@Tags		pelanggan
//	@Produce	json
//	@Success	200	{object}	respond.Amplop{data=[]HoldResponse}
//	@Security	BearerAuth
//	@Router		/hold [get]
func (h *LaporanHandler) DaftarHold(c echo.Context) error {
	rows, err := h.hold.Daftar(c.Request().Context())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	hasil := make([]HoldResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, keHoldResponse(&rows[i]))
	}
	return respond.Sukses(c, hasil, nil, "")
}

// HapusHold godoc
//
//	@Summary	Hapus hold (setelah resume sukses atau dibuang)
//	@Tags		pelanggan
//	@Produce	json
//	@Param		id	path		int	true	"ID hold"
//	@Success	200	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/hold/{id} [delete]
func (h *LaporanHandler) HapusHold(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	if err := h.hold.Hapus(c.Request().Context(), uint(id)); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Hold dihapus.")
}

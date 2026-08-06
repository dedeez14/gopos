package http

import (
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/delivery/http/middleware"
	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

type ProdukHandler struct {
	produk   *usecase.ProdukUsecase
	kategori *usecase.KategoriUsecase
}

func NewProdukHandler(produk *usecase.ProdukUsecase, kategori *usecase.KategoriUsecase) *ProdukHandler {
	return &ProdukHandler{produk: produk, kategori: kategori}
}

// ProdukResponse — kontrak API katalog, kompatibel klien Tuléh:
// tipe PRODUK|JASA (bukan enum DB), harga_efektif = harga hari ini,
// harga_beli hanya untuk manajemen (rahasia dagang — kasir menerima null).
type ProdukResponse struct {
	ID           uint     `json:"id"`
	Kode         string   `json:"kode"`
	Nama         string   `json:"nama"`
	Barcode      *string  `json:"barcode"`
	Tipe         string   `json:"tipe"` // PRODUK | JASA
	Satuan       string   `json:"satuan"`
	HargaBeli    *float64 `json:"harga_beli"` // null untuk kasir
	HargaJual    float64  `json:"harga_jual"`
	HargaPromo   *float64 `json:"harga_promo"`
	PromoAktif   bool     `json:"promo_aktif"`
	PromoMulai   *string  `json:"promo_mulai"`
	PromoSelesai *string  `json:"promo_selesai"`
	HargaEfektif float64  `json:"harga_efektif"`
	Favorit      bool     `json:"favorit"`
	KelolaStok   bool     `json:"kelola_stok"`
	Stok         float64  `json:"stok"`
	Kategori     *string  `json:"kategori"`
	KategoriID   *uint    `json:"kategori_id"`
	Aktif        bool     `json:"aktif"`
}

func keProdukResponse(p *domain.Produk, denganModal bool) ProdukResponse {
	kini := time.Now()
	res := ProdukResponse{
		ID: p.ID, Kode: p.Kode, Nama: p.Nama, Barcode: p.Barcode,
		Tipe:   "PRODUK",
		Satuan: p.Satuan, HargaJual: p.HargaJual,
		HargaPromo: p.HargaPromo, PromoAktif: p.PromoAktif(kini),
		HargaEfektif: p.HargaEfektif(kini),
		Favorit:      p.Favorit, KelolaStok: p.KelolaStok, Stok: p.Stok,
		KategoriID: p.KategoriID, Aktif: p.Aktif,
	}
	if p.Tipe == domain.TipeJasa {
		res.Tipe = "JASA"
	}
	if denganModal {
		hb := p.HargaBeli
		res.HargaBeli = &hb
	}
	if p.Kategori != nil {
		res.Kategori = &p.Kategori.Nama
	}
	if p.PromoMulai != nil {
		s := p.PromoMulai.Format("2006-01-02")
		res.PromoMulai = &s
	}
	if p.PromoSelesai != nil {
		s := p.PromoSelesai.Format("2006-01-02")
		res.PromoSelesai = &s
	}
	return res
}

// manajemen: peran pemohon boleh melihat modal / data manajemen?
func manajemen(c echo.Context) bool {
	role, _ := c.Get(middleware.CtxRole).(domain.Role)
	return role.Punya(domain.PermProdukKelola)
}

// SimpanProdukRequest dipakai create & update.
type SimpanProdukRequest struct {
	Nama         string   `json:"nama" validate:"required,min=2,max=150"`
	Kode         string   `json:"kode" validate:"omitempty,max=30"`
	Barcode      string   `json:"barcode" validate:"omitempty,max=60"`
	Tipe         string   `json:"tipe" validate:"required,oneof=PRODUK JASA"`
	Satuan       string   `json:"satuan" validate:"omitempty,max=30"`
	HargaBeli    float64  `json:"harga_beli" validate:"min=0"`
	HargaJual    float64  `json:"harga_jual" validate:"required,min=0"`
	HargaPromo   *float64 `json:"harga_promo" validate:"omitempty,min=0"`
	PromoMulai   string   `json:"promo_mulai" validate:"omitempty,datetime=2006-01-02"`
	PromoSelesai string   `json:"promo_selesai" validate:"omitempty,datetime=2006-01-02"`
	Favorit      bool     `json:"favorit"`
	KelolaStok   *bool    `json:"kelola_stok"`
	KategoriID   uint     `json:"kategori_id"`
	Aktif        *bool    `json:"aktif" validate:"required"`
}

func (r *SimpanProdukRequest) keInput() usecase.InputProduk {
	in := usecase.InputProduk{
		Nama: r.Nama, Kode: r.Kode, Barcode: r.Barcode,
		Satuan:    r.Satuan,
		HargaBeli: r.HargaBeli, HargaJual: r.HargaJual, HargaPromo: r.HargaPromo,
		Favorit: r.Favorit, KategoriID: r.KategoriID,
		Aktif:      *r.Aktif,
		KelolaStok: true,
		Tipe:       domain.TipeBarang,
	}
	if strings.EqualFold(r.Tipe, "JASA") {
		in.Tipe = domain.TipeJasa
	}
	if r.KelolaStok != nil {
		in.KelolaStok = *r.KelolaStok
	}
	if t, err := time.Parse("2006-01-02", r.PromoMulai); err == nil {
		in.PromoMulai = &t
	}
	if t, err := time.Parse("2006-01-02", r.PromoSelesai); err == nil {
		in.PromoSelesai = &t
	}
	return in
}

// Daftar godoc
//
//	@Summary		Katalog produk (paginasi + filter)
//	@Description	Kasir: hanya item aktif, tanpa harga_beli. Manajemen: + ?termasuk_nonaktif=1 dan harga_beli terisi.
//	@Tags			produk
//	@Produce		json
//	@Param			q					query		string	false	"Cari nama/kode/barcode"
//	@Param			tipe				query		string	false	"PRODUK | JASA (kosong = semua)"
//	@Param			kategori_id			query		int		false	"Filter kategori"
//	@Param			favorit				query		bool	false	"Hanya favorit"
//	@Param			promo				query		bool	false	"Hanya promo aktif hari ini"
//	@Param			termasuk_nonaktif	query		bool	false	"Hanya dihormati untuk manajemen"
//	@Param			page				query		int		false	"Halaman"
//	@Param			per_page			query		int		false	"Baris per halaman (maks 100)"
//	@Success		200					{object}	respond.Amplop{data=[]ProdukResponse,meta=respond.Meta}
//	@Security		BearerAuth
//	@Router			/produk [get]
func (h *ProdukHandler) Daftar(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))
	kategoriID, _ := strconv.ParseUint(c.QueryParam("kategori_id"), 10, 64)

	var tipe domain.TipeProduk
	switch strings.ToUpper(c.QueryParam("tipe")) {
	case "PRODUK":
		tipe = domain.TipeBarang
	case "JASA":
		tipe = domain.TipeJasa
	}

	adalahManajemen := manajemen(c)
	f := domain.FilterProduk{
		Cari:         c.QueryParam("q"),
		KategoriID:   uint(kategoriID),
		Tipe:         tipe,
		HanyaFavorit: c.QueryParam("favorit") == "1" || c.QueryParam("favorit") == "true",
		HanyaPromo:   c.QueryParam("promo") == "1" || c.QueryParam("promo") == "true",
		// Item nonaktif hanya untuk layar manajemen — katalog kasir bersih.
		TermasukNonaktif: adalahManajemen && (c.QueryParam("termasuk_nonaktif") == "1" || c.QueryParam("termasuk_nonaktif") == "true"),
		Halaman:          halaman,
		PerHal:           perHal,
	}

	rows, total, err := h.produk.Daftar(c.Request().Context(), f)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}

	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}

	hasil := make([]ProdukResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, keProdukResponse(&rows[i], adalahManajemen))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(f.Halaman, f.PerHal, total), "")
}

// Ambil godoc
//
//	@Summary	Detail produk
//	@Tags		produk
//	@Produce	json
//	@Param		id	path		int	true	"ID produk"
//	@Success	200	{object}	respond.Amplop{data=ProdukResponse}
//	@Failure	404	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/produk/{id} [get]
func (h *ProdukHandler) Ambil(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	p, err := h.produk.Ambil(c.Request().Context(), uint(id))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keProdukResponse(p, manajemen(c)), nil, "")
}

// Buat godoc
//
//	@Summary	Tambah produk (Owner/Manager)
//	@Tags		produk
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		SimpanProdukRequest	true	"Data produk"
//	@Success	201		{object}	respond.Amplop{data=ProdukResponse}
//	@Failure	409		{object}	respond.Amplop	"Kode sudah dipakai"
//	@Security	BearerAuth
//	@Router		/produk [post]
func (h *ProdukHandler) Buat(c echo.Context) error {
	var req SimpanProdukRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	p, err := h.produk.Buat(c.Request().Context(), req.keInput())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keProdukResponse(p, true), "Produk tersimpan.")
}

// Perbarui godoc
//
//	@Summary	Ubah produk (Owner/Manager)
//	@Tags		produk
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"ID produk"
//	@Param		payload	body		SimpanProdukRequest	true	"Data produk (harga_promo null = cabut promo)"
//	@Success	200		{object}	respond.Amplop{data=ProdukResponse}
//	@Failure	404		{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/produk/{id} [put]
func (h *ProdukHandler) Perbarui(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	var req SimpanProdukRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	p, err := h.produk.Perbarui(c.Request().Context(), uint(id), req.keInput())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keProdukResponse(p, true), nil, "Produk diperbarui.")
}

// Hapus godoc
//
//	@Summary	Nonaktifkan produk (bukan hapus baris — riwayat transaksi terjaga)
//	@Tags		produk
//	@Produce	json
//	@Param		id	path		int	true	"ID produk"
//	@Success	200	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/produk/{id} [delete]
func (h *ProdukHandler) Hapus(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	if err := h.produk.Nonaktifkan(c.Request().Context(), uint(id)); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Produk dinonaktifkan.")
}

// ─────────────────────────────────────────────────────────────── kategori

type KategoriResponse struct {
	ID   uint   `json:"id"`
	Nama string `json:"nama"`
}

type SimpanKategoriRequest struct {
	Nama string `json:"nama" validate:"required,min=2,max=100"`
}

// DaftarKategori godoc
//
//	@Summary	Daftar kategori
//	@Tags		produk
//	@Produce	json
//	@Success	200	{object}	respond.Amplop{data=[]KategoriResponse}
//	@Security	BearerAuth
//	@Router		/kategori [get]
func (h *ProdukHandler) DaftarKategori(c echo.Context) error {
	rows, err := h.kategori.Daftar(c.Request().Context())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	hasil := make([]KategoriResponse, 0, len(rows))
	for _, k := range rows {
		hasil = append(hasil, KategoriResponse{ID: k.ID, Nama: k.Nama})
	}
	return respond.Sukses(c, hasil, nil, "")
}

// BuatKategori godoc
//
//	@Summary	Tambah kategori (Owner/Manager)
//	@Tags		produk
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		SimpanKategoriRequest	true	"Nama kategori"
//	@Success	201		{object}	respond.Amplop{data=KategoriResponse}
//	@Security	BearerAuth
//	@Router		/kategori [post]
func (h *ProdukHandler) BuatKategori(c echo.Context) error {
	var req SimpanKategoriRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	k, err := h.kategori.Buat(c.Request().Context(), req.Nama)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, KategoriResponse{ID: k.ID, Nama: k.Nama}, "Kategori tersimpan.")
}

// HapusKategori godoc
//
//	@Summary	Hapus kategori (produk pemakainya dilepas, tidak ikut terhapus)
//	@Tags		produk
//	@Produce	json
//	@Param		id	path		int	true	"ID kategori"
//	@Success	200	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/kategori/{id} [delete]
func (h *ProdukHandler) HapusKategori(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	if err := h.kategori.Hapus(c.Request().Context(), uint(id)); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Kategori dihapus.")
}

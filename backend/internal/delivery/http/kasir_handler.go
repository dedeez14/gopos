package http

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/delivery/http/middleware"
	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

// KasirHandler — sesi kasir + checkout + riwayat transaksi.
type KasirHandler struct {
	sesi      *usecase.SesiUsecase
	transaksi *usecase.TransaksiUsecase
}

func NewKasirHandler(sesi *usecase.SesiUsecase, transaksi *usecase.TransaksiUsecase) *KasirHandler {
	return &KasirHandler{sesi: sesi, transaksi: transaksi}
}

func userID(c echo.Context) uint {
	id, _ := c.Get(middleware.CtxUserID).(uint)
	return id
}

// ─────────────────────────────────────────────────────────────── DTO

type SesiResponse struct {
	ID          uint               `json:"id"`
	Nomor       string             `json:"nomor"`
	Kasir       string             `json:"kasir"`
	Status      string             `json:"status"`
	KasAwal     float64            `json:"kas_awal"`
	KasAkhir    *float64           `json:"kas_akhir"`
	KasSistem   *float64           `json:"kas_sistem"`
	Selisih     *float64           `json:"selisih"`
	Catatan     string             `json:"catatan"`
	DibukaPada  string             `json:"dibuka_pada"`
	DitutupPada *string            `json:"ditutup_pada"`
	Total       map[string]float64 `json:"total,omitempty"` // omzet per tipe bayar (rekap)
}

func keSesiResponse(s *domain.SesiKasir, total map[string]float64) SesiResponse {
	res := SesiResponse{
		ID: s.ID, Nomor: s.Nomor, Status: s.Status,
		KasAwal: s.KasAwal, KasAkhir: s.KasAkhir, KasSistem: s.KasSistem, Selisih: s.Selisih,
		Catatan: s.Catatan, DibukaPada: s.DibukaPada.Format(time.RFC3339), Total: total,
	}
	if s.User != nil {
		res.Kasir = s.User.Nama
	}
	if s.DitutupPada != nil {
		t := s.DitutupPada.Format(time.RFC3339)
		res.DitutupPada = &t
	}
	return res
}

type ItemStruk struct {
	NamaProduk   string  `json:"nama_produk"`
	Satuan       string  `json:"satuan"`
	Harga        float64 `json:"harga"`
	Kuantitas    float64 `json:"kuantitas"`
	DiskonPersen float64 `json:"diskon_persen"`
	PajakPersen  float64 `json:"pajak_persen"`
	Subtotal     float64 `json:"subtotal"`
}

// StrukResponse — bentuk sama dengan struk Tuléh di MOVERA (klien tak boleh
// bisa membedakan siapa yang menjawab saat cutover).
type StrukResponse struct {
	ID             uint        `json:"id"`
	Nomor          string      `json:"nomor"`
	Tanggal        string      `json:"tanggal"`
	Status         string      `json:"status"`
	Kasir          string      `json:"kasir"`
	Pelanggan      string      `json:"pelanggan"`
	TipePembayaran string      `json:"tipe_pembayaran"`
	Subtotal       float64     `json:"subtotal"`
	TotalDiskon    float64     `json:"total_diskon"`
	DiskonPersen   float64     `json:"diskon_persen"`
	DiskonNominal  float64     `json:"diskon_nominal"`
	TotalPajak     float64     `json:"total_pajak"`
	Pembulatan     float64     `json:"pembulatan"`
	GrandTotal     float64     `json:"grand_total"`
	Dibayar        float64     `json:"dibayar"`
	Kembalian      float64     `json:"kembalian"`
	Catatan        string      `json:"catatan"`
	Items          []ItemStruk `json:"items"`
}

func keStruk(t *domain.Transaksi) StrukResponse {
	res := StrukResponse{
		ID: t.ID, Nomor: t.Nomor, Tanggal: t.Tanggal.Format(time.RFC3339), Status: t.Status,
		TipePembayaran: t.TipePembayaran,
		Subtotal:       t.Subtotal, TotalDiskon: t.TotalDiskon,
		DiskonPersen: t.DiskonPersen, DiskonNominal: t.DiskonNominal,
		TotalPajak: t.TotalPajak, Pembulatan: t.Pembulatan, GrandTotal: t.GrandTotal,
		Dibayar: t.Dibayar, Kembalian: t.Kembalian, Catatan: t.Catatan,
	}
	if t.User != nil {
		res.Kasir = t.User.Nama
	}
	if t.Pelanggan != nil {
		res.Pelanggan = t.Pelanggan.Nama
	}
	for _, it := range t.Items {
		res.Items = append(res.Items, ItemStruk{
			NamaProduk: it.NamaProduk, Satuan: it.Satuan, Harga: it.Harga,
			Kuantitas: it.Kuantitas, DiskonPersen: it.DiskonPersen,
			PajakPersen: it.PajakPersen, Subtotal: it.Subtotal,
		})
	}
	return res
}

// ─────────────────────────────────────────────────────────────── sesi

type BukaSesiRequest struct {
	KasAwal float64 `json:"kas_awal" validate:"min=0"`
	Catatan string  `json:"catatan" validate:"omitempty,max=255"`
}

type TutupSesiRequest struct {
	SesiID   uint    `json:"sesi_id" validate:"required"`
	KasAkhir float64 `json:"kas_akhir" validate:"min=0"`
	Catatan  string  `json:"catatan" validate:"omitempty,max=255"`
}

// BukaSesi godoc
//
//	@Summary	Buka sesi kasir (satu sesi BUKA per pengguna)
//	@Tags		kasir
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		BukaSesiRequest	true	"Kas awal laci"
//	@Success	201		{object}	respond.Amplop{data=SesiResponse}
//	@Failure	409		{object}	respond.Amplop	"Masih ada sesi terbuka"
//	@Security	BearerAuth
//	@Router		/sesi/buka [post]
func (h *KasirHandler) BukaSesi(c echo.Context) error {
	var req BukaSesiRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	s, err := h.sesi.Buka(c.Request().Context(), userID(c), req.KasAwal, req.Catatan)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keSesiResponse(s, nil), "Sesi kasir dibuka.")
}

// SesiAktif godoc
//
//	@Summary	Sesi BUKA milik pengguna saat ini (404 bila tidak ada)
//	@Tags		kasir
//	@Produce	json
//	@Success	200	{object}	respond.Amplop{data=SesiResponse}
//	@Failure	404	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/sesi/aktif [get]
func (h *KasirHandler) SesiAktif(c echo.Context) error {
	s, err := h.sesi.Aktif(c.Request().Context(), userID(c))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keSesiResponse(s, nil), nil, "")
}

// TutupSesi godoc
//
//	@Summary	Tutup sesi (hitung kas sistem = kas awal + tunai, dan selisihnya)
//	@Tags		kasir
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		TutupSesiRequest	true	"Kas fisik hasil hitung kasir"
//	@Success	200		{object}	respond.Amplop{data=SesiResponse}
//	@Failure	409		{object}	respond.Amplop	"Sesi sudah ditutup"
//	@Security	BearerAuth
//	@Router		/sesi/tutup [post]
func (h *KasirHandler) TutupSesi(c echo.Context) error {
	var req TutupSesiRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	s, err := h.sesi.Tutup(c.Request().Context(), req.SesiID, userID(c), req.KasAkhir, req.Catatan)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	// Balasan tutup sekaligus rekap — kasir langsung melihat omzet per tipe.
	rekap, total, err := h.sesi.Rekap(c.Request().Context(), s.ID, userID(c), true)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keSesiResponse(rekap, total), nil, "Sesi ditutup.")
}

// RekapSesi godoc
//
//	@Summary	Rekap sesi (pemilik sesi, atau peran ber-izin laporan)
//	@Tags		kasir
//	@Produce	json
//	@Param		id	path		int	true	"ID sesi"
//	@Success	200	{object}	respond.Amplop{data=SesiResponse}
//	@Failure	403	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/sesi/{id}/rekap [get]
func (h *KasirHandler) RekapSesi(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	role, _ := c.Get(middleware.CtxRole).(domain.Role)
	s, total, err := h.sesi.Rekap(c.Request().Context(), uint(id), userID(c), role.Punya(domain.PermLaporan))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keSesiResponse(s, total), nil, "")
}

// DaftarSesi godoc
//
//	@Summary	Daftar sesi (laporan — Owner/Manager)
//	@Tags		kasir
//	@Produce	json
//	@Param		page	query		int	false	"Halaman"
//	@Success	200		{object}	respond.Amplop{data=[]SesiResponse,meta=respond.Meta}
//	@Security	BearerAuth
//	@Router		/sesi [get]
func (h *KasirHandler) DaftarSesi(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))

	rows, total, err := h.sesi.Daftar(c.Request().Context(), domain.FilterSesi{
		Status: c.QueryParam("status"), Halaman: halaman, PerHal: perHal,
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
	hasil := make([]SesiResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, keSesiResponse(&rows[i], nil))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(halaman, perHal, total), "")
}

// ─────────────────────────────────────────────────────────────── transaksi

type ItemCheckoutRequest struct {
	ProdukID     uint    `json:"produk_id" validate:"required"`
	Kuantitas    float64 `json:"kuantitas" validate:"required,gt=0"`
	Harga        float64 `json:"harga" validate:"min=0"` // 0 = pakai harga efektif katalog
	DiskonPersen float64 `json:"diskon_persen" validate:"min=0,max=100"`
	PajakPersen  float64 `json:"pajak_persen" validate:"min=0,max=100"`
}

type CheckoutRequest struct {
	Items          []ItemCheckoutRequest `json:"items" validate:"required,min=1,dive"`
	DiskonPersen   float64               `json:"diskon_persen" validate:"min=0,max=100"`
	DiskonNominal  float64               `json:"diskon_nominal" validate:"min=0"`
	TipePembayaran string                `json:"tipe_pembayaran" validate:"required,oneof=TUNAI TRANSFER QRIS"`
	Dibayar        float64               `json:"dibayar" validate:"min=0"`
	Catatan        string                `json:"catatan" validate:"omitempty,max=500"`
	IdempotencyKey string                `json:"idempotency_key" validate:"omitempty,max=80"`
	PelangganID    uint                  `json:"pelanggan_id"`
}

// Checkout godoc
//
//	@Summary		Checkout penjualan (wajib sesi BUKA)
//	@Description	Idempotency-Key (header atau body) yang sama tidak memproses dua kali. harga=0 → server memakai harga efektif katalog (promo dihormati).
//	@Tags			kasir
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CheckoutRequest	true	"Keranjang"
//	@Success		201		{object}	respond.Amplop{data=StrukResponse}
//	@Failure		409		{object}	respond.Amplop	"Belum ada sesi terbuka"
//	@Security		BearerAuth
//	@Router			/transaksi/checkout [post]
func (h *KasirHandler) Checkout(c echo.Context) error {
	var req CheckoutRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}

	kunci := req.IdempotencyKey
	if v := c.Request().Header.Get("Idempotency-Key"); v != "" {
		kunci = v
	}

	items := make([]usecase.ItemCheckout, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, usecase.ItemCheckout{
			ProdukID: it.ProdukID, Kuantitas: it.Kuantitas, Harga: it.Harga,
			DiskonPersen: it.DiskonPersen, PajakPersen: it.PajakPersen,
		})
	}

	t, err := h.transaksi.Checkout(c.Request().Context(), userID(c), usecase.InputCheckout{
		Items: items, DiskonPersen: req.DiskonPersen, DiskonNominal: req.DiskonNominal,
		TipePembayaran: req.TipePembayaran, Dibayar: req.Dibayar,
		Catatan: req.Catatan, IdempotencyKey: kunci, PelangganID: req.PelangganID,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keStruk(t), "Transaksi tersimpan.")
}

// DaftarTransaksi godoc
//
//	@Summary	Riwayat transaksi (paginasi + filter tanggal/sesi)
//	@Tags		kasir
//	@Produce	json
//	@Param		sesi_id			query		int		false	"Filter sesi"
//	@Param		tanggal_dari	query		string	false	"YYYY-MM-DD"
//	@Param		tanggal_sampai	query		string	false	"YYYY-MM-DD"
//	@Param		page			query		int		false	"Halaman"
//	@Success	200				{object}	respond.Amplop{data=[]StrukResponse,meta=respond.Meta}
//	@Security	BearerAuth
//	@Router		/transaksi [get]
func (h *KasirHandler) DaftarTransaksi(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))
	sesiID, _ := strconv.ParseUint(c.QueryParam("sesi_id"), 10, 64)

	rows, total, err := h.transaksi.Daftar(c.Request().Context(), domain.FilterTransaksi{
		SesiKasirID: uint(sesiID),
		TanggalDari: c.QueryParam("tanggal_dari"), TanggalAkhir: c.QueryParam("tanggal_sampai"),
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
	hasil := make([]StrukResponse, 0, len(rows))
	for i := range rows {
		hasil = append(hasil, keStruk(&rows[i]))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(halaman, perHal, total), "")
}

// BatalTransaksi godoc
//
//	@Summary		Batalkan transaksi (void) — stok kembali, status DIBATALKAN
//	@Description	Hanya selama sesi transaksi masih BUKA. Kasir: transaksi sendiri; peran ber-izin laporan: milik siapa pun.
//	@Tags			kasir
//	@Produce		json
//	@Param			id	path		int	true	"ID transaksi"
//	@Success		200	{object}	respond.Amplop{data=StrukResponse}
//	@Failure		409	{object}	respond.Amplop	"Sudah dibatalkan / sesi sudah tutup"
//	@Security		BearerAuth
//	@Router			/transaksi/{id}/batal [post]
func (h *KasirHandler) BatalTransaksi(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	role, _ := c.Get(middleware.CtxRole).(domain.Role)
	t, err := h.transaksi.Batal(c.Request().Context(), uint(id), userID(c), role.Punya(domain.PermLaporan))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keStruk(t), nil, "Transaksi dibatalkan.")
}

// AmbilTransaksi godoc
//
//	@Summary	Struk satu transaksi
//	@Tags		kasir
//	@Produce	json
//	@Param		id	path		int	true	"ID transaksi"
//	@Success	200	{object}	respond.Amplop{data=StrukResponse}
//	@Failure	404	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/transaksi/{id} [get]
func (h *KasirHandler) AmbilTransaksi(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	t, err := h.transaksi.Ambil(c.Request().Context(), uint(id))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keStruk(t), nil, "")
}

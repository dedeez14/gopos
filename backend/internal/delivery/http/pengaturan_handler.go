package http

import (
	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

// PengaturanHandler — konfigurasi usaha (profil toko, struk, pajak). BACA
// terbuka untuk kasir (aplikasi butuh header/footer struk & pajak default);
// TULIS khusus Owner/Manager (izin pengaturan.kelola).
type PengaturanHandler struct {
	uc *usecase.PengaturanUsecase
}

func NewPengaturanHandler(uc *usecase.PengaturanUsecase) *PengaturanHandler {
	return &PengaturanHandler{uc: uc}
}

type PengaturanResponse struct {
	NamaToko      string  `json:"nama_toko"`
	Alamat        string  `json:"alamat"`
	Telepon       string  `json:"telepon"`
	Email         string  `json:"email"`
	LogoURL       string  `json:"logo_url"`
	MataUang      string  `json:"mata_uang"`
	StrukHeader   string  `json:"struk_header"`
	StrukFooter   string  `json:"struk_footer"`
	UkuranKertas  string  `json:"ukuran_kertas"`
	TampilkanLogo bool    `json:"tampilkan_logo"`
	PajakPersen   float64 `json:"pajak_persen"`
	PajakAktif    bool    `json:"pajak_aktif"`
	Pembulatan    int     `json:"pembulatan"`
}

func kePengaturanResponse(p *domain.Pengaturan) PengaturanResponse {
	return PengaturanResponse{
		NamaToko: p.NamaToko, Alamat: p.Alamat, Telepon: p.Telepon, Email: p.Email,
		LogoURL: p.LogoURL, MataUang: p.MataUang,
		StrukHeader: p.StrukHeader, StrukFooter: p.StrukFooter,
		UkuranKertas: p.UkuranKertas, TampilkanLogo: p.TampilkanLogo,
		PajakPersen: p.PajakPersen, PajakAktif: p.PajakAktif, Pembulatan: p.Pembulatan,
	}
}

type SimpanPengaturanRequest struct {
	NamaToko      string  `json:"nama_toko" validate:"required,min=1,max=150"`
	Alamat        string  `json:"alamat" validate:"omitempty,max=255"`
	Telepon       string  `json:"telepon" validate:"omitempty,max=30"`
	Email         string  `json:"email" validate:"omitempty,email,max=150"`
	LogoURL       string  `json:"logo_url" validate:"omitempty,url,max=255"`
	MataUang      string  `json:"mata_uang" validate:"required,max=8"`
	StrukHeader   string  `json:"struk_header" validate:"omitempty,max=255"`
	StrukFooter   string  `json:"struk_footer" validate:"omitempty,max=255"`
	UkuranKertas  string  `json:"ukuran_kertas" validate:"required,oneof=58mm 80mm"`
	TampilkanLogo *bool   `json:"tampilkan_logo" validate:"required"`
	PajakPersen   float64 `json:"pajak_persen" validate:"gte=0,lte=100"`
	PajakAktif    *bool   `json:"pajak_aktif" validate:"required"`
	Pembulatan    int     `json:"pembulatan" validate:"gte=0,lte=100000"`
}

// Ambil godoc
//
//	@Summary	Pengaturan usaha (profil toko, struk, pajak) — dibuat default bila belum ada
//	@Tags		pengaturan
//	@Produce	json
//	@Success	200	{object}	respond.Amplop{data=PengaturanResponse}
//	@Security	BearerAuth
//	@Router		/pengaturan [get]
func (h *PengaturanHandler) Ambil(c echo.Context) error {
	p, err := h.uc.Ambil(c.Request().Context())
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, kePengaturanResponse(p), nil, "")
}

// Perbarui godoc
//
//	@Summary	Simpan pengaturan usaha (Owner/Manager)
//	@Tags		pengaturan
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		SimpanPengaturanRequest	true	"Pengaturan"
//	@Success	200		{object}	respond.Amplop{data=PengaturanResponse}
//	@Security	BearerAuth
//	@Router		/pengaturan [put]
func (h *PengaturanHandler) Perbarui(c echo.Context) error {
	var req SimpanPengaturanRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	p, err := h.uc.Perbarui(c.Request().Context(), usecase.InputPengaturan{
		NamaToko: req.NamaToko, Alamat: req.Alamat, Telepon: req.Telepon, Email: req.Email,
		LogoURL: req.LogoURL, MataUang: req.MataUang,
		StrukHeader: req.StrukHeader, StrukFooter: req.StrukFooter,
		UkuranKertas: req.UkuranKertas, TampilkanLogo: *req.TampilkanLogo,
		PajakPersen: req.PajakPersen, PajakAktif: *req.PajakAktif, Pembulatan: req.Pembulatan,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, kePengaturanResponse(p), nil, "Pengaturan disimpan.")
}

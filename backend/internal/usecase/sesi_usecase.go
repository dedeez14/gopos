package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

// SesiUsecase — buka/tutup sesi kasir + rekap. Prinsip: satu sesi BUKA per
// pengguna; hanya pemilik sesi yang boleh menutupnya.
type SesiUsecase struct {
	sesi      domain.SesiRepository
	transaksi domain.TransaksiRepository
}

func NewSesiUsecase(sesi domain.SesiRepository, transaksi domain.TransaksiRepository) *SesiUsecase {
	return &SesiUsecase{sesi: sesi, transaksi: transaksi}
}

func (uc *SesiUsecase) Buka(ctx context.Context, userID uint, kasAwal float64, catatan string) (*domain.SesiKasir, error) {
	if _, err := uc.sesi.AktifMilik(ctx, userID); err == nil {
		return nil, domain.ErrSesiSudahBuka
	}

	s := &domain.SesiKasir{
		// Nomor sementara unik; nomor cantik SK-<tgl>-<id> diisi repo setelah
		// id lahir (satu transaksi DB) — tanpa balapan counter.
		Nomor:      fmt.Sprintf("SK-%d", time.Now().UnixNano()),
		UserID:     userID,
		Status:     domain.SesiBuka,
		KasAwal:    kasAwal,
		Catatan:    catatan,
		DibukaPada: time.Now(),
	}
	if err := uc.sesi.Simpan(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (uc *SesiUsecase) Aktif(ctx context.Context, userID uint) (*domain.SesiKasir, error) {
	return uc.sesi.AktifMilik(ctx, userID)
}

// Tutup menutup sesi milik user: kas sistem = kas awal + penjualan TUNAI
// (transfer/QRIS tidak menambah isi laci), selisih = kas fisik − kas sistem.
func (uc *SesiUsecase) Tutup(ctx context.Context, sesiID, userID uint, kasAkhir float64, catatan string) (*domain.SesiKasir, error) {
	s, err := uc.sesi.CariByID(ctx, sesiID)
	if err != nil {
		return nil, err
	}
	if s.UserID != userID {
		return nil, domain.ErrSesiBukanMilik
	}
	if s.Status != domain.SesiBuka {
		return nil, domain.ErrSesiSudahTutup
	}

	total, err := uc.transaksi.TotalSesi(ctx, s.ID)
	if err != nil {
		return nil, err
	}

	kini := time.Now()
	kasSistem := s.KasAwal + total[domain.TipeTunai]
	selisih := kasAkhir - kasSistem

	s.Status = domain.SesiTutup
	s.KasAkhir = &kasAkhir
	s.KasSistem = &kasSistem
	s.Selisih = &selisih
	s.DitutupPada = &kini
	if catatan != "" {
		s.Catatan = catatan
	}

	if err := uc.sesi.Perbarui(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Rekap — ringkasan sesi (kas + omzet per tipe bayar), boleh dilihat pemilik
// sesi atau peran manajemen (diputuskan pemanggil lewat bolehLintas).
func (uc *SesiUsecase) Rekap(ctx context.Context, sesiID, userID uint, bolehLintas bool) (*domain.SesiKasir, map[string]float64, error) {
	s, err := uc.sesi.CariByID(ctx, sesiID)
	if err != nil {
		return nil, nil, err
	}
	if !bolehLintas && s.UserID != userID {
		return nil, nil, domain.ErrSesiBukanMilik
	}
	total, err := uc.transaksi.TotalSesi(ctx, s.ID)
	if err != nil {
		return nil, nil, err
	}
	return s, total, nil
}

func (uc *SesiUsecase) Daftar(ctx context.Context, f domain.FilterSesi) ([]domain.SesiKasir, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.sesi.Daftar(ctx, f)
}

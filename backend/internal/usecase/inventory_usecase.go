package usecase

import (
	"context"
	"strings"

	"github.com/tuleh-pos/server/internal/domain"
)

// InventoryUsecase — stok masuk & opname. Aturan:
//   - hanya produk kelola_stok (JASA ditolak);
//   - MASUK = tambah delta positif;
//   - OPNAME = SET ke hasil hitung fisik; log menyimpan SELISIH-nya
//     (fisik − sistem) supaya riwayat memperlihatkan koreksinya, bukan
//     angka absolut yang tak bercerita.
type InventoryUsecase struct {
	produk domain.ProdukRepository
	stok   domain.StokRepository
	sesi   domain.SesiRepository
}

func NewInventoryUsecase(produk domain.ProdukRepository, stok domain.StokRepository, sesi domain.SesiRepository) *InventoryUsecase {
	return &InventoryUsecase{produk: produk, stok: stok, sesi: sesi}
}

func (uc *InventoryUsecase) Masuk(ctx context.Context, userID, produkID uint, jumlah float64, keterangan string) (*domain.StokLog, error) {
	p, err := uc.periksaProduk(ctx, produkID)
	if err != nil {
		return nil, err
	}
	_ = p

	log := &domain.StokLog{
		ProdukID: produkID, Jenis: domain.StokMasuk, Jumlah: jumlah,
		Keterangan: strings.TrimSpace(keterangan), UserID: userID,
		SesiKasirID: uc.sesiAktifID(ctx, userID),
	}
	if err := uc.stok.Masuk(ctx, produkID, jumlah, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (uc *InventoryUsecase) Opname(ctx context.Context, userID, produkID uint, stokFisik float64, keterangan string) (*domain.StokLog, error) {
	p, err := uc.periksaProduk(ctx, produkID)
	if err != nil {
		return nil, err
	}

	log := &domain.StokLog{
		ProdukID: produkID, Jenis: domain.StokOpname,
		Jumlah:     stokFisik - p.Stok, // selisih yang bercerita
		Keterangan: strings.TrimSpace(keterangan), UserID: userID,
		SesiKasirID: uc.sesiAktifID(ctx, userID),
	}
	if err := uc.stok.SetAbsolut(ctx, produkID, stokFisik, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (uc *InventoryUsecase) Riwayat(ctx context.Context, f domain.FilterStokLog) ([]domain.StokLog, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.stok.Riwayat(ctx, f)
}

func (uc *InventoryUsecase) periksaProduk(ctx context.Context, produkID uint) (*domain.Produk, error) {
	p, err := uc.produk.CariByID(ctx, produkID)
	if err != nil {
		return nil, err
	}
	if !p.KelolaStok {
		return nil, domain.ErrProdukTanpaStok
	}
	return p, nil
}

// sesiAktifID: sesi BUKA milik pencatat bila ada — pergerakan stok dari
// tengah shift tercap sesinya; di luar shift tetap boleh (nil).
func (uc *InventoryUsecase) sesiAktifID(ctx context.Context, userID uint) *uint {
	if s, err := uc.sesi.AktifMilik(ctx, userID); err == nil {
		id := s.ID
		return &id
	}
	return nil
}

package usecase

import (
	"context"
	"testing"

	"github.com/tuleh-pos/server/internal/domain"
)

// ───────────────────────────────────────────── repo palsu pelanggan & kawan

type pelangganRepoPalsu struct {
	data   map[uint]*domain.Pelanggan
	nextID uint
}

func newPelangganRepoPalsu() *pelangganRepoPalsu {
	return &pelangganRepoPalsu{data: map[uint]*domain.Pelanggan{}, nextID: 1}
}

func (r *pelangganRepoPalsu) Simpan(_ context.Context, p *domain.Pelanggan) error {
	p.ID = r.nextID
	r.nextID++
	r.data[p.ID] = p
	return nil
}
func (r *pelangganRepoPalsu) Perbarui(_ context.Context, p *domain.Pelanggan) error {
	r.data[p.ID] = p
	return nil
}
func (r *pelangganRepoPalsu) CariByID(_ context.Context, id uint) (*domain.Pelanggan, error) {
	if p, ok := r.data[id]; ok {
		return p, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *pelangganRepoPalsu) CariByTelepon(_ context.Context, tel string) (*domain.Pelanggan, error) {
	for _, p := range r.data {
		if p.Telepon != nil && *p.Telepon == tel {
			return p, nil
		}
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *pelangganRepoPalsu) Daftar(_ context.Context, _ domain.FilterPelanggan) ([]domain.Pelanggan, int64, error) {
	return nil, 0, nil
}
func (r *pelangganRepoPalsu) Nonaktifkan(_ context.Context, id uint) error {
	if p, ok := r.data[id]; ok {
		p.Aktif = false
		return nil
	}
	return domain.ErrTidakDitemukan
}

type holdRepoPalsu struct {
	data   map[uint]*domain.Hold
	nextID uint
}

func newHoldRepoPalsu() *holdRepoPalsu {
	return &holdRepoPalsu{data: map[uint]*domain.Hold{}, nextID: 1}
}

func (r *holdRepoPalsu) Simpan(_ context.Context, h *domain.Hold) error {
	h.ID = r.nextID
	r.nextID++
	r.data[h.ID] = h
	return nil
}
func (r *holdRepoPalsu) Daftar(_ context.Context) ([]domain.Hold, error) { return nil, nil }
func (r *holdRepoPalsu) CariByID(_ context.Context, id uint) (*domain.Hold, error) {
	if h, ok := r.data[id]; ok {
		return h, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *holdRepoPalsu) Hapus(_ context.Context, id uint) error {
	delete(r.data, id)
	return nil
}
func (r *holdRepoPalsu) Jumlah(_ context.Context) (int64, error) { return int64(len(r.data)), nil }

type laporanRepoPalsu struct {
	omzet   float64
	perTipe map[string]float64
	trx     int64
}

func (r *laporanRepoPalsu) PenjualanHarian(_ context.Context, _, _ string) ([]domain.PenjualanHarian, error) {
	return nil, nil
}
func (r *laporanRepoPalsu) ProdukTerlaris(_ context.Context, _, _ int) ([]domain.ProdukTerlaris, error) {
	return nil, nil
}
func (r *laporanRepoPalsu) OmzetBulan(_ context.Context, _ string) (float64, map[string]float64, int64, error) {
	return r.omzet, r.perTipe, r.trx, nil
}

type pengeluaranRepoPalsu struct {
	total float64
}

func (r *pengeluaranRepoPalsu) Simpan(_ context.Context, _ *domain.Pengeluaran) error { return nil }
func (r *pengeluaranRepoPalsu) Daftar(_ context.Context, _ domain.FilterPengeluaran) ([]domain.Pengeluaran, int64, error) {
	return nil, 0, nil
}
func (r *pengeluaranRepoPalsu) CariByID(_ context.Context, _ uint) (*domain.Pengeluaran, error) {
	return nil, domain.ErrTidakDitemukan
}
func (r *pengeluaranRepoPalsu) Hapus(_ context.Context, _ uint) error { return nil }
func (r *pengeluaranRepoPalsu) TotalBulan(_ context.Context, _ string) (float64, error) {
	return r.total, nil
}

// ───────────────────────────────────────────────────────────── kasus uji

func TestNormalisasiTelepon(t *testing.T) {
	kasus := map[string]string{
		"0812-8787-0509":  "6281287870509",
		"+62 812 111 222": "62812111222",
		"81234567890":     "6281234567890",
		"6281234567890":   "6281234567890",
		"123":             "", // terlalu pendek = tanpa telepon
	}
	for masuk, harap := range kasus {
		if got := NormalisasiTelepon(masuk); got != harap {
			t.Fatalf("NormalisasiTelepon(%q) = %q, harap %q", masuk, got, harap)
		}
	}
}

func TestQuickAddDedupByTelepon(t *testing.T) {
	uc := NewPelangganUsecase(newPelangganRepoPalsu())
	ctx := context.Background()

	p1, baru, err := uc.Quick(ctx, "Bu Sari", "0812-8787-0509")
	if err != nil || !baru {
		t.Fatalf("quick pertama harus membuat baru: %v %v", baru, err)
	}
	if *p1.Telepon != "6281287870509" {
		t.Fatalf("telepon tak ternormalisasi: %s", *p1.Telepon)
	}

	// Nomor sama (format berbeda) → pelanggan LAMA, bukan galat/duplikat.
	p2, baru, err := uc.Quick(ctx, "Sari Lain", "+62 812-8787-0509")
	if err != nil || baru || p2.ID != p1.ID {
		t.Fatalf("quick kedua harus dedup: baru=%v id=%d vs %d err=%v", baru, p2.ID, p1.ID, err)
	}
}

func TestHoldPenuhDitolak(t *testing.T) {
	repo := newHoldRepoPalsu()
	uc := NewHoldUsecase(repo)
	ctx := context.Background()

	for i := 0; i < maksHold; i++ {
		if _, err := uc.Simpan(ctx, 1, "", []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := uc.Simpan(ctx, 1, "", []byte(`{}`)); err != domain.ErrHoldPenuh {
		t.Fatalf("harap ErrHoldPenuh, dapat %v", err)
	}
}

func TestKeuanganLabaOmzetMinusPengeluaran(t *testing.T) {
	uc := NewLaporanUsecase(
		&laporanRepoPalsu{omzet: 5000000, perTipe: map[string]float64{"TUNAI": 3000000, "QRIS": 2000000}, trx: 120},
		&pengeluaranRepoPalsu{total: 1250000},
	)

	r, err := uc.Keuangan(context.Background(), "2026-08")
	if err != nil {
		t.Fatal(err)
	}
	if r.Laba != 3750000 || r.Omzet != 5000000 || r.Pengeluaran != 1250000 {
		t.Fatalf("laba=%v omzet=%v keluar=%v", r.Laba, r.Omzet, r.Pengeluaran)
	}
	if r.PerTipe["TUNAI"] != 3000000 {
		t.Fatal("rincian per tipe hilang")
	}
}

func TestCheckoutDenganPelanggan(t *testing.T) {
	produkRepo := newProdukRepoPalsu()
	sesiRepo := newSesiRepoPalsu()
	trxRepo := newTrxRepoPalsu()
	pelRepo := newPelangganRepoPalsu()

	sesiUC := NewSesiUsecase(sesiRepo, trxRepo)
	trxUC := NewTransaksiUsecase(trxRepo, sesiRepo, produkRepo, pelRepo)
	ctx := context.Background()

	if _, err := sesiUC.Buka(ctx, 7, 0, ""); err != nil {
		t.Fatal(err)
	}
	p := domain.Produk{Nama: "Z", Kode: "Z", HargaJual: 1000, Aktif: true}
	if err := produkRepo.Simpan(ctx, &p); err != nil {
		t.Fatal(err)
	}
	pel := domain.Pelanggan{Nama: "Bu Sari", Aktif: true}
	if err := pelRepo.Simpan(ctx, &pel); err != nil {
		t.Fatal(err)
	}

	trx, err := trxUC.Checkout(ctx, 7, InputCheckout{
		Items:          []ItemCheckout{{ProdukID: p.ID, Kuantitas: 1, Harga: 1000}},
		TipePembayaran: domain.TipeTunai, Dibayar: 1000, PelangganID: pel.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if trx.PelangganID == nil || *trx.PelangganID != pel.ID {
		t.Fatal("pelanggan tak tertaut ke transaksi")
	}

	// Pelanggan asing → ditolak.
	if _, err := trxUC.Checkout(ctx, 7, InputCheckout{
		Items:          []ItemCheckout{{ProdukID: p.ID, Kuantitas: 1, Harga: 1000}},
		TipePembayaran: domain.TipeTunai, PelangganID: 999,
	}); err != domain.ErrTidakDitemukan {
		t.Fatalf("pelanggan asing harus ditolak, dapat %v", err)
	}

}

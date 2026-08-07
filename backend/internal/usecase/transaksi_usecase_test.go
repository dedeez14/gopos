package usecase

import (
	"context"
	"testing"

	"github.com/tuleh-pos/server/internal/domain"
)

// ─────────────────────────────── repo palsu sesi & transaksi (in-memory)

type sesiRepoPalsu struct {
	data   map[uint]*domain.SesiKasir
	nextID uint
}

func newSesiRepoPalsu() *sesiRepoPalsu {
	return &sesiRepoPalsu{data: map[uint]*domain.SesiKasir{}, nextID: 1}
}

func (r *sesiRepoPalsu) Simpan(_ context.Context, s *domain.SesiKasir) error {
	s.ID = r.nextID
	r.nextID++
	r.data[s.ID] = s
	return nil
}
func (r *sesiRepoPalsu) Perbarui(_ context.Context, s *domain.SesiKasir) error {
	r.data[s.ID] = s
	return nil
}
func (r *sesiRepoPalsu) CariByID(_ context.Context, id uint) (*domain.SesiKasir, error) {
	if s, ok := r.data[id]; ok {
		return s, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *sesiRepoPalsu) AktifMilik(_ context.Context, userID uint) (*domain.SesiKasir, error) {
	for _, s := range r.data {
		if s.UserID == userID && s.Status == domain.SesiBuka {
			return s, nil
		}
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *sesiRepoPalsu) Daftar(_ context.Context, _ domain.FilterSesi) ([]domain.SesiKasir, int64, error) {
	return nil, 0, nil
}

type trxRepoPalsu struct {
	data         map[uint]*domain.Transaksi
	byIdem       map[string]uint
	stokDipotong map[uint]float64
	nextID       uint
}

func newTrxRepoPalsu() *trxRepoPalsu {
	return &trxRepoPalsu{data: map[uint]*domain.Transaksi{}, byIdem: map[string]uint{}, stokDipotong: map[uint]float64{}, nextID: 1}
}

func (r *trxRepoPalsu) Checkout(_ context.Context, t *domain.Transaksi, potongStok map[uint]float64) error {
	t.ID = r.nextID
	r.nextID++
	r.data[t.ID] = t
	if t.IdempotencyKey != nil {
		r.byIdem[*t.IdempotencyKey] = t.ID
	}
	for id, qty := range potongStok {
		r.stokDipotong[id] += qty
	}
	return nil
}
func (r *trxRepoPalsu) CariByID(_ context.Context, id uint) (*domain.Transaksi, error) {
	if t, ok := r.data[id]; ok {
		return t, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *trxRepoPalsu) CariByIdempotency(_ context.Context, key string) (*domain.Transaksi, error) {
	if id, ok := r.byIdem[key]; ok {
		return r.data[id], nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *trxRepoPalsu) Daftar(_ context.Context, _ domain.FilterTransaksi) ([]domain.Transaksi, int64, error) {
	return nil, 0, nil
}
func (r *trxRepoPalsu) TotalSesi(_ context.Context, sesiID uint) (map[string]float64, error) {
	total := map[string]float64{}
	for _, t := range r.data {
		if t.SesiKasirID == sesiID {
			total[t.TipePembayaran] += t.GrandTotal
		}
	}
	return total, nil
}

// perlengkapan: kasir ber-sesi + produk di katalog palsu.
func siapkanKasir(t *testing.T) (*TransaksiUsecase, *SesiUsecase, *produkRepoPalsu, *trxRepoPalsu) {
	t.Helper()
	produkRepo := newProdukRepoPalsu()
	sesiRepo := newSesiRepoPalsu()
	trxRepo := newTrxRepoPalsu()

	sesiUC := NewSesiUsecase(sesiRepo, trxRepo)
	trxUC := NewTransaksiUsecase(trxRepo, sesiRepo, produkRepo, newPelangganRepoPalsu())

	if _, err := sesiUC.Buka(context.Background(), 7, 100000, ""); err != nil {
		t.Fatal(err)
	}
	return trxUC, sesiUC, produkRepo, trxRepo
}

func tambahProduk(t *testing.T, repo *produkRepoPalsu, p domain.Produk) *domain.Produk {
	t.Helper()
	if err := repo.Simpan(context.Background(), &p); err != nil {
		t.Fatal(err)
	}
	return &p
}

// ───────────────────────────────────────────────────────────── kasus uji

func TestCheckoutRumusLengkap(t *testing.T) {
	trxUC, _, produkRepo, repo := siapkanKasir(t)
	ctx := context.Background()

	a := tambahProduk(t, produkRepo, domain.Produk{Nama: "A", Kode: "A", HargaJual: 10000, Aktif: true, KelolaStok: true, Stok: 50})
	b := tambahProduk(t, produkRepo, domain.Produk{Nama: "B", Kode: "B", HargaJual: 20000, Aktif: true, KelolaStok: true, Stok: 50})

	// A: 10 × 10.000 = 100.000, diskon item 10% = 10.000, pajak 10% dari 90.000 = 9.000
	// B:  1 × 20.000 =  20.000
	// dasar = 120.000 − 10.000 = 110.000
	// diskon keseluruhan: nominal 10.000 → sisa 100.000, persen 10% → 10.000 ⇒ 20.000
	// totalDiskon = 30.000 ; grand = 120.000 − 30.000 + 9.000 = 99.000
	trx, err := trxUC.Checkout(ctx, 7, InputCheckout{
		Items: []ItemCheckout{
			{ProdukID: a.ID, Kuantitas: 10, Harga: 10000, DiskonPersen: 10, PajakPersen: 10},
			{ProdukID: b.ID, Kuantitas: 1, Harga: 20000},
		},
		DiskonNominal: 10000, DiskonPersen: 10,
		TipePembayaran: domain.TipeTunai, Dibayar: 100000,
	})
	if err != nil {
		t.Fatal(err)
	}

	if trx.Subtotal != 120000 || trx.TotalDiskon != 30000 || trx.TotalPajak != 9000 || trx.GrandTotal != 99000 {
		t.Fatalf("rumus salah: subtotal=%v diskon=%v pajak=%v grand=%v", trx.Subtotal, trx.TotalDiskon, trx.TotalPajak, trx.GrandTotal)
	}
	if trx.Kembalian != 1000 {
		t.Fatalf("kembalian = %v, harap 1000", trx.Kembalian)
	}
	if repo.stokDipotong[a.ID] != 10 || repo.stokDipotong[b.ID] != 1 {
		t.Fatalf("stok dipotong salah: %v", repo.stokDipotong)
	}
	if trx.Items[0].NamaProduk != "A" {
		t.Fatal("snapshot nama produk hilang")
	}
}

func TestCheckoutHargaKosongPakaiHargaEfektifPromo(t *testing.T) {
	trxUC, _, produkRepo, _ := siapkanKasir(t)

	promo := 8000.0
	p := tambahProduk(t, produkRepo, domain.Produk{Nama: "Promo", Kode: "PR", HargaJual: 10000, HargaPromo: &promo, Aktif: true})

	trx, err := trxUC.Checkout(context.Background(), 7, InputCheckout{
		Items:          []ItemCheckout{{ProdukID: p.ID, Kuantitas: 2}}, // harga tidak dikirim
		TipePembayaran: domain.TipeQris, Dibayar: 16000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if trx.GrandTotal != 16000 {
		t.Fatalf("harga efektif promo tak dipakai: grand=%v", trx.GrandTotal)
	}
}

func TestCheckoutTanpaSesiDitolak(t *testing.T) {
	produkRepo := newProdukRepoPalsu()
	trxUC := NewTransaksiUsecase(newTrxRepoPalsu(), newSesiRepoPalsu(), produkRepo, newPelangganRepoPalsu())

	_, err := trxUC.Checkout(context.Background(), 99, InputCheckout{
		Items: []ItemCheckout{{ProdukID: 1, Kuantitas: 1}}, TipePembayaran: domain.TipeTunai,
	})
	if err != domain.ErrSesiBelumBuka {
		t.Fatalf("harap ErrSesiBelumBuka, dapat %v", err)
	}
}

func TestCheckoutIdempoten(t *testing.T) {
	trxUC, _, produkRepo, repo := siapkanKasir(t)
	p := tambahProduk(t, produkRepo, domain.Produk{Nama: "X", Kode: "X", HargaJual: 5000, Aktif: true, KelolaStok: true, Stok: 10})

	in := InputCheckout{
		Items:          []ItemCheckout{{ProdukID: p.ID, Kuantitas: 1, Harga: 5000}},
		TipePembayaran: domain.TipeTunai, Dibayar: 5000, IdempotencyKey: "kunci-1",
	}
	t1, err := trxUC.Checkout(context.Background(), 7, in)
	if err != nil {
		t.Fatal(err)
	}
	t2, err := trxUC.Checkout(context.Background(), 7, in) // retry
	if err != nil {
		t.Fatal(err)
	}
	if t1.ID != t2.ID {
		t.Fatal("idempotency gagal — transaksi dobel")
	}
	if repo.stokDipotong[p.ID] != 1 {
		t.Fatalf("stok terpotong %v kali — retry menggerakkan stok", repo.stokDipotong[p.ID])
	}
}

func TestCheckoutProdukNonaktifDitolak(t *testing.T) {
	trxUC, _, produkRepo, _ := siapkanKasir(t)
	p := tambahProduk(t, produkRepo, domain.Produk{Nama: "Mati", Kode: "M", HargaJual: 1000, Aktif: false})

	_, err := trxUC.Checkout(context.Background(), 7, InputCheckout{
		Items: []ItemCheckout{{ProdukID: p.ID, Kuantitas: 1}}, TipePembayaran: domain.TipeTunai,
	})
	if err != domain.ErrProdukNonaktif {
		t.Fatalf("harap ErrProdukNonaktif, dapat %v", err)
	}
}

func TestTutupSesiMenghitungKasSistem(t *testing.T) {
	trxUC, sesiUC, produkRepo, _ := siapkanKasir(t)
	ctx := context.Background()
	p := tambahProduk(t, produkRepo, domain.Produk{Nama: "Y", Kode: "Y", HargaJual: 25000, Aktif: true})

	// 2 tunai + 1 QRIS: hanya tunai yang menambah laci.
	for i := 0; i < 2; i++ {
		if _, err := trxUC.Checkout(ctx, 7, InputCheckout{
			Items:          []ItemCheckout{{ProdukID: p.ID, Kuantitas: 1, Harga: 25000}},
			TipePembayaran: domain.TipeTunai, Dibayar: 25000,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := trxUC.Checkout(ctx, 7, InputCheckout{
		Items:          []ItemCheckout{{ProdukID: p.ID, Kuantitas: 1, Harga: 25000}},
		TipePembayaran: domain.TipeQris, Dibayar: 25000,
	}); err != nil {
		t.Fatal(err)
	}

	sesi, _ := sesiUC.Aktif(ctx, 7)
	// kas awal 100.000 + tunai 50.000 = 150.000; fisik 149.000 → selisih −1.000
	ditutup, err := sesiUC.Tutup(ctx, sesi.ID, 7, 149000, "")
	if err != nil {
		t.Fatal(err)
	}
	if *ditutup.KasSistem != 150000 || *ditutup.Selisih != -1000 {
		t.Fatalf("kas sistem=%v selisih=%v", *ditutup.KasSistem, *ditutup.Selisih)
	}

	// Sesi tertutup: checkout berikutnya ditolak; tutup dua kali ditolak.
	if _, err := trxUC.Checkout(ctx, 7, InputCheckout{
		Items: []ItemCheckout{{ProdukID: p.ID, Kuantitas: 1}}, TipePembayaran: domain.TipeTunai,
	}); err != domain.ErrSesiBelumBuka {
		t.Fatalf("checkout pasca tutup harus ditolak, dapat %v", err)
	}
	if _, err := sesiUC.Tutup(ctx, sesi.ID, 7, 1, ""); err != domain.ErrSesiSudahTutup {
		t.Fatalf("tutup dua kali harus ditolak, dapat %v", err)
	}
}

func TestBukaSesiDobelDitolak(t *testing.T) {
	_, sesiUC, _, _ := siapkanKasir(t) // sudah membuka sesi utk user 7

	if _, err := sesiUC.Buka(context.Background(), 7, 0, ""); err != domain.ErrSesiSudahBuka {
		t.Fatalf("harap ErrSesiSudahBuka, dapat %v", err)
	}
}

package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

// ───────────────────────────── repo palsu (in-memory) — bukti nilai kontrak:
// usecase diuji TANPA database, murni terhadap interface domain.

type produkRepoPalsu struct {
	data   map[uint]*domain.Produk
	nextID uint
}

func newProdukRepoPalsu() *produkRepoPalsu {
	return &produkRepoPalsu{data: map[uint]*domain.Produk{}, nextID: 1}
}

func (r *produkRepoPalsu) Simpan(_ context.Context, p *domain.Produk) error {
	p.ID = r.nextID
	r.nextID++
	salinan := *p
	r.data[p.ID] = &salinan
	return nil
}

func (r *produkRepoPalsu) Perbarui(_ context.Context, p *domain.Produk) error {
	salinan := *p
	r.data[p.ID] = &salinan
	return nil
}

func (r *produkRepoPalsu) CariByID(_ context.Context, id uint) (*domain.Produk, error) {
	if p, ok := r.data[id]; ok {
		salinan := *p
		return &salinan, nil
	}
	return nil, domain.ErrTidakDitemukan
}

func (r *produkRepoPalsu) CariByKode(_ context.Context, kode string) (*domain.Produk, error) {
	for _, p := range r.data {
		if p.Kode == kode {
			salinan := *p
			return &salinan, nil
		}
	}
	return nil, domain.ErrTidakDitemukan
}

func (r *produkRepoPalsu) Daftar(_ context.Context, f domain.FilterProduk) ([]domain.Produk, int64, error) {
	var hasil []domain.Produk
	for _, p := range r.data {
		hasil = append(hasil, *p)
	}
	return hasil, int64(len(hasil)), nil
}

func (r *produkRepoPalsu) Nonaktifkan(_ context.Context, id uint) error {
	if p, ok := r.data[id]; ok {
		p.Aktif = false
		return nil
	}
	return domain.ErrTidakDitemukan
}

type kategoriRepoPalsu struct{ data map[uint]*domain.Kategori }

func (r *kategoriRepoPalsu) Simpan(_ context.Context, k *domain.Kategori) error {
	k.ID = uint(len(r.data) + 1)
	r.data[k.ID] = k
	return nil
}
func (r *kategoriRepoPalsu) Daftar(_ context.Context) ([]domain.Kategori, error) { return nil, nil }
func (r *kategoriRepoPalsu) CariByID(_ context.Context, id uint) (*domain.Kategori, error) {
	if k, ok := r.data[id]; ok {
		return k, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *kategoriRepoPalsu) Hapus(_ context.Context, id uint) error { return nil }

func ucBaru() (*ProdukUsecase, *produkRepoPalsu) {
	repo := newProdukRepoPalsu()
	return NewProdukUsecase(repo, &kategoriRepoPalsu{data: map[uint]*domain.Kategori{}}), repo
}

func tgl(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func f64(v float64) *float64 { return &v }

// ───────────────────────────────────────────────────────── logika promo

func TestPromoAktifDanHargaEfektif(t *testing.T) {
	pada := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)

	kasus := []struct {
		nama       string
		p          domain.Produk
		aktif      bool
		hargaHarap float64
	}{
		{"tanpa promo", domain.Produk{HargaJual: 10000}, false, 10000},
		{"promo tanpa batas", domain.Produk{HargaJual: 10000, HargaPromo: f64(8000)}, true, 8000},
		{"dalam periode", domain.Produk{HargaJual: 10000, HargaPromo: f64(8000), PromoMulai: tgl("2026-08-01"), PromoSelesai: tgl("2026-08-31")}, true, 8000},
		{"hari terakhir masih aktif", domain.Produk{HargaJual: 10000, HargaPromo: f64(8000), PromoSelesai: tgl("2026-08-15")}, true, 8000},
		{"belum mulai", domain.Produk{HargaJual: 10000, HargaPromo: f64(8000), PromoMulai: tgl("2026-08-16")}, false, 10000},
		{"sudah lewat", domain.Produk{HargaJual: 10000, HargaPromo: f64(8000), PromoSelesai: tgl("2026-08-14")}, false, 10000},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if got := k.p.PromoAktif(pada); got != k.aktif {
				t.Fatalf("PromoAktif = %v, harap %v", got, k.aktif)
			}
			if got := k.p.HargaEfektif(pada); got != k.hargaHarap {
				t.Fatalf("HargaEfektif = %v, harap %v", got, k.hargaHarap)
			}
		})
	}
}

// ─────────────────────────────────────────────────────── aturan usecase

func TestJasaSelaluTanpaKelolaStok(t *testing.T) {
	uc, _ := ucBaru()

	p, err := uc.Buat(context.Background(), InputProduk{
		Nama: "Cuci Kiloan", Tipe: domain.TipeJasa, HargaJual: 7000,
		KelolaStok: true, // pun diminta true — jasa tetap tanpa stok
		Aktif:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.KelolaStok {
		t.Fatal("JASA harus selalu kelola_stok=false")
	}
}

func TestKodeUnikDanGenerate(t *testing.T) {
	uc, _ := ucBaru()
	ctx := context.Background()

	p1, err := uc.Buat(ctx, InputProduk{Nama: "A", Tipe: domain.TipeBarang, Kode: "P-001", HargaJual: 1, Aktif: true, KelolaStok: true})
	if err != nil {
		t.Fatal(err)
	}
	if p1.Kode != "P-001" {
		t.Fatalf("kode = %s", p1.Kode)
	}

	// Kode sama → konflik.
	if _, err := uc.Buat(ctx, InputProduk{Nama: "B", Tipe: domain.TipeBarang, Kode: "P-001", HargaJual: 1, Aktif: true}); err != domain.ErrKodeTerpakai {
		t.Fatalf("harap ErrKodeTerpakai, dapat %v", err)
	}

	// Kosong → digenerate.
	p3, err := uc.Buat(ctx, InputProduk{Nama: "C", Tipe: domain.TipeBarang, HargaJual: 1, Aktif: true})
	if err != nil {
		t.Fatal(err)
	}
	if p3.Kode == "" {
		t.Fatal("kode kosong harus digenerate")
	}

	// Update dgn kode sendiri (tak berubah) → boleh.
	if _, err := uc.Perbarui(ctx, p1.ID, InputProduk{Nama: "A2", Tipe: domain.TipeBarang, Kode: "P-001", HargaJual: 2, Aktif: true, KelolaStok: true}); err != nil {
		t.Fatalf("update kode sendiri harus boleh: %v", err)
	}
}

func TestCabutPromoMembersihkanPeriode(t *testing.T) {
	uc, repo := ucBaru()
	ctx := context.Background()

	p, _ := uc.Buat(ctx, InputProduk{
		Nama: "X", Tipe: domain.TipeBarang, HargaJual: 10000, Aktif: true, KelolaStok: true,
		HargaPromo: f64(8000), PromoMulai: tgl("2026-08-01"), PromoSelesai: tgl("2026-08-31"),
	})

	// Cabut promo (HargaPromo nil) → periode ikut bersih.
	hasil, err := uc.Perbarui(ctx, p.ID, InputProduk{
		Nama: "X", Tipe: domain.TipeBarang, HargaJual: 10000, Aktif: true, KelolaStok: true,
		HargaPromo: nil, PromoMulai: tgl("2026-08-01"), // sisa input lama — harus diabaikan
	})
	if err != nil {
		t.Fatal(err)
	}
	if hasil.HargaPromo != nil || hasil.PromoMulai != nil || hasil.PromoSelesai != nil {
		t.Fatal("cabut promo harus membersihkan harga & periode")
	}
	tersimpan, _ := repo.CariByID(ctx, p.ID)
	if tersimpan.PromoMulai != nil {
		t.Fatal("periode promo masih tersisa di penyimpanan")
	}
}

func TestKategoriAsingDitolak(t *testing.T) {
	uc, _ := ucBaru()

	_, err := uc.Buat(context.Background(), InputProduk{
		Nama: "Y", Tipe: domain.TipeBarang, HargaJual: 1, Aktif: true, KategoriID: 999,
	})
	if err != domain.ErrTidakDitemukan {
		t.Fatalf("kategori tak ada harus ditolak, dapat %v", err)
	}
}

func TestDaftarMenjagaBatasPaginasi(t *testing.T) {
	uc, _ := ucBaru()

	// Nilai liar dinormalkan usecase — repo palsu tak peduli, yang diuji
	// adalah klamp-nya tidak meledak.
	if _, _, err := uc.Daftar(context.Background(), domain.FilterProduk{Halaman: -5, PerHal: 100000}); err != nil {
		t.Fatal(err)
	}
}

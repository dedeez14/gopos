package usecase

import (
	"context"
	"testing"

	"github.com/tuleh-pos/server/internal/domain"
)

type pengaturanRepoPalsu struct {
	data        map[uint]*domain.Pengaturan // per usaha_id
	simpanCount int
}

func newPengaturanRepoPalsu() *pengaturanRepoPalsu {
	return &pengaturanRepoPalsu{data: map[uint]*domain.Pengaturan{}}
}

func (r *pengaturanRepoPalsu) CariByUsaha(ctx context.Context) (*domain.Pengaturan, error) {
	if p, ok := r.data[domain.UsahaDari(ctx)]; ok {
		return p, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *pengaturanRepoPalsu) Simpan(ctx context.Context, p *domain.Pengaturan) error {
	if p.UsahaID == 0 {
		p.UsahaID = domain.UsahaDari(ctx)
	}
	r.simpanCount++
	p.ID = uint(len(r.data) + 1)
	r.data[p.UsahaID] = p
	return nil
}
func (r *pengaturanRepoPalsu) Perbarui(ctx context.Context, p *domain.Pengaturan) error {
	r.data[domain.UsahaDari(ctx)] = p
	return nil
}

// Ambil pada usaha yang belum punya pengaturan membuat SATU baris default,
// nama toko diambil dari nama usaha; pemanggilan ulang tidak membuat lagi.
func TestAmbilBuatDefaultSekali(t *testing.T) {
	usahaRepo := newUsahaRepoPalsu()
	_ = usahaRepo.Simpan(context.Background(), &domain.Usaha{Kode: "WB", Nama: "Warung Bahagia", Aktif: true}) // id 1
	repo := newPengaturanRepoPalsu()
	uc := NewPengaturanUsecase(repo, usahaRepo)
	ctx := domain.DenganUsaha(context.Background(), 1)

	p, err := uc.Ambil(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if p.NamaToko != "Warung Bahagia" {
		t.Fatalf("nama toko harus prefill dari usaha, dapat %q", p.NamaToko)
	}
	if p.MataUang != "Rp" || p.UkuranKertas != "80mm" {
		t.Fatalf("default salah: mata_uang=%q kertas=%q", p.MataUang, p.UkuranKertas)
	}

	if _, err := uc.Ambil(ctx); err != nil {
		t.Fatal(err)
	}
	if repo.simpanCount != 1 {
		t.Fatalf("Ambil kedua tidak boleh membuat baris baru; simpanCount=%d", repo.simpanCount)
	}
}

// Perbarui menormalkan nilai di luar rentang DAN menyimpan bool false
// (pajak nonaktif) — bukan mengabaikannya sebagai zero-value.
func TestPerbaruiNormalisasiDanBoolFalse(t *testing.T) {
	usahaRepo := newUsahaRepoPalsu()
	_ = usahaRepo.Simpan(context.Background(), &domain.Usaha{Kode: "TX", Nama: "Toko X", Aktif: true}) // id 1
	repo := newPengaturanRepoPalsu()
	uc := NewPengaturanUsecase(repo, usahaRepo)
	ctx := domain.DenganUsaha(context.Background(), 1)

	p, err := uc.Perbarui(ctx, InputPengaturan{
		NamaToko: "  Toko X  ", MataUang: "", UkuranKertas: "A4",
		PajakPersen: 150, PajakAktif: false, Pembulatan: -5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.NamaToko != "Toko X" {
		t.Fatalf("nama toko harus di-trim, dapat %q", p.NamaToko)
	}
	if p.MataUang != "Rp" || p.UkuranKertas != "80mm" {
		t.Fatalf("normalisasi salah: mata_uang=%q kertas=%q", p.MataUang, p.UkuranKertas)
	}
	if p.PajakPersen != 100 || p.Pembulatan != 0 {
		t.Fatalf("clamp salah: pajak=%v pembulatan=%v", p.PajakPersen, p.Pembulatan)
	}

	// Bool false harus benar-benar tersimpan (bukan hilang).
	lagi, err := uc.Ambil(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if lagi.PajakAktif {
		t.Fatal("pajak_aktif=false harus tersimpan")
	}
}

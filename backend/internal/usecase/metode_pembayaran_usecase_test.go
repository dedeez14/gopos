package usecase

import (
	"context"
	"testing"

	"github.com/tuleh-pos/server/internal/domain"
)

type metodeRepoPalsu struct {
	data   map[uint]*domain.MetodePembayaran
	nextID uint
}

func newMetodeRepoPalsu() *metodeRepoPalsu {
	return &metodeRepoPalsu{data: map[uint]*domain.MetodePembayaran{}, nextID: 1}
}

func (r *metodeRepoPalsu) Simpan(_ context.Context, m *domain.MetodePembayaran) error {
	m.ID = r.nextID
	r.nextID++
	r.data[m.ID] = m
	return nil
}
func (r *metodeRepoPalsu) Perbarui(_ context.Context, m *domain.MetodePembayaran) error {
	r.data[m.ID] = m
	return nil
}
func (r *metodeRepoPalsu) CariByID(_ context.Context, id uint) (*domain.MetodePembayaran, error) {
	if m, ok := r.data[id]; ok {
		return m, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *metodeRepoPalsu) Daftar(_ context.Context, hanyaAktif bool) ([]domain.MetodePembayaran, error) {
	var out []domain.MetodePembayaran
	for _, m := range r.data {
		if hanyaAktif && !m.Aktif {
			continue
		}
		out = append(out, *m)
	}
	return out, nil
}
func (r *metodeRepoPalsu) Hapus(_ context.Context, id uint) error { delete(r.data, id); return nil }

func TestMetodeValidasiPerJenis(t *testing.T) {
	uc := NewMetodePembayaranUsecase(newMetodeRepoPalsu())
	ctx := context.Background()

	// BANK tanpa nomor/atas nama → ditolak.
	if _, err := uc.Buat(ctx, InputMetode{Jenis: "BANK", Nama: "BCA", Aktif: true}); err != domain.ErrDataBayarKurang {
		t.Fatalf("BANK tanpa data → harap ErrDataBayarKurang, dapat %v", err)
	}
	// QRIS tanpa gambar → ditolak.
	if _, err := uc.Buat(ctx, InputMetode{Jenis: "QRIS", Nama: "QR", Aktif: true}); err != domain.ErrDataBayarKurang {
		t.Fatalf("QRIS tanpa gambar → harap ErrDataBayarKurang, dapat %v", err)
	}
	// Jenis asing → ditolak.
	if _, err := uc.Buat(ctx, InputMetode{Jenis: "KARTU", Nama: "Visa", Aktif: true}); err != domain.ErrJenisBayarTakDikenal {
		t.Fatalf("jenis asing → harap ErrJenisBayarTakDikenal, dapat %v", err)
	}

	// BANK lengkap → OK, jenis dinormalkan uppercase & di-trim.
	m, err := uc.Buat(ctx, InputMetode{Jenis: " bank ", Nama: "BCA", Nomor: "123", AtasNama: "Budi", Aktif: true})
	if err != nil {
		t.Fatal(err)
	}
	if m.Jenis != "BANK" {
		t.Fatalf("jenis harus dinormalkan ke BANK, dapat %q", m.Jenis)
	}
}

func TestMetodeDaftarHanyaAktif(t *testing.T) {
	uc := NewMetodePembayaranUsecase(newMetodeRepoPalsu())
	ctx := context.Background()

	aktif, _ := uc.Buat(ctx, InputMetode{Jenis: "QRIS", Nama: "QR Toko", GambarURL: "https://x/q.png", Aktif: true})
	if _, err := uc.Buat(ctx, InputMetode{Jenis: "BANK", Nama: "Mandiri", Nomor: "9", AtasNama: "A", Aktif: false}); err != nil {
		t.Fatal(err)
	}

	semua, _ := uc.Daftar(ctx, false)
	if len(semua) != 2 {
		t.Fatalf("daftar semua = %d, harap 2", len(semua))
	}
	hanyaAktif, _ := uc.Daftar(ctx, true)
	if len(hanyaAktif) != 1 || hanyaAktif[0].ID != aktif.ID {
		t.Fatalf("daftar aktif harus 1 (QR Toko), dapat %d", len(hanyaAktif))
	}
}

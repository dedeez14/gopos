package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/tuleh-pos/server/internal/domain"
)

// kotak palsu — "enkripsi" reversibel sederhana (bukan kripto nyata) supaya
// usecase teruji tanpa bergantung pkg/rahasia.
type kotakPalsu struct{}

func (kotakPalsu) Enkripsi(t string) (string, error) { return "enc:" + t, nil }
func (kotakPalsu) Dekripsi(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

type gatewayRepoPalsu struct {
	data map[uint]*domain.GatewayMidtrans
}

func newGatewayRepoPalsu() *gatewayRepoPalsu {
	return &gatewayRepoPalsu{data: map[uint]*domain.GatewayMidtrans{}}
}
func (r *gatewayRepoPalsu) CariByUsaha(ctx context.Context) (*domain.GatewayMidtrans, error) {
	if g, ok := r.data[domain.UsahaDari(ctx)]; ok {
		return g, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *gatewayRepoPalsu) Simpan(ctx context.Context, g *domain.GatewayMidtrans) error {
	if g.UsahaID == 0 {
		g.UsahaID = domain.UsahaDari(ctx)
	}
	g.ID = uint(len(r.data) + 1)
	r.data[g.UsahaID] = g
	return nil
}
func (r *gatewayRepoPalsu) Perbarui(ctx context.Context, g *domain.GatewayMidtrans) error {
	r.data[domain.UsahaDari(ctx)] = g
	return nil
}

func setUsahaMidtrans(t *testing.T, repo *usahaRepoPalsu, aktif bool) {
	t.Helper()
	_ = repo.Simpan(context.Background(), &domain.Usaha{Kode: "U", Nama: "U", Aktif: true, MidtransAktif: aktif}) // id 1
}

// Platform belum aktif → simpan ditolak; setelah aktif → tersimpan, server key
// terenkripsi & TAK pernah plaintext (hanya hint 4 digit).
func TestGatewaySaklarPlatformDanEnkripsi(t *testing.T) {
	usahaRepo := newUsahaRepoPalsu()
	setUsahaMidtrans(t, usahaRepo, false)
	uc := NewGatewayUsecase(newGatewayRepoPalsu(), usahaRepo, kotakPalsu{})
	ctx := domain.DenganUsaha(context.Background(), 1)

	// Platform mati → tolak simpan.
	if _, err := uc.Perbarui(ctx, InputGateway{Aktif: true, Env: "sandbox", ServerKey: "SB-Mid-server-ABCD1234"}); err != domain.ErrModulMidtransMati {
		t.Fatalf("platform mati → harap ErrModulMidtransMati, dapat %v", err)
	}

	// Nyalakan platform lalu simpan.
	usahaRepo.data[1].MidtransAktif = true
	k, err := uc.Perbarui(ctx, InputGateway{Aktif: true, Env: "production", ClientKey: "Mid-client-XYZ", ServerKey: "SB-Mid-server-ABCD1234"})
	if err != nil {
		t.Fatal(err)
	}
	if !k.ServerKeyTerisi || k.ServerKeyHint != "••••1234" {
		t.Fatalf("hint server key salah: terisi=%v hint=%q", k.ServerKeyTerisi, k.ServerKeyHint)
	}
	if !k.Siap {
		t.Fatal("harus siap: platform on + aktif + server key terisi")
	}
	if k.Env != "production" {
		t.Fatalf("env=%q", k.Env)
	}
}

// server_key dikosongkan saat update → key tersimpan dipertahankan (tidak
// terhapus jadi kosong).
func TestGatewayServerKeyKosongPertahankan(t *testing.T) {
	usahaRepo := newUsahaRepoPalsu()
	setUsahaMidtrans(t, usahaRepo, true)
	repo := newGatewayRepoPalsu()
	uc := NewGatewayUsecase(repo, usahaRepo, kotakPalsu{})
	ctx := domain.DenganUsaha(context.Background(), 1)

	if _, err := uc.Perbarui(ctx, InputGateway{Aktif: true, Env: "sandbox", ServerKey: "SB-Mid-server-AAAA9999"}); err != nil {
		t.Fatal(err)
	}
	// Update lain tanpa server key → tetap terisi.
	k, err := uc.Perbarui(ctx, InputGateway{Aktif: false, Env: "sandbox", ServerKey: ""})
	if err != nil {
		t.Fatal(err)
	}
	if !k.ServerKeyTerisi || k.ServerKeyHint != "••••9999" {
		t.Fatalf("server key harus dipertahankan, dapat terisi=%v hint=%q", k.ServerKeyTerisi, k.ServerKeyHint)
	}
	if k.Siap {
		t.Fatal("aktif=false → tidak siap")
	}
}

package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/tuleh-pos/server/internal/domain"
)

// kotakRahasia — kontrak enkripsi (dipenuhi pkg/rahasia.Kotak); interface di
// sisi konsumen supaya usecase mudah diuji dengan kotak palsu.
type kotakRahasia interface {
	Enkripsi(teks string) (string, error)
	Dekripsi(sandi string) (string, error)
}

// GatewayUsecase — konfigurasi Midtrans merchant. Menegakkan saklar platform
// (Usaha.MidtransAktif) sebelum mengizinkan penyimpanan, dan menjaga server
// key selalu terenkripsi.
type GatewayUsecase struct {
	repo  domain.GatewayMidtransRepository
	usaha domain.UsahaRepository
	kotak kotakRahasia
}

func NewGatewayUsecase(repo domain.GatewayMidtransRepository, usaha domain.UsahaRepository, kotak kotakRahasia) *GatewayUsecase {
	return &GatewayUsecase{repo: repo, usaha: usaha, kotak: kotak}
}

// KonfigMidtrans — bentuk AMAN untuk panel: server key tak pernah plaintext,
// hanya petunjuk 4 digit terakhir + flag terisi.
type KonfigMidtrans struct {
	PlatformAktif   bool
	Aktif           bool
	Env             string
	MerchantID      string
	ClientKey       string
	ServerKeyHint   string // "••••1234" atau kosong
	ServerKeyTerisi bool
	Siap            bool // platform && aktif && server key terisi
}

func (uc *GatewayUsecase) platformAktif(ctx context.Context) (bool, error) {
	u, err := uc.usaha.CariByID(ctx, domain.UsahaDari(ctx))
	if err != nil {
		return false, err
	}
	return u.MidtransAktif, nil
}

// Ambil mengembalikan konfigurasi (bertopeng) + status kesiapan.
func (uc *GatewayUsecase) Ambil(ctx context.Context) (*KonfigMidtrans, error) {
	platform, err := uc.platformAktif(ctx)
	if err != nil {
		return nil, err
	}
	k := &KonfigMidtrans{PlatformAktif: platform, Env: "sandbox"}

	g, err := uc.repo.CariByUsaha(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrTidakDitemukan) {
			return k, nil // belum dikonfigurasi
		}
		return nil, err
	}
	k.Aktif = g.Aktif
	k.Env = g.Env
	k.MerchantID = g.MerchantID
	k.ClientKey = g.ClientKey
	if g.ServerKeyEnc != "" {
		k.ServerKeyTerisi = true
		if plain, err := uc.kotak.Dekripsi(g.ServerKeyEnc); err == nil {
			k.ServerKeyHint = mask4(plain)
		}
	}
	k.Siap = platform && g.Aktif && g.ServerKeyEnc != ""
	return k, nil
}

type InputGateway struct {
	Aktif      bool
	Env        string
	MerchantID string
	ClientKey  string
	ServerKey  string // kosong = JANGAN ubah yang tersimpan
}

// Perbarui menyimpan konfigurasi. Ditolak bila platform belum mengaktifkan
// modul Midtrans untuk usaha ini. Server key dienkripsi sebelum disimpan;
// dikosongkan → key lama dipertahankan.
func (uc *GatewayUsecase) Perbarui(ctx context.Context, in InputGateway) (*KonfigMidtrans, error) {
	platform, err := uc.platformAktif(ctx)
	if err != nil {
		return nil, err
	}
	if !platform {
		return nil, domain.ErrModulMidtransMati
	}

	g, err := uc.repo.CariByUsaha(ctx)
	baru := false
	if err != nil {
		if !errors.Is(err, domain.ErrTidakDitemukan) {
			return nil, err
		}
		g = &domain.GatewayMidtrans{}
		baru = true
	}

	env := strings.ToLower(strings.TrimSpace(in.Env))
	if env != "production" {
		env = "sandbox"
	}
	g.Env = env
	g.MerchantID = strings.TrimSpace(in.MerchantID)
	g.ClientKey = strings.TrimSpace(in.ClientKey)
	g.Aktif = in.Aktif

	if sk := strings.TrimSpace(in.ServerKey); sk != "" {
		enc, err := uc.kotak.Enkripsi(sk)
		if err != nil {
			return nil, err
		}
		g.ServerKeyEnc = enc
	}

	if baru {
		if err := uc.repo.Simpan(ctx, g); err != nil {
			return nil, err
		}
	} else if err := uc.repo.Perbarui(ctx, g); err != nil {
		return nil, err
	}
	return uc.Ambil(ctx)
}

// KredensialAktif mengembalikan server key (PLAINTEXT, hasil dekripsi) + env
// bila gateway SIAP dipakai bertransaksi (platform on, saklar merchant on,
// server key terisi). Dipakai alur charge QRIS — TIDAK pernah mengekspos key
// ke klien; hanya usecase internal yang memanggilnya.
func (uc *GatewayUsecase) KredensialAktif(ctx context.Context) (serverKey, env string, err error) {
	platform, err := uc.platformAktif(ctx)
	if err != nil {
		return "", "", err
	}
	if !platform {
		return "", "", domain.ErrModulMidtransMati
	}
	g, err := uc.repo.CariByUsaha(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrTidakDitemukan) {
			return "", "", domain.ErrGatewayBelumSiap
		}
		return "", "", err
	}
	if !g.Aktif || g.ServerKeyEnc == "" {
		return "", "", domain.ErrGatewayBelumSiap
	}
	sk, err := uc.kotak.Dekripsi(g.ServerKeyEnc)
	if err != nil {
		return "", "", err
	}
	return sk, g.Env, nil
}

// StatusMidtrans — bentuk RINGKAS untuk aplikasi kasir: cukup untuk memutuskan
// menampilkan QRIS dinamis. TIDAK memuat server key.
type StatusMidtrans struct {
	Siap      bool
	Env       string
	ClientKey string
}

func (uc *GatewayUsecase) Status(ctx context.Context) (*StatusMidtrans, error) {
	k, err := uc.Ambil(ctx)
	if err != nil {
		return nil, err
	}
	return &StatusMidtrans{Siap: k.Siap, Env: k.Env, ClientKey: k.ClientKey}, nil
}

func mask4(s string) string {
	if len(s) <= 4 {
		return "••••"
	}
	return "••••" + s[len(s)-4:]
}

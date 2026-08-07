package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/tuleh-pos/server/internal/domain"
)

type chargerPalsu struct {
	qris       domain.HasilQris
	statusTiap []string // dikembalikan berurutan tiap StatusTransaksi
	i          int
	errCharge  error
}

func (c *chargerPalsu) ChargeQris(_ context.Context, _, _, orderID string, _ int64) (domain.HasilQris, error) {
	if c.errCharge != nil {
		return domain.HasilQris{}, c.errCharge
	}
	h := c.qris
	if h.OrderID == "" {
		h.OrderID = orderID
	}
	return h, nil
}
func (c *chargerPalsu) StatusTransaksi(_ context.Context, _, _, _ string) (string, error) {
	if c.i < len(c.statusTiap) {
		s := c.statusTiap[c.i]
		c.i++
		return s, nil
	}
	return "pending", nil
}

type tagihanRepoPalsu struct {
	data   map[uint]*domain.TagihanQris
	nextID uint
}

func newTagihanRepoPalsu() *tagihanRepoPalsu {
	return &tagihanRepoPalsu{data: map[uint]*domain.TagihanQris{}, nextID: 1}
}
func (r *tagihanRepoPalsu) Simpan(_ context.Context, t *domain.TagihanQris) error {
	t.ID = r.nextID
	r.nextID++
	r.data[t.ID] = t
	return nil
}
func (r *tagihanRepoPalsu) Perbarui(_ context.Context, t *domain.TagihanQris) error {
	r.data[t.ID] = t
	return nil
}
func (r *tagihanRepoPalsu) CariByID(_ context.Context, id uint) (*domain.TagihanQris, error) {
	if t, ok := r.data[id]; ok {
		return t, nil
	}
	return nil, domain.ErrTidakDitemukan
}

// rakit TagihanUsecase dgn gateway SIAP (platform on + gateway aktif + key).
func siapkanTagihan(t *testing.T, charger domain.GatewayCharger) (*TagihanUsecase, context.Context) {
	t.Helper()
	usahaRepo := newUsahaRepoPalsu()
	_ = usahaRepo.Simpan(context.Background(), &domain.Usaha{Kode: "U", Nama: "U", Aktif: true, MidtransAktif: true}) // id 1
	gwRepo := newGatewayRepoPalsu()
	gwRepo.data[1] = &domain.GatewayMidtrans{UsahaID: 1, Aktif: true, Env: "sandbox", ServerKeyEnc: "enc:SB-Mid-server-K"}
	gwUC := NewGatewayUsecase(gwRepo, usahaRepo, kotakPalsu{})
	uc := NewTagihanUsecase(newTagihanRepoPalsu(), gwUC, charger)
	return uc, domain.DenganUsaha(context.Background(), 1)
}

func TestBuatQrisGatewayBelumSiap(t *testing.T) {
	// Platform on tapi gateway belum ada → ErrGatewayBelumSiap.
	usahaRepo := newUsahaRepoPalsu()
	_ = usahaRepo.Simpan(context.Background(), &domain.Usaha{Kode: "U", Nama: "U", Aktif: true, MidtransAktif: true})
	gwUC := NewGatewayUsecase(newGatewayRepoPalsu(), usahaRepo, kotakPalsu{})
	uc := NewTagihanUsecase(newTagihanRepoPalsu(), gwUC, &chargerPalsu{})
	ctx := domain.DenganUsaha(context.Background(), 1)

	if _, err := uc.BuatQris(ctx, 15000); err != domain.ErrGatewayBelumSiap {
		t.Fatalf("harap ErrGatewayBelumSiap, dapat %v", err)
	}
}

func TestBuatQrisSuksesLaluPollMenjadiLunas(t *testing.T) {
	charger := &chargerPalsu{
		qris:       domain.HasilQris{QrString: "00020101", QrURL: "https://x/qr", TransactionID: "trx", StatusMentah: "pending"},
		statusTiap: []string{"pending", "settlement"},
	}
	uc, ctx := siapkanTagihan(t, charger)

	tg, err := uc.BuatQris(ctx, 15000)
	if err != nil {
		t.Fatal(err)
	}
	if tg.Status != domain.QrisPending || tg.Nominal != 15000 || tg.QrString == "" {
		t.Fatalf("tagihan awal salah: %+v", tg)
	}

	// Poll 1 → masih pending.
	if p, _ := uc.CekStatus(ctx, tg.ID); p.Status != domain.QrisPending {
		t.Fatalf("poll-1 harus PENDING, dapat %s", p.Status)
	}
	// Poll 2 → settlement → PAID.
	p2, err := uc.CekStatus(ctx, tg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p2.Status != domain.QrisPaid {
		t.Fatalf("poll-2 harus PAID, dapat %s", p2.Status)
	}
	// Poll 3 → terminal, TAK memanggil charger lagi (i tetap 2).
	if _, err := uc.CekStatus(ctx, tg.ID); err != nil {
		t.Fatal(err)
	}
	if charger.i != 2 {
		t.Fatalf("status terminal tak boleh poll lagi; panggilan charger=%d", charger.i)
	}
}

func TestBuatQrisMidtransGagalDibungkusUpstream(t *testing.T) {
	charger := &chargerPalsu{errCharge: errors.New("midtrans 401: invalid key")}
	uc, ctx := siapkanTagihan(t, charger)

	_, err := uc.BuatQris(ctx, 5000)
	if !errors.Is(err, domain.ErrGatewayUpstream) {
		t.Fatalf("kegagalan Midtrans harus dibungkus ErrGatewayUpstream, dapat %v", err)
	}
}

func TestBuatQrisNominalNol(t *testing.T) {
	uc, ctx := siapkanTagihan(t, &chargerPalsu{})
	if _, err := uc.BuatQris(ctx, 0); err != domain.ErrNominalTakSah {
		t.Fatalf("harap ErrNominalTakSah, dapat %v", err)
	}
}

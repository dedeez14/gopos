package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

// TagihanUsecase — alur pembayaran QRIS dinamis Midtrans (charge + poll).
// Kredensial merchant diambil terdekripsi dari GatewayUsecase; error API
// Midtrans dibungkus ErrGatewayUpstream agar dipetakan 502, bukan 500.
type TagihanUsecase struct {
	repo    domain.TagihanQrisRepository
	gateway *GatewayUsecase
	charger domain.GatewayCharger
}

func NewTagihanUsecase(repo domain.TagihanQrisRepository, gateway *GatewayUsecase, charger domain.GatewayCharger) *TagihanUsecase {
	return &TagihanUsecase{repo: repo, gateway: gateway, charger: charger}
}

// BuatQris membuat tagihan QRIS baru di Midtrans dan menyimpannya PENDING.
func (uc *TagihanUsecase) BuatQris(ctx context.Context, nominal int64) (*domain.TagihanQris, error) {
	if nominal <= 0 {
		return nil, domain.ErrNominalTakSah
	}
	sk, env, err := uc.gateway.KredensialAktif(ctx)
	if err != nil {
		return nil, err
	}

	orderID := fmt.Sprintf("TQR-%d-%d-%s", domain.UsahaDari(ctx), time.Now().UnixNano(), acakHex(4))
	hasil, err := uc.charger.ChargeQris(ctx, sk, env, orderID, nominal)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrGatewayUpstream, err)
	}

	t := &domain.TagihanQris{
		OrderID: hasil.OrderID, Nominal: nominal, Status: statusDari(hasil.StatusMentah),
		QrString: hasil.QrString, QrURL: hasil.QrURL, MidtransTrxID: hasil.TransactionID,
		Kedaluwarsa: hasil.KedaluwarsaISO,
	}
	if t.OrderID == "" {
		t.OrderID = orderID // jaga-jaga bila Midtrans tak mengembalikannya
	}
	if err := uc.repo.Simpan(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// CekStatus memoll status tagihan. Status terminal (PAID/EXPIRED/FAILED) tak
// perlu menanya Midtrans lagi — dikembalikan apa adanya.
func (uc *TagihanUsecase) CekStatus(ctx context.Context, id uint) (*domain.TagihanQris, error) {
	t, err := uc.repo.CariByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.QrisPending {
		return t, nil
	}
	sk, env, err := uc.gateway.KredensialAktif(ctx)
	if err != nil {
		return nil, err
	}
	mentah, err := uc.charger.StatusTransaksi(ctx, sk, env, t.OrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrGatewayUpstream, err)
	}
	if baru := statusDari(mentah); baru != t.Status {
		t.Status = baru
		if err := uc.repo.Perbarui(ctx, t); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// statusDari memetakan transaction_status Midtrans → status internal.
func statusDari(mentah string) string {
	switch mentah {
	case "settlement", "capture":
		return domain.QrisPaid
	case "expire":
		return domain.QrisExpired
	case "deny", "cancel", "failure":
		return domain.QrisFailed
	default:
		return domain.QrisPending
	}
}

func acakHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "0000"
	}
	return hex.EncodeToString(b)
}

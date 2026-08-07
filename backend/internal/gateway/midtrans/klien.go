// Package midtrans adalah implementasi konkret domain.GatewayCharger di atas
// Core API Midtrans. Ini lapisan LUAR (delivery ke pihak ketiga) — domain &
// usecase hanya bergantung pada kontrak, bukan paket ini.
//
// Gotcha Midtrans yang ditangani di sini: respons bisa HTTP 200/201 TAPI
// membawa status_code non-2xx di BODY (mis. 401 kunci salah, 404 tak ada).
// Karena itu keputusan sukses/gagal diambil dari body.status_code, bukan
// kode HTTP.
package midtrans

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

type Klien struct {
	http           *http.Client
	baseSandbox    string
	baseProduction string
}

func NewKlien() *Klien {
	return &Klien{
		http:           &http.Client{Timeout: 15 * time.Second},
		baseSandbox:    "https://api.sandbox.midtrans.com",
		baseProduction: "https://api.midtrans.com",
	}
}

func (k *Klien) base(env string) string {
	if strings.EqualFold(env, "production") {
		return k.baseProduction
	}
	return k.baseSandbox
}

// respMidtrans menampung field yang kita pakai dari charge & status.
type respMidtrans struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	QrString          string `json:"qr_string"`
	ExpiryTime        string `json:"expiry_time"`
	Actions           []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"actions"`
}

func (k *Klien) ChargeQris(ctx context.Context, serverKey, env, orderID string, nominal int64) (domain.HasilQris, error) {
	body := map[string]any{
		"payment_type": "qris",
		"transaction_details": map[string]any{
			"order_id":     orderID,
			"gross_amount": nominal, // IDR bulat
		},
		"qris": map[string]any{"acquirer": "gopay"},
	}
	var r respMidtrans
	if err := k.kirim(ctx, serverKey, http.MethodPost, k.base(env)+"/v2/charge", body, &r); err != nil {
		return domain.HasilQris{}, err
	}
	if !suksesKode(r.StatusCode) {
		return domain.HasilQris{}, fmt.Errorf("midtrans %s: %s", r.StatusCode, r.StatusMessage)
	}
	return domain.HasilQris{
		OrderID: r.OrderID, TransactionID: r.TransactionID,
		QrString: r.QrString, QrURL: urlGenerateQR(r),
		StatusMentah: r.TransactionStatus, KedaluwarsaISO: r.ExpiryTime,
	}, nil
}

func (k *Klien) StatusTransaksi(ctx context.Context, serverKey, env, orderID string) (string, error) {
	var r respMidtrans
	if err := k.kirim(ctx, serverKey, http.MethodGet, k.base(env)+"/v2/"+orderID+"/status", nil, &r); err != nil {
		return "", err
	}
	// 404 = order tak ditemukan → anggap gagal, bukan error jaringan.
	if !suksesKode(r.StatusCode) && r.StatusCode != "404" {
		return "", fmt.Errorf("midtrans %s: %s", r.StatusCode, r.StatusMessage)
	}
	return r.TransactionStatus, nil
}

func (k *Klien) kirim(ctx context.Context, serverKey, metode, url string, body any, out *respMidtrans) error {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, metode, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(serverKey, "") // Basic base64(serverKey:) — password kosong

	res, err := k.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("respons midtrans tak terbaca (http %d)", res.StatusCode)
	}
	return nil
}

// suksesKode: status_code Midtrans "2xx" (200/201) = berhasil.
func suksesKode(kode string) bool { return strings.HasPrefix(kode, "2") }

func urlGenerateQR(r respMidtrans) string {
	for _, a := range r.Actions {
		if a.Name == "generate-qr-code" {
			return a.URL
		}
	}
	return ""
}

package midtrans

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// klienUji mengarahkan base URL ke server httptest.
func klienUji(srv *httptest.Server) *Klien {
	return &Klien{http: srv.Client(), baseSandbox: srv.URL, baseProduction: srv.URL}
}

func TestChargeQrisSukses(t *testing.T) {
	var gotAuthUser, gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if u, _, ok := r.BasicAuth(); ok {
			gotAuthUser = u
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status_code":"201","transaction_id":"trx-1","order_id":"ORD-1",
			"transaction_status":"pending","qr_string":"00020101...","expiry_time":"2026-08-07 12:00:00",
			"actions":[{"name":"generate-qr-code","url":"https://x/qr.png"}]}`))
	}))
	defer srv.Close()

	h, err := klienUji(srv).ChargeQris(context.Background(), "SB-Mid-server-KEY", "sandbox", "ORD-1", 15000)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v2/charge" {
		t.Fatalf("path=%q", gotPath)
	}
	if gotAuthUser != "SB-Mid-server-KEY" {
		t.Fatalf("basic auth user=%q (server key harus jadi username)", gotAuthUser)
	}
	if gotBody["payment_type"] != "qris" {
		t.Fatalf("payment_type=%v", gotBody["payment_type"])
	}
	td, _ := gotBody["transaction_details"].(map[string]any)
	if td["order_id"] != "ORD-1" || td["gross_amount"].(float64) != 15000 {
		t.Fatalf("transaction_details salah: %v", td)
	}
	if h.QrString == "" || h.QrURL != "https://x/qr.png" || h.StatusMentah != "pending" {
		t.Fatalf("hasil salah: %+v", h)
	}
}

// Gotcha: Midtrans balas HTTP 200 TAPI status_code non-2xx di body → error.
func TestChargeQrisBodyStatusGagal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK) // HTTP 200…
		_, _ = w.Write([]byte(`{"status_code":"401","status_message":"unauthorized. invalid server key"}`))
	}))
	defer srv.Close()

	_, err := klienUji(srv).ChargeQris(context.Background(), "salah", "sandbox", "ORD-2", 1000)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("harus gagal karena status_code 401 di body, dapat %v", err)
	}
}

func TestStatusTransaksi(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"status_code":"200","transaction_status":"settlement"}`))
	}))
	defer srv.Close()

	st, err := klienUji(srv).StatusTransaksi(context.Background(), "key", "sandbox", "ORD-9")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v2/ORD-9/status" {
		t.Fatalf("path=%q", gotPath)
	}
	if st != "settlement" {
		t.Fatalf("status=%q", st)
	}
}

// Basic auth Midtrans = base64(serverKey + ":") — pastikan password kosong.
func TestBasicAuthFormat(t *testing.T) {
	var raw string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw = strings.TrimPrefix(r.Header.Get("Authorization"), "Basic ")
		_, _ = w.Write([]byte(`{"status_code":"200","transaction_status":"pending"}`))
	}))
	defer srv.Close()

	_, _ = klienUji(srv).StatusTransaksi(context.Background(), "KUNCI", "sandbox", "O")
	dec, _ := base64.StdEncoding.DecodeString(raw)
	if string(dec) != "KUNCI:" {
		t.Fatalf("basic auth decode=%q, harap 'KUNCI:'", string(dec))
	}
}

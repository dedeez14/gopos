package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

// TransaksiUsecase — checkout kasir. Rumus SAMA PERSIS dengan kontrak Tuléh
// di MOVERA (klien tidak boleh melihat angka berbeda saat cutover):
//
//	per item : subtotal = qty × harga
//	           diskon   = subtotal × diskon_persen%
//	           pajak    = (subtotal − diskon) × pajak_persen%
//	transaksi: dasar          = Σsubtotal − Σdiskon item
//	           diskon nominal dipotong dulu (di-cap ≤ dasar),
//	           lalu diskon persen dari SISANYA;
//	           pajak tidak berubah oleh diskon transaksi.
//	grand    = Σsubtotal − totalDiskon + Σpajak   (tak pernah negatif)
type TransaksiUsecase struct {
	transaksi domain.TransaksiRepository
	sesi      domain.SesiRepository
	produk    domain.ProdukRepository
	pelanggan domain.PelangganRepository
}

func NewTransaksiUsecase(
	transaksi domain.TransaksiRepository,
	sesi domain.SesiRepository,
	produk domain.ProdukRepository,
	pelanggan domain.PelangganRepository,
) *TransaksiUsecase {
	return &TransaksiUsecase{transaksi: transaksi, sesi: sesi, produk: produk, pelanggan: pelanggan}
}

// ItemCheckout — baris keranjang. Harga 0 = pakai harga efektif katalog
// (promo dihormati otomatis); terisi = harga yang dipakai kasir (kontrak
// lama Tuléh mengirim harga).
type ItemCheckout struct {
	ProdukID     uint
	Kuantitas    float64
	Harga        float64
	DiskonPersen float64
	PajakPersen  float64
}

type InputCheckout struct {
	Items          []ItemCheckout
	DiskonPersen   float64 // diskon keseluruhan (%)
	DiskonNominal  float64 // diskon keseluruhan (Rp)
	TipePembayaran string
	Dibayar        float64
	Catatan        string
	IdempotencyKey string
	PelangganID    uint // 0 = tanpa pelanggan
}

// Checkout memproses penjualan. Wajib ada sesi BUKA milik kasir; kunci
// idempoten yang sama TIDAK memproses dua kali (retry jaringan aman).
func (uc *TransaksiUsecase) Checkout(ctx context.Context, userID uint, in InputCheckout) (*domain.Transaksi, error) {
	sesi, err := uc.sesi.AktifMilik(ctx, userID)
	if err != nil {
		return nil, domain.ErrSesiBelumBuka
	}

	if in.IdempotencyKey != "" {
		if lama, err := uc.transaksi.CariByIdempotency(ctx, in.IdempotencyKey); err == nil {
			return lama, nil // pengulangan — kembalikan transaksi asli
		} else if !errors.Is(err, domain.ErrTidakDitemukan) {
			return nil, err
		}
	}

	// Pelanggan opsional — id asing ditolak, bukan diterima buta.
	var pelangganID *uint
	if in.PelangganID != 0 {
		if _, err := uc.pelanggan.CariByID(ctx, in.PelangganID); err != nil {
			return nil, err
		}
		id := in.PelangganID
		pelangganID = &id
	}

	kini := time.Now()
	var (
		subtotal, totalDiskonItem, totalPajak float64
		items                                 []domain.TransaksiItem
		potongStok                            = map[uint]float64{}
	)

	for _, baris := range in.Items {
		p, err := uc.produk.CariByID(ctx, baris.ProdukID)
		if err != nil {
			return nil, err
		}
		if !p.Aktif {
			return nil, domain.ErrProdukNonaktif
		}

		harga := baris.Harga
		if harga <= 0 {
			harga = p.HargaEfektif(kini)
		}

		itemSubtotal := baris.Kuantitas * harga
		itemDiskon := itemSubtotal * klampPersen(baris.DiskonPersen) / 100
		itemPajak := (itemSubtotal - itemDiskon) * klampPersen(baris.PajakPersen) / 100

		subtotal += itemSubtotal
		totalDiskonItem += itemDiskon
		totalPajak += itemPajak

		items = append(items, domain.TransaksiItem{
			ProdukID:     p.ID,
			NamaProduk:   p.Nama, // snapshot — riwayat kebal rename
			Satuan:       p.Satuan,
			Harga:        harga,
			Kuantitas:    baris.Kuantitas,
			DiskonPersen: klampPersen(baris.DiskonPersen),
			PajakPersen:  klampPersen(baris.PajakPersen),
			Subtotal:     bulatkan(itemSubtotal - itemDiskon + itemPajak),
		})

		if p.KelolaStok {
			potongStok[p.ID] += baris.Kuantitas
		}
	}

	// Diskon keseluruhan: nominal dulu (cap ≤ dasar), persen dari sisa.
	dasar := subtotal - totalDiskonItem
	diskonNominal := math.Min(dasar, math.Max(0, in.DiskonNominal))
	diskonPersen := klampPersen(in.DiskonPersen)
	diskonTransaksi := diskonNominal + (dasar-diskonNominal)*diskonPersen/100

	totalDiskon := totalDiskonItem + diskonTransaksi
	grand := bulatkan(subtotal - totalDiskon + totalPajak)

	t := &domain.Transaksi{
		Nomor:          fmt.Sprintf("POSTMP-%d", kini.UnixNano()), // final diisi repo: POS-<tgl>-<id>
		SesiKasirID:    sesi.ID,
		UserID:         userID,
		PelangganID:    pelangganID,
		Tanggal:        kini,
		Subtotal:       bulatkan(subtotal),
		TotalDiskon:    bulatkan(totalDiskon),
		DiskonPersen:   diskonPersen,
		DiskonNominal:  bulatkan(diskonNominal),
		TotalPajak:     bulatkan(totalPajak),
		GrandTotal:     grand,
		Dibayar:        in.Dibayar,
		Kembalian:      math.Max(0, bulatkan(in.Dibayar-grand)),
		TipePembayaran: in.TipePembayaran,
		Status:         domain.TrxSelesai,
		Catatan:        in.Catatan,
		Items:          items,
	}
	if in.IdempotencyKey != "" {
		k := in.IdempotencyKey
		t.IdempotencyKey = &k
	}

	if err := uc.transaksi.Checkout(ctx, t, potongStok); err != nil {
		return nil, err
	}
	// Muat ulang supaya relasi (Pelanggan, User) terisi di struk.
	return uc.transaksi.CariByID(ctx, t.ID)
}

func (uc *TransaksiUsecase) Ambil(ctx context.Context, id uint) (*domain.Transaksi, error) {
	return uc.transaksi.CariByID(ctx, id)
}

// Batal membatalkan transaksi SELESAI: status → DIBATALKAN + stok kembali.
// Aturan:
//   - hanya selama SESI transaksi masih BUKA — sesi yang sudah ditutup
//     kasnya sudah dihitung; koreksi setelah itu = dokumen retur (menyusul),
//     bukan mutasi diam-diam;
//   - kasir hanya boleh membatalkan transaksinya sendiri; bolehLintas
//     (peran ber-izin laporan) boleh membatalkan milik siapa pun.
func (uc *TransaksiUsecase) Batal(ctx context.Context, trxID, userID uint, bolehLintas bool) (*domain.Transaksi, error) {
	t, err := uc.transaksi.CariByID(ctx, trxID)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TrxSelesai {
		return nil, domain.ErrTrxSudahBatal
	}
	if !bolehLintas && t.UserID != userID {
		return nil, domain.ErrTrxBukanMilik
	}
	sesi, err := uc.sesi.CariByID(ctx, t.SesiKasirID)
	if err != nil {
		return nil, err
	}
	if sesi.Status != domain.SesiBuka {
		return nil, domain.ErrSesiTrxSudahTutup
	}

	// Kembalikan stok hanya untuk item yang produknya kelola stok.
	kembalikan := map[uint]float64{}
	for _, it := range t.Items {
		if p, err := uc.produk.CariByID(ctx, it.ProdukID); err == nil && p.KelolaStok {
			kembalikan[it.ProdukID] += it.Kuantitas
		}
	}

	if err := uc.transaksi.Batalkan(ctx, t, kembalikan); err != nil {
		return nil, err
	}
	return uc.transaksi.CariByID(ctx, t.ID)
}

func (uc *TransaksiUsecase) Daftar(ctx context.Context, f domain.FilterTransaksi) ([]domain.Transaksi, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.transaksi.Daftar(ctx, f)
}

func klampPersen(v float64) float64 { return math.Min(100, math.Max(0, v)) }

// bulatkan ke 2 desimal — nilai uang disimpan rapi, tanpa sisa floating.
func bulatkan(v float64) float64 { return math.Round(v*100) / 100 }

package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

// ProdukUsecase — aturan katalog. Aturan bisnis yang DITEGAKKAN di sini
// (bukan di handler, supaya berlaku dari jalur mana pun):
//   - JASA selalu kelola_stok=false (layanan tak punya fisik; menyalakannya
//     membuat item lenyap dari katalog begitu "stok"-nya nol).
//   - Kode unik; kosong → digenerate (P-<detik-unix>).
//   - Menghapus = menonaktifkan (riwayat transaksi tak boleh kehilangan nama).
type ProdukUsecase struct {
	produk   domain.ProdukRepository
	kategori domain.KategoriRepository
}

func NewProdukUsecase(produk domain.ProdukRepository, kategori domain.KategoriRepository) *ProdukUsecase {
	return &ProdukUsecase{produk: produk, kategori: kategori}
}

// InputProduk — data masuk create/update, eksplisit per field.
type InputProduk struct {
	Nama         string
	Kode         string // kosong = generate
	Barcode      string
	Tipe         domain.TipeProduk
	Satuan       string
	HargaBeli    float64
	HargaJual    float64
	HargaPromo   *float64 // nil = tanpa promo / cabut promo
	PromoMulai   *time.Time
	PromoSelesai *time.Time
	Favorit      bool
	KelolaStok   bool
	KategoriID   uint // 0 = tanpa kategori
	Aktif        bool
}

func (uc *ProdukUsecase) Daftar(ctx context.Context, f domain.FilterProduk) ([]domain.Produk, int64, error) {
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}
	return uc.produk.Daftar(ctx, f)
}

func (uc *ProdukUsecase) Ambil(ctx context.Context, id uint) (*domain.Produk, error) {
	return uc.produk.CariByID(ctx, id)
}

func (uc *ProdukUsecase) Buat(ctx context.Context, in InputProduk) (*domain.Produk, error) {
	p := &domain.Produk{}
	if err := uc.terapkan(ctx, p, in, true); err != nil {
		return nil, err
	}
	if err := uc.produk.Simpan(ctx, p); err != nil {
		return nil, err
	}
	// Muat ulang supaya relasi (Kategori) ikut terisi di respons.
	return uc.produk.CariByID(ctx, p.ID)
}

func (uc *ProdukUsecase) Perbarui(ctx context.Context, id uint, in InputProduk) (*domain.Produk, error) {
	p, err := uc.produk.CariByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := uc.terapkan(ctx, p, in, false); err != nil {
		return nil, err
	}
	if err := uc.produk.Perbarui(ctx, p); err != nil {
		return nil, err
	}
	return uc.produk.CariByID(ctx, p.ID)
}

func (uc *ProdukUsecase) Nonaktifkan(ctx context.Context, id uint) error {
	if _, err := uc.produk.CariByID(ctx, id); err != nil {
		return err
	}
	return uc.produk.Nonaktifkan(ctx, id)
}

// terapkan memindahkan input → entitas + menegakkan seluruh aturan bisnis.
// baru=true saat create (menentukan perlakuan kode unik).
func (uc *ProdukUsecase) terapkan(ctx context.Context, p *domain.Produk, in InputProduk, baru bool) error {
	kode := strings.TrimSpace(in.Kode)
	if kode == "" && baru {
		kode = fmt.Sprintf("P-%d", time.Now().UnixNano()/1e6)
	}
	if kode != "" && kode != p.Kode {
		if lain, err := uc.produk.CariByKode(ctx, kode); err == nil && lain.ID != p.ID {
			return domain.ErrKodeTerpakai
		} else if err != nil && !errors.Is(err, domain.ErrTidakDitemukan) {
			return err
		}
		p.Kode = kode
	}

	if in.KategoriID != 0 {
		// Kategori di-resolve — id asing ditolak, bukan diterima buta.
		if _, err := uc.kategori.CariByID(ctx, in.KategoriID); err != nil {
			return err
		}
		id := in.KategoriID
		p.KategoriID = &id
	} else {
		p.KategoriID = nil
	}

	p.Nama = strings.TrimSpace(in.Nama)
	p.Tipe = in.Tipe
	p.Satuan = strings.TrimSpace(in.Satuan)
	if p.Satuan == "" {
		p.Satuan = "pcs"
	}
	if b := strings.TrimSpace(in.Barcode); b != "" {
		p.Barcode = &b
	} else {
		p.Barcode = nil
	}
	p.HargaBeli = in.HargaBeli
	p.HargaJual = in.HargaJual
	p.HargaPromo = in.HargaPromo
	p.PromoMulai = in.PromoMulai
	p.PromoSelesai = in.PromoSelesai
	if in.HargaPromo == nil {
		// Cabut promo = periode ikut bersih, tak ada sisa data menyesatkan.
		p.PromoMulai, p.PromoSelesai = nil, nil
	}
	p.Favorit = in.Favorit
	p.Aktif = in.Aktif

	// Aturan: JASA tidak pernah kelola stok.
	if in.Tipe == domain.TipeJasa {
		p.KelolaStok = false
	} else {
		p.KelolaStok = in.KelolaStok
	}

	return nil
}

// ─────────────────────────────────────────────────────────────── kategori

type KategoriUsecase struct {
	repo domain.KategoriRepository
}

func NewKategoriUsecase(repo domain.KategoriRepository) *KategoriUsecase {
	return &KategoriUsecase{repo: repo}
}

func (uc *KategoriUsecase) Daftar(ctx context.Context) ([]domain.Kategori, error) {
	return uc.repo.Daftar(ctx)
}

func (uc *KategoriUsecase) Buat(ctx context.Context, nama string) (*domain.Kategori, error) {
	k := &domain.Kategori{Nama: strings.TrimSpace(nama)}
	if err := uc.repo.Simpan(ctx, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (uc *KategoriUsecase) Hapus(ctx context.Context, id uint) error {
	if _, err := uc.repo.CariByID(ctx, id); err != nil {
		return err
	}
	return uc.repo.Hapus(ctx, id)
}

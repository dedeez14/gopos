// Perintah sinkron — fase-2 strangler: menyalin data MASTER dari MOVERA
// (MySQL, system-of-record) ke Tuléh Server (PostgreSQL) untuk SATU company.
//
//	SINKRON_MYSQL_DSN=user:pass@tcp(host:3306)/movera_production?parseTime=true \
//	SINKRON_COMPANY_KODE=TULEH-DEMO \
//	go run ./cmd/sinkron
//
// Prinsip keras:
//   - MySQL HANYA DIBACA (jalankan dengan user SELECT-only, mis. tuleh_sync).
//   - Idempoten: upsert by kode (produk/kategori) & telepon-ternormalisasi
//     (pelanggan) — aman dijalankan berulang / dijadwalkan.
//   - Stok dihitung dari SUM(persediaan_lapisan_stok.sisa_kuantitas) — kolom
//     stok_saat_ini di MOVERA adalah kolom mati, JANGAN dipakai.
//
// Arah satu jalan: selama fase koeksistensi, Laravel tetap system-of-record;
// perintah ini menyegarkan replika Postgres. Tidak ada tulis-balik.
package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/tuleh-pos/server/internal/config"
	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
)

// baris hasil query produk dari MOVERA.
type produkSumber struct {
	Kode         string
	Nama         string
	Barcode      *string
	Tipe         string
	Satuan       *string
	Kategori     *string
	HargaBeli    float64
	HargaJual    float64
	HargaPromo   *float64   `gorm:"column:pos_harga_promo"`
	PromoMulai   *time.Time `gorm:"column:pos_promo_mulai"`
	PromoSelesai *time.Time `gorm:"column:pos_promo_selesai"`
	Favorit      bool       `gorm:"column:pos_favorit"`
	KelolaStok   bool
	Aktif        bool `gorm:"column:is_active"`
	Stok         float64
}

type mitraSumber struct {
	Nama    string
	Telepon *string
	Email   *string
	Aktif   bool `gorm:"column:is_active"`
}

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	cfg, err := config.Muat()
	if err != nil {
		log.Fatal().Err(err).Msg("konfigurasi tidak sah")
	}
	dsnMySQL := strings.TrimSpace(os.Getenv("SINKRON_MYSQL_DSN"))
	if dsnMySQL == "" {
		log.Fatal().Msg("SINKRON_MYSQL_DSN wajib diisi (user SELECT-only)")
	}
	kodeCompany := strings.TrimSpace(os.Getenv("SINKRON_COMPANY_KODE"))
	if kodeCompany == "" {
		kodeCompany = "TULEH-DEMO"
	}

	senyap := &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)}
	sumber, err := gorm.Open(mysql.Open(dsnMySQL), senyap)
	if err != nil {
		log.Fatal().Err(err).Msg("gagal terhubung ke MySQL MOVERA")
	}
	tujuan, err := gorm.Open(postgres.Open(cfg.DSN()), senyap)
	if err != nil {
		log.Fatal().Err(err).Msg("gagal terhubung ke PostgreSQL")
	}

	ctx := context.Background()

	// ── company ──────────────────────────────────────────────────────────
	var companyID uint
	if err := sumber.WithContext(ctx).Raw(
		`SELECT id FROM companies WHERE kode = ?`, kodeCompany).Scan(&companyID).Error; err != nil || companyID == 0 {
		log.Fatal().Err(err).Str("kode", kodeCompany).Msg("company sumber tidak ditemukan")
	}
	log.Info().Uint("company_id", companyID).Str("kode", kodeCompany).Msg("mulai sinkron")

	// ── produk (+kategori) ───────────────────────────────────────────────
	var produk []produkSumber
	if err := sumber.WithContext(ctx).Raw(`
		SELECT p.kode, p.nama, p.barcode, p.tipe,
		       s.nama AS satuan, k.nama AS kategori,
		       p.harga_beli, p.harga_jual,
		       p.pos_harga_promo, p.pos_promo_mulai, p.pos_promo_selesai,
		       p.pos_favorit, p.kelola_stok, p.is_active,
		       COALESCE((SELECT SUM(ls.sisa_kuantitas)
		                 FROM persediaan_lapisan_stok ls
		                 WHERE ls.id_produk = p.id), 0) AS stok
		FROM master_produk p
		LEFT JOIN master_satuan s ON s.id = p.satuan_id
		LEFT JOIN master_kategori_produk k ON k.id = p.kategori_id
		WHERE p.company_id = ?`, companyID).Scan(&produk).Error; err != nil {
		log.Fatal().Err(err).Msg("gagal membaca produk sumber")
	}

	kategoriID := map[string]uint{}
	var pBaru, pUbah int
	for _, src := range produk {
		// Kategori di-upsert by nama (skema Go tak mengenal id MOVERA).
		var katID *uint
		if src.Kategori != nil && *src.Kategori != "" {
			id, ok := kategoriID[*src.Kategori]
			if !ok {
				var k domain.Kategori
				err := tujuan.WithContext(ctx).Where("nama = ?", *src.Kategori).First(&k).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					k = domain.Kategori{Nama: *src.Kategori}
					if err := tujuan.WithContext(ctx).Create(&k).Error; err != nil {
						log.Fatal().Err(err).Msg("gagal upsert kategori")
					}
				} else if err != nil {
					log.Fatal().Err(err).Msg("gagal membaca kategori tujuan")
				}
				id = k.ID
				kategoriID[*src.Kategori] = id
			}
			katID = &id
		}

		tipe := domain.TipeBarang
		if strings.EqualFold(src.Tipe, "JASA") {
			tipe = domain.TipeJasa
		}
		satuan := "pcs"
		if src.Satuan != nil && *src.Satuan != "" {
			satuan = *src.Satuan
		}

		target := domain.Produk{
			Kode: src.Kode, Nama: src.Nama, Barcode: src.Barcode, Tipe: tipe,
			Satuan: satuan, HargaBeli: src.HargaBeli, HargaJual: src.HargaJual,
			HargaPromo: src.HargaPromo, PromoMulai: src.PromoMulai, PromoSelesai: src.PromoSelesai,
			Favorit: src.Favorit, KelolaStok: src.KelolaStok && tipe != domain.TipeJasa,
			Stok: src.Stok, KategoriID: katID, Aktif: src.Aktif,
		}

		var lama domain.Produk
		err := tujuan.WithContext(ctx).Where("kode = ?", src.Kode).First(&lama).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tujuan.WithContext(ctx).Create(&target).Error; err != nil {
				log.Fatal().Err(err).Str("kode", src.Kode).Msg("gagal insert produk")
			}
			pBaru++
		case err != nil:
			log.Fatal().Err(err).Msg("gagal membaca produk tujuan")
		default:
			target.ID = lama.ID
			if err := tujuan.WithContext(ctx).Model(&lama).Select("*").
				Omit("id", "created_at").Updates(&target).Error; err != nil {
				log.Fatal().Err(err).Str("kode", src.Kode).Msg("gagal update produk")
			}
			pUbah++
		}
	}
	log.Info().Int("baru", pBaru).Int("diperbarui", pUbah).Int("total_sumber", len(produk)).Msg("produk selesai")

	// ── pelanggan ────────────────────────────────────────────────────────
	var mitra []mitraSumber
	if err := sumber.WithContext(ctx).Raw(`
		SELECT nama, telepon, email, is_active
		FROM master_mitra
		WHERE company_id = ? AND tipe IN ('PELANGGAN','KEDUA')`, companyID).Scan(&mitra).Error; err != nil {
		log.Fatal().Err(err).Msg("gagal membaca pelanggan sumber")
	}

	var mBaru, mUbah, mLewati int
	for _, src := range mitra {
		tel := ""
		if src.Telepon != nil {
			tel = usecase.NormalisasiTelepon(*src.Telepon)
		}

		// Kunci pencocokan: telepon ternormalisasi; tanpa telepon → nama
		// persis. Ambigu (nama sama tanpa telepon) dilewati, tidak ditebak.
		var lama domain.Pelanggan
		var err error
		if tel != "" {
			err = tujuan.WithContext(ctx).Where("telepon = ?", tel).First(&lama).Error
		} else {
			err = tujuan.WithContext(ctx).Where("nama = ? AND telepon IS NULL", src.Nama).First(&lama).Error
		}

		target := domain.Pelanggan{Nama: src.Nama, Email: src.Email, Aktif: src.Aktif}
		if tel != "" {
			target.Telepon = &tel
		}

		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tujuan.WithContext(ctx).Create(&target).Error; err != nil {
				log.Warn().Err(err).Str("nama", src.Nama).Msg("pelanggan dilewati")
				mLewati++
				continue
			}
			mBaru++
		case err != nil:
			log.Fatal().Err(err).Msg("gagal membaca pelanggan tujuan")
		default:
			target.ID = lama.ID
			if err := tujuan.WithContext(ctx).Model(&lama).Select("*").
				Omit("id", "created_at").Updates(&target).Error; err != nil {
				log.Warn().Err(err).Str("nama", src.Nama).Msg("pelanggan gagal update")
				mLewati++
				continue
			}
			mUbah++
		}
	}
	log.Info().Int("baru", mBaru).Int("diperbarui", mUbah).Int("dilewati", mLewati).
		Int("total_sumber", len(mitra)).Msg("pelanggan selesai")

	log.Info().Msg("sinkron selesai — MOVERA tetap system-of-record, ini replika baca")
}

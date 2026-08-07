// Perintah seed — data DEFAULT Tuléh Server: akun tiga peran, kategori umum,
// dan beberapa contoh produk. Semuanya titik awal yang DINAMIS: ubah/hapus
// lewat admin panel kapan pun.
//
// Idempoten: dicocokkan by email (user) / nama (kategori) / kode (produk) —
// aman dijalankan berulang; data yang sudah Anda ubah lewat panel TIDAK
// ditimpa (seed hanya membuat yang belum ada).
//
//	set -a; source .env; set +a; go run ./cmd/seed
package main

import (
	"errors"
	"os"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/tuleh-pos/server/internal/config"
	"github.com/tuleh-pos/server/internal/domain"
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	cfg, err := config.Muat()
	if err != nil {
		log.Fatal().Err(err).Msg("konfigurasi tidak sah")
	}
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("gagal terhubung ke PostgreSQL")
	}
	if err := db.AutoMigrate(
		&domain.User{}, &domain.Kategori{}, &domain.Produk{},
		&domain.SesiKasir{}, &domain.Transaksi{}, &domain.TransaksiItem{},
		&domain.Pelanggan{}, &domain.Hold{}, &domain.Pengeluaran{}, &domain.StokLog{},
	); err != nil {
		log.Fatal().Err(err).Msg("migrasi gagal")
	}

	// ── akun default (GANTI SANDINYA lewat panel: menu Pengguna) ─────────
	akun := []struct {
		Nama, Email, Sandi string
		Role               domain.Role
	}{
		{"Pemilik Usaha", cfg.AdminEmail, cfg.AdminPass, domain.RoleOwner},
		{"Manajer Toko", "manager@tuleh.id", "gopos-manager-2026", domain.RoleManager},
		{"Kasir Satu", "kasir@tuleh.id", "gopos-kasir-2026", domain.RoleKasir},
	}
	for _, a := range akun {
		var ada domain.User
		err := db.Where("email = ?", a.Email).First(&ada).Error
		if err == nil {
			continue // sudah ada — jangan timpa sandi yang mungkin sudah diganti
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal().Err(err).Msg("gagal membaca users")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(a.Sandi), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal().Err(err).Msg("gagal hash sandi")
		}
		if err := db.Create(&domain.User{
			Nama: a.Nama, Email: a.Email, PasswordHash: string(hash),
			Role: a.Role, Aktif: true,
		}).Error; err != nil {
			log.Fatal().Err(err).Msg("gagal membuat akun")
		}
		log.Info().Str("email", a.Email).Str("role", string(a.Role)).Msg("akun dibuat")
	}

	// ── kategori umum ────────────────────────────────────────────────────
	kategoriID := map[string]uint{}
	for _, nama := range []string{"Makanan", "Minuman", "Snack", "Sembako", "Lainnya"} {
		var k domain.Kategori
		err := db.Where("nama = ?", nama).First(&k).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			k = domain.Kategori{Nama: nama}
			if err := db.Create(&k).Error; err != nil {
				log.Fatal().Err(err).Msg("gagal membuat kategori")
			}
			log.Info().Str("kategori", nama).Msg("kategori dibuat")
		} else if err != nil {
			log.Fatal().Err(err).Msg("gagal membaca kategori")
		}
		kategoriID[nama] = k.ID
	}

	// ── contoh produk (stok 0 — isi lewat menu Inventory → Tambah Stok) ──
	f := func(v float64) *float64 { return &v }
	contoh := []domain.Produk{
		{Kode: "CTH-001", Nama: "Kopi Hitam", Tipe: domain.TipeBarang, Satuan: "cup", HargaBeli: 2000, HargaJual: 5000, Favorit: true, KelolaStok: true},
		{Kode: "CTH-002", Nama: "Es Teh Manis", Tipe: domain.TipeBarang, Satuan: "cup", HargaBeli: 1500, HargaJual: 4000, Favorit: true, KelolaStok: true},
		{Kode: "CTH-003", Nama: "Indomie Goreng", Tipe: domain.TipeBarang, Satuan: "porsi", HargaBeli: 3500, HargaJual: 8000, KelolaStok: true},
		{Kode: "CTH-004", Nama: "Air Mineral 600ml", Tipe: domain.TipeBarang, Satuan: "botol", HargaBeli: 2500, HargaJual: 4000, KelolaStok: true},
		{Kode: "CTH-005", Nama: "Gula Pasir 1kg", Tipe: domain.TipeBarang, Satuan: "kg", HargaBeli: 13000, HargaJual: 16000, KelolaStok: true},
		{Kode: "CTH-006", Nama: "Teh Botol", Tipe: domain.TipeBarang, Satuan: "botol", HargaBeli: 3000, HargaJual: 5000, HargaPromo: f(4500), KelolaStok: true},
		{Kode: "CTH-007", Nama: "Jasa Bungkus Kado", Tipe: domain.TipeJasa, Satuan: "paket", HargaJual: 5000},
	}
	kategoriProduk := map[string]string{
		"CTH-001": "Minuman", "CTH-002": "Minuman", "CTH-006": "Minuman",
		"CTH-003": "Makanan", "CTH-004": "Minuman", "CTH-005": "Sembako",
		"CTH-007": "Lainnya",
	}
	for _, p := range contoh {
		var ada domain.Produk
		err := db.Where("kode = ?", p.Kode).First(&ada).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal().Err(err).Msg("gagal membaca produk")
		}
		if nk, ok := kategoriProduk[p.Kode]; ok {
			id := kategoriID[nk]
			p.KategoriID = &id
		}
		p.Aktif = true
		if err := db.Create(&p).Error; err != nil {
			log.Fatal().Err(err).Msg("gagal membuat produk contoh")
		}
		log.Info().Str("produk", p.Nama).Msg("produk contoh dibuat")
	}

	log.Info().Msg("seed selesai — semua data ini bisa diubah/dihapus lewat admin panel")
}

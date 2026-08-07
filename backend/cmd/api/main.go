// Tuléh Server — backend mandiri Tuléh POS (strangler dari MOVERA Laravel).
//
// Entry point: memuat config, membuka koneksi (PostgreSQL, Redis), merakit
// dependensi dari dalam ke luar (repository → usecase → handler), memasang
// middleware keamanan, lalu menyalakan HTTP server dengan graceful shutdown.
//
//	@title						Tuléh Server API
//	@version					0.1.0
//	@description				Backend mandiri Tuléh POS. Amplop respons: {success, data, meta, message, errors}.
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Isi dengan: Bearer <access_token>
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/tuleh-pos/server/internal/config"
	deliveryhttp "github.com/tuleh-pos/server/internal/delivery/http"
	appmw "github.com/tuleh-pos/server/internal/delivery/http/middleware"
	"github.com/tuleh-pos/server/internal/domain"
	pgrepo "github.com/tuleh-pos/server/internal/repository/postgres"
	redisrepo "github.com/tuleh-pos/server/internal/repository/redis"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/respond"
	"github.com/tuleh-pos/server/pkg/validasi"
	// Hasil `swag init` — aktifkan baris di bawah setelah docs dihasilkan:
	_ "github.com/tuleh-pos/server/docs"
)

func main() {
	// Structured logging JSON (zerolog) — konsumsi mesin (Loki/ELK);
	// mode local memakai console writer supaya enak dibaca manusia.
	log := zerolog.New(os.Stdout).With().Timestamp().Str("app", "tuleh-server").Logger()
	if os.Getenv("APP_ENV") != "production" {
		log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cfg, err := config.Muat()
	if err != nil {
		log.Fatal().Err(err).Msg("konfigurasi tidak sah")
	}

	// ── Infrastruktur ────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("gagal terhubung ke PostgreSQL")
	}

	// Migrasi skema otomatis dari struct domain. Untuk perubahan destruktif
	// (rename/drop kolom) tetap tulis migrasi manual — AutoMigrate hanya
	// MENAMBAH, tidak pernah menghapus.
	if err := db.AutoMigrate(
		&domain.Usaha{},
		&domain.User{}, &domain.Kategori{}, &domain.Produk{},
		&domain.SesiKasir{}, &domain.Transaksi{}, &domain.TransaksiItem{},
		&domain.Pelanggan{}, &domain.Hold{}, &domain.Pengeluaran{}, &domain.StokLog{},
		&domain.Pengaturan{}, &domain.MetodePembayaran{},
	); err != nil {
		log.Fatal().Err(err).Msg("migrasi gagal")
	}

	// ── bootstrap multi-tenant ───────────────────────────────────────────
	// Unique lama per-kolom diganti unique KOMPOSIT (usaha_id, kode/nama) —
	// AutoMigrate tidak menghapus index lama, jadi dibersihkan eksplisit.
	for _, q := range []string{
		`DROP INDEX IF EXISTS idx_produks_kode`,
		`DROP INDEX IF EXISTS idx_kategoris_nama`,
	} {
		if err := db.Exec(q).Error; err != nil {
			log.Fatal().Err(err).Str("q", q).Msg("gagal membersihkan index lama")
		}
	}
	// Usaha default + backfill data pra-tenant (usaha_id 0 = orphan).
	usahaDefault := domain.Usaha{Kode: "DEMO", Nama: "Tuléh Demo", Aktif: true}
	if err := db.Where(domain.Usaha{Kode: "DEMO"}).FirstOrCreate(&usahaDefault).Error; err != nil {
		log.Fatal().Err(err).Msg("gagal menyiapkan usaha default")
	}
	for _, tabel := range []string{
		"users", "kategoris", "produks", "sesi_kasirs", "transaksis",
		"pelanggans", "holds", "pengeluarans", "stok_logs", "pengaturans",
		"metode_pembayarans",
	} {
		if err := db.Exec("UPDATE "+tabel+" SET usaha_id = ? WHERE usaha_id = 0", usahaDefault.ID).Error; err != nil {
			log.Fatal().Err(err).Str("tabel", tabel).Msg("backfill usaha gagal")
		}
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr: cfg.RedisAddr, Password: cfg.RedisPass, DB: cfg.RedisDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("gagal terhubung ke Redis")
	}

	// ── Perakitan dependensi (dari dalam ke luar) ────────────────────────
	userRepo := pgrepo.NewUserRepository(db)
	tokenRepo := redisrepo.NewTokenRepository(rdb)
	usahaRepo := pgrepo.NewUsahaRepository(db)

	produkRepo := pgrepo.NewProdukRepository(db)
	kategoriRepo := pgrepo.NewKategoriRepository(db)

	userUC := usecase.NewUserUsecase(userRepo)
	authUC := usecase.NewAuthUsecase(userRepo, tokenRepo, usahaRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTTL)
	usahaUC := usecase.NewUsahaUsecase(usahaRepo, userRepo)
	produkUC := usecase.NewProdukUsecase(produkRepo, kategoriRepo)
	kategoriUC := usecase.NewKategoriUsecase(kategoriRepo)

	sesiRepo := pgrepo.NewSesiRepository(db)
	transaksiRepo := pgrepo.NewTransaksiRepository(db)
	pelangganRepo := pgrepo.NewPelangganRepository(db)
	holdRepo := pgrepo.NewHoldRepository(db)
	pengeluaranRepo := pgrepo.NewPengeluaranRepository(db)
	laporanRepo := pgrepo.NewLaporanRepository(db)
	pengaturanRepo := pgrepo.NewPengaturanRepository(db)

	sesiUC := usecase.NewSesiUsecase(sesiRepo, transaksiRepo)
	transaksiUC := usecase.NewTransaksiUsecase(transaksiRepo, sesiRepo, produkRepo, pelangganRepo, pengaturanRepo)
	laporanUC := usecase.NewLaporanUsecase(laporanRepo, pengeluaranRepo)
	pengeluaranUC := usecase.NewPengeluaranUsecase(pengeluaranRepo)
	pelangganUC := usecase.NewPelangganUsecase(pelangganRepo)
	holdUC := usecase.NewHoldUsecase(holdRepo)
	stokRepo := pgrepo.NewStokRepository(db)
	inventoryUC := usecase.NewInventoryUsecase(produkRepo, stokRepo, sesiRepo)
	pengaturanUC := usecase.NewPengaturanUsecase(pengaturanRepo, usahaRepo)
	metodeRepo := pgrepo.NewMetodePembayaranRepository(db)
	metodeUC := usecase.NewMetodePembayaranUsecase(metodeRepo)

	semaiAdmin(db, cfg, usahaDefault.ID, log)

	// Promosi SUPERADMIN sekali jalan: platform butuh minimal satu pengelola
	// tenant. Idempoten — hanya bila belum ada SUPERADMIN sama sekali.
	var adaSuper int64
	db.Model(&domain.User{}).Where("role = ?", domain.RoleSuperadmin).Count(&adaSuper)
	if adaSuper == 0 {
		if err := db.Model(&domain.User{}).Where("email = ?", cfg.AdminEmail).
			Update("role", domain.RoleSuperadmin).Error; err != nil {
			log.Fatal().Err(err).Msg("gagal mempromosikan superadmin")
		}
		log.Info().Str("email", cfg.AdminEmail).Msg("dipromosikan menjadi SUPERADMIN")
	}

	// ── HTTP server ──────────────────────────────────────────────────────
	e := echo.New()
	e.HideBanner = true
	e.Validator = validasi.Baru()

	// SATU penerjemah error → amplop baku (bukan HTML echo): kegagalan
	// validasi jadi 422 + map field→pesan; echo.HTTPError memakai kode &
	// pesannya; sisanya 500 generik (detail hanya ke log).
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var gagal *validasi.GagalValidasi
		if errors.As(err, &gagal) {
			_ = respond.Gagal(c, http.StatusUnprocessableEntity, gagal.Error(), gagal.Fields)
			return
		}

		kode := http.StatusInternalServerError
		pesan := "Terjadi kesalahan pada server."
		var he *echo.HTTPError
		if errors.As(err, &he) {
			kode = he.Code
			switch kode {
			case http.StatusNotFound:
				pesan = "Endpoint tidak ditemukan."
			case http.StatusMethodNotAllowed:
				pesan = "Metode tidak diizinkan."
			default:
				if s, ok := he.Message.(string); ok && s != "" {
					pesan = s
				}
			}
		}
		if kode >= 500 {
			log.Error().Err(err).Str("path", c.Request().URL.Path).Int("kode", kode).Msg("http error")
		}
		_ = respond.Gagal(c, kode, pesan, nil)
	}

	// Log permintaan terstruktur + recover panic.
	e.Use(echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogURI: true, LogStatus: true, LogMethod: true, LogLatency: true, LogRemoteIP: true,
		LogValuesFunc: func(_ echo.Context, v echomw.RequestLoggerValues) error {
			log.Info().Str("metode", v.Method).Str("uri", v.URI).Int("status", v.Status).
				Dur("latensi", v.Latency).Str("ip", v.RemoteIP).Msg("request")
			return nil
		},
	}))
	e.Use(echomw.Recover())
	e.Use(appmw.Keamanan(cfg.CORSOrigins, cfg.RateRPS)...)

	// Sajikan admin panel dari binari yang sama bila ADMIN_DIST diisi —
	// SPA fallback (HTML5) utk rute React Router; /api & /swagger tetap API.
	if cfg.AdminDist != "" {
		e.Use(echomw.StaticWithConfig(echomw.StaticConfig{
			Root:  cfg.AdminDist,
			HTML5: true,
			Skipper: func(c echo.Context) bool {
				p := c.Request().URL.Path
				return strings.HasPrefix(p, "/api") || strings.HasPrefix(p, "/swagger")
			},
		}))
		log.Info().Str("dist", cfg.AdminDist).Msg("admin panel disajikan dari server ini")
	}

	deliveryhttp.DaftarkanRute(e, authUC, userUC, produkUC, kategoriUC, sesiUC, transaksiUC, laporanUC, pengeluaranUC, pelangganUC, holdUC, inventoryUC, usahaUC, pengaturanUC, metodeUC)

	// ── Jalankan + graceful shutdown ─────────────────────────────────────
	go func() {
		if err := e.Start(":" + cfg.HTTPPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server berhenti")
		}
	}()
	log.Info().Str("port", cfg.HTTPPort).Msg("Tuléh Server berjalan")

	// Tunggu sinyal berhenti, beri 10 detik untuk request yang sedang jalan.
	berhenti := make(chan os.Signal, 1)
	signal.Notify(berhenti, os.Interrupt, syscall.SIGTERM)
	<-berhenti

	ctx, batal := context.WithTimeout(context.Background(), 10*time.Second)
	defer batal()
	if err := e.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown tidak mulus")
	}
	log.Info().Msg("server berhenti dengan rapi")
}

// semaiAdmin membuat pengguna pertama (OWNER) bila tabel masih kosong —
// supaya instalasi baru langsung bisa dipakai tanpa menyentuh database.
func semaiAdmin(db *gorm.DB, cfg *config.Config, usahaID uint, log zerolog.Logger) {
	var jumlah int64
	db.Model(&domain.User{}).Count(&jumlah)
	if jumlah > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("gagal menyemai admin")
	}
	admin := &domain.User{
		UsahaID: usahaID,
		Nama:    cfg.AdminNama, Email: cfg.AdminEmail,
		PasswordHash: string(hash), Role: domain.RoleOwner, Aktif: true,
	}
	if err := db.Create(admin).Error; err != nil {
		log.Fatal().Err(err).Msg("gagal menyemai admin")
	}
	log.Info().Str("email", cfg.AdminEmail).Msg("admin pertama dibuat (ganti sandinya!)")
}

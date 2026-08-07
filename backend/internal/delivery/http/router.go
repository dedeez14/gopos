package http

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	appmw "github.com/tuleh-pos/server/internal/delivery/http/middleware"
	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
)

// DaftarkanRute memasang seluruh endpoint. SATU file peta rute — programmer
// baru cukup membuka file ini untuk melihat permukaan API + gerbangnya.
func DaftarkanRute(
	e *echo.Echo,
	authUC *usecase.AuthUsecase,
	userUC *usecase.UserUsecase,
	produkUC *usecase.ProdukUsecase,
	kategoriUC *usecase.KategoriUsecase,
	sesiUC *usecase.SesiUsecase,
	transaksiUC *usecase.TransaksiUsecase,
	laporanUC *usecase.LaporanUsecase,
	pengeluaranUC *usecase.PengeluaranUsecase,
	pelangganUC *usecase.PelangganUsecase,
	holdUC *usecase.HoldUsecase,
	inventoryUC *usecase.InventoryUsecase,
) {
	authH := NewAuthHandler(authUC)
	userH := NewUserHandler(userUC)
	produkH := NewProdukHandler(produkUC, kategoriUC)
	kasirH := NewKasirHandler(sesiUC, transaksiUC)
	laporanH := NewLaporanHandler(laporanUC, pengeluaranUC, pelangganUC, holdUC)
	inventoryH := NewInventoryHandler(inventoryUC)

	// Dokumentasi Swagger (hasil `swag init`): /swagger/index.html
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := e.Group("/api/v1")

	// ── Publik ───────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	auth.POST("/login", authH.Login, appmw.RateKetat())
	auth.POST("/refresh", authH.Refresh, appmw.RateKetat())

	// ── Ber-auth (JWT) ───────────────────────────────────────────────────
	privat := v1.Group("", appmw.JWT(authUC))
	privat.POST("/auth/logout", authH.Logout)

	// RBAC per-endpoint: baca cukup users.lihat, tulis butuh users.kelola.
	privat.GET("/users", userH.Daftar, appmw.ButuhIzin(domain.PermUserLihat))
	privat.GET("/users/:id", userH.Ambil, appmw.ButuhIzin(domain.PermUserLihat))
	privat.POST("/users", userH.Buat, appmw.ButuhIzin(domain.PermUserKelola))
	privat.PUT("/users/:id", userH.Perbarui, appmw.ButuhIzin(domain.PermUserKelola))
	privat.DELETE("/users/:id", userH.Hapus, appmw.ButuhIzin(domain.PermUserKelola))

	// Katalog: BACA terbuka semua peran (kasir memakainya di layar jual —
	// harga_beli otomatis disembunyikan untuk kasir); TULIS khusus O/M.
	privat.GET("/produk", produkH.Daftar, appmw.ButuhIzin(domain.PermKasir))
	privat.GET("/produk/:id", produkH.Ambil, appmw.ButuhIzin(domain.PermKasir))
	privat.POST("/produk", produkH.Buat, appmw.ButuhIzin(domain.PermProdukKelola))
	privat.PUT("/produk/:id", produkH.Perbarui, appmw.ButuhIzin(domain.PermProdukKelola))
	privat.DELETE("/produk/:id", produkH.Hapus, appmw.ButuhIzin(domain.PermProdukKelola))

	privat.GET("/kategori", produkH.DaftarKategori, appmw.ButuhIzin(domain.PermKasir))
	privat.POST("/kategori", produkH.BuatKategori, appmw.ButuhIzin(domain.PermProdukKelola))
	privat.DELETE("/kategori/:id", produkH.HapusKategori, appmw.ButuhIzin(domain.PermProdukKelola))

	// Kasir: sesi + checkout + riwayat (semua peran); daftar SEMUA sesi
	// khusus pemegang izin laporan. Literal (/aktif) sebelum /:id.
	privat.POST("/sesi/buka", kasirH.BukaSesi, appmw.ButuhIzin(domain.PermKasir))
	privat.POST("/sesi/tutup", kasirH.TutupSesi, appmw.ButuhIzin(domain.PermKasir))
	privat.GET("/sesi/aktif", kasirH.SesiAktif, appmw.ButuhIzin(domain.PermKasir))
	privat.GET("/sesi", kasirH.DaftarSesi, appmw.ButuhIzin(domain.PermLaporan))
	privat.GET("/sesi/:id/rekap", kasirH.RekapSesi, appmw.ButuhIzin(domain.PermKasir))

	privat.POST("/transaksi/checkout", kasirH.Checkout, appmw.ButuhIzin(domain.PermKasir))
	privat.POST("/transaksi/:id/batal", kasirH.BatalTransaksi, appmw.ButuhIzin(domain.PermKasir))
	privat.GET("/transaksi", kasirH.DaftarTransaksi, appmw.ButuhIzin(domain.PermKasir))
	privat.GET("/transaksi/:id", kasirH.AmbilTransaksi, appmw.ButuhIzin(domain.PermKasir))

	// Laporan & pengeluaran — khusus pemegang izin laporan (O/M).
	privat.GET("/laporan/keuangan", laporanH.Keuangan, appmw.ButuhIzin(domain.PermLaporan))
	privat.GET("/laporan/penjualan-harian", laporanH.PenjualanHarian, appmw.ButuhIzin(domain.PermLaporan))
	privat.GET("/laporan/produk-terlaris", laporanH.ProdukTerlaris, appmw.ButuhIzin(domain.PermLaporan))
	privat.GET("/pengeluaran", laporanH.DaftarPengeluaran, appmw.ButuhIzin(domain.PermLaporan))
	privat.POST("/pengeluaran", laporanH.CatatPengeluaran, appmw.ButuhIzin(domain.PermLaporan))
	privat.DELETE("/pengeluaran/:id", laporanH.HapusPengeluaran, appmw.ButuhIzin(domain.PermLaporan))

	// Pelanggan: quick-add & baca = alat kasir; ubah/nonaktif = manajemen.
	privat.GET("/pelanggan", laporanH.DaftarPelanggan, appmw.ButuhIzin(domain.PermKasir))
	privat.POST("/pelanggan/quick", laporanH.QuickPelanggan, appmw.ButuhIzin(domain.PermKasir))
	privat.PUT("/pelanggan/:id", laporanH.PerbaruiPelanggan, appmw.ButuhIzin(domain.PermPelangganKelola))
	privat.DELETE("/pelanggan/:id", laporanH.HapusPelanggan, appmw.ButuhIzin(domain.PermPelangganKelola))

	// Inventory — restock/opname/riwayat khusus manajemen katalog (O/M).
	privat.POST("/inventory/stok-masuk", inventoryH.StokMasuk, appmw.ButuhIzin(domain.PermProdukKelola))
	privat.POST("/inventory/opname", inventoryH.Opname, appmw.ButuhIzin(domain.PermProdukKelola))
	privat.GET("/inventory/riwayat", inventoryH.Riwayat, appmw.ButuhIzin(domain.PermProdukKelola))

	// Hold — parkir keranjang, alat harian kasir.
	privat.GET("/hold", laporanH.DaftarHold, appmw.ButuhIzin(domain.PermKasir))
	privat.POST("/hold", laporanH.SimpanHold, appmw.ButuhIzin(domain.PermKasir))
	privat.DELETE("/hold/:id", laporanH.HapusHold, appmw.ButuhIzin(domain.PermKasir))
}

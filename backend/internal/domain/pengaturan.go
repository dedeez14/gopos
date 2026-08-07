package domain

import (
	"context"
	"time"
)

// Pengaturan — konfigurasi usaha yang DINAMIS: profil toko, struk, dan
// perilaku kasir. SATU baris per usaha (singleton, unik pada usaha_id).
//
// Dibaca aplikasi kasir (header/footer struk, pajak default, pembulatan) DAN
// panel admin — semuanya bisa diubah pemilik tanpa menyentuh kode. Inilah
// "semua data dinamis": tak ada nama toko/pajak yang di-hardcode.
type Pengaturan struct {
	ID      uint `gorm:"primaryKey"`
	UsahaID uint `gorm:"uniqueIndex;not null;default:0"` // satu baris per usaha

	// ── Profil toko (tampil di struk & aplikasi) ──
	NamaToko string `gorm:"size:150;not null"`
	Alamat   string `gorm:"size:255;not null"`
	Telepon  string `gorm:"size:30;not null"`
	Email    string `gorm:"size:150;not null"`
	LogoURL  string `gorm:"size:255;not null"`
	MataUang string `gorm:"size:8;not null"` // simbol, mis. "Rp"

	// ── Struk (receipt) ──
	StrukHeader   string `gorm:"size:255;not null"` // baris tambahan di atas
	StrukFooter   string `gorm:"size:255;not null"` // ucapan terima kasih, dll
	UkuranKertas  string `gorm:"size:8;not null"`   // "58mm" | "80mm"
	TampilkanLogo bool   `gorm:"not null"`

	// ── Pajak & perilaku kasir ──
	PajakPersen float64 `gorm:"not null"` // 0–100, dipakai kasir sbg default
	PajakAktif  bool    `gorm:"not null"`
	Pembulatan  int     `gorm:"not null"` // 0 = tanpa; mis. 100 = bulatkan ke 100 terdekat

	CreatedAt time.Time
	UpdatedAt time.Time
}

// PengaturanDefault menghasilkan konfigurasi awal yang MASUK AKAL untuk usaha
// baru — semua nilai bisa diubah lewat panel. Nama toko diisi dari nama usaha
// supaya struk langsung bermakna sejak hari pertama.
//
// Catatan GORM: bool di sini SENGAJA tanpa tag `default:` — tag itu membuat
// nilai false hilang saat INSERT (default DB menang). Default diatur di sini
// dan di lapisan yang menyimpan, bukan di skema.
func PengaturanDefault(usahaID uint, namaUsaha string) Pengaturan {
	return Pengaturan{
		UsahaID:       usahaID,
		NamaToko:      namaUsaha,
		MataUang:      "Rp",
		StrukFooter:   "Terima kasih telah berbelanja 🙏",
		UkuranKertas:  "80mm",
		TampilkanLogo: false,
		PajakPersen:   0,
		PajakAktif:    false,
		Pembulatan:    0,
	}
}

// PengaturanRepository — kontrak singleton per usaha (di-scope usaha_id).
// Baris default dibuat lewat Simpan saat pertama diakses (lihat usecase).
type PengaturanRepository interface {
	// CariByUsaha mengembalikan pengaturan usaha aktif; ErrTidakDitemukan
	// bila belum ada (usecase yang memutuskan membuat default).
	CariByUsaha(ctx context.Context) (*Pengaturan, error)
	Simpan(ctx context.Context, p *Pengaturan) error
	Perbarui(ctx context.Context, p *Pengaturan) error
}

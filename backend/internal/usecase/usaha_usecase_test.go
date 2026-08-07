package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/tuleh-pos/server/internal/domain"
)

type usahaRepoPalsu struct {
	data   map[uint]*domain.Usaha
	nextID uint
}

func newUsahaRepoPalsu() *usahaRepoPalsu {
	return &usahaRepoPalsu{data: map[uint]*domain.Usaha{}, nextID: 1}
}

func (r *usahaRepoPalsu) Simpan(_ context.Context, u *domain.Usaha) error {
	u.ID = r.nextID
	r.nextID++
	r.data[u.ID] = u
	return nil
}
func (r *usahaRepoPalsu) Perbarui(_ context.Context, u *domain.Usaha) error {
	r.data[u.ID] = u
	return nil
}
func (r *usahaRepoPalsu) CariByID(_ context.Context, id uint) (*domain.Usaha, error) {
	if u, ok := r.data[id]; ok {
		return u, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *usahaRepoPalsu) CariByKode(_ context.Context, kode string) (*domain.Usaha, error) {
	for _, u := range r.data {
		if u.Kode == kode {
			return u, nil
		}
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *usahaRepoPalsu) Daftar(_ context.Context, _ domain.FilterUsaha) ([]domain.Usaha, int64, error) {
	return nil, 0, nil
}

// userRepoPalsu minimal untuk usecase usaha.
type userRepoPalsu struct {
	data   map[uint]*domain.User
	nextID uint
}

func newUserRepoPalsu() *userRepoPalsu {
	return &userRepoPalsu{data: map[uint]*domain.User{}, nextID: 1}
}
func (r *userRepoPalsu) Simpan(_ context.Context, u *domain.User) error {
	u.ID = r.nextID
	r.nextID++
	r.data[u.ID] = u
	return nil
}
func (r *userRepoPalsu) Perbarui(_ context.Context, u *domain.User) error {
	r.data[u.ID] = u
	return nil
}
func (r *userRepoPalsu) Hapus(_ context.Context, id uint) error { delete(r.data, id); return nil }
func (r *userRepoPalsu) CariByID(_ context.Context, id uint) (*domain.User, error) {
	if u, ok := r.data[id]; ok {
		return u, nil
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *userRepoPalsu) CariByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range r.data {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrTidakDitemukan
}
func (r *userRepoPalsu) Daftar(_ context.Context, _ domain.FilterUser) ([]domain.User, int64, error) {
	return nil, 0, nil
}

func TestBuatUsahaSekaligusOwner(t *testing.T) {
	usahaRepo := newUsahaRepoPalsu()
	userRepo := newUserRepoPalsu()
	uc := NewUsahaUsecase(usahaRepo, userRepo)
	ctx := context.Background()

	usaha, owner, err := uc.Buat(ctx, InputUsahaBaru{
		Nama: "Warung Baru", Kode: "WB",
		OwnerNama: "Bu Owner", OwnerEmail: "Owner@WB.id", OwnerPassword: "rahasia123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if usaha.ID == 0 || owner.UsahaID != usaha.ID {
		t.Fatal("owner harus tertaut ke usaha baru")
	}
	if owner.Role != domain.RoleOwner || owner.Email != "owner@wb.id" {
		t.Fatalf("owner salah: role=%s email=%s", owner.Role, owner.Email)
	}

	// Email owner ganda → ditolak, TANPA membuat usaha yatim.
	sblm := usahaRepo.nextID
	if _, _, err := uc.Buat(ctx, InputUsahaBaru{
		Nama: "Lain", OwnerNama: "X", OwnerEmail: "owner@wb.id", OwnerPassword: "rahasia123",
	}); err != domain.ErrEmailTerpakai {
		t.Fatalf("harap ErrEmailTerpakai, dapat %v", err)
	}
	if usahaRepo.nextID != sblm {
		t.Fatal("usaha yatim terbuat padahal email owner bentrok")
	}

	// Kode usaha ganda → ditolak.
	if _, _, err := uc.Buat(ctx, InputUsahaBaru{
		Nama: "Lain2", Kode: "WB", OwnerNama: "Y", OwnerEmail: "y@lain.id", OwnerPassword: "rahasia123",
	}); err != domain.ErrKodeUsahaTerpakai {
		t.Fatalf("harap ErrKodeUsahaTerpakai, dapat %v", err)
	}
}

func TestSuspendUsahaTolakLogin(t *testing.T) {
	usahaRepo := newUsahaRepoPalsu()
	userRepo := newUserRepoPalsu()
	tokenRepo := &tokenRepoPalsu{data: map[string]uint{}}
	uUC := NewUsahaUsecase(usahaRepo, userRepo)
	authUC := NewAuthUsecase(userRepo, tokenRepo, usahaRepo, "rahasia-uji", time.Minute, time.Hour)
	ctx := context.Background()

	_, owner, err := uUC.Buat(ctx, InputUsahaBaru{
		Nama: "Toko S", Kode: "TS", OwnerNama: "O", OwnerEmail: "o@ts.id", OwnerPassword: "sandi12345",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Sebelum suspend: login OK.
	if _, _, err := authUC.Login(ctx, "o@ts.id", "sandi12345"); err != nil {
		t.Fatalf("login sebelum suspend harus OK: %v", err)
	}

	// Suspend usaha → login ditolak ErrUsahaNonaktif.
	nonaktif := false
	if _, err := uUC.Perbarui(ctx, owner.UsahaID, InputUbahUsaha{Aktif: &nonaktif}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := authUC.Login(ctx, "o@ts.id", "sandi12345"); err != domain.ErrUsahaNonaktif {
		t.Fatalf("login setelah suspend harus ditolak ErrUsahaNonaktif, dapat %v", err)
	}
}

// tokenRepoPalsu minimal (Redis).
type tokenRepoPalsu struct{ data map[string]uint }

func (r *tokenRepoPalsu) SimpanRefresh(_ context.Context, token string, userID uint, _ time.Duration) error {
	r.data[token] = userID
	return nil
}
func (r *tokenRepoPalsu) AmbilRefresh(_ context.Context, token string) (uint, error) {
	if id, ok := r.data[token]; ok {
		return id, nil
	}
	return 0, domain.ErrTokenTidakSah
}
func (r *tokenRepoPalsu) HapusRefresh(_ context.Context, token string) error {
	delete(r.data, token)
	return nil
}

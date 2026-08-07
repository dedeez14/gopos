package http

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/tuleh-pos/server/internal/domain"
	"github.com/tuleh-pos/server/internal/usecase"
	"github.com/tuleh-pos/server/pkg/apperror"
	"github.com/tuleh-pos/server/pkg/respond"
)

type UserHandler struct {
	users *usecase.UserUsecase
}

func NewUserHandler(users *usecase.UserUsecase) *UserHandler {
	return &UserHandler{users: users}
}

// UserResponse adalah DTO keluar — entitas domain TIDAK diserialisasi
// langsung supaya PasswordHash mustahil bocor dan bentuk API stabil walau
// entitas berubah.
type UserResponse struct {
	ID       uint   `json:"id"`
	Nama     string `json:"nama"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Aktif    bool   `json:"aktif"`
	DibuatKe string `json:"dibuat_pada"`
	DiubahKe string `json:"diubah_pada"`
}

func keUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID: u.ID, Nama: u.Nama, Email: u.Email,
		Role: string(u.Role), Aktif: u.Aktif,
		DibuatKe: u.CreatedAt.Format(time.RFC3339),
		DiubahKe: u.UpdatedAt.Format(time.RFC3339),
	}
}

// SimpanUserRequest dipakai create & update (password opsional saat update).
type SimpanUserRequest struct {
	Nama     string `json:"nama" validate:"required,min=2,max=150"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Password string `json:"password" validate:"omitempty,min=8,max=72"`
	Role     string `json:"role" validate:"required,oneof=OWNER MANAGER KASIR"`
	Aktif    *bool  `json:"aktif" validate:"required"`
}

// Daftar godoc
//
//	@Summary		Daftar pengguna (paginasi + pencarian)
//	@Tags			users
//	@Produce		json
//	@Param			q			query		string	false	"Cari nama/email"
//	@Param			page		query		int		false	"Halaman (mulai 1)"
//	@Param			per_page	query		int		false	"Baris per halaman (maks 100)"
//	@Success		200			{object}	respond.Amplop{data=[]UserResponse,meta=respond.Meta}
//	@Failure		403			{object}	respond.Amplop
//	@Security		BearerAuth
//	@Router			/users [get]
func (h *UserHandler) Daftar(c echo.Context) error {
	halaman, _ := strconv.Atoi(c.QueryParam("page"))
	perHal, _ := strconv.Atoi(c.QueryParam("per_page"))

	f := domain.FilterUser{Cari: c.QueryParam("q"), Halaman: halaman, PerHal: perHal}
	users, total, err := h.users.Daftar(c.Request().Context(), f)
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}

	// Normalisasi ulang nilai paginasi yang dipakai usecase (dijaga di sana).
	if f.Halaman < 1 {
		f.Halaman = 1
	}
	if f.PerHal < 1 || f.PerHal > 100 {
		f.PerHal = 20
	}

	hasil := make([]UserResponse, 0, len(users))
	for i := range users {
		hasil = append(hasil, keUserResponse(&users[i]))
	}
	return respond.Sukses(c, hasil, respond.BuatMeta(f.Halaman, f.PerHal, total), "")
}

// Ambil godoc
//
//	@Summary	Detail satu pengguna
//	@Tags		users
//	@Produce	json
//	@Param		id	path		int	true	"ID pengguna"
//	@Success	200	{object}	respond.Amplop{data=UserResponse}
//	@Failure	404	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/users/{id} [get]
func (h *UserHandler) Ambil(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	u, err := h.users.Ambil(c.Request().Context(), uint(id))
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keUserResponse(u), nil, "")
}

// Buat godoc
//
//	@Summary	Tambah pengguna
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		SimpanUserRequest	true	"Data pengguna (password wajib saat membuat)"
//	@Success	201		{object}	respond.Amplop{data=UserResponse}
//	@Failure	409		{object}	respond.Amplop	"Email sudah terdaftar"
//	@Failure	422		{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/users [post]
func (h *UserHandler) Buat(c echo.Context) error {
	var req SimpanUserRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}
	if req.Password == "" {
		return respond.Gagal(c, 422, "Data tidak valid.", map[string]string{"password": "password wajib diisi saat membuat pengguna."})
	}

	u, err := h.users.Buat(c.Request().Context(), usecase.InputUser{
		Nama: req.Nama, Email: req.Email, Password: req.Password,
		Role: domain.Role(req.Role), Aktif: *req.Aktif,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Dibuat(c, keUserResponse(u), "Pengguna tersimpan.")
}

// Perbarui godoc
//
//	@Summary	Ubah pengguna
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"ID pengguna"
//	@Param		payload	body		SimpanUserRequest	true	"Data pengguna (password kosong = tidak diganti)"
//	@Success	200		{object}	respond.Amplop{data=UserResponse}
//	@Failure	404		{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/users/{id} [put]
func (h *UserHandler) Perbarui(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	var req SimpanUserRequest
	if err := bindDanValidasi(c, &req); err != nil {
		return err
	}

	u, err := h.users.Perbarui(c.Request().Context(), uint(id), usecase.InputUser{
		Nama: req.Nama, Email: req.Email, Password: req.Password,
		Role: domain.Role(req.Role), Aktif: *req.Aktif,
	})
	if err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, keUserResponse(u), nil, "Pengguna diperbarui.")
}

// Hapus godoc
//
//	@Summary	Hapus pengguna
//	@Tags		users
//	@Produce	json
//	@Param		id	path		int	true	"ID pengguna"
//	@Success	200	{object}	respond.Amplop
//	@Failure	404	{object}	respond.Amplop
//	@Security	BearerAuth
//	@Router		/users/{id} [delete]
func (h *UserHandler) Hapus(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return respond.Gagal(c, 400, "ID tidak valid.", nil)
	}
	if err := h.users.Hapus(c.Request().Context(), uint(id)); err != nil {
		return respond.Gagal(c, apperror.Status(err), apperror.Pesan(err), nil)
	}
	return respond.Sukses(c, nil, nil, "Pengguna dihapus.")
}

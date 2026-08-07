// Package apperror menerjemahkan error domain → kode HTTP di SATU tempat.
//
// Usecase cukup mengembalikan error domain (domain.ErrTidakDitemukan, dst);
// handler tinggal memanggil apperror.Status(err) — tidak ada switch kode
// status yang diduplikasi di tiap handler (DRY).
package apperror

import (
	"errors"
	"net/http"

	"github.com/tuleh-pos/server/internal/domain"
)

// Status memetakan error ke kode HTTP. Error yang tidak dikenal = 500 —
// fail-closed, jangan pernah membocorkan detail internal ke klien.
func Status(err error) int {
	switch {
	case errors.Is(err, domain.ErrTidakDitemukan):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrEmailTerpakai),
		errors.Is(err, domain.ErrKodeTerpakai):
		return http.StatusConflict
	case errors.Is(err, domain.ErrKredensialSalah),
		errors.Is(err, domain.ErrTokenTidakSah):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrUserNonaktif),
		errors.Is(err, domain.ErrSesiBukanMilik):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrSesiSudahBuka),
		errors.Is(err, domain.ErrSesiBelumBuka),
		errors.Is(err, domain.ErrSesiSudahTutup),
		errors.Is(err, domain.ErrProdukNonaktif):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Pesan mengembalikan pesan aman untuk klien (Bahasa Indonesia).
// Error tak dikenal disembunyikan di balik pesan generik; detail aslinya
// cukup untuk log server.
func Pesan(err error) string {
	if Status(err) == http.StatusInternalServerError {
		return "Terjadi kesalahan pada server. Silakan coba lagi."
	}
	return err.Error()
}

// Package respond membakukan amplop respons JSON — SATU bentuk untuk seluruh
// API, identik dengan kontrak Tuléh yang sudah dipakai klien rilisan:
//
//	{ "success": bool, "data": ..., "meta": ..., "message": "...", "errors": ... }
//
// Kompatibilitas ini disengaja: server Go ini adalah strangler service yang
// mengambil alih endpoint dari MOVERA Laravel secara bertahap — klien Flutter
// tidak boleh bisa membedakan siapa yang menjawab.
package respond

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Amplop adalah bentuk baku seluruh respons JSON.
type Amplop struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Meta    *Meta  `json:"meta"`
	Message string `json:"message"`
	Errors  any    `json:"errors"`
}

// Meta membawa informasi paginasi. Pointer supaya respons non-list
// mengirim "meta": null persis seperti kontrak.
type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}

// BuatMeta menghitung meta paginasi dari total baris.
func BuatMeta(halaman, perHal int, total int64) *Meta {
	last := int((total + int64(perHal) - 1) / int64(perHal))
	if last < 1 {
		last = 1
	}
	return &Meta{CurrentPage: halaman, PerPage: perHal, Total: total, LastPage: last}
}

// Sukses mengirim 200 dengan data (meta boleh nil).
func Sukses(c echo.Context, data any, meta *Meta, pesan string) error {
	return c.JSON(http.StatusOK, Amplop{Success: true, Data: data, Meta: meta, Message: pesan, Errors: nil})
}

// Dibuat mengirim 201 — dipakai endpoint pembuatan sumber daya.
func Dibuat(c echo.Context, data any, pesan string) error {
	return c.JSON(http.StatusCreated, Amplop{Success: true, Data: data, Meta: nil, Message: pesan, Errors: nil})
}

// Gagal mengirim error dengan amplop baku. errors boleh nil, atau map
// field→pesan untuk kegagalan validasi.
func Gagal(c echo.Context, kode int, pesan string, errors any) error {
	return c.JSON(kode, Amplop{Success: false, Data: nil, Meta: nil, Message: pesan, Errors: errors})
}

// Package validasi membungkus go-playground/validator agar:
//  1. terpasang sebagai echo.Validator (c.Validate(&req) di handler), dan
//  2. pesan error keluar dalam Bahasa Indonesia per field — siap ditampilkan
//     langsung oleh klien tanpa penerjemahan.
package validasi

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// EchoValidator memenuhi interface echo.Validator.
type EchoValidator struct {
	v *validator.Validate
}

func Baru() *EchoValidator {
	v := validator.New(validator.WithRequiredStructEnabled())
	return &EchoValidator{v: v}
}

// Validate dipanggil echo lewat c.Validate(). Mengembalikan *GagalValidasi
// (bukan error mentah) supaya handler bisa mengirim map field→pesan.
func (ev *EchoValidator) Validate(i any) error {
	if err := ev.v.Struct(i); err != nil {
		invalid, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		gagal := &GagalValidasi{Fields: map[string]string{}}
		for _, fe := range invalid {
			gagal.Fields[strings.ToLower(fe.Field())] = pesanUntuk(fe)
		}
		return gagal
	}
	return nil
}

// GagalValidasi membawa pesan per field; memenuhi interface error.
type GagalValidasi struct {
	Fields map[string]string
}

func (g *GagalValidasi) Error() string { return "Data tidak valid." }

// pesanUntuk menerjemahkan tag validator → kalimat Indonesia. Tambahkan case
// baru di sini saat memakai tag lain — satu tempat, bukan di tiap handler.
func pesanUntuk(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi.", field)
	case "email":
		return fmt.Sprintf("%s harus berupa alamat email yang sah.", field)
	case "min":
		return fmt.Sprintf("%s minimal %s karakter.", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter.", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s harus salah satu dari: %s.", field, fe.Param())
	default:
		return fmt.Sprintf("%s tidak valid.", field)
	}
}

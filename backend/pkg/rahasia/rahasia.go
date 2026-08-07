// Package rahasia mengenkripsi rahasia at-rest (mis. server key Midtrans
// milik merchant) dengan AES-256-GCM. Nonce acak per-pesan diprepend ke
// ciphertext, seluruhnya di-base64.
//
// Kunci diturunkan dari passphrase env lewat SHA-256 → 32 byte, jadi panjang
// ENC_KEY bebas. GCM memberi kerahasiaan + integritas (auth tag) sekaligus.
package rahasia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

type Kotak struct {
	gcm cipher.AEAD
}

// Baru membuat kotak enkripsi dari passphrase. Passphrase kosong ditolak —
// enkripsi tanpa kunci nyata adalah keamanan palsu.
func Baru(passphrase string) (*Kotak, error) {
	if passphrase == "" {
		return nil, errors.New("passphrase enkripsi kosong")
	}
	sum := sha256.Sum256([]byte(passphrase)) // 32 byte → AES-256
	blok, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blok)
	if err != nil {
		return nil, err
	}
	return &Kotak{gcm: gcm}, nil
}

// Enkripsi mengembalikan base64(nonce || ciphertext).
func (k *Kotak) Enkripsi(teks string) (string, error) {
	nonce := make([]byte, k.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sandi := k.gcm.Seal(nonce, nonce, []byte(teks), nil)
	return base64.StdEncoding.EncodeToString(sandi), nil
}

// Dekripsi membalik Enkripsi; gagal bila data rusak/kunci salah (auth tag).
func (k *Kotak) Dekripsi(sandi string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(sandi)
	if err != nil {
		return "", err
	}
	ns := k.gcm.NonceSize()
	if len(raw) < ns {
		return "", errors.New("ciphertext terlalu pendek")
	}
	nonce, data := raw[:ns], raw[ns:]
	teks, err := k.gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(teks), nil
}

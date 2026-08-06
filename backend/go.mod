module github.com/tuleh-pos/server

go 1.23

// Jalankan `go mod tidy` setelah clone — versi di bawah adalah pin awal yang
// dikenal baik; tidy akan melengkapi dependensi transitif.
require (
	github.com/go-playground/validator/v10 v10.22.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/labstack/echo/v4 v4.12.0
	github.com/redis/go-redis/v9 v9.7.0
	github.com/rs/zerolog v1.33.0
	github.com/swaggo/echo-swagger v1.4.1
	github.com/swaggo/swag v1.16.4
	golang.org/x/crypto v0.31.0
	golang.org/x/time v0.8.0
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.12
)

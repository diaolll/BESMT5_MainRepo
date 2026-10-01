package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"siakaduts/utils"
)

// LoginFailLimiter memenuhi spesifikasi endpoint 1:
// "429 jika gagal login lebih dari 5 kali per menit".
//
// Hanya KEGAGALAN login yang dihitung (bukan semua request),
// dikelompokkan per IP, jendela geser 60 detik.
// Setelah 5 kegagalan dalam jendela, request berikutnya ditolak 429
// sampai kegagalan tertua kedaluwarsa.
type LoginFailLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
}

func NewLoginFailLimiter() *LoginFailLimiter {
	return &LoginFailLimiter{failures: map[string][]time.Time{}}
}

// Check dipanggil di AWAL handler login: tolak bila sudah ≥5 gagal/menit.
func (l *LoginFailLimiter) Check(c *fiber.Ctx) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	key := c.IP()
	kept := l.failures[key][:0]
	for _, t := range l.failures[key] {
		if now.Sub(t) < time.Minute {
			kept = append(kept, t)
		}
	}
	l.failures[key] = kept
	if len(kept) >= 5 {
		c.Set("Retry-After", "60")
		_ = utils.Err(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan login yang gagal, coba lagi dalam satu menit")
		return false
	}
	return true
}

// RecordFailed dipanggil setiap kredensial salah / validasi login gagal.
func (l *LoginFailLimiter) RecordFailed(c *fiber.Ctx) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures[c.IP()] = append(l.failures[c.IP()], time.Now())
}

// Reset dipanggil saat login BERHASIL agar user yang akhirnya ingat
// password tidak terus dihukum oleh kegagalan sebelumnya.
func (l *LoginFailLimiter) Reset(c *fiber.Ctx) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, c.IP())
}

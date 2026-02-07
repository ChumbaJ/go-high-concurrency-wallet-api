package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type Limiter struct {
	limit       int
	period      time.Duration
	connections map[string]*Connection
	mu          sync.RWMutex
}

type Connection struct {
	LastUpdateTime time.Time
	Tries          int
}

func NewLimiter(limit int, period time.Duration) *Limiter {
	return &Limiter{
		limit:       limit,
		period:      period,
		connections: make(map[string]*Connection),
	}
}

func (l *Limiter) CheckConn(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	conn, exists := l.connections[ip]
	if !exists {
		conn = &Connection{LastUpdateTime: now, Tries: 0}
		l.connections[ip] = conn
	}

	if time.Since(conn.LastUpdateTime) > l.period {
		conn.LastUpdateTime = now
		conn.Tries = 0
	}

	conn.Tries++

	return conn.Tries > l.limit
}

func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if l.CheckConn(ip) {
			http.Error(w, "too many requests, try again later", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		return realIP
	}
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

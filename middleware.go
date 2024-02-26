package echos

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// CustomResponseWriter custom ResponseWriter
type CustomResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (s *CustomResponseWriter) Write(b []byte) (int, error) {
	return s.Writer.Write(b)
}

type Middleware struct {
	log       *slog.Logger
	ipLimiter *IpRateLimiter
}

func NewMiddleware(
	log *slog.Logger,
) *Middleware {
	tmp := &Middleware{
		log: log,
	}
	// ip限流器 每秒钟产生10个Token, Token桶容量为:20
	tmp.ipLimiter = NewIpRateLimiter(10, 20)
	return tmp
}

// Logger for logger http log
func (s *Middleware) Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// if false {
		// 	return next(c)
		// }
		reqBodyBuffer := &bytes.Buffer{}
		if _, err := io.Copy(reqBodyBuffer, c.Request().Body); err != nil {
			return err
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBodyBuffer.Bytes()))
		resBodyBuffer := bytes.NewBuffer(nil)
		multiWriter := io.MultiWriter(c.Response().Writer, resBodyBuffer)
		writer := &CustomResponseWriter{
			Writer:         multiWriter,
			ResponseWriter: c.Response().Writer,
		}
		c.Response().Writer = writer
		uri := c.Request().RequestURI
		method := c.Request().Method
		status := c.Response().Status
		ip := c.RealIP()
		defer s.log.Debug(uri, "request_method", method, "request_ip", ip, "response_status", fmt.Sprintf("%d", status), "request_body", reqBodyBuffer.String(), "response_body", resBodyBuffer.String())
		return next(c)
	}
}

// IpLimiter ip限流中间件
func (s *Middleware) IpLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		limiter := s.ipLimiter.GetLimiter(c.RealIP())
		if !limiter.Allow() {
			return c.NoContent(http.StatusTooManyRequests)
		}
		return next(c)
	}
}

// IpRateLimiter .
type IpRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIpRateLimiter .
func NewIpRateLimiter(r rate.Limit, b int) *IpRateLimiter {
	s := &IpRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
	return s
}

// AddIp creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (s *IpRateLimiter) AddIp(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 限流器对象; r: 每秒钟产生多少Token, b: Token桶的容量
	limiter := rate.NewLimiter(s.r, s.b)
	s.ips[ip] = limiter
	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists,
// Otherwise calls AddIP to add IP address to the map
func (s *IpRateLimiter) GetLimiter(ip string) *rate.Limiter {
	s.mu.RLock()
	limiter, exists := s.ips[ip]
	if !exists {
		s.mu.RUnlock()
		return s.AddIp(ip)
	}
	s.mu.RUnlock()
	return limiter
}

package rlproxy

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiter manages rate limiting for a single domain.
type limiter struct {
	mu    sync.Mutex
	rl    *rate.Limiter
	reset time.Time
}

// Proxy is a concurrency-safe, per-domain, rate-limited HTTP client.
type Proxy struct {
	client   *http.Client
	limiters sync.Map // map[string]*limiter
}

// New creates a new Proxy.
// It can be customized with a different http.Client.
func New(client *http.Client) *Proxy {
	if client == nil {
		client = &http.Client{}
	}
	return &Proxy{
		client: client,
	}
}

// Do sends an HTTP request, waiting for the rate limiter for the request's host.
func (p *Proxy) Do(req *http.Request) (*http.Response, error) {
	host := req.URL.Hostname()
	l := p.getLimiter(host)

	err := l.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	p.updateLimiterFromHeaders(l, resp.Header)

	return resp, nil
}

func (p *Proxy) getLimiter(host string) *limiter {
	// Default limiter: 10 req/s, burst of 30.
	// This is a reasonable default based on bsky's PDS rate limits (3000/300s).
	newLimiter := &limiter{
		rl: rate.NewLimiter(rate.Limit(10), 30),
	}

	val, ok := p.limiters.Load(host)
	if ok {
		return val.(*limiter)
	}

	val, _ = p.limiters.LoadOrStore(host, newLimiter)
	return val.(*limiter)
}

func (l *limiter) Wait(ctx context.Context) error {
	l.mu.Lock()
	resetTime := l.reset
	l.mu.Unlock()

	// If we are in a hard reset period, wait for it to pass.
	if time.Now().Before(resetTime) {
		select {
		case <-time.After(time.Until(resetTime)):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return l.rl.Wait(ctx)
}

func (p *Proxy) updateLimiterFromHeaders(l *limiter, h http.Header) {
	policyStr := h.Get("ratelimit-policy")
	remainingStr := h.Get("ratelimit-remaining")
	resetStr := h.Get("ratelimit-reset")

	l.mu.Lock()
	defer l.mu.Unlock()

	if policyStr != "" {
		// Example: 3000;w=300
		parts := strings.Split(policyStr, ";")
		if len(parts) == 2 {
			reqsStr := parts[0]
			winStr := strings.TrimPrefix(parts[1], "w=")

			reqs, err1 := strconv.Atoi(reqsStr)
			winSecs, err2 := strconv.Atoi(winStr)

			if err1 == nil && err2 == nil && winSecs > 0 {
				newRate := rate.Limit(float64(reqs) / float64(winSecs))
				if l.rl.Limit() != newRate {
					l.rl.SetLimit(newRate)
				}
			}
		}
	}

	if remainingStr != "" && resetStr != "" {
		remaining, err := strconv.Atoi(remainingStr)
		if err == nil && remaining == 0 {
			resetUnix, err := strconv.ParseInt(resetStr, 10, 64)
			if err == nil {
				l.reset = time.Unix(resetUnix, 0)
			}
		}
	}
}

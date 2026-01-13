package service

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type LoadBalancer struct {
	backends        []*Backend
	healthCheckPath string
	counter         uint64
	mu              sync.RWMutex
}

func NewLoadBalancer(urls []string, healthCheckPath string) *LoadBalancer {
	backends := make([]*Backend, 0, len(urls))
	for _, raw := range urls {
		u, err := url.Parse(raw)
		if err != nil {
			panic(fmt.Sprintf("invalid backend URL: %s", raw))
		}
		b := &Backend{URL: u}
		b.Healthy.Store(true)
		backends = append(backends, b)
	}

	return &LoadBalancer{
		backends:        backends,
		healthCheckPath: healthCheckPath,
	}
}

func (lb *LoadBalancer) getNextBackend() *Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	n := len(lb.backends)
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		idx := int(atomic.AddUint64(&lb.counter, 1) % uint64(n))
		b := lb.backends[idx]
		if b.Healthy.Load() {
			return b
		}
	}
	return nil
}

func (lb *LoadBalancer) StartHealthChecks(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			for _, backend := range lb.backends {
				go backend.CheckBackend(lb.healthCheckPath)
			}
		}
	}()
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/healthz" {
		slog.Debug("health check request received", "header", r.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		return
    }
	
	backend := lb.getNextBackend()
	if backend == nil {
		http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
		return
	}

	r.Host = backend.URL.Host
	slog.Info("forwarding request", "backend", backend.URL.String(), "path", r.URL.Path)

	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		backend.Healthy.Store(false)
		slog.Error("proxy error", "error", err, "backend", backend.URL.String())
		http.Error(w, "backend error", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}

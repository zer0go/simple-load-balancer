package service

import (
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL     *url.URL
	Healthy atomic.Bool
}

func (b *Backend) CheckBackend(healthCheckPath string) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(b.URL.String() + healthCheckPath)
	if err != nil {
		slog.Warn("Backend health check failed", "error", err.Error())
		b.Healthy.Store(false)
		return
	}
	_ = resp.Body.Close()

	b.Healthy.Store(resp.StatusCode >= 200 && resp.StatusCode < 300)
}

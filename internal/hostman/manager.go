package hostman

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
	"github.com/temoto/robotstxt"
	"golang.org/x/time/rate"
)

// HostInfo stores crawl policy & limiter for one host.
type HostInfo struct {
	robots  *robotstxt.RobotsData // nil if fetch failed
	limiter *rate.Limiter        // per-host token bucket
	fetched bool                 // robots.txt attempted
}

// Manager holds HostInfo for every domain we touch.
type Manager struct {
	mu        sync.RWMutex
	hosts     map[string]*HostInfo
	userAgent string
	rps       float64         // requests per second
	timeout   time.Duration   // robots.txt download timeout
}

// New returns a ready Manager.
func New(ua string, rps float64, robotsTimeout time.Duration) *Manager {
	return &Manager{
		userAgent: ua,
		hosts:     make(map[string]*HostInfo),
		rps:       rps,
		timeout:   robotsTimeout,
	}
}

// Check returns (allowed, waitFn).  waitFn blocks on the token bucket.
func (m *Manager) Check(u *url.URL) (bool, func(ctx context.Context) error) {
	m.mu.RLock()
	h, ok := m.hosts[u.Host]
	m.mu.RUnlock()

	if !ok {
		// Lazily create HostInfo
		h = &HostInfo{
			limiter: rate.NewLimiter(rate.Limit(m.rps), int(m.rps)), // burst = rps
		}
		// Fetch robots.txt once
		h.robots = fetchRobots(u.Scheme, u.Host, m.timeout)
		h.fetched = true

		m.mu.Lock()
		m.hosts[u.Host] = h
		m.mu.Unlock()
	}

	// robots allow/deny
	allowed := true
	if h.robots != nil {
		grp := h.robots.FindGroup(m.userAgent)
		allowed = grp.Test(u.Path)
	}

	return allowed, h.limiter.Wait
}

// --- helpers -------------------------------------------------------------

func fetchRobots(scheme, host string, timeout time.Duration) *robotstxt.RobotsData {
	robotsURL := scheme + "://" + host + "/robots.txt"

	req, _ := http.NewRequest("GET", robotsURL, nil)
	req.Header.Set("User-Agent", "GoCrawler-Robots/0.2")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		return nil // treat as no robots file
	}
	defer resp.Body.Close()

	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		return nil
	}
	return robots
}

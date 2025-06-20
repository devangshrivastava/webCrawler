package crawler

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"crawler-go/internal/frontier"
)

type Options struct {
	Seeds          []string
	MaxPages       int
	Workers        int
	Strategy       string
	RequestsPerHost float64       // NEW (rps)
	RobotsTimeout  time.Duration  // NEW
	UserAgent      string         // NEW
	mixPct        int            // NEW (0 = pure BFS, 100 = pure DFS)
}

func (o *Options) initStrategy() {
	s := strings.ToLower(o.Strategy)
	switch {
	case s == "bfs":
		o.mixPct = 0
	case s == "dfs":
		o.mixPct = 100
	case strings.HasPrefix(s, "mixed"):
		if n, err := strconv.Atoi(s[5:]); err == nil {
			if n < 0 {
				n = 0
			}
			if n > 100 {
				n = 100
			}
			o.mixPct = n
		}
	default:
		o.mixPct = 0
	}
}

// SelectURL pops from front or back according to mixPct.
func (o *Options) SelectURL(q *frontier.Queue, rng *rand.Rand) (string, bool) {
	switch {
	case o.mixPct == 0:   // BFS
		return q.PopFront()
	case o.mixPct == 100: // DFS
		return q.PopBack()
	default: // mixed
		if rng.Intn(100) < o.mixPct {
			return q.PopBack()
		}
		return q.PopFront()
	}
}

// -----------------------------------------------------------------------------
// helper called by engine once
// -----------------------------------------------------------------------------
func (o *Options) prepare() *rand.Rand {
	o.initStrategy()
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
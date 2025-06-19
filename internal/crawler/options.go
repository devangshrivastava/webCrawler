package crawler

import (
	"time"
)

type Options struct {
	Seeds          []string
	MaxPages       int
	Workers        int
	Strategy       string
	RequestsPerHost float64       // NEW (rps)
	RobotsTimeout  time.Duration  // NEW
	UserAgent      string         // NEW
}

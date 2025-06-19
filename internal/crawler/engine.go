// internal/crawler/engine.go
package crawler

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"crawler-go/internal/frontier"
	"crawler-go/internal/hostman"
	"crawler-go/internal/parser"
	"crawler-go/internal/storage"

	"github.com/joho/godotenv"
)

// -----------------------------------------------------------------------------
// Public entry-point
// -----------------------------------------------------------------------------
func Run(opts Options) error {
	_ = godotenv.Load()

	// ----- MongoDB -----------------------------------------------------------
	access := os.Getenv("MONGODB_URI") != ""
	fmt.Println("Connecting to DB at:", os.Getenv("MONGODB_URI"))
	if !access {
		fmt.Println("MongoDB access disabled, running in no-op mode")
	}
	store, err := storage.New(os.Getenv("MONGODB_URI"), access)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer store.Close()

	// ----- Frontier / visited sets ------------------------------------------
	queue   := frontier.NewQueue()
	visited := frontier.NewVisited()
	for _, s := range opts.Seeds {
		queue.Enqueue(s)
	}

	// ----- Host politeness manager ------------------------------------------
	hm := hostman.New(opts.UserAgent, opts.RequestsPerHost, opts.RobotsTimeout)

	// ----- Strategy parsing (bfs | dfs | mixedN) -----------------------------
	strat   := strings.ToLower(opts.Strategy)
	mixPct  := 0 // 0 = pure BFS, 100 = pure DFS
	switch {
	case strat == "bfs":
		mixPct = 0
	case strat == "dfs":
		mixPct = 100
	case strings.HasPrefix(strat, "mixed"):
		if n, err := strconv.Atoi(strat[5:]); err == nil {
			if n < 0 {
				n = 0
			}
			if n > 100 {
				n = 100
			}
			mixPct = n
		}
	default:
		fmt.Printf("unknown strategy %q, defaulting to bfs\n", strat)
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ----- Worker pool channels ---------------------------------------------
	jobs := make(chan string, opts.Workers*2)
	wg   := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	// ----- Dispatcher --------------------------------------------------------
	go func() {
		for {
			// global stop?
			if visited.Size() >= opts.MaxPages {
				cancel()
				close(jobs)
				return
			}

			// choose front/back according to strategy
			var webURL string
			var ok bool
			if rng.Intn(100) < mixPct {      // DFS sample
				webURL, ok = queue.PopBack()
			} else {                         // BFS default
				webURL, ok = queue.PopFront()
			}
			if !ok {                         // frontier empty
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// politeness checks (robots + rate-limit)
			parsed, err := url.Parse(webURL)
			if err != nil {
				continue
			}
			if allow, wait := hm.Check(parsed); !allow {
				continue
			} else if err := wait(ctx); err != nil {
				continue
			}

			jobs <- parsed.String()          // enqueue for workers
		}
	}()

	// ----- Workers -----------------------------------------------------------
	for i := 0; i < opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case u, ok := <-jobs:
					if !ok {
						return
					}
					if visited.Has(u) {
						continue
					}
					visited.Add(u)

					body := fetch(u)
					if len(body) == 0 {
						continue
					}

					wp, links := parser.ParseHTML(u, body)
					store.Insert(wp)

					for _, l := range links {
						if !visited.Has(l) {
							queue.Enqueue(l)
						}
					}
				}
			}
		}()
	}

	// ----- Stats ticker ------------------------------------------------------
	start  := time.Now()
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for t := range ticker.C {
			fmt.Printf("[%.0f min] crawled=%d queued=%d\n",
				t.Sub(start).Minutes(), visited.Size(), queue.Size())
		}
	}()

	wg.Wait()
	ticker.Stop()

	fmt.Println("------- FINAL STATS -------")
	fmt.Printf("Crawled: %d pages\n", visited.Size())
	fmt.Printf("Queued : %d (never visited)\n", queue.Size())
	return nil
}

// -----------------------------------------------------------------------------
// HTTP fetch helper (shared client with timeout)
// -----------------------------------------------------------------------------
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func fetch(u string) []byte {
	resp, err := httpClient.Get(u)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	const max = 1 << 20 // 1 MiB safety cap
	b, _ := io.ReadAll(io.LimitReader(resp.Body, max))
	return b
}

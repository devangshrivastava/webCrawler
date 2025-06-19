package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
	"crawler-go/internal/frontier"
	"crawler-go/internal/storage"
	"github.com/joho/godotenv"
	"crawler-go/internal/parser"
	"net/url"
	"crawler-go/internal/hostman"
	
)


func Run(opts Options) error{
	_ = godotenv.Load()
	access := os.Getenv(("MONGODB_URI")) != ""
	fmt.Println("Connecting to DB at:", os.Getenv("MONGODB_URI"))
	if !access {
		fmt.Println("MongoDB access disabled, running in no-op mode")
	}
	store, err := storage.New(os.Getenv("MONGODB_URI"), access)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer store.Close()
	queue := frontier.NewQueue()
	visited := frontier.NewVisited()

	for _, s := range opts.Seeds {
		queue.Enqueue(s)
	}

	// --- Worker pool ---
	jobs   := make(chan string, opts.Workers*2)
	wg     := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
    hm     := hostman.New(opts.UserAgent, opts.RequestsPerHost, opts.RobotsTimeout)

	// Dispatcher
	go func() {               // its own goroutine
		for {
			if visited.Size() >= opts.MaxPages {   // global stop condition
				cancel()        // ⬅ broadcasts via ctx
				close(jobs)     // no more URLs will be sent
				return
			}

			webUrl, ok := queue.Dequeue()
			if !ok {                      // queue empty? wait a bit
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// host politeness gate ↓
			parsed, _ := url.Parse(webUrl)
			if allow, wait := hm.Check(parsed); !allow {
				continue                 // disallowed by robots.txt
			} else {                     // may be rate-limited
				wait(ctx)                // blocks for token
			}

			jobs <- parsed.String()      // ⚙️  conveyor belt send (blocks if full)
		}
	}()

	// Workers
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

	// --- Stats ticker (same as old) ---
	start := time.Now()
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


var httpClient = &http.Client{
    Timeout: 15 * time.Second, // hard deadline for everything (dial + TLS + body)
    // Optionally add Transport with IdleConnTimeout etc.
}

func fetch(u string) []byte {
    resp, err := httpClient.Get(u)
    if err != nil {
        return nil
    }
    defer resp.Body.Close()

    // cheap safeguard: ignore >1 MB pages
    const max = 1 << 20
    b, _ := io.ReadAll(io.LimitReader(resp.Body, max))
    return b
}



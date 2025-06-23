package crawler


import (
	"context"
	"io"
	"net/http"
	"time"

	"crawler-go/internal/frontier"
	"crawler-go/internal/parser"
	"crawler-go/internal/storage"
	"crawler-go/internal/metrics"
)

// -----------------------------------------------------------------------------
// shared HTTP client (15-s hard timeout)
// -----------------------------------------------------------------------------
var httpClient = &http.Client{Timeout: 15 * time.Second}

// -----------------------------------------------------------------------------
// runWorker handles the whole life-cycle for one goroutine.
// -----------------------------------------------------------------------------
func runWorker(
	ctx context.Context,
	jobs <-chan string,
	visited *frontier.Visited,
	queue *frontier.Queue,
	store *storage.Store,
	tokenToParse int,
) {
	for {
		select {
		case <-ctx.Done():
			return

		case u, ok := <-jobs:
			if !ok { // channel closed
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
			title, clean, wc := parser.Extract(body, tokenToParse)
			_, links := parser.ParseHTML(u, body, tokenToParse)
			
			wp := storage.Webpage{
				URL:       u,
				Title:     title,
				Content:   clean,
				WordCount: wc,
			}
			store.Insert(wp)

			for _, l := range links {
				if !visited.Has(l) {
					queue.Enqueue(l)
				}
			}
		}
	}
}


// -----------------------------------------------------------------------------
// tiny helper — capped response reader
// -----------------------------------------------------------------------------


func fetch(u string) []byte {
	resp, err := httpClient.Get(u)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	const max = 1 << 20 // 1 MiB safety cap
	b, _ := io.ReadAll(io.LimitReader(resp.Body, max))
	metrics.BytesFetched.Add(float64(len(b)))
	metrics.PagesFetched.Inc()

	return b
}

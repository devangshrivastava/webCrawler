package crawler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"crawler-go/internal/frontier"
	"crawler-go/internal/storage"
	"golang.org/x/net/html"
	"github.com/joho/godotenv"
)


	
// 	_ = godotenv.Load()
// access := os.Getenv("MONGODB_URI") != ""
// fmt.Println("Connecting to DB at:", os.Getenv("MONGODB_URI"))

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

	// Dispatcher
	go func() {
		for {
			if visited.Size() >= opts.MaxPages {
				cancel() // broadcast stop
				close(jobs)
				return
			}
			u, ok := queue.Dequeue()
			if !ok {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			jobs <- u
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
					wp, links := parseHTML(u, body)
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




func parseHTML(currURL string, content []byte) (storage.Webpage, []string) {
	z := html.NewTokenizer(bytes.NewReader(content))
	tokenCount := 0
	bodyStarted := false
	textLen := 0
	wp := storage.Webpage{URL: currURL}
	links := make([]string, 0)

	for {
		if z.Next() == html.ErrorToken || tokenCount > 500 {
			break
		}
		t := z.Token()

		if t.Type == html.StartTagToken {
			switch t.Data {
			case "title":
				z.Next()
				wp.Title = z.Token().Data
			case "body":
				bodyStarted = true
			case "script", "style":
				z.Next() // skip contents
			case "a":
				if href := absoluteHref(t); href != "" {
					links = append(links, href)
				}
			}
		}

		if bodyStarted && t.Type == html.TextToken && textLen < 500 {
			txt := strings.TrimSpace(t.Data)
			wp.Content += txt
			textLen += len(txt)
		}
		tokenCount++
	}
	return wp, links
}

func fetch(u string) []byte {
	res, err := http.Get(u)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return b
}


func absoluteHref(tok html.Token) string {
	for _, a := range tok.Attr {
		if a.Key == "href" && strings.HasPrefix(a.Val, "http") {
			return a.Val
		}
	}
	return ""
}
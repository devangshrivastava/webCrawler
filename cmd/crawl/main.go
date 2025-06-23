package main

import (
	"flag"
	"log"
	"time"
	"crawler-go/internal/crawler"
)

func main() {
	seed    := flag.String("seed", "https://www.cc.gatech.edu/", "initial URL to start crawling from")
	max     := flag.Int   ("maxPages", 5000, "stop after N pages")
	workers := flag.Int   ("workers", 32, "number of parallel fetchers")
	strat   := flag.String("strategy", "bfs", "can choose bfs (breadth-first) or dfs (depth-first) crawling strategy or mixedN (bfs+dfs)")
	rps     := flag.Float64("maxPerHost", 2.0,  "max requests/sec to one host")
	ua      := flag.String("userAgent",  "GoCrawler/0.2", "HTTP User-Agent string")
	rt      := flag.Int   ("robotsTimeout", 5,  "robots.txt timeout (sec)")
	// tokens to parse in the HTML content
	token   := flag.Int("tokens", 5000, "max number of tokens to parse in HTML content")

	flag.Parse()

	opts := crawler.Options{
		Seeds:           []string{*seed},
		MaxPages:        *max,
		Workers:         *workers,
		Strategy:        *strat,
		RequestsPerHost: *rps,
		RobotsTimeout:   time.Duration(*rt) * time.Second,
		UserAgent:       *ua,
		Tokens:          *token,
	}

	if err := crawler.Run(opts); err != nil {
		log.Fatal(err)
	}
}
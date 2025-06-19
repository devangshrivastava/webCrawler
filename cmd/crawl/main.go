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
	strat   := flag.String("strategy", "bfs", "bfs (only choice for now)")
	rps     := flag.Float64("maxPerHost", 2.0,  "max requests/sec to one host")
	ua      := flag.String("userAgent",  "GoCrawler/0.2", "HTTP User-Agent string")
	rt      := flag.Int   ("robotsTimeout", 5,  "robots.txt timeout (sec)")

	flag.Parse()

	opts := crawler.Options{
		Seeds:           []string{*seed},
		MaxPages:        *max,
		Workers:         *workers,
		Strategy:        *strat,
		RequestsPerHost: *rps,
		RobotsTimeout:   time.Duration(*rt) * time.Second,
		UserAgent:       *ua,
	}

	if err := crawler.Run(opts); err != nil {
		log.Fatal(err)
	}
}
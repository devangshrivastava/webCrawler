package main

import (
	"flag"
	"log"

	"crawler-go/internal/crawler"
)

func main() {
	seed    := flag.String("seed", "https://www.cc.gatech.edu/", "initial URL to start crawling from")
	max     := flag.Int   ("maxPages", 5000, "stop after N pages")
	workers := flag.Int   ("workers", 32, "number of parallel fetchers")
	strat   := flag.String("strategy", "bfs", "bfs (only choice for now)")
	flag.Parse()

	opts := crawler.Options{
		Seeds:      []string{*seed},
		MaxPages:   *max,
		Workers:    *workers,
		Strategy:   *strat,
	}

	if err := crawler.Run(opts); err != nil {
		log.Fatal(err)
	}
}
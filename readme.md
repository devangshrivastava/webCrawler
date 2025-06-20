# Go Web Crawler

A concurrent and configurable web crawler built in Go. This project supports BFS, DFS, and mixed crawling strategies, and uses Prometheus for performance monitoring and MongoDB for storage.

## ✅ What I Have Done

- Built a modular crawler engine using Go
- Implemented:
  - BFS, DFS, and mixed strategy queue management
  - Per-host politeness via rate limiting and `robots.txt` handling
  - URL frontier and deduplication system
- Connected to MongoDB for persistent storage
- Added Prometheus instrumentation to expose metrics (e.g. pages fetched, bytes downloaded, GC stats)
- Spawned workers using goroutines for parallel crawling
- Built CLI with flags for seeds, max pages, strategy, concurrency, and rate limits

## 🚀 Features

- **Concurrency**: Efficiently uses goroutines for multiple parallel workers
- **Crawl Strategies**:
  - `bfs`: Breadth-first crawling
  - `dfs`: Depth-first crawling
  - `mixedN`: Blend of BFS and DFS where `N` is a percentage
- **Host Politeness**:
  - Respects `robots.txt`
  - Enforces per-host rate limiting
- **Metrics**:
  - Prometheus `/metrics` endpoint
  - Tracks pages crawled, bytes downloaded, memory usage, and more
- **Storage**:
  - Pluggable MongoDB support (auto-disabled if URI not set)
  - Stores visited pages and metadata

## 🛠️ How to Run

```bash
go run ./cmd/crawl \
  -seed https://www.cc.gatech.edu/ \
  -maxPages 100 \
  -workers 8 \
  -strategy bfs \
  -maxPerHost 2.0

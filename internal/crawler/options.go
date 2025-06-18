package crawler

type Options struct {
	Seeds []string
	MaxPages int 
	Workers int
	Strategy string // bfs or dfs
}
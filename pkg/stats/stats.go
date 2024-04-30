package stats

import (
	"fmt"
	"os"
	"sync"

	"github.com/g0ldencybersec/Caduceus/pkg/types"
)

type Stats struct {
	hits   int64
	misses int64
	total  int64
	mu     sync.Mutex
}

func (s *Stats) HitPercentage() float64 {
	if s.total == 0 {
		return 0.0
	}
	return float64(s.hits) / float64(s.total) * 100
}

func (s *Stats) Update(result types.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.total++
	if result.Hit {
		s.hits++
	} else {
		s.misses++
	}
}

func (s *Stats) Display() {
	// Return to the start of the line, clear it, and re-display the stats
	fmt.Fprintf(os.Stdout, "\r\033[KHits: %d, Misses: %d, Total: %d, Hit Rate: %.2f%%", s.hits, s.misses, s.total, float64(s.hits)/float64(s.total)*100)
}

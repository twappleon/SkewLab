package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/replay"
)

// replayReviewPublic is an on-demand read-only aggregate. Access is cached for
// 3 minutes so a public visitor cannot repeatedly parse the 50MiB archive.
// No trade IDs, market slugs, token IDs or paths leave this handler.
func (a *App) replayReviewPublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	a.replayMu.Lock()
	defer a.replayMu.Unlock()
	if a.replayCachedAt.IsZero() || time.Since(a.replayCachedAt) > 3*time.Minute {
		a.mu.RLock()
		positions := append([]paper.Position(nil), a.ledger.Positions...)
		cfg := a.ledger.Config
		a.mu.RUnlock()
		ledger := paper.Ledger{Positions: positions, Config: cfg}
		a.mu.RLock()
		ledger.Cash = a.ledger.Cash
		ledger.RealizedPnL = a.ledger.RealizedPnL
		ledger.TotalClosed = a.ledger.TotalClosed
		a.mu.RUnlock()
		archive := replay.Archive{}
		summary := replay.BuildReview(positions, cfg, archive)
		paths, err := filepath.Glob(filepath.Join(a.dataDir, "observations.jsonl*"))
		if err == nil {
			sort.Strings(paths)
			var bytes int64
			for _, p := range paths {
				st, e := os.Stat(p)
				if e != nil || !st.Mode().IsRegular() {
					err = os.ErrInvalid
					break
				}
				bytes += st.Size()
				if bytes > 80<<20 {
					err = os.ErrInvalid
					break
				}
			}
		}
		if err == nil && len(paths) > 0 {
			var data replay.Archive
			data, err = replay.Load(paths)
			if err == nil {
				summary = replay.BuildReview(positions, cfg, data)
			}
		}
		if err != nil {
			summary.Status = "OBSERVATION_ARCHIVE_UNAVAILABLE_OR_INCOMPLETE"
		}
		gate := replay.BuildEvidenceGate(&ledger, archive, summary)
		summary.Evidence = &gate
		a.replayCached = summary
		a.replayCachedAt = time.Now()
	}
	_ = json.NewEncoder(w).Encode(a.replayCached)
}

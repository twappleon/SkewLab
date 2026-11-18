package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/twappleon/SkewLab/internal/edge"
)

// edgeRun uses public official closed outcomes only. It is independent of
// new-entry permission and never sends an order or alters the trading ledger.
func (a *App) edgeRun(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, record := range a.edge.Due(time.Now().UTC(), 8) {
				if ctx.Err() != nil {
					return
				}
				outcome, err := a.api.Settlement(ctx, record.Slug, record.ConditionID)
				if err != nil {
					continue
				} // unresolved, unavailable or contradictory => no label
				if outcome.Winner != "UP" && outcome.Winner != "DOWN" {
					continue
				}
				if err = a.edge.Resolve(record.Slug, record.ConditionID, outcome.Winner, time.Now().UTC()); err != nil {
					log.Printf("edge outcome journal unavailable; research halted: %v", err)
					break
				}
			}
		}
	}
}

func (a *App) edgeResearchPrivate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	records := a.edge.Snapshot()
	st := a.edge.Status()
	st.PaperEntriesEnabled = paperEntriesEnabled()
	// The study is **not** permission to enable entries: its model assumes
	// settle-to-expiry, unlike the legacy paper strategy's early exits.
	out := struct {
		Status      edge.RecordStatus `json:"status"`
		Quality edge.Quality `json:"quality"`
		WalkForward edge.WalkForward  `json:"walk_forward"`
		Recent      []edge.Candidate  `json:"recent_private_samples"`
		Notice      string            `json:"notice"`
	}{st, a.edge.AssessQuality(time.Now().UTC()), edge.WalkForwardCheck(records, .002, 25), records, "Research only; no real trading, no paper-entry approval. Gamma is not independent Chainlink evidence."}
	if len(out.Recent) > 100 {
		out.Recent = out.Recent[len(out.Recent)-100:]
	}
	_ = json.NewEncoder(w).Encode(out)
}

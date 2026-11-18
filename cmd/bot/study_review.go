package main

import (
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/validation"
	"net/http"
	"time"
)

// studyReviewPublic exposes aggregate research diagnostics only; no raw
// observations, account credentials, market slugs, token IDs or order controls.
func (a *App) studyReviewPublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	review := validation.BuildReview(a.comparison.Snapshot(), time.Now().UTC())
	_ = json.NewEncoder(w).Encode(review)
}

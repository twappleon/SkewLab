// Package replay audits recorded quotes and paper positions without creating fills
// or trading decisions. Missing observations always remain missing evidence.
package replay

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"time"

	"github.com/twappleon/SkewLab/internal/polymarket"
)

const MaxFrames = 120000
const MaxLineBytes = 4 << 20

type Frame struct {
	Time    time.Time           `json:"time"`
	BTC     float64             `json:"btc"`
	Source  string              `json:"source"`
	Markets []polymarket.Market `json:"markets"`
}

type Archive struct {
	Frames     []Frame
	Files      int
	TotalLines int
}

// Load reads archived JSONL segments in lexical order and fails closed on malformed
// records, duplicate timestamps, oversized archives and invalid spot quotes.
// Callers must supply actual, privately recorded files; no synthetic fallback.
func Load(paths []string) (Archive, error) {
	out := Archive{Frames: make([]Frame, 0)}
	if len(paths) == 0 {
		return out, errors.New("no observation archive files")
	}
	for _, path := range paths {
		if path == "" {
			return Archive{}, errors.New("empty observation path")
		}
		f, err := os.Open(path)
		if err != nil {
			return Archive{}, fmt.Errorf("open archive: %w", err)
		}
		err = readFrames(f, &out)
		closeErr := f.Close()
		if err != nil {
			return Archive{}, fmt.Errorf("invalid archive: %w", err)
		}
		if closeErr != nil {
			return Archive{}, closeErr
		}
		out.Files++
	}
	if len(out.Frames) == 0 {
		return Archive{}, errors.New("no valid observations")
	}
	sort.Slice(out.Frames, func(i, j int) bool { return out.Frames[i].Time.Before(out.Frames[j].Time) })
	for i := 1; i < len(out.Frames); i++ {
		if !out.Frames[i].Time.After(out.Frames[i-1].Time) {
			return Archive{}, errors.New("duplicate or nonmonotonic observation timestamp")
		}
	}
	return out, nil
}
func readFrames(rd io.Reader, a *Archive) error {
	sc := bufio.NewScanner(rd)
	sc.Buffer(make([]byte, 8192), MaxLineBytes)
	for sc.Scan() {
		a.TotalLines++
		if len(a.Frames) >= MaxFrames {
			return errors.New("archive sample limit exceeded")
		}
		var f Frame
		if err := json.Unmarshal(sc.Bytes(), &f); err != nil {
			return fmt.Errorf("line %d JSON: %w", a.TotalLines, err)
		}
		if f.Time.IsZero() || f.Source == "" || !(f.BTC > 100) || math.IsNaN(f.BTC) || math.IsInf(f.BTC, 0) {
			return fmt.Errorf("line %d: invalid frame", a.TotalLines)
		}
		a.Frames = append(a.Frames, f)
	}
	return sc.Err()
}

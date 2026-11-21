package research100

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// Dataset is an immutable snapshot of validated observations. One identity
// can have at most one timestamp; a duplicate is an error, never a second win.
type Dataset struct {
	Items []Observation `json:"items"`
}

func BuildDataset(raw []Observation) (Dataset, error) {
	if len(raw) > 50000 {
		return Dataset{}, errors.New("observation capacity exceeded")
	}
	seen := make(map[string]bool, len(raw))
	out := Dataset{Items: make([]Observation, 0, len(raw))}
	for i, o := range raw {
		if e := o.Validate(); e != nil {
			return Dataset{}, fmt.Errorf("observation %d: %w", i, e)
		}
		if seen[o.ID] {
			return Dataset{}, fmt.Errorf("repeated market identity at observation %d", i)
		}
		seen[o.ID] = true
		out.Items = append(out.Items, o)
	}
	sort.Slice(out.Items, func(i, j int) bool { return out.Items[i].CapturedAt.Before(out.Items[j].CapturedAt) })
	return out, nil
}
func (d Dataset) Bounds() (*time.Time, *time.Time) {
	if len(d.Items) == 0 {
		return nil, nil
	}
	min, max := d.Items[0].CapturedAt, d.Items[len(d.Items)-1].CapturedAt
	return &min, &max
}

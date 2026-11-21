package research100

import (
	"time"
)

type Coverage struct {
	Observed       int        `json:"observed"`
	Expected       *int       `json:"expected,omitempty"`
	Fraction       *float64   `json:"fraction,omitempty"`
	Missing        int        `json:"missing"`
	MonitoredFrom  *time.Time `json:"monitored_from,omitempty"`
	MonitoredUntil *time.Time `json:"monitored_until,omitempty"`
	Status         string     `json:"status"`
}

// AssessCoverage reports a denominator ONLY if the operator supplies an
// independently recorded monitoring interval. A data-only sample cannot tell
// which slots were missed during outages or before the study started.
func AssessCoverage(d Dataset, window time.Duration, from, to *time.Time) Coverage {
	c := Coverage{Observed: len(d.Items), Status: "UNKNOWN_MONITORING_INTERVAL"}
	if from == nil || to == nil || from.IsZero() || to.IsZero() || !from.Before(*to) || window <= 0 {
		return c
	}
	// Count UTC-aligned market starts within [from,to). A market with an
	// observation outside the monitoring interval is not in this denominator.
	step := int64(window.Seconds())
	if step < 1 {
		return c
	}
	first := (from.Unix() + step - 1) / step * step
	last := (to.Unix() - 1) / step * step
	expected := 0
	if last >= first {
		expected = int((last-first)/step + 1)
	}
	if expected > 1000000 {
		return c
	}
	observed := 0
	for _, o := range d.Items {
		if !o.CapturedAt.Before(*from) && o.CapturedAt.Before(*to) {
			observed++
		}
	}
	if observed > expected {
		return c
	} // wrong sampling policy or conflicting window
	c.Expected = &expected
	c.Observed = observed
	c.MonitoredFrom = from
	c.MonitoredUntil = to
	c.Missing = expected - observed
	c.Status = "MEASURED_WINDOW_ONLY"
	if expected > 0 {
		v := float64(observed) / float64(expected)
		c.Fraction = &v
	}
	return c
}

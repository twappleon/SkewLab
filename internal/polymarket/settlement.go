package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
)

// Settlement is the outcome reported by the official Polymarket Gamma API.
// It is NOT an independently captured Chainlink TWAP stream.
type Settlement struct {
	Slug        string `json:"slug"`
	ConditionID string `json:"condition_id"`
	Winner      string `json:"winner"`
	Source      string `json:"source"`
}

// Settlement returns a confirmed binary result only when the official market
// reports closed AND the published outcome prices unequivocally identify a
// single winner. Any contradictory or missing response fails closed.
func (c *Client) Settlement(ctx context.Context, slug, conditionID string) (Settlement, error) {
	if slug == "" || conditionID == "" {
		return Settlement{}, errors.New("market slug and condition ID are required")
	}
	var m gmSettlement
	if err := c.get(ctx, c.Gamma+"/markets/slug/"+url.PathEscape(slug), &m); err != nil {
		return Settlement{}, err
	}
	if m.Slug != slug || !strings.EqualFold(m.ConditionID, conditionID) {
		return Settlement{}, fmt.Errorf("resolved market identity mismatch for %s", slug)
	}
	if !m.Closed {
		return Settlement{}, errors.New("market not yet closed")
	}
	outcomes, err := asStrings(m.Outcomes)
	if err != nil {
		return Settlement{}, fmt.Errorf("unreadable outcome labels: %w", err)
	}
	prices, err := asStrings(m.OutcomePrices)
	if err != nil {
		return Settlement{}, fmt.Errorf("unreadable outcome prices: %w", err)
	}
	if len(outcomes) != 2 || len(prices) != 2 {
		return Settlement{}, errors.New("not a binary outcome")
	}
	winner := ""
	for i, label := range outcomes {
		name := strings.ToUpper(strings.TrimSpace(label))
		if name != "UP" && name != "DOWN" {
			return Settlement{}, errors.New("unexpected outcome label")
		}
		p, e := parseProbability(prices[i])
		if e != nil {
			return Settlement{}, e
		}
		if p == 1 {
			if winner != "" {
				return Settlement{}, errors.New("multiple winners")
			}
			winner = name
		}
		if p != 0 && p != 1 {
			return Settlement{}, errors.New("unresolved intermediate outcome price")
		}
	}
	if winner == "" {
		return Settlement{}, errors.New("no published winner")
	}
	return Settlement{Slug: slug, ConditionID: conditionID, Winner: winner, Source: "Polymarket Gamma closed market outcome (not independent TWAP proof)"}, nil
}

type gmSettlement struct {
	Slug          string          `json:"slug"`
	ConditionID   string          `json:"conditionId"`
	Closed        bool            `json:"closed"`
	Outcomes      json.RawMessage `json:"outcomes"`
	OutcomePrices json.RawMessage `json:"outcomePrices"`
}

func parseProbability(value string) (float64, error) {
	switch value {
	case "1", "1.0", "1.00", "1.000":
		return 1, nil
	case "0", "0.0", "0.00", "0.000":
		return 0, nil
	default:
		return math.NaN(), fmt.Errorf("non-final outcome price %q", value)
	}
}

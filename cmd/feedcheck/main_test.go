package main

import (
	"testing"
	"time"

	"github.com/twappleon/SkewLab/internal/polymarket"
)

func TestValidPairRequiresTwoFreshBooksAndActiveMarket(t *testing.T) {
	now := time.Now().UTC()
	m := polymarket.Market{
		Active: true, UpToken: "up", DownToken: "down", End: now.Add(time.Minute),
		Up:   polymarket.Quote{Bid: .45, Ask: .46, BidDepth: 10, AskDepth: 20, UpdatedAt: now},
		Down: polymarket.Quote{Bid: .52, Ask: .54, BidDepth: 30, AskDepth: 40, UpdatedAt: now},
	}
	if !validPair(m, now) {
		t.Fatal("valid pair rejected")
	}
	bad := m
	bad.Down.UpdatedAt = now.Add(-13 * time.Second)
	if validPair(bad, now) {
		t.Fatal("stale token accepted")
	}
	bad = m
	bad.Up.Ask = bad.Up.Bid
	if validPair(bad, now) {
		t.Fatal("crossed token accepted")
	}
	bad = m
	bad.Down.AskDepth = 0
	if validPair(bad, now) {
		t.Fatal("empty ask accepted")
	}
	bad = m
	bad.Active = false
	if validPair(bad, now) {
		t.Fatal("inactive market accepted")
	}
	bad = m
	bad.DownToken = bad.UpToken
	if validPair(bad, now) {
		t.Fatal("duplicate tokens accepted")
	}
	bad = m
	bad.End = now
	if validPair(bad, now) {
		t.Fatal("expired market accepted")
	}
}

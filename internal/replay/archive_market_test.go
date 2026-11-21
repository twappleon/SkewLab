package replay

import (
	"strings"
	"testing"
)

func TestDuplicateMarketInSingleRecordedSnapshotRejected(t *testing.T) {
	data := `{"time":"2026-11-18T01:00:00Z","btc":90000,"source":"Coinbase","markets":[{"slug":"x"},{"slug":"x"}]}` + "\n"
	a := Archive{}
	if e := readFrames(strings.NewReader(data), &a); e == nil || !strings.Contains(e.Error(), "duplicate market") {
		t.Fatalf("duplicate market accepted %v", e)
	}
	data = `{"time":"2026-11-18T01:00:00Z","btc":90000,"source":"Coinbase","markets":[{"slug":"x"},{"slug":"y"}]}` + "\n"
	a = Archive{}
	if e := readFrames(strings.NewReader(data), &a); e != nil {
		t.Fatal(e)
	}
}

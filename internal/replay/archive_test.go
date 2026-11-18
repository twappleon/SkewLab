package replay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadArchiveNoSyntheticOrSilentRepair(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "observations.jsonl")
	if _, e := Load(nil); e == nil {
		t.Fatal("missing path accepted")
	}
	t1 := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	sample := `{"time":"` + t1.Format(time.RFC3339Nano) + `","btc":95000,"source":"Binance BTCUSDT spot","markets":[]}`
	if e := os.WriteFile(f, []byte(sample+"\n"), 0600); e != nil {
		t.Fatal(e)
	}
	a, e := Load([]string{f})
	if e != nil || len(a.Frames) != 1 || a.Files != 1 || a.TotalLines != 1 {
		t.Fatalf("load=%+v %v", a, e)
	}
	if e := os.WriteFile(f, []byte(sample+"\n"+sample+"\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = Load([]string{f}); e == nil || !strings.Contains(e.Error(), "duplicate") {
		t.Fatalf("duplicate not flagged %v", e)
	}
	if e := os.WriteFile(f, []byte(sample+"\n{not-json\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = Load([]string{f}); e == nil {
		t.Fatal("truncated JSON accepted")
	}
}

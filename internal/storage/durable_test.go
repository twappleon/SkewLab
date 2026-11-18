package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicAndRestartRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	for _, content := range []string{"{\"cash\":900}", "{\"cash\":915}"} {
		if err := Atomic(path, []byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	got, e := os.ReadFile(path)
	if e != nil || string(got) != "{\"cash\":915}" {
		t.Fatalf("atomic failed: %s %v", got, e)
	}
	fi, _ := os.Stat(path)
	if fi.Mode().Perm() != 0600 {
		t.Fatalf("insecure file permissions: %s", fi.Mode())
	}
}
func TestAppendAndTruncatedRecoveryRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ticks.jsonl")
	for n := 0; n < 4; n++ {
		if e := AppendJSONL(path, map[string]int{"price": n}, 10000); e != nil {
			t.Fatal(e)
		}
	}
	rows, e := ReadRecentJSONL[map[string]int](path, 2)
	if e != nil || len(rows) != 2 || rows[0]["price"] != 2 || rows[1]["price"] != 3 {
		t.Fatalf("recovery bad %+v: %v", rows, e)
	}
	file, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	_, _ = file.WriteString("{corrupted\n")
	_ = file.Close()
	if _, e := ReadRecentJSONL[map[string]int](path, 10); e == nil {
		t.Fatal("corrupted final line was accepted")
	}
}

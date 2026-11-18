// Package storage contains local, sync-to-disk persistence primitives.
// Intended for a single bot process per volume (no distributed write locks).
package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Atomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	f, e := os.CreateTemp(dir, ".skewlab-*.tmp")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if e = f.Chmod(0600); e != nil {
		f.Close()
		return e
	}
	if _, e = f.Write(data); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(tmp, path); e != nil {
		return e
	}
	if d, e := os.Open(dir); e == nil {
		defer d.Close()
		return d.Sync()
	}
	return nil
}

// AppendJSONL rotates oversized segments, and fsyncs every appended record.
// A partial final line after a crash is never silently interpreted as data.
func AppendJSONL(path string, value any, maxSize int64) error {
	if maxSize <= 0 {
		return errors.New("invalid max size")
	}
	if stat, e := os.Stat(path); e == nil && stat.Size() >= maxSize {
		archive := fmt.Sprintf("%s.%s", path, time.Now().UTC().Format("20060102T150405.000000000"))
		if e = os.Rename(path, archive); e != nil {
			return e
		}
	} else if e != nil && !os.IsNotExist(e) {
		return e
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	if e = json.NewEncoder(f).Encode(value); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	return f.Close()
}
func ReadRecentJSONL[T any](path string, limit int) ([]T, error) {
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}
	f, e := os.Open(path)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4<<20)
	records := make([]T, 0)
	for sc.Scan() {
		var x T
		if e = json.Unmarshal(sc.Bytes(), &x); e != nil {
			return nil, fmt.Errorf("corrupted jsonl record: %w", e)
		}
		records = append(records, x)
		if len(records) > limit {
			records = records[1:]
		}
	}
	if e = sc.Err(); e != nil {
		return nil, e
	}
	return records, nil
}

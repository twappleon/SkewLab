package wirews

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func sendFrame(w io.Writer, opcode byte, payload string) error {
	b := []byte(payload)
	h := []byte{0x80 | opcode}
	if len(b) < 126 {
		h = append(h, byte(len(b)))
	} else {
		h = append(h, 126, byte(len(b)>>8), byte(len(b)))
	}
	_, e := w.Write(append(h, b...))
	return e
}
func TestWebsocketHandshakeMaskAndPing(t *testing.T) {
	done := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := w.(http.Hijacker)
		if !ok {
			t.Error("no hijack")
			return
		}
		sock, buf, e := h.Hijack()
		if e != nil {
			t.Error(e)
			return
		}
		defer sock.Close()
		digest := sha1.Sum([]byte(r.Header.Get("Sec-WebSocket-Key") + magic))
		_, _ = fmt.Fprintf(sock, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", base64.StdEncoding.EncodeToString(digest[:]))
		hdr := make([]byte, 2)
		_, e = io.ReadFull(buf, hdr)
		if e != nil {
			t.Error(e)
			return
		}
		if hdr[1]&0x80 == 0 {
			t.Error("client must mask frames")
		}
		size := int(hdr[1] & 0x7f)
		if size == 126 {
			length := make([]byte, 2)
			_, _ = io.ReadFull(buf, length)
			size = int(binary.BigEndian.Uint16(length))
		}
		mask := make([]byte, 4)
		_, _ = io.ReadFull(buf, mask)
		msg := make([]byte, size)
		_, _ = io.ReadFull(buf, msg)
		for i := range msg {
			msg[i] ^= mask[i%4]
		}
		done <- string(msg)
		_ = sendFrame(sock, 9, "hello")
		_ = sendFrame(sock, 1, `{"ok":true}`)
		_ = buf.Flush()
		// Keep socket open long enough to receive the client's masked PONG.
		_ = sock.SetReadDeadline(time.Now().Add(2 * time.Second))
		var pongHeader [2]byte
		if _, e = io.ReadFull(buf, pongHeader[:]); e == nil && pongHeader[0]&0x0f == 10 {
			count := int(pongHeader[1] & 0x7f)
			mask := make([]byte, 4)
			_, _ = io.ReadFull(buf, mask)
			pong := make([]byte, count)
			_, _ = io.ReadFull(buf, pong)
		}
	}))
	defer server.Close()
	c, err := Dial(context.Background(), strings.Replace(server.URL, "http://", "ws://", 1))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.WriteText([]byte(`{"type":"market"}`)); err != nil {
		t.Fatal(err)
	}
	select {
	case x := <-done:
		if x != `{"type":"market"}` {
			t.Fatal(x)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no masked frame")
	}
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	b, err := c.ReadText()
	if err != nil || string(b) != `{"ok":true}` {
		t.Fatalf("read %s %v", b, err)
	}
}
func TestBadUpgradeFails(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("no")) }))
	defer s.Close()
	_, err := Dial(context.Background(), strings.Replace(s.URL, "http://", "ws://", 1))
	if err == nil {
		t.Fatal("accepted invalid upgrade")
	}
}

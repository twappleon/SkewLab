// Package wirews is a minimal RFC6455 client for read-only public market
// streams. It has no order-signing, private-wallet or trading functionality.
package wirews

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const maxFrame = 2 << 20
const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type Conn struct {
	socket  net.Conn
	reader  *bufio.Reader
	writeMu sync.Mutex
}

func Dial(ctx context.Context, raw string) (*Conn, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "wss" && u.Scheme != "ws" {
		return nil, errors.New("only ws/wss supported")
	}
	if u.Host == "" || u.User != nil {
		return nil, errors.New("invalid websocket host")
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "wss" {
			port = "443"
		} else {
			port = "80"
		}
	}
	d := &net.Dialer{Timeout: 8 * time.Second, KeepAlive: 30 * time.Second}
	var sock net.Conn
	if u.Scheme == "wss" {
		sock, err = (&tls.Dialer{NetDialer: d, Config: &tls.Config{MinVersion: tls.VersionTLS12}}).DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	} else {
		sock, err = d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = sock.Close()
		}
	}()
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	key := base64.StdEncoding.EncodeToString(nonce[:])
	u.Fragment = ""
	req := &http.Request{Method: http.MethodGet, URL: u, Host: u.Host, Header: make(http.Header)}
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", key)
	req.Header.Set("Sec-WebSocket-Version", "13")
	_ = sock.SetDeadline(time.Now().Add(10 * time.Second))
	if err = req.Write(sock); err != nil {
		return nil, err
	}
	br := bufio.NewReader(sock)
	res, readErr := http.ReadResponse(br, req)
	if readErr != nil {
		err = readErr
		return nil, err
	}
	defer res.Body.Close()
	digest := sha1.Sum([]byte(key + magic))
	accept := base64.StdEncoding.EncodeToString(digest[:])
	if res.StatusCode != 101 || !strings.EqualFold(res.Header.Get("Upgrade"), "websocket") || !strings.Contains(strings.ToLower(res.Header.Get("Connection")), "upgrade") || res.Header.Get("Sec-WebSocket-Accept") != accept {
		err = fmt.Errorf("websocket handshake rejected: HTTP %d", res.StatusCode)
		return nil, err
	}
	_ = sock.SetDeadline(time.Time{})
	return &Conn{socket: sock, reader: br}, nil
}

func (c *Conn) Close() error                      { return c.socket.Close() }
func (c *Conn) SetReadDeadline(t time.Time) error { return c.socket.SetReadDeadline(t) }
func (c *Conn) WriteText(b []byte) error          { return c.write(1, b) }
func (c *Conn) write(opcode byte, b []byte) error {
	if len(b) > maxFrame {
		return errors.New("websocket frame too large")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	header := []byte{0x80 | opcode}
	n := len(b)
	switch {
	case n < 126:
		header = append(header, 0x80|byte(n))
	case n <= 65535:
		header = append(header, 0x80|126, byte(n>>8), byte(n))
	default:
		header = append(header, 0x80|127, 0, 0, 0, 0, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	}
	mask := make([]byte, 4)
	if _, err := rand.Read(mask); err != nil {
		return err
	}
	header = append(header, mask...)
	masked := make([]byte, len(b))
	for i, v := range b {
		masked[i] = v ^ mask[i%4]
	}
	if _, err := c.socket.Write(header); err != nil {
		return err
	}
	_, err := c.socket.Write(masked)
	return err
}

// ReadText consumes fragmented messages, responds to ping frames and rejects
// unexpectedly masked server frames. Call from exactly one reader goroutine.
func (c *Conn) ReadText() ([]byte, error) {
	var data []byte
	var fragmenting bool
	for {
		var hdr [2]byte
		if _, err := io.ReadFull(c.reader, hdr[:]); err != nil {
			return nil, err
		}
		fin := hdr[0]&0x80 != 0
		op := hdr[0] & 0x0f
		if hdr[0]&0x70 != 0 || hdr[1]&0x80 != 0 {
			return nil, errors.New("invalid server websocket frame")
		}
		size := uint64(hdr[1] & 0x7f)
		if size == 126 {
			var b [2]byte
			if _, err := io.ReadFull(c.reader, b[:]); err != nil {
				return nil, err
			}
			size = uint64(binary.BigEndian.Uint16(b[:]))
		}
		if size == 127 {
			var b [8]byte
			if _, err := io.ReadFull(c.reader, b[:]); err != nil {
				return nil, err
			}
			size = binary.BigEndian.Uint64(b[:])
		}
		if size > maxFrame || size+uint64(len(data)) > maxFrame {
			return nil, errors.New("websocket payload too large")
		}
		if op >= 8 && (!fin || size > 125) {
			return nil, errors.New("invalid control frame")
		}
		chunk := make([]byte, int(size))
		if _, err := io.ReadFull(c.reader, chunk); err != nil {
			return nil, err
		}
		switch op {
		case 8:
			return nil, io.EOF
		case 9:
			if err := c.write(10, chunk); err != nil {
				return nil, err
			}
			continue
		case 10:
			continue
		case 1, 2:
			if fragmenting {
				return nil, errors.New("overlapping websocket fragments")
			}
			data = append(data, chunk...)
			fragmenting = !fin
		case 0:
			if !fragmenting {
				return nil, errors.New("unexpected continuation")
			}
			data = append(data, chunk...)
			fragmenting = !fin
		default:
			return nil, errors.New("unknown websocket opcode")
		}
		if !fragmenting {
			return data, nil
		}
	}
}

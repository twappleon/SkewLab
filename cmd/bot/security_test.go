package main

import "testing"

func TestLocalFirstListenAddr(t *testing.T) {
	cases := []struct {
		name, host, port         string
		railway, password, proxy bool
		want                     string
		bad                      bool
	}{
		{"default local", "", "", false, false, false, "127.0.0.1:8080", false},
		{"override local", "localhost", "8181", false, false, false, "localhost:8181", false},
		{"ipv6 local", "::1", "8081", false, false, false, "[::1]:8081", false},
		{"insecure LAN", "0.0.0.0", "8181", false, false, false, "", true},
		{"password but no https", "0.0.0.0", "8181", false, true, false, "", true},
		{"authenticated proxy", "0.0.0.0", "8181", false, true, true, "0.0.0.0:8181", false},
		{"railway missing password", "", "8080", true, false, false, "", true},
		{"railway authenticated", "", "8080", true, true, false, "0.0.0.0:8080", false},
		{"invalid port", "", "bogus", false, false, false, "", true},
		{"zero port", "", "0", false, false, false, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := localFirstListenAddr(tc.host, tc.port, tc.railway, tc.password, tc.proxy)
			if (err != nil) != tc.bad || got != tc.want {
				t.Fatalf("got %q, %v; want %q, bad %v", got, err, tc.want, tc.bad)
			}
		})
	}
}

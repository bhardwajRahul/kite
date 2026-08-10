package connector

import (
	"errors"
	"io"
	"net/http"
	"testing"
)

// recordingRoundTripper records whether it was called and returns a fixed
// response. It is used to verify which transport the upgradeAwareTransport
// delegates to.
type recordingRoundTripper struct {
	name    string
	calls   int
	lastReq *http.Request
	respErr error
}

func (r *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls++
	r.lastReq = req
	if r.respErr != nil {
		return nil, r.respErr
	}
	// Return a minimal non-nil response so ReverseProxy is satisfied when used
	// in integration-style tests; here we only inspect call counts.
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

func TestUpgradeAwareTransportRouting(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		upgrade    string
		setEmpty   bool // explicitly set Upgrade: "" (empty value header)
		wantNormal int
		wantH1     int
	}{
		{name: "no upgrade header - GET", method: http.MethodGet, wantNormal: 1, wantH1: 0},
		{name: "no upgrade header - POST", method: http.MethodPost, wantNormal: 1, wantH1: 0},
		{name: "no upgrade header - DELETE", method: http.MethodDelete, wantNormal: 1, wantH1: 0},
		{name: "websocket upgrade", method: http.MethodGet, upgrade: "websocket", wantNormal: 0, wantH1: 1},
		{name: "spdy/3.1 upgrade", method: http.MethodGet, upgrade: "SPDY/3.1", wantNormal: 0, wantH1: 1},
		{name: "spdy/4 upgrade", method: http.MethodGet, upgrade: "SPDY/4", wantNormal: 0, wantH1: 1},
		{name: "post exec upgrade", method: http.MethodPost, upgrade: "SPDY/3.1", wantNormal: 0, wantH1: 1},
		// An empty-valued Upgrade header is treated by http.Header.Get as ""
		// and should route to the normal transport.
		{name: "empty upgrade header value", method: http.MethodGet, setEmpty: true, wantNormal: 1, wantH1: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normal := &recordingRoundTripper{name: "normal"}
			http1Only := &recordingRoundTripper{name: "http1Only"}
			transport := &upgradeAwareTransport{
				normal:    normal,
				http1Only: http1Only,
			}

			req, err := http.NewRequest(tt.method, "https://k8s.example.com/api/v1/namespaces/default/pods", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			if tt.setEmpty {
				req.Header.Set("Upgrade", "")
			} else if tt.upgrade != "" {
				req.Header.Set("Upgrade", tt.upgrade)
			}

			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("RoundTrip error: %v", err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			if normal.calls != tt.wantNormal {
				t.Errorf("normal transport calls = %d, want %d", normal.calls, tt.wantNormal)
			}
			if http1Only.calls != tt.wantH1 {
				t.Errorf("http1Only transport calls = %d, want %d", http1Only.calls, tt.wantH1)
			}
		})
	}
}

// TestUpgradeAwareTransportErrorPropagation verifies that errors from the
// underlying transport are returned to the caller, not swallowed.
func TestUpgradeAwareTransportErrorPropagation(t *testing.T) {
	wantErr := errors.New("transport failure")
	normal := &recordingRoundTripper{name: "normal", respErr: wantErr}
	http1Only := &recordingRoundTripper{name: "http1Only"}
	transport := &upgradeAwareTransport{
		normal:    normal,
		http1Only: http1Only,
	}

	req, err := http.NewRequest(http.MethodGet, "https://k8s.example.com/api/v1", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("expected error %v, got %v", wantErr, err)
	}
	if http1Only.calls != 0 {
		t.Errorf("http1Only should not be called for non-upgrade request, got %d calls", http1Only.calls)
	}
}

// TestUpgradeAwareTransportRequestPassthrough verifies that the original
// request is forwarded to the underlying transport unmodified.
func TestUpgradeAwareTransportRequestPassthrough(t *testing.T) {
	normal := &recordingRoundTripper{name: "normal"}
	http1Only := &recordingRoundTripper{name: "http1Only"}
	transport := &upgradeAwareTransport{
		normal:    normal,
		http1Only: http1Only,
	}

	req, err := http.NewRequest(http.MethodPost, "https://k8s.example.com/api/v1/namespaces/default/pods/x/exec", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Upgrade", "SPDY/3.1")
	req.Header.Set("X-Custom", "keep-me")

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error: %v", err)
	}
	_ = resp.Body.Close()

	if http1Only.lastReq != req {
		t.Error("http1Only transport did not receive the original request pointer")
	}
	if got := http1Only.lastReq.Header.Get("X-Custom"); got != "keep-me" {
		t.Errorf("custom header not preserved, got %q", got)
	}
	if got := http1Only.lastReq.Header.Get("Upgrade"); got != "SPDY/3.1" {
		t.Errorf("Upgrade header not preserved, got %q", got)
	}
	if normal.calls != 0 {
		t.Errorf("normal transport should not be called for upgrade request, got %d calls", normal.calls)
	}
}

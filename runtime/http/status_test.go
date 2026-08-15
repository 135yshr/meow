package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	meowhttp "github.com/135yshr/meow/runtime/http"
	"github.com/135yshr/meow/runtime/meowrt"
)

// newStatusServer serves a chosen status code at /status/<code>.
func newStatusServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status/401":
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
		case "/status/404":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		case "/status/500":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("boom"))
		default:
			w.Header().Set("X-Meow", "nyan")
			w.Write([]byte("ok body"))
		}
	}))
}

// A failed request must not look like a successful one. Returning the error
// body as an ordinary String left 401 indistinguishable from 200 unless the
// caller matched on the body text.
func TestVerbsReportNon2xxAsFurball(t *testing.T) {
	srv := newStatusServer()
	defer srv.Close()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"unauthorized", "/status/401", "401"},
		{"not found", "/status/404", "404"},
		{"server error", "/status/500", "500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowhttp.Pounce(meowrt.NewString(srv.URL + tt.path))
			f, ok := got.(*meowrt.Furball)
			if !ok {
				t.Fatalf("expected a Furball, got %T (%s)", got, got.String())
			}
			if !strings.Contains(f.Message, tt.want) {
				t.Errorf("expected the status %s in %q", tt.want, f.Message)
			}
		})
	}
}

func TestVerbsStillReturnBodyOnSuccess(t *testing.T) {
	srv := newStatusServer()
	defer srv.Close()

	got := meowhttp.Pounce(meowrt.NewString(srv.URL + "/ok"))
	if _, ok := got.(*meowrt.Furball); ok {
		t.Fatalf("unexpected Furball: %s", got.String())
	}
	if got.String() != "ok body" {
		t.Errorf("got %q, want %q", got.String(), "ok body")
	}
}

// chase reports what happened rather than failing, which is what a reachability
// check needs.
func TestChaseExposesStatus(t *testing.T) {
	srv := newStatusServer()
	defer srv.Close()

	tests := []struct {
		name   string
		path   string
		status string
		ok     string
	}{
		{"success", "/ok", "200", "true"},
		{"unauthorized", "/status/401", "401", "false"},
		{"not found", "/status/404", "404", "false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowhttp.Chase(meowrt.NewString("GET"), meowrt.NewString(srv.URL+tt.path))
			m, ok := got.(*meowrt.Map)
			if !ok {
				t.Fatalf("expected a Map, got %T (%s)", got, got.String())
			}
			if v, _ := m.Get("status"); v.String() != tt.status {
				t.Errorf("status: got %s, want %s", v.String(), tt.status)
			}
			if v, _ := m.Get("ok"); v.String() != tt.ok {
				t.Errorf("ok: got %s, want %s", v.String(), tt.ok)
			}
		})
	}
}

func TestChaseReturnsBodyAndHeaders(t *testing.T) {
	srv := newStatusServer()
	defer srv.Close()

	got := meowhttp.Chase(meowrt.NewString("GET"), meowrt.NewString(srv.URL+"/ok"))
	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("expected a Map, got %T", got)
	}
	if v, _ := m.Get("body"); v.String() != "ok body" {
		t.Errorf("body: got %q, want %q", v.String(), "ok body")
	}
	headers, _ := m.Get("headers")
	hm, ok := headers.(*meowrt.Map)
	if !ok {
		t.Fatalf("expected headers to be a Map, got %T", headers)
	}
	if v, _ := hm.Get("X-Meow"); v == nil || v.String() != "nyan" {
		t.Errorf("expected the X-Meow header to be reported, got %v", v)
	}
}

// The body is positional so that it cannot be mistaken for the options Map.
func TestChaseWithBody(t *testing.T) {
	var gotBody, gotType, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotType = r.Header.Get("Content-Type")
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = string(buf)
		w.Write([]byte("stored"))
	}))
	defer srv.Close()

	got := meowhttp.Chase(
		meowrt.NewString("POST"),
		meowrt.NewString(srv.URL),
		meowrt.NewMap(map[string]meowrt.Value{"marker": meowrt.NewString("abc")}),
	)
	if _, ok := got.(*meowrt.Furball); ok {
		t.Fatalf("unexpected Furball: %s", got.String())
	}
	if gotMethod != "POST" {
		t.Errorf("method: got %q, want POST", gotMethod)
	}
	if gotType != "application/json" {
		t.Errorf("content type: got %q, want application/json", gotType)
	}
	if !strings.Contains(gotBody, "abc") {
		t.Errorf("expected the JSON body to be sent, got %q", gotBody)
	}
}

func TestChaseWithoutBodyAndWithOptions(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	got := meowhttp.Chase(
		meowrt.NewString("GET"),
		meowrt.NewString(srv.URL),
		meowrt.NewNil(),
		meowrt.NewMap(map[string]meowrt.Value{
			"headers": meowrt.NewMap(map[string]meowrt.Value{"Authorization": meowrt.NewString("Bearer t0ken")}),
		}),
	)
	if _, ok := got.(*meowrt.Furball); ok {
		t.Fatalf("unexpected Furball: %s", got.String())
	}
	if gotAuth != "Bearer t0ken" {
		t.Errorf("authorization: got %q, want %q", gotAuth, "Bearer t0ken")
	}
}

func TestChaseRejectsBadArguments(t *testing.T) {
	tests := []struct {
		name string
		args []meowrt.Value
	}{
		{"no arguments", nil},
		{"method only", []meowrt.Value{meowrt.NewString("GET")}},
		{"empty method", []meowrt.Value{meowrt.NewString("  "), meowrt.NewString("http://example.invalid")}},
		{"non-string method", []meowrt.Value{meowrt.NewInt(1), meowrt.NewString("http://example.invalid")}},
		{"bad body type", []meowrt.Value{
			meowrt.NewString("POST"), meowrt.NewString("http://example.invalid"), meowrt.NewInt(1),
		}},
		{"bad options type", []meowrt.Value{
			meowrt.NewString("GET"), meowrt.NewString("http://example.invalid"),
			meowrt.NewNil(), meowrt.NewString("nope"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := meowhttp.Chase(tt.args...).(*meowrt.Furball); !ok {
				t.Error("expected a Furball")
			}
		})
	}
}

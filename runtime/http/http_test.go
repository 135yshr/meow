package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	meowhttp "github.com/135yshr/meow/runtime/http"
	"github.com/135yshr/meow/runtime/meowrt"
)

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Write([]byte("pounce ok"))
	})
	mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		ct := r.Header.Get("Content-Type")
		w.Write([]byte("toss:" + ct + ":" + string(body)))
	})
	mux.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		ct := r.Header.Get("Content-Type")
		w.Write([]byte("knead:" + ct + ":" + string(body)))
	})
	mux.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Write([]byte("swat ok"))
	})
	mux.HandleFunc("/options", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "OPTIONS" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.Write([]byte("prowl ok"))
	})
	mux.HandleFunc("/large", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("x", 2048)))
	})
	mux.HandleFunc("/echo-header", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		custom := r.Header.Get("X-Custom")
		w.Write([]byte("auth=" + auth + ",custom=" + custom))
	})
	return httptest.NewServer(mux)
}

func TestPounce(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Pounce(meowrt.NewString(srv.URL + "/get"))
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Val != "pounce ok" {
		t.Errorf("expected %q, got %q", "pounce ok", s.Val)
	}
}

func TestPounceConnectionRefused(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	srv := newTestServer()
	url := srv.URL
	srv.Close() // ensure connection refused
	meowhttp.Pounce(meowrt.NewString(url + "/get"))
}

func TestPounceNonString(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	meowhttp.Pounce(meowrt.NewInt(42))
}

func TestTossStringBody(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Toss(
		meowrt.NewString(srv.URL+"/post"),
		meowrt.NewString("hello"),
	)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "toss::hello"
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestTossStringBodyWithHeaders(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Toss(
		meowrt.NewString(srv.URL+"/post"),
		meowrt.NewString(`{"name":"Tama"}`),
		meowrt.NewMap(map[string]meowrt.Value{
			"headers": meowrt.NewMap(map[string]meowrt.Value{
				"Content-Type": meowrt.NewString("application/json"),
			}),
		}),
	)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `toss:application/json:{"name":"Tama"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestTossWrongArgCount(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	meowhttp.Toss(meowrt.NewString("http://example.com"))
}

func TestKnead(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Knead(
		meowrt.NewString(srv.URL+"/put"),
		meowrt.NewMap(map[string]meowrt.Value{
			"name": meowrt.NewString("Mochi"),
		}),
	)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `knead:application/json:{"name":"Mochi"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestSwat(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Swat(meowrt.NewString(srv.URL + "/delete"))
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Val != "swat ok" {
		t.Errorf("expected %q, got %q", "swat ok", s.Val)
	}
}

func TestSwatNonString(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	meowhttp.Swat(meowrt.NewInt(42))
}

func TestProwl(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Prowl(meowrt.NewString(srv.URL + "/options"))
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Val != "prowl ok" {
		t.Errorf("expected %q, got %q", "prowl ok", s.Val)
	}
}

func TestProwlNonString(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	meowhttp.Prowl(meowrt.NewInt(42))
}

func TestPounceWithOptions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	opts := meowrt.NewMap(map[string]meowrt.Value{
		"maxBodyBytes": meowrt.NewInt(4096),
	})
	result := meowhttp.Pounce(meowrt.NewString(srv.URL+"/get"), opts)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Val != "pounce ok" {
		t.Errorf("expected %q, got %q", "pounce ok", s.Val)
	}
}

func TestPounceExceedsMaxBodyBytes(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for body exceeding maxBodyBytes")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(msg, "exceeds") {
			t.Errorf("expected truncation error, got %q", msg)
		}
	}()

	opts := meowrt.NewMap(map[string]meowrt.Value{
		"maxBodyBytes": meowrt.NewInt(16),
	})
	// /large returns 2048 bytes, limit is 16
	meowhttp.Pounce(meowrt.NewString(srv.URL+"/large"), opts)
}

func TestTossWithOptions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	opts := meowrt.NewMap(map[string]meowrt.Value{
		"maxBodyBytes": meowrt.NewInt(4096),
		"headers": meowrt.NewMap(map[string]meowrt.Value{
			"Content-Type": meowrt.NewString("application/json"),
		}),
	})
	result := meowhttp.Toss(
		meowrt.NewString(srv.URL+"/post"),
		meowrt.NewString(`{"name":"Tama"}`),
		opts,
	)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `toss:application/json:{"name":"Tama"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestTossMapBody(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	m := meowrt.NewMap(map[string]meowrt.Value{
		"name": meowrt.NewString("Tama"),
	})
	result := meowhttp.Toss(meowrt.NewString(srv.URL+"/post"), m)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `toss:application/json:{"name":"Tama"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestTossMapBodyWithContentTypeOverride(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	m := meowrt.NewMap(map[string]meowrt.Value{
		"name": meowrt.NewString("Tama"),
	})
	result := meowhttp.Toss(
		meowrt.NewString(srv.URL+"/post"),
		m,
		meowrt.NewMap(map[string]meowrt.Value{
			"headers": meowrt.NewMap(map[string]meowrt.Value{
				"Content-Type": meowrt.NewString("text/plain"),
			}),
		}),
	)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `toss:text/plain:{"name":"Tama"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestTossMapBodyWithOptions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	m := meowrt.NewMap(map[string]meowrt.Value{
		"name": meowrt.NewString("Tama"),
	})
	opts := meowrt.NewMap(map[string]meowrt.Value{
		"maxBodyBytes": meowrt.NewInt(4096),
	})
	result := meowhttp.Toss(meowrt.NewString(srv.URL+"/post"), m, opts)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `toss:application/json:{"name":"Tama"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestKneadMapBody(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	m := meowrt.NewMap(map[string]meowrt.Value{
		"name": meowrt.NewString("Mochi"),
	})
	result := meowhttp.Knead(meowrt.NewString(srv.URL+"/put"), m)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `knead:application/json:{"name":"Mochi"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestTossNestedMapBody(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	m := meowrt.NewMap(map[string]meowrt.Value{
		"cats": meowrt.NewList(
			meowrt.NewString("Tama"),
			meowrt.NewString("Mochi"),
		),
		"count": meowrt.NewInt(2),
	})
	result := meowhttp.Toss(meowrt.NewString(srv.URL+"/post"), m)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `toss:application/json:{"cats":["Tama","Mochi"],"count":2}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestSwatWithOptions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	opts := meowrt.NewMap(map[string]meowrt.Value{
		"maxBodyBytes": meowrt.NewInt(4096),
	})
	result := meowhttp.Swat(meowrt.NewString(srv.URL+"/delete"), opts)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Val != "swat ok" {
		t.Errorf("expected %q, got %q", "swat ok", s.Val)
	}
}

// --- Custom header tests ---

func TestPounceWithHeaders(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	opts := meowrt.NewMap(map[string]meowrt.Value{
		"headers": meowrt.NewMap(map[string]meowrt.Value{
			"Authorization": meowrt.NewString("Bearer token123"),
			"X-Custom":      meowrt.NewString("hello"),
		}),
	})
	result := meowhttp.Pounce(meowrt.NewString(srv.URL+"/echo-header"), opts)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "auth=Bearer token123,custom=hello"
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestSwatWithHeaders(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		w.Write([]byte("auth=" + auth))
	})
	headerSrv := httptest.NewServer(mux)
	defer headerSrv.Close()

	opts := meowrt.NewMap(map[string]meowrt.Value{
		"headers": meowrt.NewMap(map[string]meowrt.Value{
			"Authorization": meowrt.NewString("Bearer swat-token"),
		}),
	})
	result := meowhttp.Swat(meowrt.NewString(headerSrv.URL+"/"), opts)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "auth=Bearer swat-token"
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestProwlWithHeaders(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		w.Write([]byte("auth=" + auth))
	})
	headerSrv := httptest.NewServer(mux)
	defer headerSrv.Close()

	opts := meowrt.NewMap(map[string]meowrt.Value{
		"headers": meowrt.NewMap(map[string]meowrt.Value{
			"Authorization": meowrt.NewString("Bearer prowl-token"),
		}),
	})
	result := meowhttp.Prowl(meowrt.NewString(headerSrv.URL+"/"), opts)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "auth=Bearer prowl-token"
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestTossMapBodyWithHeaders(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		auth := r.Header.Get("Authorization")
		ct := r.Header.Get("Content-Type")
		w.Write([]byte("auth=" + auth + ",ct=" + ct + ",body=" + string(body)))
	})
	headerSrv := httptest.NewServer(mux)
	defer headerSrv.Close()

	m := meowrt.NewMap(map[string]meowrt.Value{
		"name": meowrt.NewString("Tama"),
	})
	opts := meowrt.NewMap(map[string]meowrt.Value{
		"headers": meowrt.NewMap(map[string]meowrt.Value{
			"Authorization": meowrt.NewString("Bearer post-token"),
		}),
	})
	result := meowhttp.Toss(meowrt.NewString(headerSrv.URL+"/"), m, opts)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := `auth=Bearer post-token,ct=application/json,body={"name":"Tama"}`
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

func TestKneadStringBodyWithHeaders(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Knead(
		meowrt.NewString(srv.URL+"/put"),
		meowrt.NewString("raw data"),
		meowrt.NewMap(map[string]meowrt.Value{
			"headers": meowrt.NewMap(map[string]meowrt.Value{
				"Content-Type": meowrt.NewString("text/plain"),
			}),
		}),
	)
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	expected := "knead:text/plain:raw data"
	if s.Val != expected {
		t.Errorf("expected %q, got %q", expected, s.Val)
	}
}

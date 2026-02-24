package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
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

func TestPounceInvalidHost(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	meowhttp.Pounce(meowrt.NewString("http://invalid.host.that.does.not.exist.example:1/path"))
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

func TestToss(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	result := meowhttp.Toss(
		meowrt.NewString(srv.URL+"/post"),
		meowrt.NewString(`{"name":"Tama"}`),
		meowrt.NewString("application/json"),
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
		meowrt.NewString(`{"name":"Mochi"}`),
		meowrt.NewString("application/json"),
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

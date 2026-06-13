package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/135yshr/meow/runtime/meowrt"
)

const defaultMaxBodyBytes int64 = 1 << 20 // 1 MiB
const userAgent = "meow-http-client/2.0"

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type options struct {
	maxBodyBytes int64
	headers      map[string]string
}

func parseOptions(args []meowrt.Value, requiredArgs int) ([]meowrt.Value, options, error) {
	opts := options{maxBodyBytes: defaultMaxBodyBytes}
	if len(args) > requiredArgs {
		if m, ok := args[len(args)-1].(*meowrt.Map); ok {
			o, err := extractOptions(m)
			if err != nil {
				return nil, opts, err
			}
			return args[:len(args)-1], o, nil
		}
	}
	return args, opts, nil
}

// expectString extracts a Go string from a meowrt.Value, returning an error on
// type mismatch instead of panicking.
func expectString(funcName string, v meowrt.Value) (string, error) {
	if v == nil {
		return "", fmt.Errorf("%s expects a String, got <nil>", funcName)
	}
	if f, ok := v.(*meowrt.Furball); ok {
		return "", errors.New(f.Message)
	}
	s, ok := v.(*meowrt.String)
	if !ok {
		return "", fmt.Errorf("%s expects a String, got %s", funcName, v.Type())
	}
	return s.Val, nil
}

// readResponse reads the response body (up to limit bytes) and returns it as a
// meowrt.String, or an error if the body exceeds the limit / I/O fails.
func readResponse(resp *http.Response, limit int64) (meowrt.Value, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return meowrt.NewString(string(body)), nil
}

// extractOptions reads option fields from a Map value.
func extractOptions(m *meowrt.Map) (options, error) {
	opts := options{maxBodyBytes: defaultMaxBodyBytes}
	if v, found := m.Get("maxBodyBytes"); found {
		if n, ok := v.(*meowrt.Int); ok {
			if n.Val <= 0 {
				return opts, fmt.Errorf("maxBodyBytes must be positive, got %d", n.Val)
			}
			opts.maxBodyBytes = n.Val
		}
	}
	if v, found := m.Get("headers"); found {
		if hm, ok := v.(*meowrt.Map); ok {
			opts.headers = make(map[string]string, len(hm.Items))
			for k, val := range hm.Items {
				if s, ok := val.(*meowrt.String); ok {
					opts.headers[k] = s.Val
				}
			}
		}
	}
	return opts, nil
}

// newRequest creates an http.Request with the default User-Agent set.
func newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

// applyHeaders sets custom headers from options on the request.
func applyHeaders(req *http.Request, opts options) {
	for k, v := range opts.headers {
		req.Header.Set(k, v)
	}
}

// furball wraps an error as a Meow Furball value with the "Hiss! ... nya~" form.
func furball(err error) meowrt.Value {
	return &meowrt.Furball{Message: "Hiss! " + err.Error() + ", nya~"}
}

// firstFurball returns the first Furball value among args, or nil.
func firstFurball(args []meowrt.Value) meowrt.Value {
	for _, a := range args {
		if f, ok := a.(*meowrt.Furball); ok {
			return f
		}
	}
	return nil
}

// doWithBody performs an HTTP request with a body (POST/PUT).
// Supported argument patterns:
//
//	(url, mapBody)       → JSON body, Content-Type: application/json
//	(url, strBody)       → string body, no Content-Type
//	(url, mapBody, opts) → JSON body, ct=application/json, headers can override
//	(url, strBody, opts) → string body, Content-Type via headers
func doWithBody(funcName, method string, args []meowrt.Value) meowrt.Value {
	if f := firstFurball(args); f != nil {
		return f
	}
	if len(args) < 2 || len(args) > 3 {
		return furball(fmt.Errorf("%s expects 2-3 arguments, got %d", funcName, len(args)))
	}

	u, err := expectString(funcName, args[0])
	if err != nil {
		return furball(err)
	}

	var body string
	var ct string
	opts := options{maxBodyBytes: defaultMaxBodyBytes}

	switch b := args[1].(type) {
	case *meowrt.Map:
		body = meowrt.ToJSON(b)
		ct = "application/json"
	case *meowrt.String:
		body = b.Val
	default:
		return furball(fmt.Errorf("%s: body must be String or Map, got %s", funcName, args[1].Type()))
	}

	if len(args) == 3 {
		optsMap, ok := args[2].(*meowrt.Map)
		if !ok {
			return furball(fmt.Errorf("%s: 3rd argument must be Map, got %s", funcName, args[2].Type()))
		}
		o, oerr := extractOptions(optsMap)
		if oerr != nil {
			return furball(oerr)
		}
		opts = o
	}

	req, err := newRequest(method, u, strings.NewReader(body))
	if err != nil {
		return furball(err)
	}
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	applyHeaders(req, opts)

	resp, err := client.Do(req)
	if err != nil {
		return furball(err)
	}
	defer resp.Body.Close()
	v, err := readResponse(resp, opts.maxBodyBytes)
	if err != nil {
		return furball(err)
	}
	return v
}

// doSimple handles the common GET/DELETE/OPTIONS pattern: single URL argument
// plus optional headers/options map.
func doSimple(funcName, method string, args []meowrt.Value) meowrt.Value {
	if f := firstFurball(args); f != nil {
		return f
	}
	posArgs, opts, err := parseOptions(args, 1)
	if err != nil {
		return furball(err)
	}
	if len(posArgs) != 1 {
		return furball(fmt.Errorf("%s expects 1 argument (url), got %d", funcName, len(posArgs)))
	}
	u, err := expectString(funcName, posArgs[0])
	if err != nil {
		return furball(err)
	}
	req, err := newRequest(method, u, nil)
	if err != nil {
		return furball(err)
	}
	applyHeaders(req, opts)
	resp, err := client.Do(req)
	if err != nil {
		return furball(err)
	}
	defer resp.Body.Close()
	v, err := readResponse(resp, opts.maxBodyBytes)
	if err != nil {
		return furball(err)
	}
	return v
}

// Pounce performs an HTTP GET request and returns the response body as a String.
func Pounce(args ...meowrt.Value) meowrt.Value {
	return doSimple("pounce", "GET", args)
}

// Toss performs an HTTP POST request.
func Toss(args ...meowrt.Value) meowrt.Value {
	return doWithBody("toss", "POST", args)
}

// Knead performs an HTTP PUT request.
func Knead(args ...meowrt.Value) meowrt.Value {
	return doWithBody("knead", "PUT", args)
}

// Swat performs an HTTP DELETE request and returns the response body as a String.
func Swat(args ...meowrt.Value) meowrt.Value {
	return doSimple("swat", "DELETE", args)
}

// Prowl performs an HTTP OPTIONS request and returns the response body as a String.
func Prowl(args ...meowrt.Value) meowrt.Value {
	return doSimple("prowl", "OPTIONS", args)
}

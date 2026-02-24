package http

import (
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

func parseOptions(args []meowrt.Value, requiredArgs int) ([]meowrt.Value, options) {
	opts := options{maxBodyBytes: defaultMaxBodyBytes}
	if len(args) > requiredArgs {
		if m, ok := args[len(args)-1].(*meowrt.Map); ok {
			opts = extractOptions(m)
			return args[:len(args)-1], opts
		}
	}
	return args, opts
}

// expectString extracts a Go string from a meowrt.Value or panics with a type error.
func expectString(funcName string, v meowrt.Value) string {
	if v == nil {
		panic(fmt.Sprintf("Hiss! %s expects a String, got <nil>, nya~", funcName))
	}
	s, ok := v.(*meowrt.String)
	if !ok {
		panic(fmt.Sprintf("Hiss! %s expects a String, got %s, nya~", funcName, v.Type()))
	}
	return s.Val
}

// readResponse reads the response body (up to limit bytes) and returns it as a meowrt.String.
// It panics if the body exceeds the limit.
func readResponse(resp *http.Response, limit int64) meowrt.Value {
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	if int64(len(body)) > limit {
		panic(fmt.Sprintf("Hiss! response body exceeds %d bytes, nya~", limit))
	}
	return meowrt.NewString(string(body))
}

// extractOptions reads option fields from a Map value.
func extractOptions(m *meowrt.Map) options {
	opts := options{maxBodyBytes: defaultMaxBodyBytes}
	if v, found := m.Get("maxBodyBytes"); found {
		if n, ok := v.(*meowrt.Int); ok {
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
	return opts
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

// doWithBody performs an HTTP request with a body (POST/PUT).
// Supported argument patterns:
//
//	(url, mapBody)       → JSON body, Content-Type: application/json
//	(url, strBody)       → string body, no Content-Type
//	(url, mapBody, opts) → JSON body, ct=application/json, headers can override
//	(url, strBody, opts) → string body, Content-Type via headers
func doWithBody(funcName, method string, args []meowrt.Value) meowrt.Value {
	if len(args) < 2 || len(args) > 3 {
		panic(fmt.Sprintf("Hiss! %s expects 2-3 arguments, got %d, nya~", funcName, len(args)))
	}

	u := expectString(funcName, args[0])

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
		panic(fmt.Sprintf("Hiss! %s: body must be String or Map, got %s, nya~", funcName, args[1].Type()))
	}

	if len(args) == 3 {
		optsMap, ok := args[2].(*meowrt.Map)
		if !ok {
			panic(fmt.Sprintf("Hiss! %s: 3rd argument must be Map, got %s, nya~", funcName, args[2].Type()))
		}
		opts = extractOptions(optsMap)
	}

	req, err := newRequest(method, u, strings.NewReader(body))
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	applyHeaders(req, opts)

	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp, opts.maxBodyBytes)
}

// Pounce performs an HTTP GET request and returns the response body as a String.
func Pounce(args ...meowrt.Value) meowrt.Value {
	posArgs, opts := parseOptions(args, 1)
	if len(posArgs) != 1 {
		panic(fmt.Sprintf("Hiss! pounce expects 1 argument (url), got %d, nya~", len(posArgs)))
	}
	u := expectString("pounce", posArgs[0])
	req, err := newRequest("GET", u, nil)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	applyHeaders(req, opts)
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp, opts.maxBodyBytes)
}

// Toss performs an HTTP POST request.
// Arguments: url, body [, options].
func Toss(args ...meowrt.Value) meowrt.Value {
	return doWithBody("toss", "POST", args)
}

// Knead performs an HTTP PUT request.
// Arguments: url, body [, options].
func Knead(args ...meowrt.Value) meowrt.Value {
	return doWithBody("knead", "PUT", args)
}

// Swat performs an HTTP DELETE request and returns the response body as a String.
func Swat(args ...meowrt.Value) meowrt.Value {
	posArgs, opts := parseOptions(args, 1)
	if len(posArgs) != 1 {
		panic(fmt.Sprintf("Hiss! swat expects 1 argument (url), got %d, nya~", len(posArgs)))
	}
	u := expectString("swat", posArgs[0])
	req, err := newRequest("DELETE", u, nil)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	applyHeaders(req, opts)
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp, opts.maxBodyBytes)
}

// Prowl performs an HTTP OPTIONS request and returns the response body as a String.
func Prowl(args ...meowrt.Value) meowrt.Value {
	posArgs, opts := parseOptions(args, 1)
	if len(posArgs) != 1 {
		panic(fmt.Sprintf("Hiss! prowl expects 1 argument (url), got %d, nya~", len(posArgs)))
	}
	u := expectString("prowl", posArgs[0])
	req, err := newRequest("OPTIONS", u, nil)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	applyHeaders(req, opts)
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp, opts.maxBodyBytes)
}

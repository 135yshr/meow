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

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type options struct {
	maxBodyBytes int64
}

func parseOptions(args []meowrt.Value, requiredArgs int) ([]meowrt.Value, options) {
	opts := options{maxBodyBytes: defaultMaxBodyBytes}
	if len(args) > requiredArgs {
		if m, ok := args[len(args)-1].(*meowrt.Map); ok {
			if v, found := m.Get("maxBodyBytes"); found {
				if n, ok := v.(*meowrt.Int); ok {
					opts.maxBodyBytes = n.Val
				}
			}
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

// doWithBody performs an HTTP request with a body (POST/PUT).
func doWithBody(funcName, method string, args []meowrt.Value) meowrt.Value {
	posArgs, opts := parseOptions(args, 3)
	if len(posArgs) != 3 {
		panic(fmt.Sprintf("Hiss! %s expects 3 arguments (url, body, contentType), got %d, nya~", funcName, len(posArgs)))
	}
	u := expectString(funcName, posArgs[0])
	b := expectString(funcName, posArgs[1])
	ct := expectString(funcName, posArgs[2])

	req, err := http.NewRequest(method, u, strings.NewReader(b))
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	req.Header.Set("Content-Type", ct)

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
	resp, err := client.Get(u)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp, opts.maxBodyBytes)
}

// Toss performs an HTTP POST request.
// Arguments: url, body, contentType [, options].
func Toss(args ...meowrt.Value) meowrt.Value {
	return doWithBody("toss", "POST", args)
}

// Knead performs an HTTP PUT request.
// Arguments: url, body, contentType [, options].
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
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
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
	req, err := http.NewRequest("OPTIONS", u, nil)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp, opts.maxBodyBytes)
}

package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/135yshr/meow/runtime/meowrt"
)

// expectString extracts a Go string from a meowrt.Value or panics with a type error.
func expectString(funcName string, v meowrt.Value) string {
	s, ok := v.(*meowrt.String)
	if !ok {
		panic(fmt.Sprintf("Hiss! %s expects a String, got %s, nya~", funcName, v.Type()))
	}
	return s.Val
}

// readResponse reads the entire response body and returns it as a meowrt.String.
func readResponse(resp *http.Response) meowrt.Value {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	return meowrt.NewString(string(body))
}

// doWithBody performs an HTTP request with a body (POST/PUT).
// It expects exactly 3 arguments: url, body, contentType.
func doWithBody(funcName, method string, args []meowrt.Value) meowrt.Value {
	if len(args) != 3 {
		panic(fmt.Sprintf("Hiss! %s expects 3 arguments (url, body, contentType), got %d, nya~", funcName, len(args)))
	}
	u := expectString(funcName, args[0])
	b := expectString(funcName, args[1])
	ct := expectString(funcName, args[2])

	req, err := http.NewRequest(method, u, strings.NewReader(b))
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp)
}

// Pounce performs an HTTP GET request and returns the response body as a String.
func Pounce(url meowrt.Value) meowrt.Value {
	u := expectString("pounce", url)
	resp, err := http.Get(u)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp)
}

// Toss performs an HTTP POST request.
// Arguments: url, body, contentType.
func Toss(args ...meowrt.Value) meowrt.Value {
	return doWithBody("toss", "POST", args)
}

// Knead performs an HTTP PUT request.
// Arguments: url, body, contentType.
func Knead(args ...meowrt.Value) meowrt.Value {
	return doWithBody("knead", "PUT", args)
}

// Swat performs an HTTP DELETE request and returns the response body as a String.
func Swat(url meowrt.Value) meowrt.Value {
	u := expectString("swat", url)
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp)
}

// Prowl performs an HTTP OPTIONS request and returns the response body as a String.
func Prowl(url meowrt.Value) meowrt.Value {
	u := expectString("prowl", url)
	req, err := http.NewRequest("OPTIONS", u, nil)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer resp.Body.Close()
	return readResponse(resp)
}

package controllers

import (
	"net"
	"rest/models/mysql"
	"rest/models/redis"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

// TestSubstringHandler tests SubstringHandler
func TestSubstringHandler(t *testing.T) {
	r := NewRouter(
		&MyServer{
			db:        &mysql.MySQL{},
			redisConn: &redis.RedisCache{},
		},
	)
	ln := fasthttputil.NewInmemoryListener()
	defer func() {
		_ = ln.Close()
	}()

	s := &fasthttp.Server{
		Handler: r.Handler,
	}
	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://test.com/rest/substr")
	if err := c.Do(req, res); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected %d but got %d", fasthttp.StatusOK, res.StatusCode())
	}
	expectedBody := "To get the longest substring, follow the /find endpoint."
	if body := string(res.Body()); body != (expectedBody) {
		t.Errorf("expected %q but got %q", expectedBody, body)
	}
}

var testTable = []struct {
	number             int
	body               string
	expectedOutput     string
	expectedStatusCode int
	method             string
}{
	{0, `"abcda"`, "abcd", fasthttp.StatusOK, fasthttp.MethodPost},
	{1, `"вдаьц"`, "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{2, `"abcda1"`, "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{3, `""`, "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{4, `"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcde"`, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", fasthttp.StatusOK, fasthttp.MethodPost},
	{5, `"pwwke"`, "wke", fasthttp.StatusOK, fasthttp.MethodPost},
	{6, `"nnnnnnn"`, "n", fasthttp.StatusOK, fasthttp.MethodPost},
	{7, `"a"`, "a", fasthttp.StatusOK, fasthttp.MethodPost},
	{8, `"ab"`, "ab", fasthttp.StatusOK, fasthttp.MethodPost},
	{9, `"0128917"`, "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{10, `"abcda"`, "", fasthttp.StatusMethodNotAllowed, fasthttp.MethodGet},
	{11, `"вдаьц"`, "", fasthttp.StatusMethodNotAllowed, fasthttp.MethodGet},
}

// TestGetSubstring tests GetSubstring
func TestGetSubstring(t *testing.T) {
	r := NewRouter(
		&MyServer{
			db:        &mysql.MySQL{},
			redisConn: &redis.RedisCache{},
		},
	)
	ln := fasthttputil.NewInmemoryListener()
	defer func() {
		_ = ln.Close()
	}()

	s := &fasthttp.Server{
		Handler: r.Handler,
	}
	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.SetRequestURI("http://test.com/rest/substr/find")
	for _, testCase := range testTable {
		switch testCase.method {
		case fasthttp.MethodGet:
			req.Header.SetMethod(fasthttp.MethodGet)

			if err := c.Do(req, res); err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != testCase.expectedStatusCode {
				t.Errorf("for test #%d, expected %d but got %d", testCase.number, testCase.expectedStatusCode, res.StatusCode())
			}
		case fasthttp.MethodPost:
			req.Header.SetMethod(fasthttp.MethodPost)
			req.Header.SetContentType("text/plain")
			req.SetBody([]byte(testCase.body))
			if err := c.Do(req, res); err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != testCase.expectedStatusCode {
				t.Errorf("for test #%d, expected %d but got %d", testCase.number, testCase.expectedStatusCode, res.StatusCode())
			}
			if body, exp := string(res.Body()), testCase.expectedOutput; body != exp {
				t.Errorf("for test #%d, expected %q but got %q", testCase.number, exp, body)
			}
		}

	}

}

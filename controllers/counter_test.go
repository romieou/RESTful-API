package controllers

import (
	"net"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

// TestGetCounter tests GetCounter
func TestGetCounter(t *testing.T) {
	r := NewRouter(
		&MyServer{
			db:        &testDB{},
			redisConn: &testRedis{},
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
	req.SetRequestURI("http://test.com/rest/counter/val")
	req.Header.SetMethod(fasthttp.MethodGet)

	if err := c.Do(req, res); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected %d but got %d", fasthttp.StatusOK, res.StatusCode())
	}
	expBody := "counter value is 0"
	if body := string(res.Body()); body != expBody {
		t.Errorf("expected %q but got %q", expBody, body)
	}
}

var addCounterTests = []struct {
	number             int
	addVal             string
	expectedOutput     string
	expectedStatusCode int
	method             string
}{
	{0, "0", "Success! Counter is now 0", fasthttp.StatusOK, fasthttp.MethodPost},
	{1, "1", "Success! Counter is now 1", fasthttp.StatusOK, fasthttp.MethodPost},
	{2, "2", "Something went wrong. Please try again later.", fasthttp.StatusInternalServerError, fasthttp.MethodPost},
	{3, "7", "Success! Counter is now 7", fasthttp.StatusOK, fasthttp.MethodPost},
	{4, "фвфв", "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{5, "lejew", "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{6, "+9", "Success! Counter is now 9", fasthttp.StatusOK, fasthttp.MethodPost},
	{7, "14", "", fasthttp.StatusMethodNotAllowed, fasthttp.MethodGet},
}

// TestAddCounter tests AddCounter
func TestAddCounter(t *testing.T) {
	r := NewRouter(
		&MyServer{
			db:        &testDB{},
			redisConn: &testRedis{},
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
	req.SetRequestURI("http://test.com/rest/counter/add/")
	for _, testCase := range addCounterTests {
		switch testCase.method {
		case fasthttp.MethodGet:
			req.Header.SetMethod(fasthttp.MethodGet)
			req.SetRequestURI("http://test.com/rest/counter/add/" + testCase.addVal)
			if err := c.Do(req, res); err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != testCase.expectedStatusCode {
				t.Errorf("for test #%d, expected %d but got %d", testCase.number, testCase.expectedStatusCode, res.StatusCode())
			}
		case fasthttp.MethodPost:
			req.Header.SetMethod(fasthttp.MethodPost)
			req.SetRequestURI("http://test.com/rest/counter/add/" + testCase.addVal)
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

var subCounterTests = []struct {
	number             int
	subVal             string
	expectedOutput     string
	expectedStatusCode int
	method             string
}{
	{0, "0", "Success! Counter is now 0", fasthttp.StatusOK, fasthttp.MethodPost},
	{1, "1", "Success! Counter is now 0", fasthttp.StatusOK, fasthttp.MethodPost},
	{2, "3", "Success! Counter is now 2", fasthttp.StatusOK, fasthttp.MethodPost},
	{3, "1234567", "input exceeds counter: counter cannot be negative", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{4, "фвфв", "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{5, "lejew", "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{6, "-9", "Success! Counter is now 9", fasthttp.StatusOK, fasthttp.MethodPost},
	{7, "14", "", fasthttp.StatusMethodNotAllowed, fasthttp.MethodGet},
	{8, "2", "Something went wrong. Please try again later.", fasthttp.StatusInternalServerError, fasthttp.MethodPost},
	{9, "--2", "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
}

// TestSubCounter tests SubCounter
func TestSubCounter(t *testing.T) {
	r := NewRouter(
		&MyServer{
			db:        &testDB{},
			redisConn: &testRedis{},
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
	req.SetRequestURI("http://test.com/rest/counter/sub/")
	for _, testCase := range subCounterTests {
		switch testCase.method {
		case fasthttp.MethodGet:
			req.Header.SetMethod(fasthttp.MethodGet)
			req.SetRequestURI("http://test.com/rest/counter/sub/" + testCase.subVal)
			if err := c.Do(req, res); err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != testCase.expectedStatusCode {
				t.Errorf("for test #%d, expected %d but got %d", testCase.number, testCase.expectedStatusCode, res.StatusCode())
			}
		case fasthttp.MethodPost:
			req.Header.SetMethod(fasthttp.MethodPost)
			req.SetRequestURI("http://test.com/rest/counter/sub/" + testCase.subVal)
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

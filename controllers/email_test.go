package controllers

import (
	"net"
	"rest/models/mysql"
	"rest/models/redis"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var emailTestTable = []struct {
	number             int
	body               string
	expectedOutput     string
	expectedStatusCode int
	method             string
}{
	{0, `"Email:__email@gmail.com\nEmail:__\n__\nram.osp98@gmail.com\n__dog$@krispie.hrEmail:__dog@krispie.hr Email:__________________ram.osp98@krispie.hr\n"`, "email@gmail.com, ram.osp98@gmail.com, dog@krispie.hr, ram.osp98@krispie.hr", fasthttp.StatusOK, fasthttp.MethodPost},
	{1, `"Email:__valid@sss.com"`, "valid@sss.com", fasthttp.StatusOK, fasthttp.MethodPost},
	{2, `"Email:__ывлыв@sss.com"`, "invalid input", fasthttp.StatusNotFound, fasthttp.MethodPost},
	{3, `""`, "invalid input", fasthttp.StatusNotFound, fasthttp.MethodPost},
	{4, `""`, "invalid input", fasthttp.StatusNotFound, fasthttp.MethodPost},
	{5, `"вдаьц"`, "", fasthttp.StatusMethodNotAllowed, fasthttp.MethodGet},
}

// TestEmailHandler tests EmailHandler
func TestEmailHandler(t *testing.T) {
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
	req.SetRequestURI("http://test.com/rest/email")
	if err := c.Do(req, res); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected %d but got %d", fasthttp.StatusOK, res.StatusCode())
	}
	expectedBody := "To parse emails, follow the /check endpoint."
	if body := string(res.Body()); body != (expectedBody) {
		t.Errorf("expected %q but got %q", expectedBody, body)
	}
}

// TestGetEmail tests GetEmail
func TestGetEmail(t *testing.T) {
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
	req.SetRequestURI("http://test.com/rest/email/check")
	for _, testCase := range emailTestTable {
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

var IINTestTable = []struct {
	number             int
	body               string
	expectedOutput     string
	expectedStatusCode int
	method             string
}{
	{0, `"IIN:__980124450084\nIIN:__\n__\n980124450084\n__91891IIN:__111111111111 IIN:__________________98012445008444\n"`, "980124450084 980124450084 ", fasthttp.StatusOK, fasthttp.MethodPost},
	{1, `"IIN:__980124450084"`, "980124450084", fasthttp.StatusOK, fasthttp.MethodPost},
	{2, `"IIN:__ывлыв  IIN:___\n\n90813901824218947"`, "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{3, ``, "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{4, `""`, "invalid input", fasthttp.StatusBadRequest, fasthttp.MethodPost},
	{5, `"вдаьц"`, "", fasthttp.StatusMethodNotAllowed, fasthttp.MethodGet},
}

// TestGetIIN tests GetIIN
func TestGetIIN(t *testing.T) {
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
	req.SetRequestURI("http://test.com/rest/iin/check")
	for _, testCase := range IINTestTable {
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

// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package qthttptest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/juju/qthttptest"
)

// handlerResponse holds the body of a testing handler response.
type handlerResponse struct {
	URL    string
	Method string
	Body   string
	Auth   bool
	Header http.Header
}

func makeHandler(c *qt.C, status int, ctype string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		c.Assert(err, qt.Equals, nil)
		hasAuth := req.Header.Get("Authorization") != ""
		for _, h := range []string{"User-Agent", "Content-Length", "Accept-Encoding", "Authorization"} {
			delete(req.Header, h)
		}
		// Create the response.
		response := handlerResponse{
			URL:    req.URL.String(),
			Method: req.Method,
			Body:   string(body),
			Header: req.Header,
			Auth:   hasAuth,
		}
		// Write the response.
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		enc := json.NewEncoder(w)
		err = enc.Encode(response)
		c.Assert(err, qt.Equals, nil)
	})
}

var assertJSONCallTests = []struct {
	about  string
	params qthttptest.JSONCallParams
}{{
	about: "simple request",
	params: qthttptest.JSONCallParams{
		Method: "GET",
		URL:    "/",
	},
}, {
	about: "method not specified",
	params: qthttptest.JSONCallParams{
		URL: "/",
	},
}, {
	about: "POST request with a body",
	params: qthttptest.JSONCallParams{
		Method: "POST",
		URL:    "/my/url",
		Body:   strings.NewReader("request body"),
	},
}, {
	about: "GET request with custom headers",
	params: qthttptest.JSONCallParams{
		Method: "GET",
		URL:    "/my/url",
		Header: http.Header{
			"Custom1": {"header1", "header2"},
			"Custom2": {"foo"},
		},
	},
}, {
	about: "POST request with a JSON body",
	params: qthttptest.JSONCallParams{
		Method:   "POST",
		URL:      "/my/url",
		JSONBody: map[string]int{"hello": 99},
	},
}, {
	about: "authentication",
	params: qthttptest.JSONCallParams{
		URL:          "/",
		Method:       "PUT",
		Username:     "who",
		Password:     "bad-wolf",
		ExpectStatus: http.StatusOK,
	},
}, {
	about: "test for ExceptHeader in response",
	params: qthttptest.JSONCallParams{
		URL: "/",
		Do: func(req *http.Request) (*http.Response, error) {
			resp, err := http.DefaultClient.Do(req)
			resp.StatusCode = http.StatusOK
			resp.Header["Custom"] = []string{"value1", "value2"}
			resp.Header["Ignored"] = []string{"value3", "value3"}
			return resp, err
		},
		ExpectStatus: http.StatusOK,
		ExpectHeader: http.Header{
			"Custom": {"value1", "value2"},
		},
	},
}, {
	about: "test case insensitive for ExceptHeader in response",
	params: qthttptest.JSONCallParams{
		URL: "/",
		Do: func(req *http.Request) (*http.Response, error) {
			resp, err := http.DefaultClient.Do(req)
			resp.StatusCode = http.StatusOK
			resp.Header["Custom"] = []string{"value1", "value2"}
			resp.Header["Ignored"] = []string{"value3", "value3"}
			return resp, err
		},
		ExpectStatus: http.StatusOK,
		ExpectHeader: http.Header{
			"CUSTOM": {"value1", "value2"},
		},
	},
}, {
	about: "error status",
	params: qthttptest.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusBadRequest,
	},
}, {
	about: "custom Do",
	params: qthttptest.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusTeapot,
		Do: func(req *http.Request) (*http.Response, error) {
			resp, err := http.DefaultClient.Do(req)
			resp.StatusCode = http.StatusTeapot
			return resp, err
		},
	},
}, {
	about: "custom Do with seekable JSON body",
	params: qthttptest.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusTeapot,
		JSONBody:     123,
		Do: func(req *http.Request) (*http.Response, error) {
			r, ok := req.Body.(io.ReadSeeker)
			if !ok {
				return nil, fmt.Errorf("body is not seeker")
			}
			data, err := ioutil.ReadAll(r)
			if err != nil {
				panic(err)
			}
			if string(data) != "123" {
				panic(fmt.Errorf(`unexpected body content, got %q want "123"`, data))
			}
			r.Seek(0, 0)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			resp.StatusCode = http.StatusTeapot
			return resp, err
		},
	},
}, {
	about: "expect error",
	params: qthttptest.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusTeapot,
		Do: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("some error")
		},
		ExpectError: "some error",
	},
}, {
	about: "expect error regexp",
	params: qthttptest.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusTeapot,
		Do: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("some bad error")
		},
		ExpectError: "some .* error",
	},
}}

func TestAssertJSONCall(t *testing.T) {
	c := qt.New(t)
	for _, test := range assertJSONCallTests {
		c.Run(test.about, func(c *qt.C) {
			params := test.params

			// A missing status is assumed to be http.StatusOK.
			status := params.ExpectStatus
			if status == 0 {
				status = http.StatusOK
			}

			// Create the HTTP handler for this test.
			params.Handler = makeHandler(c, status, "application/json")

			// Populate the expected body parameter.
			expectBody := handlerResponse{
				URL:    params.URL,
				Method: params.Method,
				Header: params.Header,
			}

			// A missing method is assumed to be "GET".
			if expectBody.Method == "" {
				expectBody.Method = "GET"
			}
			expectBody.Header = make(http.Header)
			if params.JSONBody != nil {
				expectBody.Header.Set("Content-Type", "application/json")
			}
			for k, v := range params.Header {
				expectBody.Header[k] = v
			}
			if params.JSONBody != nil {
				data, err := json.Marshal(params.JSONBody)
				c.Assert(err, qt.Equals, nil)
				expectBody.Body = string(data)
				params.Body = bytes.NewReader(data)
			} else if params.Body != nil {
				// Handle the request body parameter.
				body, err := ioutil.ReadAll(params.Body)
				c.Assert(err, qt.Equals, nil)
				expectBody.Body = string(body)
				params.Body = bytes.NewReader(body)
			}

			// Handle basic HTTP authentication.
			if params.Username != "" || params.Password != "" {
				expectBody.Auth = true
			}
			params.ExpectBody = expectBody
			qthttptest.AssertJSONCall(c, params)
		})
	}
}

func TestAssertJSONCallWithBodyAsserter(t *testing.T) {
	c := qt.New(t)
	called := false
	params := qthttptest.JSONCallParams{
		URL:     "/",
		Handler: makeHandler(c, http.StatusOK, "application/json"),
		ExpectBody: qthttptest.BodyAsserter(func(c1 *qt.C, body json.RawMessage) {
			c.Assert(c1, qt.Equals, c)
			c.Assert(string(body), qthttptest.JSONEquals, handlerResponse{
				URL:    "/",
				Method: "GET",
				Header: make(http.Header),
			})
			called = true
		}),
	}
	qthttptest.AssertJSONCall(c, params)
	c.Assert(called, qt.Equals, true)
}

func TestAssertJSONCallWithHostedURL(t *testing.T) {
	c := qt.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("%q", "ok "+req.URL.Path)))
	}))
	defer srv.Close()
	qthttptest.AssertJSONCall(c, qthttptest.JSONCallParams{
		URL:        srv.URL + "/foo",
		ExpectBody: "ok /foo",
	})
}

var bodyReaderFuncs = []func(string) io.Reader{
	func(s string) io.Reader {
		return strings.NewReader(s)
	},
	func(s string) io.Reader {
		return bytes.NewBufferString(s)
	},
	func(s string) io.Reader {
		return bytes.NewReader([]byte(s))
	},
}

func TestDoRequestWithInferrableContentLength(t *testing.T) {
	c := qt.New(t)
	text := "hello, world"
	for i, f := range bodyReaderFuncs {
		c.Logf("test %d", i)
		called := false
		qthttptest.DoRequest(c, qthttptest.DoRequestParams{
			Handler: http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
				c.Check(req.ContentLength, qt.Equals, int64(len(text)))
				called = true
			}),
			Body: f(text),
		})
		c.Assert(called, qt.Equals, true)
	}
}

// The TestAssertJSONCall above exercises the testing.AssertJSONCall succeeding
// calls. Failures are already massively tested in practice. DoRequest and
// AssertJSONResponse are also indirectly tested as they are called by
// AssertJSONCall.

func TestTransport(t *testing.T) {
	c := qt.New(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.String()))
	}))
	defer server.Close()
	transport := qthttptest.URLRewritingTransport{
		MatchPrefix: "http://example.com",
		Replace:     server.URL,
	}
	client := http.Client{
		Transport: &transport,
	}
	resp, err := client.Get("http://example.com/path")
	c.Assert(err, qt.Equals, nil)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, qt.Equals, nil)
	resp.Body.Close()
	c.Assert(resp.Request.URL.String(), qt.Equals, "http://example.com/path")
	c.Assert(string(body), qt.Equals, "/path")

	transport.RoundTripper = &http.Transport{}
	resp, err = client.Get(server.URL + "/otherpath")
	c.Assert(err, qt.Equals, nil)
	body, err = ioutil.ReadAll(resp.Body)
	c.Assert(err, qt.Equals, nil)
	resp.Body.Close()
	c.Assert(resp.Request.URL.String(), qt.Equals, server.URL+"/otherpath")
	c.Assert(string(body), qt.Equals, "/otherpath")
}

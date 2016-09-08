// Package srest contains tools for sites and web services creation.
/*
	RESTfuler interface:
		Create(w http.ResponseWriter, r *http.Request)
		One(w http.ResponseWriter, r *http.Request)
		List(w http.ResponseWriter, r *http.Request)
		Update(w http.ResponseWriter, r *http.Request)
		Delete(w http.ResponseWriter, r *http.Request)

	Modeler interface:
		IsValid() error
*/
// The MIT License (MIT)
//
// Copyright (c) 2016 Angel Del Castillo
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package srest

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"syscall"
	"testing"
)

// API satisfies RESTfuler interface
type API struct {
	T *testing.T
}

// Create test
func (a *API) Create(w http.ResponseWriter, r *http.Request) {}

// One test
func (a *API) One(w http.ResponseWriter, r *http.Request) {}

// List test
func (a *API) List(w http.ResponseWriter, r *http.Request) {}

// Update test
func (a *API) Update(w http.ResponseWriter, r *http.Request) {}

// Delete test
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {}

// Model struct satisfies Modeler interface
type Model struct {
	Name string `schema:"name"`
}

// IsValid modeler interface
func (m *Model) IsValid() error {
	return nil
}

// Modelfail struct
type Modelfail struct {
	Name string `schema:"name"`
}

// IsValid modeler interface
func (m *Modelfail) IsValid() error {
	return errors.New("this must fail")
}

func TestBind(t *testing.T) {
	p := url.Values{}
	p.Add("name", "x")
	var x Model
	err := Bind(p, &x)
	if err != nil {
		t.Logf("bind err [%s]", err)
		t.Fail()
	}
}

func TestBindFail(t *testing.T) {
	p := url.Values{}
	p.Add("name", "x")
	var x struct {
		Name string `schema:"name"`
	}
	err := Bind(p, &x)
	if err == nil {
		t.Logf("bind : err [%s]", err)
		t.Fail()
	}
}

func TestModelerFail(t *testing.T) {
	p := url.Values{}
	p.Add("name", "x")
	var x Modelfail
	err := Bind(p, &x)
	if err == nil {
		t.Logf("bind : err [%s]", err)
		t.Fail()
	}
}

func TestBindDecoderFail(t *testing.T) {
	p := url.Values{}
	var x Modelfail
	err := Bind(p, x)
	if err == nil {
		t.Logf("bind : err [%s]", err)
		t.Fail()
	}
}

func sampleMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// DO NOTHING
		h.ServeHTTP(w, r)
	})
}

func TestServer(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("panics : err [%s]", err)
		}
	}()

	m := New(nil)
	m.Get("/static", Static("/static", "mydir"))
	m.Use("/v1/api/friends", &API{t})
	m.Use("/v1/api/others", &API{t}, sampleMW)
	c := m.Run(9999)
	go func() {
		c <- syscall.SIGTERM
	}()
}

type TM struct {
	Input              Input
	ExpectedBody       string
	ExpectedStatusCode int
}

type Input struct {
	Options *Options
	Port    int
	Handler func(http.ResponseWriter, *http.Request)
	MW      []func(http.Handler) http.Handler
}

func TestMiddlewareTable(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("recover : err [%s]", err)
			t.Fail()
		}
	}()
	table := []TM{
		TM{
			Input: Input{
				Options: &Options{
					UseTLS: true,
				},
				Handler: func(w http.ResponseWriter, r *http.Request) {
					JSON(w, "x")
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							h.ServeHTTP(w, r)
						})
					},
				},
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `"x"`,
		},
		TM{
			Input: Input{
				Options: &Options{},
				Handler: func(w http.ResponseWriter, r *http.Request) {
					JSON(w, true)
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.Write([]byte("one"))
							h.ServeHTTP(w, r)
						})
					},
				},
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `onetrue`,
		},
		TM{
			Input: Input{
				Options: nil,
				Handler: func(w http.ResponseWriter, r *http.Request) {
					JSON(w, true)
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.Write([]byte("one"))
							h.ServeHTTP(w, r)
						})
					},
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.Write([]byte("two"))
							h.ServeHTTP(w, r)
						})
					},
				},
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `onetwotrue`,
		},
		TM{
			Input: Input{
				Options: &Options{},
				Handler: func(w http.ResponseWriter, r *http.Request) {
					JSON(w, 1)
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							// skip
							return
							h.ServeHTTP(w, r)
							w.Write([]byte("one"))
						})
					},
				},
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       ``,
		},
		TM{
			Input: Input{
				Options: &Options{},
				Handler: func(w http.ResponseWriter, r *http.Request) {
					JSON(w, 2)
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							// skip
							w.WriteHeader(http.StatusBadRequest)
							return
							h.ServeHTTP(w, r)
							w.Write([]byte("one"))
						})
					},
				},
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedBody:       ``,
		},
	}
	for i := range table {
		x := table[i]

		n := New(x.Input.Options)
		n.Get("/", http.HandlerFunc(x.Input.Handler), x.Input.MW...)
		ts := httptest.NewServer(n.Mux)
		defer ts.Close()

		res, err := http.Get(ts.URL)
		if err != nil {
			t.Logf("get : err [%s]", err)
			t.Fail()
			continue
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Logf("get : err [%s]", err)
			t.Fail()
		}

		if res.StatusCode != x.ExpectedStatusCode {
			t.Logf("expected [%v] actual [%v]", x.ExpectedStatusCode, res.StatusCode)
			t.Fail()
		}

		actual := string(body)
		if len(body) > 0 {
			actual = actual[:len(actual)-1]
		}
		if actual != x.ExpectedBody {
			t.Logf("expected [%s] actual [%s]", x.ExpectedBody, actual)
			t.Fail()
		}

		c := n.Run(x.Input.Port)
		go func() {
			c <- syscall.SIGTERM
		}()
	}
}

func TestLoadViews(t *testing.T) {
	tmplInited = false
	err := LoadViews("mock2fail", map[string]interface{}{})
	if err != nil {
		t.Logf("LoadViews : err [%s]", err)
		t.Fail()
	}
}

func TestLoadViewsFail(t *testing.T) {
	tmplInited = false
	err := LoadViews("mock2fail", map[string]interface{}{})
	if err != nil {
		t.Logf("LoadViews : err [%s]", err)
		t.Fail()
	}
}

func TestRender(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Logf("get pwd : err [%s]", err)
		t.Fail()
	}

	tmplInited = false
	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		t.Logf("LoadViews : err [%s]", err)
		t.Fail()
		return
	}

	w := httptest.NewRecorder()
	// mock/index.html file must exists o this will panic
	// index.html content will be:
	// {{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	if err != nil {
		t.Logf("Render : err [%s]", err)
		t.Fail()
		return
	}

	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Logf("read body : err [%s]", err)
		t.Fail()
		return
	}

	expected := []byte("I am lowercase-eqs:true")
	// remove additional \r
	actual = actual[:len(actual)-1]
	if string(actual) != string(expected) {
		t.Logf("expected [%s] actual [%s]", string(expected), string(actual))
		t.Fail()
		return
	}
}

func TestRenderFail(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Logf("get pwd : err [%s]", err)
		t.Fail()
	}

	tmplInited = false
	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		t.Logf("LoadViews : err [%s]", err)
		t.Fail()
		return
	}

	w := httptest.NewRecorder()
	// mock/index.html file must exists o this will panic
	// index.html content will be: {{cap "i am lowercase"}}
	err = Render(w, "notfound.html", map[string]interface{}{"x": 1})
	if err != ErrTemplateNotFound {
		t.Logf("Render : err [%s]", err)
		t.Fail()
		return
	}

	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Logf("read body : err [%s]", err)
		t.Fail()
		return
	}
	expected := []byte("template not found")
	if string(actual) != string(expected) {
		t.Logf("expected [%s] actual [%s]", string(expected), string(actual))
		t.Fail()
		return
	}
}

func TestRenderDebug(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Logf("get pwd : err [%s]", err)
		t.Fail()
	}

	tmplInited = false
	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		t.Logf("LoadViews : err [%s]", err)
		t.Fail()
		return
	}
	Debug(true)

	w := httptest.NewRecorder()
	// mock/index.html file must exists o this will panic
	// index.html content will be:
	// {{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	if err != nil {
		t.Logf("Render : err [%s]", err)
		t.Fail()
		return
	}

	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Logf("read body : err [%s]", err)
		t.Fail()
		return
	}

	expected := []byte("I am lowercase-eqs:true")
	// remove additional \r
	actual = actual[:len(actual)-1]
	if string(actual) != string(expected) {
		t.Logf("expected [%s] actual [%s]", string(expected), string(actual))
		t.Fail()
		return
	}
}

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := JSON(w, `this is string`)
	if err != nil {
		t.Logf("JSON : err [%s]", err)
		t.Fail()
		return
	}
	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Logf("read body : err [%s]", err)
		t.Fail()
		return
	}
	expected := []byte(`"this is string"`)
	actual = actual[:len(actual)-1]
	if string(actual) != string(expected) {
		t.Logf("expected [%v] actual [%v]", string(expected), string(actual))
		t.Fail()
		return
	}
}

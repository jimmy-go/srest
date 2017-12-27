// Package srest contains tools for sites and web services creation.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

//	RESTfuler interface:
//		List(w http.ResponseWriter, r *http.Request)
//		One(w http.ResponseWriter, r *http.Request)
//		Create(w http.ResponseWriter, r *http.Request)
//		Update(w http.ResponseWriter, r *http.Request)
//		Delete(w http.ResponseWriter, r *http.Request)
//
//	Modeler interface:
//		IsValid() error

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("get pwd : err [%s]", err)
		return
	}

	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		log.Printf("LoadViews : err [%s]", err)
		return
	}

	v := m.Run()
	// TODO; verify counter, go1.4.2 reports 22, go1.5.3 reports 30
	gos := runtime.NumGoroutine()
	if gos > 50 {
		log.Printf("goroutines [%v]", gos)
		debug.PrintStack()
		panic("blocked goroutines")
	}

	os.Exit(v)
}

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
		t.Errorf("err [%s]", err)
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
		t.Errorf("err [%s]", err)
	}
}

func TestModelerFail(t *testing.T) {
	p := url.Values{}
	p.Add("name", "x")
	var x Modelfail
	err := Bind(p, &x)
	if err == nil {
		t.Errorf("err [%s]", err)
	}
}

func TestBindDecoderFail(t *testing.T) {
	p := url.Values{}
	var x Modelfail
	err := Bind(p, x)
	if err == nil {
		t.Errorf("err [%s]", err)
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
	Name               string
	ExpectedError      error
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
			t.Errorf("err [%s]", err)
		}
	}()
	table := []TM{
		TM{
			Input: Input{
				Options: &Options{
					UseTLS: true,
				},
				Handler: func(w http.ResponseWriter, r *http.Request) {
					err := JSON(w, "x")
					if err != nil {
						log.Printf("mw : err [%s]", err)
					}
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
					err := JSON(w, true)
					if err != nil {
						log.Printf("mw : err [%s]", err)
					}

				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							_, err := w.Write([]byte("one"))
							if err != nil {
								log.Printf("mw : err [%s]", err)
							}
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
					err := JSON(w, true)
					if err != nil {
						log.Printf("err [%s]", err)
					}
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							_, err := w.Write([]byte("one"))
							if err != nil {
								log.Printf("mw : err [%s]", err)
							}
							h.ServeHTTP(w, r)
						})
					},
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							_, err := w.Write([]byte("two"))
							if err != nil {
								log.Printf("mw : err [%s]", err)
							}
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
					err := JSON(w, 1)
					if err != nil {
						log.Printf("mw : err [%s]", err)
					}
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							if "" == " " {
								h.ServeHTTP(w, r)
								_, err := w.Write([]byte("one"))
								if err != nil {
									log.Printf("mw : err [%s]", err)
								}
							}

							// skip
							return
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
					err := JSON(w, 2)
					if err != nil {
						log.Printf("mw : err [%s]", err)
					}
				},
				MW: []func(http.Handler) http.Handler{
					func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							if "" == " " {
								h.ServeHTTP(w, r)
								_, err := w.Write([]byte("one"))
								if err != nil {
									log.Printf("mw : err [%s]", err)
								}
							}

							// skip
							w.WriteHeader(http.StatusBadRequest)
							return
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
			t.Errorf("get : err [%s]", err)
			continue
		}
		defer func() {
			err := res.Body.Close()
			if err != nil {
				log.Printf("close file err [%s]", err)
			}
		}()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("get : err [%s]", err)
		}

		if res.StatusCode != x.ExpectedStatusCode {
			t.Errorf("expected [%v] actual [%v]", x.ExpectedStatusCode, res.StatusCode)
		}

		actual := string(body)
		if len(body) > 0 {
			actual = actual[:len(actual)-1]
		}
		if actual != x.ExpectedBody {
			t.Errorf("expected [%s] actual [%s]", x.ExpectedBody, actual)
		}

		c := n.Run(x.Input.Port)
		go func() {
			c <- syscall.SIGTERM
		}()
	}
}

func TestLoadViews(t *testing.T) {
	dir, err := os.Getwd()
	assert.Nil(t, err)

	err = LoadViews(dir+"/mock", DefaultFuncMap)
	assert.Nil(t, err)
}

func TestLoadViewsFail(t *testing.T) {
	err := LoadViews("mock2fail", map[string]interface{}{})
	assert.NotNil(t, err)
	assert.EqualValues(t, "lstat mock2fail: no such file or directory", fmt.Sprintf("%s", err))
}

func TestRender(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("get pwd : err [%s]", err)
	}

	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		t.Errorf("LoadViews : err [%s]", err)
		return
	}

	w := httptest.NewRecorder()
	// mock/index.html file must exists o this will panic
	// index.html content will be:
	// {{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	if err != nil {
		t.Errorf("Render : err [%s]", err)
		return
	}

	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf("read body : err [%s]", err)
		return
	}

	expected := []byte("I am lowercase-eqs:true")
	if string(actual) != string(expected) {
		t.Errorf("expected [%s] actual [%s]", string(expected), string(actual))
		return
	}
}

func TestRenderFail(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("get pwd : err [%s]", err)
	}

	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		t.Errorf("LoadViews : err [%s]", err)
		return
	}

	w := httptest.NewRecorder()
	// mock/index.html file must exists o this will panic
	// index.html content will be: {{cap "i am lowercase"}}
	err = Render(w, "notfound.html", map[string]interface{}{"x": 1})
	if err != ErrTemplateNotFound {
		t.Errorf("Render : err [%s]", err)
		return
	}

	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf("read body : err [%s]", err)
		return
	}
	actual := b[:len(b)-1]
	expected := []byte("template view not found")
	if string(actual) != string(expected) {
		t.Errorf("expected [%s] actual [%s]", string(expected), string(actual))
		return
	}
}

func TestRenderDebug(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("get pwd : err [%s]", err)
	}

	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		t.Errorf("LoadViews : err [%s]", err)
		return
	}
	Debug(true)

	w := httptest.NewRecorder()
	// mock/index.html file must exists o this will panic
	// index.html content will be:
	// {{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	if err != nil {
		t.Errorf("Render : err [%s]", err)
		return
	}

	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf("read body : err [%s]", err)
		return
	}

	expected := []byte("I am lowercase-eqs:true")
	if string(actual) != string(expected) {
		t.Errorf("expected [%s] actual [%s]", string(expected), string(actual))
		return
	}
}

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := JSON(w, `this is string`)
	if err != nil {
		t.Errorf("err [%s]", err)
		return
	}
	actual := w.Body.String()
	expected := []byte(`"this is string"`)
	actual = actual[:len(actual)-1]
	if string(actual) != string(expected) {
		t.Errorf("expected [%v] actual [%v]", string(expected), string(actual))
	}
}

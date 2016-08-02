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
		t.Logf("TestBind : bind err [%s]", err)
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
		t.Logf("TestBindFail : bind : err [%s]", err)
		t.Fail()
	}
}

func TestModeler(t *testing.T) {
}

func TestModelerFail(t *testing.T) {
	p := url.Values{}
	p.Add("name", "x")
	var x Modelfail
	err := Bind(p, &x)
	if err == nil {
		t.Logf("TestModelerFail : bind : err [%s]", err)
		t.Fail()
	}
}

func TestBindDecoder(t *testing.T) {
}

func TestBindDecoderFail(t *testing.T) {
	// Bind decoder fail
	p := url.Values{}
	var x Modelfail
	err := Bind(p, x)
	if err == nil {
		t.Logf("TestBindDecoderFail : bind : err [%s]", err)
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
			t.Logf("TestServer : panics : err [%s]", err)
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

func TestMiddleware(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("TestMiddleware : panics : err [%s]", err)
		}
	}()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSON(w, true)
	})
	MidOne := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}
	MidTwo := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}

	// TODO; test middleware behavior is working.
	m := New(nil)
	m.Get("/homei/", handler, MidOne, MidTwo)
	m.Post("/homeii/", handler, MidOne, MidTwo)
	m.Put("/homeiii/", handler, MidOne, MidTwo)
	m.Del("/homeiv/", handler, MidOne, MidTwo)
	c := m.Run(9998)
	go func() {
		c <- syscall.SIGTERM
	}()
}

func TestRenderFail(t *testing.T) {
	w := httptest.NewRecorder()
	err := Render(w, "none", true)
	if err == nil {
		t.Logf("TestRenderFail : render : err [%s]", err)
		t.Fail()
	}
}

func TestLoadViews(t *testing.T) {
	tmplInited = false
	err := LoadViews("mock2fail", map[string]interface{}{})
	if err != nil {
		t.Logf("TestLoadViews : LoadViews : err [%s]", err)
		t.Fail()
	}
}

func TestLoadViewsFail(t *testing.T) {
	tmplInited = false
	err := LoadViews("mock2fail", map[string]interface{}{})
	if err != nil {
		t.Logf("TestLoadViewsFail : LoadViews : err [%s]", err)
		t.Fail()
	}
}

func TestRender(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Logf("TestRender : get pwd : err [%s]", err)
		t.Fail()
	}

	tmplInited = false
	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		t.Logf("TestRender : LoadViews : err [%s]", err)
		t.Fail()
	}

	w := httptest.NewRecorder()
	// mock/index.html file must exists o this will panic
	// index.html content will be: {{cap "i am lowercase"}}
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	if err != nil {
		t.Logf("TestRender : Render : err [%s]", err)
		t.Fail()
	}

	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Logf("TestRender : read body : err [%s]", err)
		t.Fail()
	}

	expected := []byte("I am lowercase")
	// take first 14 chars because readAll adds and aditional \r
	if string(actual[:13]) != string(expected[:13]) {
		t.Logf("TestRender : expected [%s] actual [%s]", string(expected), string(actual))
		t.Fail()
	}
}

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := JSON(w, `this is string`)
	if err != nil {
		t.Logf("TestJSON : JSON : err [%s]", err)
		t.Fail()
	}
	actual, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Logf("TestJSON : read body : err [%s]", err)
		t.Fail()
	}
	expected := []byte(`"this is string"`)
	if string(actual[:len(actual)-1]) != string(expected) {
		t.Logf("TestJSON : expected [%v] actual [%v]", string(expected), string(actual[:len(actual)-1]))
		t.Fail()
	}
}

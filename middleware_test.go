package srest

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Input struct {
	Options *Options
	Port    int
	Handler func(http.ResponseWriter, *http.Request)
	MW      []func(http.Handler) http.Handler
}

func TestMiddlewareHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = JSON(w, true)
	})
	mid1 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("mid1", "middleware one")
			h.ServeHTTP(w, r)
		})
	}

	m := New(nil)
	m.Get("/", handler, mid1)
	err := m.registerHandlers()
	assert.Nil(t, err)
	ts := httptest.NewServer(m.Mux)

	res, err := http.Get(ts.URL)
	uerr, ok := err.(*url.Error)
	if ok {
		log.Printf("err [%s]", uerr)
	}
	assert.Nil(t, err)
	assert.EqualValues(t, 200, res.StatusCode)
	hmsg := res.Header.Get("mid1")
	assert.EqualValues(t, `middleware one`, hmsg)

	b, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = res.Body.Close()
	assert.Nil(t, err)

	var actual string
	if len(b) > 0 {
		actual = string(b[:len(b)-1])
	}
	assert.EqualValues(t, `true`, actual)
	ts.Close()
}

func TestMiddlewareHandlerFail(t *testing.T) {
	defer func() {
		err := recover()
		assert.EqualValues(t, "Run : register handlers : err [method not found: ]", err)
	}()
	m := New(nil)
	m.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m.Get("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m.handlers = append(m.handlers, tmpHandler{})
	err := m.registerHandlers()
	assert.NotNil(t, err)
	m.Run(9001)
}

func TestMiddleware(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("err [%s]", err)
		}
	}()
	table := []struct {
		Purpose string
		Input   Input
		ExpBody string
		Code    int
	}{
		{
			"1. OK",
			Input{
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
			`"x"`,
			http.StatusOK,
		},
		{
			"2. OK",
			Input{
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
			`onetrue`,
			http.StatusOK,
		},
		{
			"3. OK",
			Input{
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
			`onetwotrue`,
			http.StatusOK,
		},
		{
			"4. OK",
			Input{
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
			``,
			http.StatusOK,
		},
		{
			"5. Fail",
			Input{
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
			``,
			http.StatusBadRequest,
		},
	}
	for i := range table {
		x := table[i]

		m := New(x.Input.Options)
		m.Get("/", http.HandlerFunc(x.Input.Handler), x.Input.MW...)
		err := m.registerHandlers()
		assert.Nil(t, err, x.Purpose)
		ts := httptest.NewServer(m.Mux)
		defer ts.Close()

		res, err := http.Get(ts.URL)
		uerr, ok := err.(*url.Error)
		if ok {
			log.Printf("[%s] err [%s]", x.Purpose, uerr)
		}
		assert.Nil(t, err, x.Purpose)
		assert.EqualValues(t, x.Code, res.StatusCode, x.Purpose)

		b, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, x.Purpose)
		err = res.Body.Close()
		assert.Nil(t, err, x.Purpose)

		var actual string
		if len(b) > 0 {
			actual = string(b[:len(b)-1])
		}
		assert.EqualValues(t, x.ExpBody, actual, x.Purpose)
	}
}

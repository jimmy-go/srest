package srest

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
)

type Input struct {
	Options *Options
	Port    int
	Handler func(http.ResponseWriter, *http.Request)
	MW      []func(http.Handler) http.Handler
}

func TestMiddleware(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("err [%s]", err)
		}
	}()
	table := []struct {
		Purpose            string
		Input              Input
		ExpectedBody       string
		ExpectedStatusCode int
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

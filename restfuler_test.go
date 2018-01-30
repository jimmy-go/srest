package srest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// API satisfies RESTfuler interface
type API struct{}

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

func sampleMid(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func TestRESTFuler(t *testing.T) {
	defer func() {
		err := recover()
		assert.EqualValues(t, nil, err)
	}()
	m := New(nil)
	m.Get("/static", Static("/static", "mydir"))
	m.Use("/v1/api/friends", &API{})
	m.Use("/v1/api/others", &API{}, sampleMid)
}

func TestRESTDuplicated(t *testing.T) {
	table := []struct {
		Purpose, Method, URI string
	}{
		{"1. OK", "GET", "/hello"},
		{"2. OK", "POST", "/hello"},
		{"3. OK", "PUT", "/hello"},
		{"4. OK", "DELETE", "/hello"},
	}
	for _, x := range table {
		func(method, uri string) {
			defer func() {
				err := recover()
				assert.EqualValues(t, fmt.Sprintf("duplicated definition: %s %s", method, uri), err)
			}()
			m := New(nil)
			switch method {
			case "GET":
				m.Get(uri, http.HandlerFunc(helloHandler))
				m.Get(uri, http.HandlerFunc(helloHandler))
			case "POST":
				m.Post(uri, http.HandlerFunc(helloHandler))
				m.Post(uri, http.HandlerFunc(helloHandler))
			case "PUT":
				m.Put(uri, http.HandlerFunc(helloHandler))
				m.Put(uri, http.HandlerFunc(helloHandler))
			case "DELETE":
				m.Del(uri, http.HandlerFunc(helloHandler))
				m.Del(uri, http.HandlerFunc(helloHandler))
			default:
				t.Fail()
			}
			m.Run(9002)
		}(x.Method, x.URI)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {}

func TestRESTDuplicatedVars(t *testing.T) {
	defer func() {
		err := recover()
		assert.EqualValues(t, "duplicated definition: GET /me/:c/name", err)
	}()
	m := New(nil)
	m.Get("/me/:id/name", http.HandlerFunc(helloHandler))
	m.Get("/me/:c/name", http.HandlerFunc(helloHandler))
}

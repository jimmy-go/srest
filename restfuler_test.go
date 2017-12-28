package srest

import (
	"net/http"
	"testing"
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
	m := New(nil)
	m.Get("/static", Static("/static", "mydir"))
	m.Use("/v1/api/friends", &API{})
	m.Use("/v1/api/others", &API{}, sampleMid)
}

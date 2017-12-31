package srest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := JSON(w, `this is string`)
	assert.Nil(t, err)

	s := w.Body.String()

	expected := []byte(`"this is string"`)
	actual := s[:len(s)-1]
	assert.EqualValues(t, expected, actual)
}

func TestRemoveVars(t *testing.T) {
	table := []struct {
		Purpose, S, X string
	}{
		{"1. OK", "/a/:b/c", "/a/*/c"},
	}
	for _, x := range table {
		actual := removeVars(x.S)
		assert.EqualValues(t, x.X, actual, x.Purpose)
	}
}

func TestRegisterHandler(t *testing.T) {
	m := New(nil)
	table := []struct {
		Purpose string
		Error   error
		HS      []tmpHandler
	}{
		{
			"1. OK: GET, POST, PUT, DELETE",
			nil,
			[]tmpHandler{
				tmpHandler{
					"GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = fmt.Fprintln(w, "active")
					}),
				},
				tmpHandler{
					"POST", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = fmt.Fprintln(w, "active")
					}),
				},
				tmpHandler{
					"PUT", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = fmt.Fprintln(w, "active")
					}),
				},
				tmpHandler{
					"DELETE", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = fmt.Fprintln(w, "active")
					}),
				},
			},
		},
		{
			"2. FAIL: Not found",
			errors.New("method not found: NOT"),
			[]tmpHandler{
				tmpHandler{
					"NOT", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = fmt.Fprintln(w, "active")
					}),
				},
			},
		},
	}
	for _, x := range table {
		err := registerHandlers(m.Mux, x.HS)
		assert.EqualValues(t, x.Error, err, x.Purpose)
	}
}

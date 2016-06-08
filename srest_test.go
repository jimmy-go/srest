package srest

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"testing"
)

// API test struct
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

// Model struct
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

func TestNew(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// Bind
	{
		p := url.Values{}
		p.Add("name", "x")
		var x Model
		err := Bind(p, &x)
		if err != nil {
			log.Printf("Bind err [%s]", err)
			t.Fail()
		}
	}
	// Bind fail
	{
		p := url.Values{}
		p.Add("name", "x")
		var x struct {
			Name string `schema:"name"`
		}
		err := Bind(p, &x)
		if err == nil {
			log.Printf("Bind err [%s]", err)
			t.Fail()
		}
	}
	// IsValid fail
	{
		p := url.Values{}
		p.Add("name", "x")
		var x Modelfail
		err := Bind(p, &x)
		if err == nil {
			log.Printf("Bind err [%s]", err)
			t.Fail()
		}
	}
	// Bind decoder fail
	{
		p := url.Values{}
		var x Modelfail
		err := Bind(p, x)
		if err == nil {
			log.Printf("Bind err [%s]", err)
			t.Fail()
		}
	}
	// Server
	{
		m := New(nil)
		m.Use("/v1/api/friends", &API{t})
		m.Static("/static", "static")
		m.Run(-1)
	}
	// Render Fail
	{
		var w http.ResponseWriter
		err := Render(w, "none", true)
		if err == nil {
			log.Printf("Render err [%s]", err)
			t.Fail()
		}
	}
	// LoadViews Fail
	{
		err := LoadViews("fail")
		if err != nil {
			log.Printf("LoadViews err [%s]", err)
			t.Fail()
		}
	}
}

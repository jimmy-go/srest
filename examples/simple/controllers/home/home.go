package home

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jimmy-go/srest"
)

// API struct
type API struct{}

// Create func
func (a *API) Create(w http.ResponseWriter, r *http.Request) {}

// One func
func (a *API) One(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "a" {
		v := map[string]interface{}{"Title": "One A", "Message": "hola soy template"}
		srest.Render(w, "home/show.html", v)
		return
	}
	v := map[string]interface{}{"Title": "One Normal", "Message": "hola soy template"}
	srest.Render(w, "home/edit.html", v)
}

// List func
func (a *API) List(w http.ResponseWriter, r *http.Request) {
	v := map[string]interface{}{"Title": "List", "Message": "hola soy template"}
	srest.Render(w, "home/index.html", v)
}

// Update func
func (a *API) Update(w http.ResponseWriter, r *http.Request) {}

// Delete func
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {}

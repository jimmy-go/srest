package users

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/jimmy-go/srest"
	"github.com/jimmy-go/srest/examples/simple/dai"
)

// User model
type User struct {
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
}

// IsValid satisfies modeler interface.
func (u *User) IsValid() bool {
	// TODO
	return true
}

// API satisfies RESTfuler interface
type API struct{}

// Create func
func (a *API) Create(w http.ResponseWriter, r *http.Request) {
	var m *User
	err := srest.Bind(r, &m)
	if err != nil {
		srest.JSON(w, err)
		return
	}

	var id string
	err = dai.Db.Get(&id, "INSERT INTO users (name, email) VALUES($1, $2) RETURNING id", m)
	if err != nil {
		srest.JSON(w, err)
		return
	}

	srest.JSON(w, true)
}

// One func
func (a *API) One(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	id := p["id"]
	log.Printf("id [%v]", id)
	if id == "1" {
		err := errors.New("OK BAD")
		srest.JSON(w, &E{Error: err.Error()})
		return
	}

	var u User
	err := dai.Db.Get(&u, "SELECT name, email FROM users WHERE id=$1", id)
	if err != nil {
		srest.JSON(w, &E{Error: err.Error()})
		return
	}

	srest.JSON(w, &Result{u})
}

// List func
func (a *API) List(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, true)
}

// Update func
func (a *API) Update(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, true)
}

// Delete func
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, true)
}

// E error api struct
type E struct {
	Error string `json:"error"`
}

// Result generic response
type Result struct {
	Response interface{} `json:"result"`
}

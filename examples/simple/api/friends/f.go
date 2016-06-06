package friends

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/jimmy-go/srest"
	"github.com/jimmy-go/srest/examples/simple/dai"
)

// Friend model
type Friend struct {
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
}

// IsValid satisfies modeler interface.
func (u *Friend) IsValid() bool {
	// TODO
	return true
}

// API struct
type API struct{}

// New inits configuration
// copy session or reuse database connection, logic belongs to user.
func New(s string) *API {
	return &API{}
}

// Create func
func (a *API) Create(w http.ResponseWriter, r *http.Request) {
	var m *Friend
	err := srest.Bind(r, &m)
	if err != nil {
		srest.JSON(w, err)
		return
	}

	var id string
	err = dai.Db.Get(&id, "INSERT INTO users (name, email) VALUES($1, $2) RETURNING id", m)
	if err != nil {
		srest.JSON(w, &E{Error: err.Error()})
		return
	}

	srest.JSON(w, "item created: "+id)
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

	var u Friend
	err := dai.Db.Get(&u, "SELECT name, email FROM users WHERE id=$1", id)
	if err != nil {
		srest.JSON(w, &E{Error: err.Error()})
		return
	}

	srest.JSON(w, &Result{u})
}

// List func
func (a *API) List(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, &Result{"friends list"})
}

// Update func
func (a *API) Update(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, &Result{"friends update"})
}

// Delete func
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, &Result{"friends delete"})
}

// E error api struct
type E struct {
	Error string `json:"error"`
}

// Result generic response
type Result struct {
	Response interface{} `json:"result"`
}

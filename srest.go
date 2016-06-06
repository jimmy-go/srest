package srest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
)

var (
	errModeler = errors.New("modeler interface not found")
)

// Options struct
type Options struct {
}

// Multi struct.
type Multi struct {
	Mux  *mux.Router
	Done chan struct{}
}

// New returns a new server.
func New(opts *Options) *Multi {
	m := &Multi{
		Mux:  mux.NewRouter(),
		Done: make(chan struct{}, 1),
	}
	return m
}

// Static dir.
func (m *Multi) Static(dir string) {
	m.Mux.Handle("/", http.FileServer(http.Dir("static")))
}

// Use adds a module.
func (m *Multi) Use(path string, n RESTfuler) {
	m.Mux.HandleFunc(path, n.Create).Methods("POST")
	m.Mux.HandleFunc(path+"/{id}", n.One).Methods("GET")
	m.Mux.HandleFunc(path, n.List).Methods("GET")
	m.Mux.HandleFunc(path, n.Update).Methods("PUT")
	m.Mux.HandleFunc(path, n.Delete).Methods("DELETE")
}

// Run run multi on port.
func (m *Multi) Run(port int) chan os.Signal {
	log.Printf("listening port [%v]", port)
	go http.ListenAndServe(fmt.Sprintf(":%v", port), m.Mux)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

// RESTfuler interface
type RESTfuler interface {
	Create(w http.ResponseWriter, r *http.Request)
	One(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// Bind func.
// TODO; Bind must cast request.Values to v interface, actually is not working
func Bind(r *http.Request, v interface{}) error {
	// check model is valid
	_, ok := v.(Modeler)
	if !ok {
		return errModeler
	}

	b, err := json.Marshal(r.Form)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, v)
	if err != nil {
		return err
	}

	return nil
}

// JSON func.
func JSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// Modeler interface
type Modeler interface {
	IsValid() bool
}

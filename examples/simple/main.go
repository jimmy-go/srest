// Package main contains a srest example using rpc calls and RESTfuler interface.
package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/jimmy-go/srest"
)

var (
	port   = flag.Int("port", 7000, "Listen port")
	views  = flag.String("templates", "templates", "Templates dir full path.")
	static = flag.String("static", "static", "Static dir full path.")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(0)
	log.Printf("templates dir [%v]", *views)
	log.Printf("static dir [%v]", *static)
	log.SetFlags(log.Lshortfile)

	// load template views only for this project
	err := srest.LoadViews(*views, srest.DefaultFuncMap)
	if err != nil {
		panic(err)
	}

	m := srest.New(nil)
	m.Get("/static", srest.Static("/static", *static))
	m.Get("/", http.HandlerFunc(homeHandler))
	m.Get("/home", http.HandlerFunc(homeHandler), midAuth)
	m.Use("/v1/api/friends", &FriendAPI{})
	m.Use("/v1/api/mid", &FriendAPI{}, midAuth, midOne) // same last line with middlewares

	/*
		// other form to pass middlewares.
		c := []func(http.Handler) http.Handler{
				midAuth,
				midOne,
		}
		m.Use("/v1/api/mid", &FriendAPI{}, c...)
	*/

	<-m.Run(*port)
	log.Printf("Done")
}

// Response type for common API responses.
type Response struct {
	Data interface{} `json:"response"`
}

// FriendAPI implements srest.RESTfuler interface
type FriendAPI struct{}

// One implements rest.RESTfuler interface.
func (f *FriendAPI) One(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	user, ok := mockusers[id]
	if !ok {
		writeBad(w)
		return
	}

	_ = srest.JSON(w, &Response{user})
}

// List implements rest.RESTfuler interface.
func (f *FriendAPI) List(w http.ResponseWriter, r *http.Request) {
	_ = srest.JSON(w, &Response{mockusers})
}

// Params implements srest.Modeler interface.
// Every Params must be unique per handler but for demostration purposes Create
// and Update method uses the same.
type Params struct {
	Name  string
	Email string
}

// IsValid implements srest.Modeler interface.
func (p *Params) IsValid() error {
	if len(p.Name) < 1 {
		return errors.New("invalid name")
	}
	if len(p.Email) < 1 {
		return errors.New("invalid email")
	}
	return nil
}

// Create implements rest.RESTfuler interface.
func (f *FriendAPI) Create(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	var m Params
	err := srest.Bind(r.Form, &m)
	if err != nil {
		writeBad(w)
		return
	}

	mockusers[m.Email] = &Fake{
		Name:  m.Name,
		Email: m.Email,
	}

	_ = srest.JSON(w, &Response{true})
}

// Update implements rest.RESTfuler interface.
func (f *FriendAPI) Update(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	id := r.URL.Query().Get("id")

	var m Params
	err := srest.Bind(r.Form, &m)
	if err != nil {
		writeBad(w)
		return
	}

	_, ok := mockusers[id]
	if !ok {
		writeBad(w)
		return
	}

	mockusers[id] = &Fake{
		Name:  m.Name,
		Email: m.Email,
	}

	_ = srest.JSON(w, &Response{true})
}

// Delete implements rest.RESTfuler interface.
func (f *FriendAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	_, ok := mockusers[id]
	if !ok {
		writeBad(w)
		return
	}

	delete(mockusers, id)

	_ = srest.JSON(w, &Response{true})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// v can be some struct type too.
	v := map[string]interface{}{}
	v["some"] = "Hello World"
	v["thing"] = 1
	_ = srest.Render(w, "home.html", v)
}

func midAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer 123456" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("unauthorized access"))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func midOne(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("midOne : just here doing nothing :)")
		h.ServeHTTP(w, r)
	})
}

func writeBad(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte("bad request"))
}

var (
	mockusers = map[string]*Fake{}
)

// Fake type represents database user.
type Fake struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

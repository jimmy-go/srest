// Package friends contains friends endpoint
// API
// /v1/api/friends GET
//
// The MIT License (MIT)
//
// Copyright (c) 2016 Angel Del Castillo
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package friends

import (
	"errors"
	"net/http"

	"github.com/jimmy-go/srest"
	"github.com/satori/go.uuid"
)

// Friend model
type Friend struct {
	Name  string `db:"name" schema:"name" json:"name"`
	Email string `db:"email" schema:"email" json:"email"`
}

// IsValid satisfies modeler interface.
func (u *Friend) IsValid() error {
	if len(u.Email) > 20 {
		return errors.New("email must be less than 20 chars")
	}
	if len(u.Name) < 3 {
		return errors.New("name must be greater than 3 chars")
	}
	return nil
}

// API struct
type API struct{}

// New inits configuration
// copy session or reuse database connection, logic belongs to user.
func New() *API {
	return &API{}
}

// Create func
func (a *API) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		srest.JSON(w, &Result{err.Error()})
		return
	}

	var m Friend
	err = srest.Bind(r.PostForm, &m)
	if err != nil {
		srest.JSON(w, &Result{err.Error()})
		return
	}

	id := uuid.NewV4().String()
	srest.JSON(w, "item created: "+id)
}

// One func
func (a *API) One(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	if id == "1" {
		err := errors.New("OK BAD")
		srest.JSON(w, &E{Error: err.Error()})
		return
	}

	u := Friend{
		Name:  "some name",
		Email: "some email",
	}

	srest.JSON(w, &Result{u})
}

// List func
func (a *API) List(w http.ResponseWriter, r *http.Request) {
	var list []*Friend

	list = append(list, &Friend{
		Name:  "some name",
		Email: "some email",
	})

	srest.JSON(w, &Result{list})
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

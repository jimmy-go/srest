// Package srest contains tools for sites and web services creation.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"syscall"
	"testing"
)

func TestMain(m *testing.M) {
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("get pwd : err [%s]", err)
		return
	}

	err = LoadViews(dir+"/mock", DefaultFuncMap)
	if err != nil {
		log.Printf("LoadViews : err [%s]", err)
		return
	}

	v := m.Run()
	gos := runtime.NumGoroutine()
	if gos > 50 {
		log.Printf("goroutines [%v]", gos)
		debug.PrintStack()
		panic("blocked goroutines")
	}

	os.Exit(v)
}

// API satisfies RESTfuler interface
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

func sampleMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// DO NOTHING
		h.ServeHTTP(w, r)
	})
}

func TestServer(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("panics : err [%s]", err)
		}
	}()

	m := New(nil)
	m.Get("/static", Static("/static", "mydir"))
	m.Use("/v1/api/friends", &API{t})
	m.Use("/v1/api/others", &API{t}, sampleMW)
	c := m.Run(9999)
	go func() {
		c <- syscall.SIGTERM
	}()
}

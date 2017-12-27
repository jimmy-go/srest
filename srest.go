// Package srest contains tools for REST services and web sites.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/bmizerany/pat"
)

// Options type.
type Options struct {
	UseTLS  bool
	TLSCert string
	TLSKey  string
}

// SREST type.
type SREST struct {
	Mux     *pat.PatternServeMux
	Options *Options
}

// New returns a new server.
func New(options *Options) *SREST {
	if options == nil {
		options = &Options{}
	}
	m := &SREST{
		Mux:     pat.New(),
		Options: options,
	}
	return m
}

// Get wrapper register a GET endpoint with optional middlewares. It will
// generate endpoints for `uri` and `uri/` because some pat unexpected behaviour.
func (m *SREST) Get(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	s := path.Clean(uri)
	m.Mux.Get(s, chainHandler(hf, mws...))
	m.Mux.Get(s+"/", chainHandler(hf, mws...))
}

// Post wrapper register a POST endpoint with optional middlewares.
func (m *SREST) Post(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Post(path.Clean(uri), chainHandler(hf, mws...))
}

// Put wrapper register a PUT endpoint with optional middlewares.
func (m *SREST) Put(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Put(path.Clean(uri), chainHandler(hf, mws...))
}

// Del wrapper register a DELETE endpoint with optional middlewares.
func (m *SREST) Del(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Del(path.Clean(uri), chainHandler(hf, mws...))
}

// Use receives a RESTfuler interface and generates endpoints for:
// One : GET		path/:id
// List : GET		path/
// Create : POST	path/
// Update : PUT		path/:id
// Remove : DELETE	path/:id
func (m *SREST) Use(uri string, n RESTfuler, mws ...func(http.Handler) http.Handler) {
	m.Get(uri+"/:id", http.HandlerFunc(n.One), mws...)
	m.Get(uri, http.HandlerFunc(n.List), mws...)
	m.Post(uri, http.HandlerFunc(n.Create), mws...)
	m.Put(uri+"/:id", http.HandlerFunc(n.Update), mws...)
	m.Del(uri+"/:id", http.HandlerFunc(n.Delete), mws...)
}

// Run starts the server with http.ListenAndServe or http.ListenAndServeTLS
// returns a channel binded it to SIGTERM and SIGINT signal.
func (m *SREST) Run(port int) chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		addrs := fmt.Sprintf(":%v", port)
		log.Printf("srest : listen on : %v", addrs)
		var err error
		if m.Options.UseTLS {
			err = http.ListenAndServeTLS(addrs, m.Options.TLSCert, m.Options.TLSKey, m.Mux)
		} else {
			err = http.ListenAndServe(addrs, m.Mux)
		}
		if err != nil {
			log.Printf("srest : Run : ListenAndServe : err [%s]", err)
		}
	}()
	return c
}

// RESTfuler interface.
type RESTfuler interface {
	Create(w http.ResponseWriter, r *http.Request)
	One(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

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
	"sort"
	"syscall"

	"github.com/gorilla/pat"
)

// RESTfuler interface.
type RESTfuler interface {
	Create(w http.ResponseWriter, r *http.Request)
	One(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// Options type.
type Options struct {
	UseTLS  bool
	TLSCert string
	TLSKey  string
}

// SREST type.
type SREST struct {
	Mux      *pat.Router
	Options  *Options
	Map      map[string]bool
	handlers []tmpHandler
}

// New returns a new server.
func New(options *Options) *SREST {
	if options == nil {
		options = &Options{}
	}
	m := &SREST{
		Mux:     pat.New(),
		Options: options,
		Map:     make(map[string]bool),
	}
	return m
}

// Get wrapper register a GET endpoint with optional middlewares. It will
// generate endpoints for `uri` and `uri/` because some pat unexpected behaviour.
func (m *SREST) Get(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	s := path.Clean(uri)
	checkDuplicate(m, "GET", s)
	h := chainHandler(hf, mws...)
	m.handlers = append(m.handlers, tmpHandler{"GET", s, h})
	if s != "/" {
		m.handlers = append(m.handlers, tmpHandler{"GET", s + "/", h})
	}
}

// Post wrapper register a POST endpoint with optional middlewares.
func (m *SREST) Post(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	s := path.Clean(uri)
	checkDuplicate(m, "POST", s)
	h := chainHandler(hf, mws...)
	m.handlers = append(m.handlers, tmpHandler{"POST", s, h})
	if s != "/" {
		m.handlers = append(m.handlers, tmpHandler{"POST", s + "/", h})
	}
}

// Put wrapper register a PUT endpoint with optional middlewares.
func (m *SREST) Put(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	s := path.Clean(uri)
	checkDuplicate(m, "PUT", s)
	h := chainHandler(hf, mws...)
	m.handlers = append(m.handlers, tmpHandler{"PUT", s, h})
	if s != "/" {
		m.handlers = append(m.handlers, tmpHandler{"PUT", s + "/", h})
	}
}

// Del wrapper register a DELETE endpoint with optional middlewares.
func (m *SREST) Del(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	s := path.Clean(uri)
	checkDuplicate(m, "DELETE", s)
	h := chainHandler(hf, mws...)
	m.handlers = append(m.handlers, tmpHandler{"DELETE", s, h})
	if s != "/" {
		m.handlers = append(m.handlers, tmpHandler{"DELETE", s + "/", h})
	}
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

// registerHandlers sorts and register the handlers on Mux. Erases the map and
// slice from SREST in order to free memory. It's called once by Run method.
func (m *SREST) registerHandlers() error {
	// Sort handlers.
	sort.Sort(ByURIDesc(m.handlers))

	// Register pat endpoints.
	if err := registerHandlers(m.Mux, m.handlers); err != nil {
		return err
	}
	m.Map = nil
	m.handlers = nil
	return nil
}

// Run starts the server with http.ListenAndServe or http.ListenAndServeTLS
// returns a channel binded it to SIGTERM and SIGINT signal.
func (m *SREST) Run(port int) chan os.Signal {
	if err := m.registerHandlers(); err != nil {
		panic(fmt.Sprintf("Run : register handlers : err [%s]", err))
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		var err error
		addrs := fmt.Sprintf(":%v", port)
		if m.Options.UseTLS {
			err = http.ListenAndServeTLS(addrs, m.Options.TLSCert, m.Options.TLSKey, m.Mux)
		} else {
			err = http.ListenAndServe(addrs, m.Mux)
		}
		if err != nil {
			log.Printf("srest : Run : err [%s]", err)
		}
	}()
	return c
}

type tmpHandler struct {
	Method, URI string
	Handler     http.Handler
}

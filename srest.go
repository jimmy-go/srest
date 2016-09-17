// Package srest contains utilyties for sites creation and web services.
/*
	RESTfuler interface:
		One(w http.ResponseWriter, r *http.Request)
		List(w http.ResponseWriter, r *http.Request)
		Create(w http.ResponseWriter, r *http.Request)
		Update(w http.ResponseWriter, r *http.Request)
		Delete(w http.ResponseWriter, r *http.Request)

	Modeler interface:
		IsValid() error
*/
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
package srest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/bmizerany/pat"
	"github.com/gorilla/schema"
)

var (
	debug        bool
	templatesDir string

	// DefaultFuncMap can be used with LoadViews for common template tasks like:
	//
	// cap: capitalize strings
	// eqs: compare value of two types.
	DefaultFuncMap = deffuncmap()

	// ErrModeler error returned when modeler interface is
	// not implemented.
	ErrModeler = errors.New("srest: modeler interface not found")

	// ErrTemplatesInited error returned when LoadViews
	// function is called twice.
	ErrTemplatesInited = errors.New("srest: templates already inited")

	// ErrTemplatesNil error returned when not template files
	// were loaded.
	ErrTemplatesNil = errors.New("srest: not templates found")

	// ErrTemplateNotFound error returned when template name
	// is not present.
	ErrTemplateNotFound = errors.New("srest: template not found")
)

func deffuncmap() template.FuncMap {
	return template.FuncMap{
		"cap": func(s string) string {
			if len(s) < 1 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		// eqs validates x and y are equal no matter type.
		"eqs": func(x, y interface{}) bool {
			return fmt.Sprintf("%v", x) == fmt.Sprintf("%v", y)
		},
	}
}

// Options struct
type Options struct {
	UseTLS bool
	TLSCer string
	TLSKey string
}

var (
	// DefaultConf contains default configuration without TLS.
	DefaultConf = &Options{
		UseTLS: false,
	}
)

// SREST struct.
type SREST struct {
	Mux     *pat.PatternServeMux
	Options *Options
}

// New returns a new server.
func New(options *Options) *SREST {
	if options == nil {
		options = DefaultConf
	}
	m := &SREST{
		Mux:     pat.New(),
		Options: options,
	}
	return m
}

// Get wrapper allows GET endpoints and middlewares. It will
// generate endpoints for `resource` and `resource/` because
// some services requires both endpoints.
func (m *SREST) Get(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	// FIXME; allow both o remove one?
	m.Mux.Get(path.Clean(uri), ChainHandler(hf, mws...))
	m.Mux.Get(path.Clean(uri)+"/", ChainHandler(hf, mws...))
}

// Post wrapper useful for add middleware like Use method.
func (m *SREST) Post(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Post(path.Clean(uri), ChainHandler(hf, mws...))
}

// Put wrapper useful for add middleware like Use method.
func (m *SREST) Put(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Put(path.Clean(uri), ChainHandler(hf, mws...))
}

// Del wrapper useful for add middleware like Use method.
func (m *SREST) Del(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Del(path.Clean(uri), ChainHandler(hf, mws...))
}

// Use receives a RESTfuler interface and generates endpoints for:
//
// GET /:id
// GET /
// POST /
// PUT /:id
// DELETE /:id
func (m *SREST) Use(uri string, n RESTfuler, mws ...func(http.Handler) http.Handler) {
	m.Get(uri+"/:id", http.HandlerFunc(n.One), mws...)
	m.Get(uri, http.HandlerFunc(n.List), mws...)
	m.Post(uri, http.HandlerFunc(n.Create), mws...)
	m.Put(uri+"/:id", http.HandlerFunc(n.Update), mws...)
	m.Del(uri+"/:id", http.HandlerFunc(n.Delete), mws...)
}

// Run start a server listening with http.ListenAndServe or http.ListenAndServeTLS
// returns a channel bind it to SIGTERM and SIGINT signal
// you will block this way: <-m.Run()
func (m *SREST) Run(port int) chan os.Signal {
	// TODO; change logic to allow server stop without leaking a goroutine and handle graceful shutdown.
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		addrs := fmt.Sprintf(":%v", port)
		log.Printf("srest: Run %v", addrs)
		var err error
		if m.Options.UseTLS {
			err = http.ListenAndServeTLS(addrs, m.Options.TLSCer, m.Options.TLSKey, m.Mux)
		} else {
			err = http.ListenAndServe(addrs, m.Mux)
		}
		if err != nil {
			log.Printf("srest: Run : ListenAndServe : err [%s]", err)
		}
	}()
	return c
}

// Debug enables template files reload on every request.
func (m *SREST) Debug(ok bool) {
	debug = ok
}

// Debug enables template files reload on every request.
func Debug(ok bool) {
	debug = ok
}

// Static handler.
//
// Usage:
// Get("/public", Static("/public", "mydir"))
func Static(uri, dir string) http.Handler {
	uri = path.Clean(uri) + "/"
	dir = path.Clean(dir) + "/"
	return http.StripPrefix(uri, http.FileServer(http.Dir(dir)))
}

// ChainHandler concats multiple handlers in one http.Handler.
func ChainHandler(fh http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	// no middlewares then return handler
	if len(mws) < 1 {
		return fh
	}

	var cs []func(http.Handler) http.Handler
	cs = append(cs, mws...)
	var h http.Handler
	h = fh // this disable linter warning
	for i := range cs {
		h = cs[len(cs)-1-i](h)
	}
	return h
}

// RESTfuler interface
type RESTfuler interface {
	Create(w http.ResponseWriter, r *http.Request)
	One(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

var (
	// schDecoder default gorilla schema decoder.
	schDecoder = schema.NewDecoder()
)

// Bind implements gorilla schema and runs IsValid method from data.
func Bind(vars url.Values, dst interface{}) error {
	err := schDecoder.Decode(dst, vars)
	if err != nil {
		return err
	}
	// check model is valid
	mo, ok := dst.(Modeler)
	if !ok {
		return ErrModeler
	}
	return mo.IsValid()
}

// JSON writes v to response writer.
func JSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

var (
	// templates collection.
	templates  = map[string]*template.Template{}
	tmplInited bool
	mut        sync.RWMutex
)

// LoadViews read html files on dir tree and parses it
// as templates.
// In order to render templates you need to call Render
// function passing <file.html> or <subdir>/<file.html>
// as name for template.
//
// funcMap will overwrite DefaultFuncMap.
func LoadViews(dir string, funcMap template.FuncMap) error {
	if tmplInited {
		return ErrTemplatesInited
	}

	dir = filepath.Clean(dir)
	templatesDir = dir

	var files []string
	var data []byte
	err := filepath.Walk(dir, func(name string, info os.FileInfo, err error) error {
		// take template name from subdir+filename
		tname := strings.Replace(name, dir+"/", "", -1)
		ext := filepath.Ext(name)
		// ommit files not .html
		if ext != ".html" {
			return nil
		}
		b, err := ioutil.ReadFile(name)
		if err != nil {
			return err
		}
		// append to unique template data
		data = append(data, []byte(fmt.Sprintf(`{{define "%s"}}`, tname))...)
		data = append(data, b...)
		data = append(data, []byte(`{{end}}`)...)
		// wee need this after for template parsing
		files = append(files, tname)
		return nil
	})
	if err != nil {
		return err
	}

	DefaultFuncMap = funcMap
	for _, k := range files {
		// template parsing
		templates[k] = template.Must(template.New(k).Funcs(funcMap).Parse(string(data)))
	}

	tmplInited = true
	return nil
}

// Render writes a template to http response.
// In order to render templates you need to call Render
// function passing <file.html> or <subdir>/<file.html>
// as name for template.
func Render(w http.ResponseWriter, name string, v interface{}) error {
	// for now use a mutex, later implementations can use sync.Pool of templates.
	mut.RLock()
	defer mut.RUnlock()

	if debug {
		// clean templates
		for k := range templates {
			delete(templates, k)
		}
		tmplInited = false
		// load templates again
		// this generates a race condition. TODO; check later if a really trouble
		// on debug mode, this is not expected to be turned on into production.
		err := LoadViews(templatesDir, DefaultFuncMap)
		if err != nil {
			return err
		}
	}
	if !tmplInited {
		return ErrTemplatesNil
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// write template to buffer to make sure is working.
	var buf bytes.Buffer
	t, ok := templates[name]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("template not found"))
		return ErrTemplateNotFound
	}
	err := t.ExecuteTemplate(&buf, name, v)
	if err != nil {
		return err
	}
	// buffer writing was done without errors. Write to http
	// response.
	_, err = buf.WriteTo(w)
	return err
}

// Modeler interface
type Modeler interface {
	IsValid() error
}

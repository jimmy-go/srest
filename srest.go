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
	srestDebug   bool
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
	m.Mux.Get(path.Clean(uri), chainHandler(hf, mws...))
	m.Mux.Get(path.Clean(uri)+"/", chainHandler(hf, mws...))
}

// Post wrapper useful for add middleware like Use method.
func (m *SREST) Post(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Post(path.Clean(uri), chainHandler(hf, mws...))
}

// Put wrapper useful for add middleware like Use method.
func (m *SREST) Put(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Put(path.Clean(uri), chainHandler(hf, mws...))
}

// Del wrapper useful for add middleware like Use method.
func (m *SREST) Del(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	m.Mux.Del(path.Clean(uri), chainHandler(hf, mws...))
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
	srestDebug = ok
}

// Debug enables template files reload on every request.
func Debug(ok bool) {
	srestDebug = ok
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

// chainHandler concats multiple handlers in one http.Handler.
func chainHandler(fh http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
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

	mut  sync.RWMutex
	xmut sync.Mutex
	cmut sync.RWMutex
)

// LoadViews read html files on dir tree and parses to templates.
// In order to render templates you need to call Render function passing
// <file.html> or <subdir>/<file.html> as name.
//
// funcMap will overwrite DefaultFuncMap.
func LoadViews(dirs string, funcMap template.FuncMap) error {
	xmut.Lock()
	defer xmut.Unlock()

	if tmplInited {
		return ErrTemplatesInited
	}

	// clean templates map.
	for k := range templates {
		delete(templates, k)
	}
	templatesDir = dirs

	// dir = filepath.Clean(dir)

	// buftmpl contains all the data from templates in dir and subdirectories.
	var buftmpl bytes.Buffer

	// we need to keep the names for later template parsing.
	var names []string

	// TODO; update doc

	x := strings.Split(dirs, ",")
	for i := range x {
		dir := filepath.Clean(x[i])

		var dirPrefix string
		if fs := strings.Split(dir, "/"); len(fs) > 0 {
			dirPrefix = fs[len(fs)-1]
		}
		if i != 0 {
			log.Printf("LoadViews: using secondary directory [%s] as [%s]", dir, dirPrefix)
		}

		if err := filepath.Walk(dir, func(name string, info os.FileInfo, err error) error {

			// take template name from subdir+filename

			tname := strings.Replace(name, dir+"/", "", -1)
			if i != 0 {
				tname = dirPrefix + "/" + tname
			}
			ext := filepath.Ext(name)
			// ommit files not .html
			if ext != ".html" {
				return nil
			}

			// don't parse emtpy files
			if info.Size() <= 1 {
				return fmt.Errorf("empty file: %s", name)
			}

			if _, err := buftmpl.Write([]byte(`{{define "` + tname + `"}}`)); err != nil {
				return err
			}
			f, err := os.Open(name)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Printf("error closing file : name [%s] err [%s]", tname, err)
				}
			}()

			if _, err := buftmpl.ReadFrom(f); err != nil {
				return err
			}

			// clean \r
			buftmpl.Truncate(buftmpl.Len() - 1)
			if _, err := buftmpl.Write([]byte(`{{end}}`)); err != nil {
				return err
			}

			names = append(names, tname)
			return nil
		}); err != nil {
			return err
		}
	}

	for _, name := range names {
		// load template
		templates[name] = template.Must(template.New(name).Funcs(funcMap).Parse(buftmpl.String()))
	}
	DefaultFuncMap = funcMap
	tmplInited = true
	return nil
}

// Render writes a template name to w.
// In order to render templates you need to call Render function passing
// <file.html> or <subdir>/<file.html> as name.
//
// Future implementations are ahead to improve render time without locking.
func Render(w http.ResponseWriter, name string, v interface{}) error {
	// for now use a mutex, later implementations can use sync.Pool of templates.
	mut.Lock()
	defer mut.Unlock()

	if srestDebug {
		tmplInited = false
		// load templates again
		// this generates a race condition. TODO; check later if a really trouble
		// on debug mode, this is not expected to be turned on into production.
		if err := LoadViews(templatesDir, DefaultFuncMap); err != nil {
			return err
		}
	}
	if !tmplInited {
		return ErrTemplatesNil
	}

	// write template to buffer to make sure is working.
	t, ok := templates[name]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("template not found"))
		if err != nil {
			log.Printf("err [%s]", err)
		}
		return ErrTemplateNotFound
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	err := t.ExecuteTemplate(w, name, v)
	if err != nil {
		return err
	}
	return nil
}

// Modeler interface
type Modeler interface {
	IsValid() error
}

// Stresser interface
type Stresser interface {
	Modeler

	Factory() error

	Fuzz() error
}

// Stressor interface
type Stressor interface {
	Execute(Stresser, string)
}

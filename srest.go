// Package srest contains utilyties for sites creation and web services.
/*
	RESTfuler interface:
		Create(w http.ResponseWriter, r *http.Request)
		One(w http.ResponseWriter, r *http.Request)
		List(w http.ResponseWriter, r *http.Request)
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
	// Cap: capitalize
	DefaultFuncMap = deffuncmap()

	errModeler          = errors.New("srest: modeler interface not found")
	errTemplatesInited  = errors.New("srest: templates already inited")
	errTemplatesNil     = errors.New("srest: not templates found")
	errTemplateNotFound = errors.New("srest: template not found")
)

func deffuncmap() template.FuncMap {
	// TODO; add common functions for templates
	return template.FuncMap{
		"cap": func(s string) string {
			if len(s) < 1 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
	}
}

// Options struct
type Options struct {
	UseTLS bool
	TLSCer string
	TLSKey string
}

// Multi struct.
type Multi struct {
	Mux *pat.PatternServeMux
}

// New returns a new server.
func New(opts *Options) *Multi {
	m := &Multi{
		Mux: pat.New(),
	}
	return m
}

// Static replaces old Multi method.
//
// Usage with Get("/example", Static("/example", "mydir"))
func Static(uri, dir string) http.Handler {
	return http.StripPrefix(uri, http.FileServer(http.Dir(dir)))
}

// Get conveniense.
func (m *Multi) Get(uri string, hf http.Handler, mws ...func(http.Handler) http.Handler) {
	// TODO; move hfs outside
	hfs := func(fh http.Handler) http.Handler {
		var cs []func(http.Handler) http.Handler
		cs = append(cs, mws...)
		var h http.Handler
		h = fh
		for i := range cs {
			h = cs[len(cs)-1-i](h)
		}
		return h
	}
	m.Mux.Get(uri, hfs(hf))
}

// Use adds endpoints RESTful
func (m *Multi) Use(uri string, n RESTfuler, mws ...func(http.Handler) http.Handler) {
	uri = path.Clean(uri)
	if len(mws) < 1 {
		m.Mux.Get(uri+"/:id", http.HandlerFunc(n.One))
		m.Mux.Get(uri, http.HandlerFunc(n.List))
		m.Mux.Post(uri, http.HandlerFunc(n.Create))
		m.Mux.Put(uri, http.HandlerFunc(n.Update))
		m.Mux.Del(uri+"/:id", http.HandlerFunc(n.Delete))
		return
	}

	// TODO; move hfs outside
	hfs := func(fh http.HandlerFunc) http.Handler {
		var cs []func(http.Handler) http.Handler
		cs = append(cs, mws...)
		var h http.Handler = http.HandlerFunc(fh)
		for i := range cs {
			h = cs[len(cs)-1-i](h)
		}
		return h
	}

	m.Mux.Get(uri+"/:id", hfs(n.One))
	m.Mux.Get(uri, hfs(n.List))
	m.Mux.Post(uri, hfs(n.Create))
	m.Mux.Put(uri, hfs(n.Update))
	m.Mux.Del(uri+"/:id", hfs(n.Delete))
}

// Run run multi on port.
func (m *Multi) Run(port int) chan os.Signal {
	http.Handle("/", m.Mux)
	go log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

// Debug enables templates reload with every petition.
func (m *Multi) Debug(ok bool) {
	debug = ok
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
		return errModeler
	}
	if err := mo.IsValid(); err != nil {
		return fmt.Errorf("srest: %v", err)
	}
	return nil
}

// JSON func.
func JSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

var (
	templates  = map[string]*template.Template{}
	tmplInited bool
	mut        sync.RWMutex
)

// LoadViews func
//
// funcMap overwrites DefaultFuncMap
func LoadViews(dir string, funcMap template.FuncMap) error {
	if tmplInited {
		return errTemplatesInited
	}

	dir = filepath.Clean(dir)
	templatesDir = dir

	var files []string
	var data []byte
	err := filepath.Walk(dir, func(name string, info os.FileInfo, err error) error {
		// take template name from subdir+filename
		tname := strings.Replace(name, dir+"/", "", -1)
		ext := filepath.Ext(name)
		if ext != ".html" {
			// We need to ommit file is not html
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
func Render(w http.ResponseWriter, view string, v interface{}) error {
	if debug {
		// clean templates
		for k := range templates {
			delete(templates, k)
		}
		tmplInited = false
		// load templates again
		// this generates a race condition. TODO; check later if a really trouble
		// on debug mode, this is not expected to be turned on to production.
		err := LoadViews(templatesDir, DefaultFuncMap)
		if err != nil {
			return err
		}
	}
	if !tmplInited {
		return errTemplatesNil
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var buf bytes.Buffer
	mut.RLock()
	t, ok := templates[view]
	mut.RUnlock()
	if !ok {
		w.Write([]byte("template not found"))
		return errTemplateNotFound
	}
	err := t.ExecuteTemplate(&buf, view, v)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}
	return nil
}

// Modeler interface
type Modeler interface {
	IsValid() error
}

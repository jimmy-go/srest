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
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
)

var (
	errModeler          = errors.New("modeler interface not found")
	errTemplatesInited  = errors.New("templates already inited")
	errTemplatesNil     = errors.New("not templates found")
	errTemplateNotFound = errors.New("template not found")
)

// Options struct
type Options struct {
	UseTLS bool
	TLSCer string
	TLSKey string
}

// Multi struct.
type Multi struct {
	Mux *mux.Router
}

// New returns a new server.
func New(opts *Options) *Multi {
	m := &Multi{
		Mux: mux.NewRouter(),
	}
	return m
}

// Static func wrapper for mux.PathPrefix(path).Handler(http.StripPrefix(path, http.FileServer(http.Dir(dir))))
func (m *Multi) Static(path, dir string) {
	m.Mux.PathPrefix(path).Handler(http.StripPrefix(path, http.FileServer(http.Dir(dir))))
}

// Use adds a module.
func (m *Multi) Use(uri string, n RESTfuler) {
	m.Mux.HandleFunc(uri, n.Create).Methods("POST")
	m.Mux.HandleFunc(uri+"/{id}", n.One).Methods("GET")
	m.Mux.HandleFunc(uri, n.List).Methods("GET")
	m.Mux.HandleFunc(uri, n.Update).Methods("PUT")
	m.Mux.HandleFunc(uri, n.Delete).Methods("DELETE")
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

var (
	templates  = map[string]*template.Template{}
	tmplInited bool
	mut        sync.RWMutex
)

// LoadViews func
// if recursive enabled then scans all subdirs
func LoadViews(dir string, recursive bool) error {
	if tmplInited {
		return errTemplatesInited
	}

	dir = filepath.Clean(dir)

	var files []string
	var data []byte
	err := filepath.Walk(dir, func(name string, info os.FileInfo, err error) error {
		// take template name from subdir+filename
		tname := strings.Replace(name, dir+"/", "", -1)
		ext := filepath.Ext(name)
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
		log.Printf("LoadViews : err [%s]", err)
		return err
	}

	// log.Printf("all templates [%s]", string(data))

	for _, k := range files {
		// template parsing
		templates[k] = template.Must(template.New(k).Parse(string(data)))
	}

	tmplInited = true
	return nil
}

// Render writes a template to http response.
func Render(w http.ResponseWriter, view string, v interface{}) error {
	if !tmplInited {
		log.Printf("Render : err [%s]", errTemplatesNil)
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
		log.Printf("Render : err [%s]", err)
		return err
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("Render : err [%s]", err)
		return err
	}
	return nil
}

// Modeler interface
type Modeler interface {
	IsValid() bool
}

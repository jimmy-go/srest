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

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var (
	errModeler          = errors.New("srest: modeler interface not found")
	errTemplatesInited  = errors.New("srest: templates already inited")
	errTemplatesNil     = errors.New("srest: not templates found")
	errTemplateNotFound = errors.New("srest: template not found")
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

// Static func wrapper for
// mux.PathPrefix(uri).Handler(http.StripPrefix(uri, http.FileServer(http.Dir(dir))))
func (m *Multi) Static(uri, dir string) {
	m.Mux.PathPrefix(uri).Handler(http.StripPrefix(uri, http.FileServer(http.Dir(dir))))
}

// Use adds endpoints RESTful
func (m *Multi) Use(uri string, n RESTfuler, mws ...http.Handler) {
	uri = path.Clean(uri)
	log.Printf("uri clean [%v]", uri)

	nmws := func(met http.HandlerFunc) http.Handler {
		return met
	}
	m.Mux.Handle(uri+"/{id}", nmws(n.One)).Methods("GET")
	m.Mux.Handle(uri, nmws(n.List)).Methods("GET")
	m.Mux.Handle(uri, nmws(n.Create)).Methods("POST")
	m.Mux.Handle(uri, nmws(n.Update)).Methods("PUT")
	m.Mux.Handle(uri+"/{id}", nmws(n.Delete)).Methods("DELETE")

	//	m.Mux.HandleFunc(uri+"/{id}", n.One).Methods("GET")
	//	m.Mux.HandleFunc(uri, n.List).Methods("GET")
	//	m.Mux.HandleFunc(uri, n.Create).Methods("POST")
	//	m.Mux.HandleFunc(uri, n.Update).Methods("PUT")
	//	m.Mux.HandleFunc(uri+"/{id}", n.Delete).Methods("DELETE")
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
func LoadViews(dir string) error {
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
	IsValid() error
}

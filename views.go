// Package srest contains tools for REST services and web sites.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// ErrTemplateNotFound error returned when template is not found.
	ErrTemplateNotFound = errors.New("srest: template not found")

	debugb       bool
	templatesDir string

	// templates collection.
	templates = map[string]*template.Template{}

	mut  sync.RWMutex
	xmut sync.Mutex
	cmut sync.RWMutex
)

// LoadViews parses html files on dir as templates.
func LoadViews(dirs string, funcMap template.FuncMap) error {
	xmut.Lock()
	defer xmut.Unlock()

	// Clean templates map.
	for k := range templates {
		delete(templates, k)
	}
	templatesDir = dirs

	var buf bytes.Buffer

	// We need to keep the names for later template parsing.
	var names []string

	paths := strings.Split(dirs, ",")
	for i := range paths {
		cpath := filepath.Clean(paths[i])

		var prefix string
		folders := strings.Split(cpath, "/")
		if len(folders) > 0 && i > 0 {
			prefix = folders[len(folders)-1]
		}
		if err := filepath.Walk(cpath, func(name string, info os.FileInfo, ferr error) error {
			if ferr != nil {
				return ferr
			}
			if info.Size() == 0 {
				return fmt.Errorf("empty file: %s", name)
			}

			newName, err := parseFile(cpath, name, prefix, &buf)
			if err != nil {
				return err
			}
			names = append(names, newName)

			return nil
		}); err != nil {
			return err
		}
	}

	for _, name := range names {
		// load template
		templates[name] = template.Must(template.New(name).Funcs(funcMap).Parse(buf.String()))
	}
	DefaultFuncMap = funcMap
	return nil
}

// Render writes a template to w.
// In order to render templates you need to call Render function passing
// <file.html> or <subdir>/<file.html> as name.
func Render(w http.ResponseWriter, name string, v interface{}) error {
	mut.Lock()
	defer mut.Unlock()

	if debugb {
		// Load templates again
		// this generates a race condition. TODO; check later if a really trouble
		// on debug mode, this is not expected to be turned on into production.
		if err := LoadViews(templatesDir, DefaultFuncMap); err != nil {
			return err
		}
	}

	// Write template to buffer to make sure is working.
	t, ok := templates[name]
	if !ok {
		http.Error(w, "template view not found", http.StatusInternalServerError)
		return ErrTemplateNotFound
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if err := t.ExecuteTemplate(w, name, v); err != nil {
		return err
	}
	return nil
}

// Debug enables template files reload on every request.
func Debug(ok bool) {
	debugb = ok
}

func parseFile(dir, name, prefix string, buf *bytes.Buffer) (string, error) {
	// log.Printf("parseFile : dir [%s] name [%s] prefix [%s]", dir, name, prefix)
	// Take template name from subdir+filename
	tname := strings.Replace(name, dir+"/", "", -1)
	if prefix != "" {
		tname = prefix + "/" + tname
	}

	// Ommit files not .html
	if ext := filepath.Ext(name); ext != ".html" {
		return "", nil
	}
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	// Benchmark this cost.
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	body := string(b)
	if _, err := fmt.Fprintf(buf, `{{define "%s"}}%s{{end}}`, tname, body); err != nil {
		return "", err
	}
	return tname, nil
}

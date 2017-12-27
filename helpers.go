// Package srest contains tools for REST services and web sites.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"
)

var (
	// DefaultFuncMap can be used with LoadViews for common template tasks like:
	//	cap: capitalize strings
	//	eqs: compare value of two types.
	DefaultFuncMap = deffuncmap()
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

// Static handler for static files.
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
	// No middlewares then return handler
	if len(mws) < 1 {
		return fh
	}

	var cs []func(http.Handler) http.Handler
	cs = append(cs, mws...)
	var h http.Handler
	// This disable linter warning
	h = fh
	for i := range cs {
		h = cs[len(cs)-1-i](h)
	}
	return h
}

// JSON writes v to response writer.
func JSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	return json.NewEncoder(w).Encode(v)
}

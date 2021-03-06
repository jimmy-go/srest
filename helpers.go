package srest

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/mux"
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

func checkDuplicate(m *SREST, method, uri string) {
	// Validate path vars.
	s := method + ":" + removeVars(uri)
	if _, ok := m.Map[s]; ok {
		panic(fmt.Sprintf("duplicated definition: %s %s", method, uri))
	}
	m.Map[s] = true
}

func removeVars(uri string) string {
	var res []string
	s := strings.Split(uri, "/")
	for _, x := range s {
		if strings.Contains(x, ":") {
			x = "*"
		}
		res = append(res, strings.TrimSpace(x))
	}
	return strings.Join(res, "/")
}

// ByURIDesc implements sort.Interface for []tmpHandler based on the URI field.
type ByURIDesc []tmpHandler

func (a ByURIDesc) Len() int           { return len(a) }
func (a ByURIDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByURIDesc) Less(i, j int) bool { return removeVars(a[i].URI) > removeVars(a[j].URI) }

func registerHandlers(r *mux.Router, hs []tmpHandler) error {
	for _, x := range hs {
		switch x.Method {
		case "GET", "POST", "PUT", "DELETE":
			uri := paramsToGorilla(x.URI)
			h := varsWrap(x.Handler)
			route := r.NewRoute()
			route.Path(uri).Handler(h).Methods(x.Method)
		default:
			return fmt.Errorf("method not found: %s", x.Method)
		}
	}
	return nil
}

func varsWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		registerVars(r, vars)
		h.ServeHTTP(w, r)
	})
}

// registerVars adds the matched route variables to the URL query.
// taken from: https://github.com/gorilla/pat/blob/master/pat.go#L95
func registerVars(r *http.Request, vars map[string]string) {
	parts, i := make([]string, len(vars)), 0
	for key, value := range vars {
		parts[i] = url.QueryEscape(":"+key) + "=" + url.QueryEscape(value)
		i++
	}
	q := strings.Join(parts, "&")
	if r.URL.RawQuery == "" {
		r.URL.RawQuery = q
	} else {
		r.URL.RawQuery += "&" + q
	}
}

// paramsToGorilla change old notation ':param' to '{param}'.
func paramsToGorilla(uri string) string {
	var res []string
	s := strings.Split(uri, "/")
	for _, x := range s {
		if strings.Contains(x, ":") {
			x = "{" + x[1:] + "}"
		}
		res = append(res, strings.TrimSpace(x))
	}
	return strings.Join(res, "/")
}

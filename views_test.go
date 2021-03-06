package srest

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadViews(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)
	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)
}

func TestLoadViewsFail(t *testing.T) {
	err := LoadViews("mock2fail", map[string]interface{}{})
	assert.NotNil(t, err)
	assert.EqualValues(t, "lstat mock2fail: no such file or directory", fmt.Sprintf("%s", err))
}

func TestRender(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	// a/index.html file must exists o this will panic
	// index.html content will be:
	// {{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	assert.Nil(t, err)

	b, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)

	actual := string(b)
	expected := "I am lowercase-eqs:true"
	assert.EqualValues(t, expected, actual)
}

func TestRenderFail(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	// a/index.html file must exists o this will panic
	// index.html content will be: {{cap "i am lowercase"}}
	err = Render(w, "notfound.html", map[string]interface{}{"x": 1})
	assert.EqualValues(t, ErrTemplateNotFound, err)

	b, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)

	actual := string(b[:len(b)-1])
	expected := "template view not found"
	assert.EqualValues(t, expected, actual)
}

func TestRenderDebug(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)
	Debug(true)

	w := httptest.NewRecorder()
	// a/index.html file must exists o this will panic
	// index.html content will be:
	// {{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	assert.Nil(t, err)

	b, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)

	actual := string(b)
	expected := "I am lowercase-eqs:true"
	assert.EqualValues(t, expected, actual)
}

func TestRenderDebugFail(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	Debug(true)
	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	// a/index.html file must exists o this will panic
	// index.html content will be:
	// {{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}
	templatesDir = "2fail"
	err = Render(w, "index.html", map[string]interface{}{"x": 1})
	assert.EqualValues(t, "lstat 2fail: no such file or directory", fmt.Sprintf("%s", err))

	// Validate error when writer is nil.
	wr := &WR{}
	Debug(false)
	err = LoadViews(dir+"/a", DefaultFuncMap)
	err = Render(wr, "index.html", nil)
	assert.EqualValues(t, "expected fail", fmt.Sprintf("%s", err))
}

// WR used for failed write validations.
type WR struct{}

// Header implements http.ResponseWriter.
func (wr *WR) Header() http.Header {
	h := http.Header{}
	return h
}

// WriteHeader implements http.ResponseWriter.
func (wr *WR) WriteHeader(code int) {}

// Write implements io.Writer.
func (wr *WR) Write(b []byte) (int, error) {
	return 0, errors.New("expected fail")
}

func TestParseFile(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	var buf bytes.Buffer
	s, err := parseFile(dir+"/a/all", dir+"/a/all/all.html", "", &buf)
	oerr, ok := err.(*os.PathError)
	if ok {
		log.Printf("path error [%s]", oerr)
	}
	assert.Nil(t, err)
	assert.EqualValues(t, "all.html", s)

	b, err := ioutil.ReadAll(&buf)
	assert.Nil(t, err)

	actual := string(b)
	expected := `{{define "all.html"}}before_index::{{template "index.html" .}}::after_index.before_menu::{{template "menu.html" .}}::after_menu{{end}}`
	assert.EqualValues(t, expected, actual)
}

func TestParseFileFail(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	var buf bytes.Buffer
	_, err = parseFile(dir+"/c", dir+"/c/not_exists.html", "", &buf)
	assert.NotNil(t, err)
	oerr, ok := err.(*os.PathError)
	assert.EqualValues(t, true, ok)
	assert.EqualValues(t, "no such file or directory", fmt.Sprintf("%s", oerr.Err))
}

func TestLoadViewsMultipleDirs(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)
	err = LoadViews(dir+"/a,"+dir+"/b", DefaultFuncMap)
	assert.Nil(t, err)
}

func TestLoadViewsEmptyFile(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)
	err = LoadViews(dir+"/c", DefaultFuncMap)
	assert.NotNil(t, err)
}

package srest

import (
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRender(b *testing.B) {
	dir, err := getTempDir()
	assert.Nil(b, err)

	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(b, err)

	w := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := struct{}{}
		err := Render(w, "index.html", v)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSON(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	w := httptest.NewRecorder()
	res := struct{}{}
	for i := 0; i < b.N; i++ {
		err := JSON(w, res)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func init() {
	t := template.Must(template.New("hello").Parse(aCase))
	var err error
	tA, err = t.Parse(header)
	if err != nil {
		panic(err)
	}

	tB = template.Must(template.New("hello").Parse(bCase))
}

var (
	tA *template.Template
	tB *template.Template
)

func BenchmarkCaseA(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		err := tA.Execute(w, "")
		if err != nil {
			b.Fatal(err)
		}
		// log.Printf("body [%s]", w.Body.String())
	}
}

func BenchmarkCaseB(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		err := tB.Execute(w, "")
		if err != nil {
			b.Fatal(err)
		}
		// log.Printf("body [%s]", w.Body.String())
	}
}

const (
	header = `{{define "header"}}hello im header.{{end}}`

	// aCase template execution.
	aCase = `
		hello world A
		{{template "header" . }}
	`

	// bCase html inside.
	bCase = `
		hello world B
		hello im header
	`
)

package srest

import (
	"net/http/httptest"
	"testing"
)

func BenchmarkRender(b *testing.B) {
	//	dir, err := os.Getwd()
	//	if err != nil {
	//		b.Errorf("get pwd : err [%s]", err)
	//	}
	//
	//	tmplInited = false
	//	funcm := deffuncmap()
	//	err = LoadViews(dir+"/mock", funcm)
	//	if err != nil {
	//		b.Errorf("LoadViews : err [%s]", err)
	//		return
	//	}

	w := httptest.NewRecorder()
	v := struct{}{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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

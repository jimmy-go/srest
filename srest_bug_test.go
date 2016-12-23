// Package srest contains test for bug fixes.
//
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
	"log"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestBugRaceRender(t *testing.T) {
	racerender(t, 1000)
	log.Printf("render done")
}

func TestBugRaceRenderDebug(t *testing.T) {
	Debug(true)
	racerender(t, 1000)
	log.Printf("render debug done")
}

func racerender(t *testing.T, l int) {
	var wg sync.WaitGroup
	wg.Add(l)
	for i := 0; i < l; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			err := Render(w, "index.html", map[string]interface{}{"x": 1})
			if err != nil {
				t.Errorf("Render : err [%s]", err)
				return
			}

			actual := w.Body.String()
			expected := "I am lowercase-eqs:true"
			if actual != expected {
				t.Errorf("expected [%s] actual [%s]", expected, actual)
				return
			}
		}()
	}
	wg.Wait()
}

// TestBugAllViewsLoaded demonstrates all views are loaded.
func TestBugAllViewsLoaded(t *testing.T) {

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("get pwd : err [%s]", err)
		return
	}

	tmplInited = false
	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		log.Printf("LoadViews : err [%s]", err)
		return
	}

	table := []TM{
		TM{
			Name:          "all/all.html",
			ExpectedError: nil,
			ExpectedBody:  `before_index::I am lowercase-eqs:true::after_index.before_menu::menu::after_menu`,
		},
		TM{
			Name:          "index.html",
			ExpectedError: nil,
			ExpectedBody:  `I am lowercase-eqs:true`,
		},
		TM{
			Name:          "menu.html",
			ExpectedError: nil,
			ExpectedBody:  `menu`,
		},
	}
	for i := range table {
		x := table[i]

		w := httptest.NewRecorder()
		err := Render(w, x.Name, map[string]interface{}{"x": 1})
		if err != x.ExpectedError {
			t.Errorf("expected [%s] actual [%s] view [%s]", x.ExpectedError, err, x.Name)
			continue
		}

		actual := w.Body.String()
		if actual != x.ExpectedBody {
			t.Errorf("expected [%s] actual [%s] view [%s]", x.ExpectedBody, actual, x.Name)
			continue
		}
	}
}

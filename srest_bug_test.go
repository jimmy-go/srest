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
			expected := []byte("I am lowercase-eqs:true")
			// remove additional \r
			actual = actual[:len(actual)-1]
			if string(actual) != string(expected) {
				t.Errorf("expected [%s] actual [%s]", string(expected), string(actual))
				return
			}
		}()
	}
	wg.Wait()
}

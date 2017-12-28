// Package srest contains test for bug fixes.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"log"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Nil(t, err)
			if err != nil {
				t.FailNow()
			}

			actual := w.Body.String()
			expected := "I am lowercase-eqs:true"
			assert.EqualValues(t, expected, actual)
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

	funcm := deffuncmap()
	err = LoadViews(dir+"/mock", funcm)
	if err != nil {
		log.Printf("LoadViews : err [%s]", err)
		return
	}

	table := []struct {
		Purpose  string
		Name     string
		ExpError error
		ExpBody  string
	}{
		{
			"1. OK",
			"all/all.html",
			nil,
			`before_index::I am lowercase-eqs:true::after_index.before_menu::menu::after_menu`,
		},
		{
			"2. OK",
			"index.html",
			nil,
			`I am lowercase-eqs:true`,
		},
		{
			"3. OK",
			"menu.html",
			nil,
			`menu`,
		},
	}
	for i := range table {
		x := table[i]

		w := httptest.NewRecorder()
		err := Render(w, x.Name, map[string]interface{}{"x": 1})
		if err != x.ExpError {
			t.Errorf("expected [%s] actual [%s] view [%s]", x.ExpError, err, x.Name)
			continue
		}

		actual := w.Body.String()
		if actual != x.ExpBody {
			t.Errorf("expected [%s] actual [%s] view [%s]", x.ExpBody, actual, x.Name)
			continue
		}
	}
}

// TestBugEmpty demonstrate empty templates return error.
func TestBugEmpty(t *testing.T) {

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("get pwd : err [%s]", err)
		return
	}

	err = LoadViews(dir+"/mock_empty", DefaultFuncMap)
	assert.NotNil(t, err)
}
